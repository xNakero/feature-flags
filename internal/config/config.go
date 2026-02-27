package config

import (
	"errors"
	"os"
)

// Config holds all application configuration.
type Config struct {
	HTTPAddr    string
	PostgresDSN string
	RedisAddr   string
	LogLevel    string
}

// Load reads configuration from environment variables and returns a Config.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN environment variable is required")
	}

	cfg := &Config{
		PostgresDSN: postgresDSN,
		HTTPAddr:    getEnvOrDefault("HTTP_ADDR", ":8080"),
		RedisAddr:   getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
