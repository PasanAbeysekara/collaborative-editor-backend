package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/handlers"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
)

func main() {
	// Initialize storage
	store := storage.NewMemoryStore()

	// Initialize handlers
	userHandler := &handlers.UserHandler{Store: store}

	// Initialize router
	r := chi.NewRouter()

	// A good base middleware stack
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

		// Example protected route to test authentication
		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := r.Context().Value(auth.UserIDKey).(string)
			w.Write([]byte("Hello, authenticated user with ID: " + userID))
		})
	})

	port := "8081"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}