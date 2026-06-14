package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	AppPort     string
	DatabaseURL string
	LogLevel    string
}

// Load reads configuration from environment variables, applying sensible
// defaults where a value is not provided.
func Load() *Config {
	cfg := &Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", buildDefaultDBURL()),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
	return cfg
}

func buildDefaultDBURL() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "userdb")
	ssl := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, name, ssl)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
