package grpcclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserServiceClient provides methods to call user-service via gRPC
type UserServiceClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewUserServiceClient creates a new gRPC client for user-service
func NewUserServiceClient(address string) (*UserServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user-service: %w", err)
	}

	return &UserServiceClient{
		client: pb.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *UserServiceClient) Close() error {
	return c.conn.Close()
}

// GetUserByEmail fetches a user by email
func (c *UserServiceClient) GetUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{Email: email})
}

// GetUserByID fetches a user by ID
func (c *UserServiceClient) GetUserByID(ctx context.Context, id string) (*pb.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetUserByID(ctx, &pb.GetUserByIDRequest{Id: id})
}

// Register creates a new user
func (c *UserServiceClient) Register(ctx context.Context, email, password string) (*pb.RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
}

// Login authenticates a user
func (c *UserServiceClient) Login(ctx context.Context, email, password string) (*pb.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
}
