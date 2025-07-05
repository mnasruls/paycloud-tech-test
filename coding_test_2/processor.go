package main

import (
	"coding_test_2/internal/config"
	"coding_test_2/internal/models"
	"coding_test_2/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

func StartReportProcessor(ctx context.Context, wg *sync.WaitGroup, ch *amqp.Channel, rdb *redis.Client) error {
	s := services.NewConsumerService(rdb, ch)
	log.Println("[Consumer] Starting report processor...")

	// Set QoS to limit unacknowledged messages per worker
	err := ch.Qos(
		1,     // prefetch count - only 1 unacknowledged message per worker
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		config.QUEUE_NAME, // queue
		"",                // consumer
		false,             // auto-ack (we'll manually ack)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Create channels for worker communication
	results := make(chan models.ReportResult, config.NUM_WORKERS*2) // Buffered channel
	deliveries := make(map[string]amqp.Delivery)
	// var deliveriesMu sync.Mutex

	// Start result acknowledgment handler
	go s.ResultAckHandler(ctx, results, deliveries)

	// Start workers
	workerMsgs := make(chan amqp.Delivery, config.NUM_WORKERS)

	for i := 1; i <= config.NUM_WORKERS; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.ReportWorker(ctx, workerID, workerMsgs, results)
		}(i)
	}

	// Message distributor
	go func() {
		defer close(workerMsgs)
		defer close(results)

		for {
			select {
			case <-ctx.Done():
				log.Println("[Consumer] Context cancelled, stopping message distribution")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("[Consumer] Message channel closed")
					return
				}

				// Parse request to get ID for tracking
				var request models.ReportRequest
				if err := json.Unmarshal(msg.Body, &request); err != nil {
					log.Printf("[Consumer] Failed to parse message for tracking: %v", err)
					msg.Nack(false, false)
					continue
				}

				// Store delivery for later acknowledgment
				deliveries[request.ID] = msg

				// Update status to PENDING
				s.UpdateReportStatus(ctx, request.ID, models.StatusPending, "", "")

				// Forward to workers
				select {
				case workerMsgs <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	log.Printf("[Consumer] Started %d workers, waiting for messages...", config.NUM_WORKERS)

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("[Consumer] Shutting down...")

	// Wait for workers to finish
	log.Println("[Consumer] All workers stopped")

	return ctx.Err()
}
