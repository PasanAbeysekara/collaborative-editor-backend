package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/config"
	"github.com/pasanAbeysekara/collaborative-editor/internal/handlers"
	"github.com/pasanAbeysekara/collaborative-editor/internal/realtime"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
)

func main() {
	cfg := config.Load()
	rtManager := realtime.NewManager()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

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

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)

		r.Post("/api/documents", docHandler.CreateDocument)

		r.Get("/ws/doc/{documentID}", rtManager.ServeWS)

		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := r.Context().Value(auth.UserIDKey).(string)
			user, err := store.GetUserByID(userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			w.Write([]byte("Hello, authenticated user: " + user.Email))
		})
	})

	port := strings.TrimSpace(cfg.Port)
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	log.Printf("Starting server on %s...", port)
}
