package answer7

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	pb "knowladge-test/answer_7/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Status  string      `json:"status"`
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	conn, err := grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Create gRPC client
	client := pb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcResp, err := client.GetUsers(ctx, &pb.GetUsersRequest{})
	if err != nil {
		http.Error(w, "Failed to get users from gRPC service", http.StatusInternalServerError)
		return
	}

	users := make([]User, len(grpcResp.Users))
	for i, u := range grpcResp.Users {
		users[i] = User{
			ID:    int(u.Id),
			Name:  u.Name,
			Email: u.Email,
		}
	}

	response := Response{
		Message: grpcResp.Message,
		Data:    users,
		Status:  "success",
	}

	json.NewEncoder(w).Encode(response)
}

func StartRESTServer() {
	if err := InitRabbitMQ(); err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ: %v", err)
		log.Println("RabbitMQ endpoints will not work properly")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/users", getUsersHandler)
	mux.HandleFunc("/send-message", sendMessageHandler)

	log.Println("Starting REST API server on :7080")
	log.Println("Available endpoints:")
	log.Println("  GET  http://localhost:7080/users")
	log.Println("  POST http://localhost:7080/send-message")

	if err := http.ListenAndServe(":7080", mux); err != nil {
		log.Fatalf("Failed to start REST server: %v", err)
	}
}

func RestEx() {
	fmt.Println("=== REST API Example ===")

	// Start server in a goroutine
	go StartRESTServer()
}
