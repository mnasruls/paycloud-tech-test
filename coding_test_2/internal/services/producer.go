package services

import (
	"coding_test_2/internal/config"
	"coding_test_2/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type ProducerInterface interface {
	ProduceReportRequests(ctx context.Context, wg *sync.WaitGroup) error
}

type ProducerService struct {
	ch *amqp.Channel
}

func NewProducerService(ch *amqp.Channel) ProducerInterface {
	return &ProducerService{
		ch: ch,
	}
}

// produceReportRequests creates and publishes report requests to RabbitMQ
func (s *ProducerService) ProduceReportRequests(ctx context.Context, wg *sync.WaitGroup) error {
	log.Println("[Producer] Starting to produce report requests...")

	reportTypes := []string{"sales", "inventory", "financial", "user_activity"}

	for i := 0; i < config.NUM_PRODUCER_REQUESTS; i++ {
		select {
		case <-ctx.Done():
			log.Println("[Producer] Context cancelled, stopping production")
			return ctx.Err()
		default:
		}

		request := models.ReportRequest{
			ID:         fmt.Sprintf("report-%d", i+1),
			ReportType: reportTypes[i%len(reportTypes)],
			Parameters: map[string]string{
				"start_date": "2024-01-01",
				"end_date":   "2024-01-31",
				"format":     "PDF",
			},
			CreatedAt: time.Now(),
		}

		body, err := json.Marshal(request)
		if err != nil {
			log.Printf("[Producer] Failed to marshal request %s: %v", request.ID, err)
			continue
		}

		err = s.ch.Publish(
			"",                // exchange
			config.QUEUE_NAME, // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent, // Make message persistent
				ContentType:  "application/json",
				Body:         body,
			},
		)

		if err != nil {
			log.Printf("[Producer] Failed to publish request %s: %v", request.ID, err)
			continue
		}

		log.Printf("[Producer] Published request: %s (Type: %s)", request.ID, request.ReportType)

		// Wait before sending next message
		select {
		case <-time.After(config.PUBLISH_INTERVAL):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	log.Printf("[Producer] Finished producing %d report requests", config.NUM_PRODUCER_REQUESTS)
	return nil
}
