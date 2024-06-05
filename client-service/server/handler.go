package server

import (
	"client-service/pb"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
)

type ClientServiceServer struct {
	pb.UnimplementedClientServiceServer
	db *sql.DB
}

func NewClientService() *ClientServiceServer {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "user=postgres password=postgres dbname=queue_management sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return &ClientServiceServer{db: db}
}

func (s *ClientServiceServer) RegisterClient(ctx context.Context, req *pb.RegisterClientRequest) (*pb.RegisterClientResponse, error) {
	if req.QueueId == 0 || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Queue ID and client name are required")
	}

	_, err := s.db.Exec("INSERT INTO clients (queue_id, name) VALUES ($1, $2)", req.QueueId, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RegisterClientResponse{Success: true, Message: "Client registered successfully"}, nil
}

func (s *ClientServiceServer) GetClientStatus(ctx context.Context, req *pb.GetClientStatusRequest) (*pb.GetClientStatusResponse, error) {
	if req.QueueId == 0 {
		return nil, status.Error(codes.InvalidArgument, "Queue ID is required")
	}

	rows, err := s.db.Query("SELECT name FROM clients WHERE queue_id = $1", req.QueueId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	var clients []string
	for rows.Next() {
		var clientName string
		if err := rows.Scan(&clientName); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		clients = append(clients, clientName)
	}
	return &pb.GetClientStatusResponse{Clients: clients, Message: "Client status retrieved successfully"}, nil
}
