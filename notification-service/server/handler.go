package server

import (
	"context"
	"fmt"
	"log"
	"notification-service/pb"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/mail.v2"
)

type NotificationServiceServer struct {
	pb.UnimplementedNotificationServiceServer
	dialer *mail.Dialer
}

func NewNotificationService(dialer *mail.Dialer) *NotificationServiceServer {
	return &NotificationServiceServer{dialer: dialer}
}

func (s *NotificationServiceServer) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	if req.Message == "" || req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "Message and channel are required")
	}

	log.Printf("Sending %s notification: %s", req.Channel, req.Message)

	if req.Channel == "email" {
		if req.Email == "" {
			return nil, status.Error(codes.InvalidArgument, "Email is required for email notifications")
		}
		err := s.sendEmail(req.Email, req.Message)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Implement other notification logic (e.g., send SMS, etc.)

	return &pb.SendNotificationResponse{Success: true, Message: "Notification sent successfully"}, nil
}

func (s *NotificationServiceServer) sendEmail(to string, message string) error {
	from := os.Getenv("MAILTRAP_USER")
	if from == "" {
		return fmt.Errorf("MAILTRAP_USER environment variable is not set")
	}

	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Queue Notification")
	m.SetBody("text/plain", message)

	// Send the email using the provided dialer
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
