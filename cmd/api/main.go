// cmd/api/main.go

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
	store := storage.NewMemoryStore()

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

	port := "8081"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
