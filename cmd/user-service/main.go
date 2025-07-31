package main
// This main.go only contains routes for /auth/*
// It uses the same internal packages as before.
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
    userHandler := &handlers.UserHandler{Store: store}

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Post("/auth/register", userHandler.Register)
    r.Post("/auth/login", userHandler.Login)

    log.Printf("Starting user-service on port %s...\n", cfg.Port)
    http.ListenAndServe(":"+cfg.Port, r)
}