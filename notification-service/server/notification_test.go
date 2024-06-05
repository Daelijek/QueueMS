package server

import (
	"context"
	"net"
	"os"
	"testing"

	"notification-service/pb"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/mail.v2"
)

func mockDialer() *mail.Dialer {
	return &mail.Dialer{
		Host:     "smtp.mailtrap.io",
		Port:     587,
		Username: os.Getenv("MAILTRAP_USER"),
		Password: os.Getenv("MAILTRAP_PASSWORD"),
	}
}

func TestSendEmailNotification(t *testing.T) {
	defer gock.Off()

	// Ensure environment variables are set
	os.Setenv("MAILTRAP_USER", "ernar")
	os.Setenv("MAILTRAP_PASSWORD", "password")

	// Mock Mailtrap API response
	gock.New("https://smtp.mailtrap.io").
		Post("/api/v1/inboxes").
		Reply(200).
		JSON(map[string]string{"status": "success"})

	dialer := mockDialer()
	server := NewNotificationService(dialer)
	req := &pb.SendNotificationRequest{
		Message: "Test email message",
		Channel: "email",
		Email:   "client@example.com",
	}

	_, err := server.SendNotification(context.Background(), req)
	assert.NoError(t, err)
}

// Test invalid arguments
func TestSendNotificationInvalidArgs(t *testing.T) {
	os.Setenv("MAILTRAP_USER", "ernar")
	os.Setenv("MAILTRAP_PASSWORD", "password")

	dialer := mockDialer()
	server := NewNotificationService(dialer)

	req := &pb.SendNotificationRequest{
		Channel: "email",
		Email:   "client@example.com",
	}
	_, err := server.SendNotification(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	req = &pb.SendNotificationRequest{
		Message: "Test message",
		Email:   "client@example.com",
	}
	_, err = server.SendNotification(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSendNotificationIntegration(t *testing.T) {
	os.Setenv("MAILTRAP_USER", "ernar")
	os.Setenv("MAILTRAP_PASSWORD", "password")

	server := grpc.NewServer()
	dialer := mockDialer()
	pb.RegisterNotificationServiceServer(server, NewNotificationService(dialer))

	lis, err := net.Listen("tcp", ":0") // dynamically allocate a port
	assert.NoError(t, err)

	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewNotificationServiceClient(conn)
	req := &pb.SendNotificationRequest{
		Message: "Test integration message",
		Channel: "email",
		Email:   "client@example.com",
	}

	resp, err := client.SendNotification(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}
