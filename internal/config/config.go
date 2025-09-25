package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates runtime settings for the service.
type Config struct {
	HTTP struct {
		Host string
		Port int
	}

	Database struct {
		DSN string
	}

	Auth struct {
		Username string
		Password string
	}

	FRC struct {
		BaseURL         string
		UploadAPIKey    string
		RecognizeAPIKey string
		TenantID        string
		RequestTimeout  time.Duration
	}

	Verification struct {
		DistanceThreshold   float64
		SimilarityThreshold float64
	}

	Liveness struct {
		Enabled bool
	}
}

// Load builds a Config using environment variables while applying sane defaults.
func Load() (*Config, error) {
	// Load local .env when present so API keys and other secrets are automatically available.
	_ = godotenv.Load(".env")

	cfg := &Config{}

	cfg.HTTP.Host = getEnv("HTTP_HOST", "0.0.0.0")
	portStr := getEnv("HTTP_PORT", "9800")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}
	cfg.HTTP.Port = port

	cfg.Database.DSN = getEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable")

	cfg.Auth.Username = getEnv("BASIC_AUTH_USERNAME", "")
	cfg.Auth.Password = getEnv("BASIC_AUTH_PASSWORD", "")

	cfg.FRC.BaseURL = getEnv("FRCORE_BASE_URL", "http://localhost:8000")
	cfg.FRC.UploadAPIKey = os.Getenv("FRCORE_UPLOAD_API_KEY")
	cfg.FRC.RecognizeAPIKey = os.Getenv("FRCORE_RECOGNIZE_API_KEY")
	cfg.FRC.TenantID = os.Getenv("FRCORE_TENANT_ID")

	timeoutStr := getEnv("FRCORE_TIMEOUT_SECONDS", "10")
	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid FRCORE_TIMEOUT_SECONDS: %w", err)
	}
	cfg.FRC.RequestTimeout = time.Duration(timeoutSeconds) * time.Second

	distanceStr := getEnv("VERIFICATION_DISTANCE_THRESHOLD", "0.6")
	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid VERIFICATION_DISTANCE_THRESHOLD: %w", err)
	}
	cfg.Verification.DistanceThreshold = distance

	similarityStr := getEnv("VERIFICATION_SIMILARITY_THRESHOLD", "75")
	similarity, err := strconv.ParseFloat(similarityStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid VERIFICATION_SIMILARITY_THRESHOLD: %w", err)
	}
	cfg.Verification.SimilarityThreshold = similarity

	cfg.Liveness.Enabled = getEnv("LIVENESS_ENABLED", "true") == "true"

	if cfg.Auth.Username == "" || cfg.Auth.Password == "" {
		return nil, fmt.Errorf("BASIC_AUTH_USERNAME and BASIC_AUTH_PASSWORD must be set")
	}

	if cfg.FRC.UploadAPIKey == "" {
		return nil, fmt.Errorf("FRCORE_UPLOAD_API_KEY is required")
	}
	if cfg.FRC.RecognizeAPIKey == "" {
		return nil, fmt.Errorf("FRCORE_RECOGNIZE_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
