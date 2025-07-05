package services

import (
	"coding_test_2/internal/config"
	"coding_test_2/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

type ConsumerServiceInterface interface {
	UpdateReportStatus(ctx context.Context, requestID string, status models.ReportStatus, reportData string, errMsg string) (*models.ReportResult, error)
	ReportWorker(ctx context.Context, workerID int, msgs <-chan amqp.Delivery, results chan<- models.ReportResult)
	ResultAckHandler(ctx context.Context, results <-chan models.ReportResult, deliveries map[string]amqp.Delivery)
	SimulateReportGeneration(ctx context.Context, request models.ReportRequest) (string, error)
}

type ConsumerService struct {
	rdb *redis.Client
	ch  *amqp.Channel
}

func NewConsumerService(rdb *redis.Client, ch *amqp.Channel) ConsumerServiceInterface {
	return &ConsumerService{
		rdb: rdb,
		ch:  ch,
	}
}

// UpdateReportStatus updates the status of a report in Redis
func (s *ConsumerService) UpdateReportStatus(ctx context.Context, requestID string, status models.ReportStatus, reportData string, errMsg string) (*models.ReportResult, error) {
	result := models.ReportResult{
		RequestID:   requestID,
		Status:      status,
		GeneratedAt: time.Now().UTC(),
		ReportData:  reportData,
		Error:       errMsg,
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Printf("[Redis] Failed to marshal result for %s: %s", requestID, err)
		return nil, err
	}

	key := config.KEY_PREFIX_REPORT_STATUS + requestID
	err = s.rdb.Set(ctx, key, string(resultJson), 24*time.Hour).Err() // TTL: 24 hours
	if err != nil {
		log.Printf("[Redis] Failed to update status for %s: %s", requestID, status)
		return nil, err
	}

	if status == models.StatusCompleted || status == models.StatusFailed {
		key := config.KEY_PREFIX_REPORT_DATA + requestID
		err := s.rdb.Set(ctx, key, string(resultJson), 24*time.Hour).Err() // TTL: 24 hours
		if err != nil {
			log.Printf("[Redis] Failed to update report data for %s: %s", requestID, reportData)
			return nil, err
		}
	}

	log.Printf("[Redis] Updated status for %s: %s", requestID, status)
	return &result, nil
}

// reportWorker processes report requests from RabbitMQ
// It updates status in Redis and simulates report generation
func (s *ConsumerService) ReportWorker(ctx context.Context, workerID int, msgs <-chan amqp.Delivery, results chan<- models.ReportResult) {
	log.Printf("[Worker %d] Started", workerID)
	defer log.Printf("[Worker %d] Stopped", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Context cancelled", workerID)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("[Worker %d] Message channel closed", workerID)
				return
			}

			// Parse the request
			var request models.ReportRequest
			if err := json.Unmarshal(msg.Body, &request); err != nil {
				log.Printf("[Worker %d] Failed to unmarshal message: %v", workerID, err)
				msg.Nack(false, false) // Don't requeue malformed messages
				continue
			}

			log.Printf("[Worker %d] Processing request: %s", workerID, request.ID)

			// Update status to IN_PROGRESS
			result, err := s.UpdateReportStatus(ctx, request.ID, models.StatusInProgress, "", "")
			if err != nil {
				log.Printf("[Worker %d] Failed to update status for request %s: %v", workerID, request.ID, err)
				continue
			}

			// Create a timeout context for this specific task
			taskCtx, cancel := context.WithTimeout(ctx, config.WORKER_TIMEOUT)

			// Process the report
			reportData, err := s.SimulateReportGeneration(taskCtx, request)
			cancel() // Clean up the timeout context
			if err != nil {
				result.Status = models.StatusFailed
				result.Error = err.Error()
			} else {
				result.Status = models.StatusCompleted
			}

			// Create result based on processing outcome
			result, err = s.UpdateReportStatus(ctx, result.RequestID, result.Status, reportData, result.Error)
			if err != nil {
				log.Printf("[Worker %d] Failed to update status for request %s: %v", workerID, request.ID, err)
				continue
			}

			// Send result for acknowledgment handling
			select {
			case results <- *result:
			case <-ctx.Done():
				log.Printf("[Worker %d] Context cancelled while sending result", workerID)
				return
			}

			log.Printf("[Worker %d] Finished processing request: %s (Status: %s)",
				workerID, request.ID, result.Status)
		}
	}
}

// resultAckHandler handles RabbitMQ message acknowledgments based on processing results
func (s *ConsumerService) ResultAckHandler(ctx context.Context, results <-chan models.ReportResult, deliveries map[string]amqp.Delivery) {
	log.Println("[AckHandler] Started")
	defer log.Println("[AckHandler] Stopped")

	for {
		select {
		case <-ctx.Done():
			log.Println("[AckHandler] Context cancelled")
			return
		case result, ok := <-results:
			if !ok {
				log.Println("[AckHandler] Results channel closed")
				return
			}

			delivery, exists := deliveries[result.RequestID]
			if exists {
				delete(deliveries, result.RequestID)
			}

			if !exists {
				log.Printf("[AckHandler] No delivery found for request: %s", result.RequestID)
				continue
			}

			if result.Status == models.StatusCompleted {
				if err := delivery.Ack(false); err != nil {
					log.Printf("[AckHandler] Failed to ack message for %s: %v", result.RequestID, err)
				} else {
					log.Printf("[AckHandler] Acknowledged successful processing of %s", result.RequestID)
				}
			} else {
				// For failed processing, we nack without requeue
				if err := delivery.Nack(false, false); err != nil {
					log.Printf("[AckHandler] Failed to nack message for %s: %v", result.RequestID, err)
				} else {
					log.Printf("[AckHandler] Nacked failed processing of %s", result.RequestID)
				}
			}
		}
	}
}

// StartReportProcessor orchestrates the consumer side
// It connects to RabbitMQ, starts workers, and handles message delivery

// SimulateReportGeneration simulates the actual report generation process
// It includes random delays and a chance of failure
// It also respects its own context for timeout/cancellation
func (s *ConsumerService) SimulateReportGeneration(ctx context.Context, request models.ReportRequest) (string, error) {
	log.Printf("[Worker: %s] Starting report generation for ID: %s (Type: %s)",
		request.ID, request.ID, request.ReportType)

	// Simulate CPU-intensive work or external API calls
	delay := time.Duration(1+rand.Intn(5)) * time.Second // 1 to 5 seconds
	select {
	case <-time.After(delay):
		// Continue processing
	case <-ctx.Done():
		log.Printf("[Worker: %s] Context cancelled during simulation for ID: %s",
			request.ID, request.ID)
		return "", ctx.Err() // Propagate context cancellation error
	}

	// Simulate random failure (e.g., database error, invalid parameters)
	if rand.Intn(100) < 20 { // 20% chance of failure
		log.Printf("[Worker: %s] Simulated failure for ID: %s", request.ID, request.ID)
		return "", fmt.Errorf("simulated report generation error for ID %s", request.ID)
	}

	reportData := fmt.Sprintf("Report %s - Type: %s, Generated On: %s, Data: Random Value %d",
		request.ID, request.ReportType, time.Now().Format(time.RFC3339), rand.Intn(1000))

	log.Printf("[Worker: %s] Successfully generated report for ID: %s", request.ID, request.ID)
	return reportData, nil
}
