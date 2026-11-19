package app

import (
	"fmt"
	"os"
)

// Config holds application configuration.
type Config struct {
	DatabasePath string
	S3Bucket     string
	S3Region     string
	S3PathPrefix string
	HTTPPort     string
	LogLevel     string
	Environment  string
}

// LoadConfig loads configuration from environment variables or config files.
func LoadConfig() (*Config, error) {
	config := &Config{
		DatabasePath: getEnvOrDefault("DB_PATH", "./rtsprec.db"),
		S3Bucket:     getEnvOrDefault("S3_BUCKET", "rtsprec-videos"),
		S3Region:     getEnvOrDefault("S3_REGION", "us-east-1"),
		S3PathPrefix: getEnvOrDefault("S3_PATH_PREFIX", "surveillance"),
		HTTPPort:     getEnvOrDefault("HTTP_PORT", "8080"),
		LogLevel:     getEnvOrDefault("LOG_LEVEL", "info"),
		Environment:  getEnvOrDefault("ENVIRONMENT", "development"),
	}

	// Validate required configuration
	if config.DatabasePath == "" {
		return nil, fmt.Errorf("missing required config: DB_PATH")
	}

	return config, nil
}

// getEnvOrDefault returns an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultValue
}
