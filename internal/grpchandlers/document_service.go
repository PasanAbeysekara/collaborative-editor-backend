package grpchandlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DocumentServiceServer struct {
	pb.UnimplementedDocumentServiceServer
	Store       storage.Store
	AMQPChannel *amqp091.Channel
}

func (s *DocumentServiceServer) CreateDocument(ctx context.Context, req *pb.CreateDocumentRequest) (*pb.Document, error) {
	doc, err := s.Store.CreateDocument(req.Title, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create document: %v", err)
	}

	return &pb.Document{
		Id:        doc.ID,
		Title:     doc.Title,
		OwnerId:   doc.OwnerID,
		Content:   doc.Content,
		Version:   int32(doc.Version),
		CreatedAt: doc.CreatedAt.String(),
		UpdatedAt: doc.UpdatedAt.String(),
	}, nil
}

func (s *DocumentServiceServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
	doc, err := s.Store.GetDocument(req.DocumentId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "document not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get document: %v", err)
	}

	return &pb.Document{
		Id:        doc.ID,
		Title:     doc.Title,
		OwnerId:   doc.OwnerID,
		Content:   doc.Content,
		Version:   int32(doc.Version),
		CreatedAt: doc.CreatedAt.String(),
		UpdatedAt: doc.UpdatedAt.String(),
	}, nil
}

func (s *DocumentServiceServer) GetUserDocuments(ctx context.Context, req *pb.GetUserDocumentsRequest) (*pb.DocumentList, error) {
	documents, err := s.Store.GetUserDocuments(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user documents: %v", err)
	}

	var pbDocs []*pb.Document
	for _, doc := range documents {
		pbDocs = append(pbDocs, &pb.Document{
			Id:        doc.ID,
			Title:     doc.Title,
			OwnerId:   doc.OwnerID,
			Content:   doc.Content,
			Version:   int32(doc.Version),
			CreatedAt: doc.CreatedAt.String(),
			UpdatedAt: doc.UpdatedAt.String(),
		})
	}

	return &pb.DocumentList{Documents: pbDocs}, nil
}

func (s *DocumentServiceServer) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentRequest) (*pb.UpdateDocumentResponse, error) {
	err := s.Store.UpdateDocument(req.DocumentId, req.Content, int(req.Version))
	if err != nil {
		return &pb.UpdateDocumentResponse{
			Success: false,
			Message: err.Error(),
		}, status.Errorf(codes.Internal, "failed to update document: %v", err)
	}

	return &pb.UpdateDocumentResponse{
		Success: true,
		Message: "Document updated successfully",
	}, nil
}

func (s *DocumentServiceServer) ShareDocument(ctx context.Context, req *pb.ShareDocumentRequest) (*pb.ShareDocumentResponse, error) {
	// Get target user by email
	targetUser, err := s.Store.GetUserByEmail(req.TargetUserEmail)
	if err != nil {
		return &pb.ShareDocumentResponse{
			Success: false,
			Message: "Target user not found",
		}, status.Errorf(codes.NotFound, "target user not found")
	}

	// Share the document
	err = s.Store.ShareDocument(req.DocumentId, req.OwnerId, targetUser.ID, req.Role)
	if err != nil {
		if err.Error() == "permission denied: only the owner can share this document" {
			return &pb.ShareDocumentResponse{
				Success: false,
				Message: err.Error(),
			}, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return &pb.ShareDocumentResponse{
			Success: false,
			Message: "Failed to share document",
		}, status.Errorf(codes.Internal, "failed to share document: %v", err)
	}

	// Publish event to RabbitMQ
	if s.AMQPChannel != nil {
		eventBody, _ := json.Marshal(map[string]string{
			"document_id":         req.DocumentId,
			"shared_with_user_id": targetUser.ID,
			"shared_by_user_id":   req.OwnerId,
		})

		err = s.AMQPChannel.PublishWithContext(ctx,
			"events",       // exchange
			"user.invited", // routing key
			false,          // mandatory
			false,          // immediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        eventBody,
			})

		if err != nil {
			log.Printf("WARN: Failed to publish user.invited event for doc %s: %v", req.DocumentId, err)
		}
	}

	return &pb.ShareDocumentResponse{
		Success: true,
		Message: "Document shared successfully",
	}, nil
}

func (s *DocumentServiceServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	hasPermission, err := s.Store.CheckDocumentPermission(req.DocumentId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check permission: %v", err)
	}

	return &pb.CheckPermissionResponse{
		HasPermission: hasPermission,
	}, nil
}
