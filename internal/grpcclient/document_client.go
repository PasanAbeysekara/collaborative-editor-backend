package grpcclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DocumentServiceClient provides methods to call document-service via gRPC
type DocumentServiceClient struct {
	client pb.DocumentServiceClient
	conn   *grpc.ClientConn
}

// NewDocumentServiceClient creates a new gRPC client for document-service
func NewDocumentServiceClient(address string) (*DocumentServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to document-service: %w", err)
	}

	return &DocumentServiceClient{
		client: pb.NewDocumentServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *DocumentServiceClient) Close() error {
	return c.conn.Close()
}

// CreateDocument creates a new document
func (c *DocumentServiceClient) CreateDocument(ctx context.Context, title, userID, content string) (*pb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.CreateDocument(ctx, &pb.CreateDocumentRequest{
		Title:   title,
		UserId:  userID,
		Content: content,
	})
}

// GetDocument fetches a document by ID
func (c *DocumentServiceClient) GetDocument(ctx context.Context, documentID string) (*pb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetDocument(ctx, &pb.GetDocumentRequest{DocumentId: documentID})
}

// GetUserDocuments fetches all documents for a user
func (c *DocumentServiceClient) GetUserDocuments(ctx context.Context, userID string) (*pb.DocumentList, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetUserDocuments(ctx, &pb.GetUserDocumentsRequest{UserId: userID})
}

// UpdateDocument updates a document's content
func (c *DocumentServiceClient) UpdateDocument(ctx context.Context, documentID, content string, version int32) (*pb.UpdateDocumentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.UpdateDocument(ctx, &pb.UpdateDocumentRequest{
		DocumentId: documentID,
		Content:    content,
		Version:    version,
	})
}

// ShareDocument shares a document with another user
func (c *DocumentServiceClient) ShareDocument(ctx context.Context, documentID, ownerID, targetEmail, role string) (*pb.ShareDocumentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ShareDocument(ctx, &pb.ShareDocumentRequest{
		DocumentId:      documentID,
		OwnerId:         ownerID,
		TargetUserEmail: targetEmail,
		Role:            role,
	})
}

// CheckPermission checks if a user has permission to access a document
func (c *DocumentServiceClient) CheckPermission(ctx context.Context, documentID, userID string) (*pb.CheckPermissionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.CheckPermission(ctx, &pb.CheckPermissionRequest{
		DocumentId: documentID,
		UserId:     userID,
	})
}
