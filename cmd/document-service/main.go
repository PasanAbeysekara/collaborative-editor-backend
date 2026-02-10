package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/config"
	"github.com/pasanAbeysekara/collaborative-editor/internal/grpchandlers"
	"github.com/pasanAbeysekara/collaborative-editor/internal/handlers"
	customMiddleware "github.com/pasanAbeysekara/collaborative-editor/internal/middleware"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	conn, err := amqp091.Dial(cfg.RabbitMQ_URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()
	auth.Initialize(cfg)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	var store storage.Store = storage.NewPostgresStore(pool)

	// Start gRPC server
	grpcPort := "50052"
	go startGRPCServer(store, ch, grpcPort)

	// Keep HTTP server for backwards compatibility (REST API)
	docHandler := &handlers.DocumentHandler{Store: store, AMQPChannel: ch}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.CORSMiddleware(customMiddleware.GetAllowedOrigins()))

	r.Handle("/metrics", promhttp.Handler())

	// Internal route for other services, no auth middleware needed
	r.Get("/documents/{documentID}/permissions/{userID}", docHandler.CheckPermission)
	r.Get("/documents/{documentID}", docHandler.GetDocument)
	r.Put("/documents/{documentID}", docHandler.SaveDocument)

	// Public-facing routes with auth
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)
		r.Get("/documents", docHandler.GetUserDocuments)
		r.Post("/documents", docHandler.CreateDocument)
		r.Post("/documents/{documentID}/share", docHandler.ShareDocument)
	})

	log.Printf("Starting document-service HTTP on port %s...\n", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, r)
}

func startGRPCServer(store storage.Store, amqpCh *amqp091.Channel, port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	docService := &grpchandlers.DocumentServiceServer{Store: store, AMQPChannel: amqpCh}
	pb.RegisterDocumentServiceServer(grpcServer, docService)

	log.Printf("Starting document-service gRPC on port %s...\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
