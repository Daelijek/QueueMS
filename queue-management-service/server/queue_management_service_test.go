package server

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"queue-management-system/queue-management-service/pb"
)

var testDB *sql.DB

func TestMain(m *testing.M) {

	dsn := "user=postgres password=root dbname=testDB sslmode=disable"
	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	setupTestTables()

	code := m.Run()

	testDB.Close()

	os.Exit(code)
}

func setupTestTables() {
	_, err := testDB.Exec(`
	CREATE TABLE IF NOT EXISTS queues (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS clients (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		queue_id INTEGER REFERENCES queues(id)
	);
	`)
	if err != nil {
		log.Fatalf("failed to create test database tables: %v", err)
	}
}

func setupTestDB() {
	_, err := testDB.Exec("TRUNCATE TABLE clients, queues RESTART IDENTITY")
	if err != nil {
		log.Fatalf("failed to truncate test database tables: %v", err)
	}
}

func TestCreateQueue(t *testing.T) {
	setupTestDB()
	server := NewQueueManagementService(testDB)

	t.Run("Success", func(t *testing.T) {
		req := &pb.CreateQueueRequest{Name: "Test Queue"}
		resp, err := server.CreateQueue(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Queue created successfully", resp.Message)
	})

	t.Run("EmptyName", func(t *testing.T) {
		req := &pb.CreateQueueRequest{Name: ""}
		_, err := server.CreateQueue(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "Queue name is required", status.Convert(err).Message())
	})
}

func TestUpdateQueue(t *testing.T) {
	setupTestDB()
	server := NewQueueManagementService(testDB)

	// First, create a queue to update
	_, err := server.CreateQueue(context.Background(), &pb.CreateQueueRequest{Name: "Test Queue"})
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		req := &pb.UpdateQueueRequest{Id: 1, Name: "Updated Queue"}
		resp, err := server.UpdateQueue(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Queue updated successfully", resp.Message)
	})

	t.Run("InvalidArguments", func(t *testing.T) {
		req := &pb.UpdateQueueRequest{Id: 0, Name: ""}
		_, err := server.UpdateQueue(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "Queue ID and name are required", status.Convert(err).Message())
	})
}

func TestDeleteQueue(t *testing.T) {
	setupTestDB()
	server := NewQueueManagementService(testDB)

	// First, create a queue to delete
	_, err := server.CreateQueue(context.Background(), &pb.CreateQueueRequest{Name: "Test Queue"})
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		req := &pb.DeleteQueueRequest{Id: 1}
		resp, err := server.DeleteQueue(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Queue deleted successfully", resp.Message)
	})

	t.Run("InvalidID", func(t *testing.T) {
		req := &pb.DeleteQueueRequest{Id: 0}
		_, err := server.DeleteQueue(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "Queue ID is required", status.Convert(err).Message())
	})
}

func TestGetQueueStatus(t *testing.T) {
	setupTestDB()
	server := NewQueueManagementService(testDB)

	_, err := server.CreateQueue(context.Background(), &pb.CreateQueueRequest{Name: "Test Queue"})
	require.NoError(t, err)
	_, err = testDB.Exec("INSERT INTO clients (name, queue_id) VALUES ($1, $2), ($3, $4), ($5, $6)",
		"Client A", 1, "Client B", 1, "Client C", 1)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		req := &pb.GetQueueStatusRequest{Id: 1}
		resp, err := server.GetQueueStatus(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int32(1), resp.Id)
		assert.Equal(t, "Test Queue", resp.Name)
		assert.ElementsMatch(t, []string{"Client A", "Client B", "Client C"}, resp.Clients)
		assert.Equal(t, "Queue status retrieved successfully", resp.Message)
	})

	t.Run("WithFilter", func(t *testing.T) {
		req := &pb.GetQueueStatusRequest{Id: 1, ClientNameFilter: "Client B"}
		resp, err := server.GetQueueStatus(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int32(1), resp.Id)
		assert.Equal(t, "Test Queue", resp.Name)
		assert.ElementsMatch(t, []string{"Client B"}, resp.Clients)
		assert.Equal(t, "Queue status retrieved successfully", resp.Message)
	})

	t.Run("WithPagination", func(t *testing.T) {
		req := &pb.GetQueueStatusRequest{Id: 1, Limit: 2, Offset: 1}
		resp, err := server.GetQueueStatus(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int32(1), resp.Id)
		assert.Equal(t, "Test Queue", resp.Name)
		assert.ElementsMatch(t, []string{"Client B", "Client C"}, resp.Clients)
		assert.Equal(t, "Queue status retrieved successfully", resp.Message)
	})

	t.Run("WithSorting", func(t *testing.T) {
		req := &pb.GetQueueStatusRequest{Id: 1, SortBy: "name", SortOrder: "desc"}
		resp, err := server.GetQueueStatus(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int32(1), resp.Id)
		assert.Equal(t, "Test Queue", resp.Name)
		assert.Equal(t, []string{"Client C", "Client B", "Client A"}, resp.Clients)
		assert.Equal(t, "Queue status retrieved successfully", resp.Message)
	})

	t.Run("InvalidID", func(t *testing.T) {
		req := &pb.GetQueueStatusRequest{Id: 0}
		_, err := server.GetQueueStatus(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "Queue ID is required", status.Convert(err).Message())
	})

	t.Run("QueueNotFound", func(t *testing.T) {
		req := &pb.GetQueueStatusRequest{Id: 999}
		_, err := server.GetQueueStatus(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Equal(t, "Queue not found", status.Convert(err).Message())
	})
}
