// Example usage of gRPC clients in the collaborative editor
package main

import (
	"context"
	"log"

	"github.com/pasanAbeysekara/collaborative-editor/internal/grpcclient"
)

func main() {
	demonstrateUserService()
	demonstrateDocumentService()
	demonstrateNotificationService()
}

func demonstrateUserService() {
	log.Println("=== User Service gRPC Examples ===")

	// Connect to user-service
	client, err := grpcclient.NewUserServiceClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Register a new user
	registerResp, err := client.Register(ctx, "newuser@example.com", "password123")
	if err != nil {
		log.Printf("Register error: %v", err)
	} else {
		log.Printf("User registered: ID=%s, Email=%s", registerResp.Id, registerResp.Email)
	}

	// Example 2: Login
	loginResp, err := client.Login(ctx, "user@example.com", "password")
	if err != nil {
		log.Printf("Login error: %v", err)
	} else {
		log.Printf("Login successful: Token=%s", loginResp.Token)
	}

	// Example 3: Get user by email
	user, err := client.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		log.Printf("GetUserByEmail error: %v", err)
	} else {
		log.Printf("User found: ID=%s, Email=%s", user.Id, user.Email)
	}

	// Example 4: Get user by ID
	user, err = client.GetUserByID(ctx, "user-123")
	if err != nil {
		log.Printf("GetUserByID error: %v", err)
	} else {
		log.Printf("User found: ID=%s, Email=%s", user.Id, user.Email)
	}
}

func demonstrateDocumentService() {
	log.Println("\n=== Document Service gRPC Examples ===")

	// Connect to document-service
	client, err := grpcclient.NewDocumentServiceClient("localhost:50052")
	if err != nil {
		log.Fatalf("Failed to create document client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Create a document
	doc, err := client.CreateDocument(ctx, "My First Document", "user-123", "Hello World!")
	if err != nil {
		log.Printf("CreateDocument error: %v", err)
	} else {
		log.Printf("Document created: ID=%s, Title=%s", doc.Id, doc.Title)
	}

	// Example 2: Get a specific document
	doc, err = client.GetDocument(ctx, "doc-123")
	if err != nil {
		log.Printf("GetDocument error: %v", err)
	} else {
		log.Printf("Document retrieved: ID=%s, Title=%s, Content=%s", doc.Id, doc.Title, doc.Content)
	}

	// Example 3: Get all user documents
	docList, err := client.GetUserDocuments(ctx, "user-123")
	if err != nil {
		log.Printf("GetUserDocuments error: %v", err)
	} else {
		log.Printf("Retrieved %d documents for user", len(docList.Documents))
		for i, doc := range docList.Documents {
			log.Printf("  %d. %s (ID: %s)", i+1, doc.Title, doc.Id)
		}
	}

	// Example 4: Update a document
	updateResp, err := client.UpdateDocument(ctx, "doc-123", "Updated content", 2)
	if err != nil {
		log.Printf("UpdateDocument error: %v", err)
	} else {
		log.Printf("Document updated: Success=%v, Message=%s", updateResp.Success, updateResp.Message)
	}

	// Example 5: Share a document
	shareResp, err := client.ShareDocument(ctx, "doc-123", "user-123", "colleague@example.com", "editor")
	if err != nil {
		log.Printf("ShareDocument error: %v", err)
	} else {
		log.Printf("Document shared: Success=%v, Message=%s", shareResp.Success, shareResp.Message)
	}

	// Example 6: Check permission
	permResp, err := client.CheckPermission(ctx, "doc-123", "user-456")
	if err != nil {
		log.Printf("CheckPermission error: %v", err)
	} else {
		log.Printf("Permission check: HasPermission=%v", permResp.HasPermission)
	}
}

func demonstrateNotificationService() {
	log.Println("\n=== Notification Service gRPC Examples ===")

	// Connect to notification-service
	client, err := grpcclient.NewNotificationServiceClient("localhost:50053")
	if err != nil {
		log.Fatalf("Failed to create notification client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Send a document share notification
	resp, err := client.SendNotification(
		ctx,
		"user-456",
		"You have been invited to collaborate on a document",
		"document_share",
		map[string]string{
			"document_id":   "doc-123",
			"document_name": "Project Proposal",
			"shared_by":     "user-123",
		},
	)
	if err != nil {
		log.Printf("SendNotification error: %v", err)
	} else {
		log.Printf("Notification sent: Success=%v, Message=%s", resp.Success, resp.Message)
	}

	// Example 2: Send a comment notification
	resp, err = client.SendNotification(
		ctx,
		"user-123",
		"Someone commented on your document",
		"document_comment",
		map[string]string{
			"document_id":  "doc-789",
			"comment_by":   "user-456",
			"comment_text": "Great work!",
		},
	)
	if err != nil {
		log.Printf("SendNotification error: %v", err)
	} else {
		log.Printf("Notification sent: Success=%v, Message=%s", resp.Success, resp.Message)
	}

	// Example 3: Send a system notification
	resp, err = client.SendNotification(
		ctx,
		"user-123",
		"Your subscription is about to expire",
		"system_alert",
		map[string]string{
			"expires_in": "7 days",
		},
	)
	if err != nil {
		log.Printf("SendNotification error: %v", err)
	} else {
		log.Printf("Notification sent: Success=%v, Message=%s", resp.Success, resp.Message)
	}
}
