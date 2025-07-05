package answer7

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

type MessageRequest struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

type RabbitMQService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQService() (*RabbitMQService, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:8081/") // change base on your url
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		"hello_world_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	return &RabbitMQService{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (r *RabbitMQService) PublishMessage(message string) error {
	body := map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"id":        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
	}

	msgBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = r.channel.Publish(
		"",
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBody,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	log.Printf("Sent message to RabbitMQ: %s", message)
	return nil
}

func (r *RabbitMQService) ConsumeMessages() {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		return
	}

	log.Println("Starting RabbitMQ consumer...")
	go func() {
		for d := range msgs {
			log.Printf("Received message from RabbitMQ: %s", d.Body)
			// Process the message here
			fmt.Printf("Processing: %s\n", d.Body)
		}
	}()
}

func (r *RabbitMQService) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

var rabbitMQService *RabbitMQService

func InitRabbitMQ() error {
	var err error
	rabbitMQService, err = NewRabbitMQService()
	if err != nil {
		return err
	}
	// Start consuming messages
	rabbitMQService.ConsumeMessages()
	return nil
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body provided, use default "Hello World" message
		req.Message = "Hello World"
	}

	if req.Message == "" {
		req.Message = "Hello World"
	}

	go func() {
		if err := rabbitMQService.PublishMessage(req.Message); err != nil {
			log.Printf("Error publishing message: %v", err)
		}
	}()

	response := MessageResponse{
		Status:  "success",
		Message: fmt.Sprintf("Message '%s' sent to RabbitMQ asynchronously", req.Message),
		ID:      fmt.Sprintf("req_%d", time.Now().UnixNano()),
	}

	json.NewEncoder(w).Encode(response)
}
