package app

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration.
// All values are read from environment variables; see docs/deployment.md.
type Config struct {
	DatabasePath string
	HTTPPort     string
	LogLevel     string
	Environment  string

	// Encryption key used by the config store
	EncryptionKeyPath string

	// HLS streaming
	HLSOutputPath            string
	HLSSegmentDurationSecond int
	HLSSegmentsPerList       int

	// Authentication
	AuthMethod    string
	OIDCProvider  string
	OIDCClientID  string
	OIDCSecret    string
	BasicUsername string
	BasicPassword string

	// S3 storage
	S3Bucket      string
	S3Region      string
	S3PathPrefix  string
	S3AccessKeyID string
	S3SecretKey   string
}

// LoadConfig loads configuration from environment variables, falling back to defaults.
func LoadConfig() (*Config, error) {
	segmentDuration, err := getEnvInt("HLS_SEGMENT_DURATION_SECONDS", 2)
	if err != nil {
		return nil, err
	}
	segmentsPerList, err := getEnvInt("HLS_SEGMENTS_PER_LIST", 5)
	if err != nil {
		return nil, err
	}

	config := &Config{
		DatabasePath:      getEnvOrDefault("DB_PATH", "./rtsprec.db"),
		HTTPPort:          getEnvOrDefault("HTTP_PORT", "8080"),
		LogLevel:          getEnvOrDefault("LOG_LEVEL", "info"),
		Environment:       getEnvOrDefault("ENVIRONMENT", "development"),
		EncryptionKeyPath: getEnvOrDefault("ENCRYPTION_KEY_PATH", "/opt/rtsprec/.encryption.key"),

		HLSOutputPath:            getEnvOrDefault("HLS_OUTPUT_PATH", "/var/cache/rtsprec/hls"),
		HLSSegmentDurationSecond: segmentDuration,
		HLSSegmentsPerList:       segmentsPerList,

		AuthMethod:    getEnvOrDefault("AUTH_METHOD", "basic"),
		OIDCProvider:  getEnvOrDefault("AUTH_OIDC_PROVIDER", "google"),
		OIDCClientID:  getEnvOrDefault("AUTH_OIDC_CLIENT_ID", ""),
		OIDCSecret:    getEnvOrDefault("AUTH_OIDC_SECRET", ""),
		BasicUsername: getEnvOrDefault("AUTH_BASIC_USERNAME", "admin"),
		BasicPassword: getEnvOrDefault("AUTH_BASIC_PASSWORD", "changeme"),

		S3Bucket:      getEnvOrDefault("S3_BUCKET", "rtsprec-videos"),
		S3Region:      getEnvOrDefault("S3_REGION", "us-east-1"),
		S3PathPrefix:  getEnvOrDefault("S3_PATH_PREFIX", "surveillance"),
		S3AccessKeyID: getEnvOrDefault("S3_ACCESS_KEY_ID", ""),
		S3SecretKey:   getEnvOrDefault("S3_SECRET_ACCESS_KEY", ""),
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

// getEnvInt returns an integer environment variable or a default value.
func getEnvInt(key string, defaultValue int) (int, error) {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}
