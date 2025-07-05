package main

import (
	"coding_test_2/internal/services"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting Report Processing System...")
	defer log.Println("Report Processing System stopped gracefully")

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	// Connect to Redis
	rdb, err := connectRedis(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		log.Println("Closing Redis connection...")
		rdb.Close()
		log.Println("Redis connection closed")
	}()

	// Connect to RabbitMQ
	conn, ch, err := connectRabbitMQ(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		log.Println("Closing RabbitMQ connection...")
		conn.Close()
		log.Println("RabbitMQ connection closed")
	}()

	// Create producer and consumer services
	producer := services.NewProducerService(ch)

	// Start both producer and consumer concurrently
	var wg sync.WaitGroup

	// Start consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := StartReportProcessor(ctx, &wg, ch, rdb); err != nil && err != context.Canceled {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Give consumer time to start
	time.Sleep(2 * time.Second)

	// Start producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := producer.ProduceReportRequests(ctx, &wg); err != nil && err != context.Canceled {
			log.Printf("Producer error: %v", err)
		}
	}()

	// Wait for all goroutines
	wg.Wait()
}
