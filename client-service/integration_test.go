package main

import (
	"context"
	"log"
	"net"
	"testing"
	"client-service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterClientServiceServer(s, NewServer())
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestRegisterClientIntegration(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()
	client := pb.NewClientServiceClient(conn)

	req := &pb.RegisterClientRequest{
		QueueId: 1,
		Name:    "Dias Ermek",
	}
	resp, err := client.RegisterClient(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "Client registered with ID")
}

func TestGetClientStatusIntegration(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()
	client := pb.NewClientServiceClient(conn)

	// Register a client first
	registerReq := &pb.RegisterClientRequest{
		QueueId: 1,
		Name:    "Dias Ermek",
	}
	_, err = client.RegisterClient(ctx, registerReq)
	assert.NoError(t, err)

	// Now get the client status
	statusReq := &pb.GetClientStatusRequest{
		QueueId: 1,
	}
	statusResp, err := client.GetClientStatus(ctx, statusReq)
	assert.NoError(t, err)
	assert.NotNil(t, statusResp)
	assert.Equal(t, []string{"Dias Ermek"}, statusResp.Clients)
	assert.Equal(t, "Clients retrieved successfully", statusResp.Message)
}
