package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"client-service/models"
	pb "client-service/pb"
)

// Server struct to implement the gRPC server methods
type server struct {
	pb.UnimplementedClientServiceServer
	clients map[int32]*models.Client
	queues  map[int32]*models.Queue
}

// NewServer creates a new server instance with initialized data
func NewServer() *server {
	return &server{
		clients: make(map[int32]*models.Client),
		queues:  make(map[int32]*models.Queue),
	}
}

// RegisterClient registers a new client in the specified queue
func (s *server) RegisterClient(ctx context.Context, req *pb.RegisterClientRequest) (*pb.RegisterClientResponse, error) {
	clientID := int32(len(s.clients) + 1)
	client := &models.Client{
		ID:      clientID,
		QueueID: req.GetQueueId(),
		Name:    req.GetName(),
	}
	s.clients[clientID] = client

	// Assuming the queue exists. You can add more error handling here.
	if _, exists := s.queues[req.GetQueueId()]; !exists {
		return &pb.RegisterClientResponse{
			Success: false,
			Message: "Queue not found",
		}, nil
	}

	return &pb.RegisterClientResponse{
		Success: true,
		Message: fmt.Sprintf("Client registered with ID %d", clientID),
	}, nil
}

// GetClientStatus retrieves the status of clients in the specified queue
func (s *server) GetClientStatus(ctx context.Context, req *pb.GetClientStatusRequest) (*pb.GetClientStatusResponse, error) {
	queueID := req.GetQueueId()
	var clientNames []string

	for _, client := range s.clients {
		if client.QueueID == queueID {
			clientNames = append(clientNames, client.Name)
		}
	}

	if len(clientNames) == 0 {
		return &pb.GetClientStatusResponse{
			Clients: clientNames,
			Message: "No clients found in the specified queue",
		}, nil
	}

	return &pb.GetClientStatusResponse{
		Clients: clientNames,
		Message: "Clients retrieved successfully",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	s := NewServer()
	pb.RegisterClientServiceServer(grpcServer, s)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Printf("Server listening on port %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
