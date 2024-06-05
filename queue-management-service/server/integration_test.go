package server_test

import (
	"context"
	"database/sql"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"queue-management-system/queue-management-service/pb"
	"queue-management-system/queue-management-service/server"
)

var db *sql.DB

func init() {
	var err error
	dsn := "user=postgres password=root dbname=testDB sslmode=disable"
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	// Clean up test database before running tests
	_, err = db.Exec(`DROP TABLE IF EXISTS queues, clients`)
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE queues (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE clients (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			queue_id INT REFERENCES queues(id)
		);
	`)
	if err != nil {
		log.Fatalf("failed to set up test database: %v", err)
	}
}

func setupServer() *server.QueueManagementServiceServer {
	return server.NewQueueManagementService(db)
}

func TestCreateQueue(t *testing.T) {
	s := setupServer()
	ctx := context.Background()

	// Test creating a queue with a valid name
	req := &pb.CreateQueueRequest{Name: "Test Queue"}
	res, err := s.CreateQueue(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.Success)

	// Test creating a queue with an empty name
	req = &pb.CreateQueueRequest{Name: ""}
	res, err = s.CreateQueue(ctx, req)
	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestUpdateQueue(t *testing.T) {
	s := setupServer()
	ctx := context.Background()

	// Create a queue to update
	_, err := db.Exec("INSERT INTO queues (name) VALUES ($1)", "Old Queue")
	require.NoError(t, err)

	var queueID int
	err = db.QueryRow("SELECT id FROM queues WHERE name = $1", "Old Queue").Scan(&queueID)
	require.NoError(t, err)

	// Test updating the queue with a valid name
	req := &pb.UpdateQueueRequest{Id: int32(queueID), Name: "Updated Queue"}
	res, err := s.UpdateQueue(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.Success)

	// Test updating the queue with an empty name
	req = &pb.UpdateQueueRequest{Id: int32(queueID), Name: ""}
	res, err = s.UpdateQueue(ctx, req)
	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDeleteQueue(t *testing.T) {
	s := setupServer()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO queues (name) VALUES ($1)", "Queue To Delete")
		require.NoError(t, err)

		var queueID int
		err = db.QueryRow("SELECT id FROM queues WHERE name = $1", "Queue To Delete").Scan(&queueID)
		require.NoError(t, err)

		req := &pb.DeleteQueueRequest{Id: int32(queueID)}
		res, err := s.DeleteQueue(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.True(t, res.Success)
	})

	t.Run("InvalidID", func(t *testing.T) {
		// Test deleting a queue with an invalid ID (ID = 0)
		req := &pb.DeleteQueueRequest{Id: 0}
		res, err := s.DeleteQueue(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("NonExistentID", func(t *testing.T) {
		// Test deleting a non-existent queue (e.g., ID = 9999)
		req := &pb.DeleteQueueRequest{Id: 9999}
		res, err := s.DeleteQueue(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
		require.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestGetQueueStatus(t *testing.T) {
	s := setupServer()
	ctx := context.Background()

	// Create a queue and some clients
	_, err := db.Exec("INSERT INTO queues (name) VALUES ($1)", "Status Queue")
	require.NoError(t, err)

	var queueID int
	err = db.QueryRow("SELECT id FROM queues WHERE name = $1", "Status Queue").Scan(&queueID)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO clients (name, queue_id) VALUES ($1, $2), ($3, $2)", "Client1", queueID, "Client2")
	require.NoError(t, err)

	// Test getting queue status
	req := &pb.GetQueueStatusRequest{Id: int32(queueID)}
	res, err := s.GetQueueStatus(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int32(queueID), res.Id)
	require.Equal(t, "Status Queue", res.Name)
	require.Len(t, res.Clients, 2)

	// Test getting queue status with a client name filter
	req = &pb.GetQueueStatusRequest{Id: int32(queueID), ClientNameFilter: "Client1"}
	res, err = s.GetQueueStatus(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Clients, 1)
	require.Equal(t, "Client1", res.Clients[0])

	// Test getting queue status for a non-existent queue
	req = &pb.GetQueueStatusRequest{Id: int32(queueID + 1)}
	res, err = s.GetQueueStatus(ctx, req)
	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, codes.NotFound, status.Code(err))
}
