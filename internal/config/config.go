package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port           string
	FrontendOrigin string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	OAuthRedirectURL   string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		Auth: AuthConfig{
			GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			OAuthRedirectURL:   getEnv("OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "dev-secret-change-me"),
		},
	}

	// Validate critical configuration
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL must be set")
	}

	if cfg.JWT.Secret == "dev-secret-change-me" && os.Getenv("GO_ENV") == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
