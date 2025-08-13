package config

import (
	"os"
)

// Config captures environment-driven configuration for the API server.
type Config struct {
	DatabaseURL string
	Addr       string
}

// Load reads configuration from environment with sensible defaults.
func Load() Config {
	return Config{
        DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/go_gateway?sslmode=disable"),
		Addr:       ":" + getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}


