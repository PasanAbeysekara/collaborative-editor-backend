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
	"github.com/redis/go-redis/v9"
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.Load()
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Successfully connected to the database!")

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Could not parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}
	log.Println("Successfully connected to Redis!")

	auth.Initialize(cfg)

	var store storage.Store = storage.NewPostgresStore(pool)
	var cache storage.Cache = storage.NewRedisCache(redisClient)

	rtManager := realtime.NewManager(store, cache)

	userHandler := &handlers.UserHandler{Store: store}
	docHandler := &handlers.DocumentHandler{Store: store}

	r := chi.NewRouter()

	// Add CORS middleware first
	r.Use(corsMiddleware)
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

		r.Post("/documents", docHandler.CreateDocument)

		r.Post("/documents/{documentID}/share", docHandler.ShareDocument)

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
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}