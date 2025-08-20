package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/config"
	"github.com/pasanAbeysekara/collaborative-editor/internal/handlers"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	docHandler := &handlers.DocumentHandler{Store: store}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

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

	log.Printf("Starting document-service on port %s...\n", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, r)
}
