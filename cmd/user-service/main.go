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
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	auth.Initialize(cfg)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	var store storage.Store = storage.NewPostgresStore(pool)

	// Start gRPC server
	grpcPort := "50051"
	go startGRPCServer(store, grpcPort)

	// Keep HTTP server for backwards compatibility (REST API)
	userHandler := &handlers.UserHandler{Store: store}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.CORSMiddleware(customMiddleware.GetAllowedOrigins()))

	r.Handle("/metrics", promhttp.Handler())

	r.Post("/auth/register", userHandler.Register)
	r.Post("/auth/login", userHandler.Login)

	log.Printf("Starting user-service HTTP on port %s...\n", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, r)
}

func startGRPCServer(store storage.Store, port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	userService := &grpchandlers.UserServiceServer{Store: store}
	pb.RegisterUserServiceServer(grpcServer, userService)

	log.Printf("Starting user-service gRPC on port %s...\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
