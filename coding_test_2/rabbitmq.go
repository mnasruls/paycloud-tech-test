package main

import (
	"coding_test_2/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// connectRabbitMQ establishes a connection and channel to RabbitMQ
func connectRabbitMQ(ctx context.Context) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(config.RABBITMQ_URL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		config.QUEUE_NAME, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ and declared queue:", config.QUEUE_NAME)
	return conn, ch, nil
}
