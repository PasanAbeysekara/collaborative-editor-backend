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
)

func main() {
    cfg := config.Load()
    auth.Initialize(cfg)

    pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
    if err != nil { log.Fatalf("Unable to connect to database: %v\n", err) }
    defer pool.Close()

    var store storage.Store = storage.NewPostgresStore(pool)
    docHandler := &handlers.DocumentHandler{Store: store}

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // Internal route for other services, no auth middleware needed
    r.Get("/documents/{documentID}/permissions/{userID}", docHandler.CheckPermission)
    r.Get("/documents/{documentID}", docHandler.GetDocument)
    r.Put("/documents/{documentID}", docHandler.SaveDocument)

    // Public-facing routes with auth
    r.Group(func(r chi.Router) {
        r.Use(auth.JWTMiddleware)
        r.Post("/documents", docHandler.CreateDocument)
        r.Post("/documents/{documentID}/share", docHandler.ShareDocument)
    })

    log.Println("Starting document-service on port :8080...")
    http.ListenAndServe(":8080", r)
}