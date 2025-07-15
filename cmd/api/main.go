package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/config"
	"github.com/pasanAbeysekara/collaborative-editor/internal/handlers"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
)

func main() {
	cfg := config.Load()

	// Establish a database connection pool
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Ping the database to ensure connectivity
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Successfully connected to the database!")

	auth.Initialize(cfg)

	var store storage.Store = storage.NewPostgresStore(pool)

	userHandler := &handlers.UserHandler{Store: store}
	docHandler := &handlers.DocumentHandler{Store: store}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes for authentication
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
	})

	// Protected routes - any route in here requires a valid JWT
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)

		// Document routes
		r.Post("/api/documents", docHandler.CreateDocument)

		// Example protected route to test authentication
		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := r.Context().Value(auth.UserIDKey).(string)
			// For demonstration, let's also fetch the user details
			user, err := store.GetUserByID(userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			w.Write([]byte("Hello, authenticated user: " + user.Email))
		})
	})

	listenAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s...", listenAddr)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
