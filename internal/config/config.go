package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port               string `envconfig:"PORT" default:"8080"`
	DatabaseURL        string `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret          string `envconfig:"JWT_SECRET" required:"true"`
	RedisURL           string `envconfig:"REDIS_URL" required:"true"`
	DocumentServiceURL string `envconfig:"DOCUMENT_SERVICE_URL"`
	RabbitMQ_URL       string `envconfig:"RABBITMQ_URL" required:"true"`
}

func Load() *Config {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	return &cfg
}
