package answer7

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "knowladge-test/answer_7/proto"

	"google.golang.org/grpc"
)

// UserService implementation
type userServiceServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *userServiceServer) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	users := []*pb.User{
		{Id: 1, Name: "John Doe", Email: "john@example.com"},
		{Id: 2, Name: "Jane Smith", Email: "jane@example.com"},
		{Id: 3, Name: "Bob Johnson", Email: "bob@example.com"},
	}

	return &pb.GetUsersResponse{
		Users:   users,
		Message: "Users retrieved successfully via gRPC",
	}, nil
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer() {
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userService := &userServiceServer{}
	pb.RegisterUserServiceServer(s, userService)

	log.Println("Starting gRPC server on :8090")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// GrpcEx demonstrates gRPC API usage
func GrpcEx() {
	fmt.Println("=== gRPC API Example ===")

	// Start gRPC server in a goroutine
	go StartGRPCServer()

}
