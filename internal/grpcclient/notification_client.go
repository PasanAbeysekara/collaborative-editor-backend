package grpcclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NotificationServiceClient provides methods to call notification-service via gRPC
type NotificationServiceClient struct {
	client pb.NotificationServiceClient
	conn   *grpc.ClientConn
}

// NewNotificationServiceClient creates a new gRPC client for notification-service
func NewNotificationServiceClient(address string) (*NotificationServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification-service: %w", err)
	}

	return &NotificationServiceClient{
		client: pb.NewNotificationServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *NotificationServiceClient) Close() error {
	return c.conn.Close()
}

// SendNotification sends a notification to a user
func (c *NotificationServiceClient) SendNotification(ctx context.Context, userID, message, notifType string, metadata map[string]string) (*pb.NotificationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.SendNotification(ctx, &pb.NotificationRequest{
		UserId:   userID,
		Message:  message,
		Type:     notifType,
		Metadata: metadata,
	})
}
