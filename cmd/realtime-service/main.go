package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/config"
	"github.com/pasanAbeysekara/collaborative-editor/internal/realtime"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
	"github.com/redis/go-redis/v9"
)

func main() {
    cfg := config.Load()
    auth.Initialize(cfg)

    redisOpts, _ := redis.ParseURL(cfg.RedisURL)
    redisClient := redis.NewClient(redisOpts)
    var cache storage.Cache = storage.NewRedisCache(redisClient)
    
    // Pass the service URL from config/env
    rtManager := realtime.NewManager(cache, cfg.DocumentServiceURL)

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Group(func(r chi.Router) {
        r.Use(auth.JWTMiddleware)
        r.Get("/ws/doc/{documentID}", rtManager.ServeWS)
    })

    log.Println("Starting realtime-service on port :8081...")
    http.ListenAndServe(":8081", r)
}