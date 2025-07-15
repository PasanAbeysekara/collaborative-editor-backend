package config

import "os"

type Config struct {
	DatabaseURL string
}

func Load() *Config {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@host:port/db_name?sslmode=disable"
	}

	return &Config{
		DatabaseURL: dbURL,
	}
}
