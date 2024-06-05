package server

import (
	"client-service/pb"
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterClient(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	server := &ClientServiceServer{db: db}

	req := &pb.RegisterClientRequest{
		QueueId: 1,
		Name:    "Ermek Dias",
	}

	mock.ExpectExec("INSERT INTO clients").WithArgs(req.QueueId, req.Name).WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := server.RegisterClient(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Client registered successfully", resp.Message)
}

func TestGetClientStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	server := &ClientServiceServer{db: db}

	req := &pb.GetClientStatusRequest{
		QueueId: 1,
	}

	rows := sqlmock.NewRows([]string{"name"}).AddRow("Dias Ermek").AddRow("Ernar Asherbekov")
	mock.ExpectQuery("SELECT name FROM clients WHERE queue_id = \\$1").WithArgs(req.QueueId).WillReturnRows(rows)

	resp, err := server.GetClientStatus(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"Dias Ermek", "Ernar Asherbekov"}, resp.Clients)
	assert.Equal(t, "Client status retrieved successfully", resp.Message)
}
