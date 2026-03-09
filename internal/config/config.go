package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the service.
// Values are loaded from environment variables; Load() returns an
// error if any required field is missing or invalid.
type Config struct {
	// Shared
	RedisAddr   string
	DatabaseURL string

	// Server
	GRPCPort string

	// Worker
	WorkerCount int
	AssetsDir   string
}

// Load reads configuration from environment variables.
// It returns an error if a required variable is absent or invalid.
func Load() (*Config, error) {
	cfg := &Config{
		RedisAddr:   env("REDIS_ADDR", "localhost:6379"),
		GRPCPort:    env("GRPC_PORT", "50051"),
		WorkerCount: 2,
	}

	var missing []string

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	cfg.AssetsDir = os.Getenv("ASSETS_DIRECTORY")
	if cfg.AssetsDir == "" {
		missing = append(missing, "ASSETS_DIRECTORY")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	if v := os.Getenv("WORKER_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("WORKER_COUNT must be a positive integer, got %q", v)
		}
		cfg.WorkerCount = n
	}

	return cfg, nil
}

// env returns the value of the environment variable named by key,
// or fallback if the variable is not set.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
