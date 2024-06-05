package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"queue-management-system/queue-management-service/models"
	"queue-management-system/queue-management-service/pb"

	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueueManagementServiceServer struct {
	pb.UnimplementedQueueManagementServiceServer
	db *sql.DB
}

func NewQueueManagementService(db *sql.DB) *QueueManagementServiceServer {
	if db == nil {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "user=postgres password=root dbname=d.abaevDB sslmode=disable"
		}
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		return &QueueManagementServiceServer{db: db}
	}
	return &QueueManagementServiceServer{db: db}
}

func (s *QueueManagementServiceServer) CreateQueue(ctx context.Context, req *pb.CreateQueueRequest) (*pb.CreateQueueResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Queue name is required")
	}

	_, err := s.db.Exec("INSERT INTO queues (name) VALUES ($1)", req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateQueueResponse{Success: true, Message: "Queue created successfully"}, nil
}

func (s *QueueManagementServiceServer) UpdateQueue(ctx context.Context, req *pb.UpdateQueueRequest) (*pb.UpdateQueueResponse, error) {
	if req.Id == 0 || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Queue ID and name are required")
	}

	_, err := s.db.Exec("UPDATE queues SET name = $1 WHERE id = $2", req.Name, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateQueueResponse{Success: true, Message: "Queue updated successfully"}, nil
}

func (s *QueueManagementServiceServer) DeleteQueue(ctx context.Context, req *pb.DeleteQueueRequest) (*pb.DeleteQueueResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "Queue ID is required")
	}

	_, err := s.db.Exec("DELETE FROM queues WHERE id = $1", req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteQueueResponse{Success: true, Message: "Queue deleted successfully"}, nil
}

func (s *QueueManagementServiceServer) GetQueueStatus(ctx context.Context, req *pb.GetQueueStatusRequest) (*pb.GetQueueStatusResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "Queue ID is required")
	}

	row := s.db.QueryRow("SELECT id, name FROM queues WHERE id = $1", req.Id)
	var queue models.Queue
	if err := row.Scan(&queue.ID, &queue.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "Queue not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	query := "SELECT name FROM clients WHERE queue_id = $1"
	args := []interface{}{req.Id}

	paramIndex := 2

	if req.ClientNameFilter != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", paramIndex)
		args = append(args, "%"+req.ClientNameFilter+"%")
		paramIndex++
	}

	if req.SortBy != "" {
		order := "ASC"
		if req.SortOrder == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", req.SortBy, order)
	} else {
		query += " ORDER BY name ASC"
	}

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		args = append(args, req.Limit)
		paramIndex++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramIndex)
		args = append(args, req.Offset)
		paramIndex++
	}

	rows, err := s.db.Query(query, args...)
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
	return &pb.GetQueueStatusResponse{Id: queue.ID, Name: queue.Name, Clients: clients, Message: "Queue status retrieved successfully"}, nil
}
