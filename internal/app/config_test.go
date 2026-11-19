package app

import (
	"os"
	"testing"
)

// allConfigEnv lists every environment variable read by LoadConfig.
var allConfigEnv = []string{
	"DB_PATH", "HTTP_PORT", "LOG_LEVEL", "ENVIRONMENT", "ENCRYPTION_KEY_PATH",
	"HLS_OUTPUT_PATH", "HLS_SEGMENT_DURATION_SECONDS", "HLS_SEGMENTS_PER_LIST",
	"AUTH_METHOD", "AUTH_OIDC_PROVIDER", "AUTH_OIDC_CLIENT_ID", "AUTH_OIDC_SECRET",
	"AUTH_BASIC_USERNAME", "AUTH_BASIC_PASSWORD",
	"S3_BUCKET", "S3_REGION", "S3_PATH_PREFIX", "S3_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY",
}

// clearConfigEnv unsets all config variables for the duration of the test.
// t.Setenv + Unsetenv makes the test restore the previous values on cleanup.
func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, k := range allConfigEnv {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	checks := map[string][2]any{
		"DatabasePath":             {cfg.DatabasePath, "./rtsprec.db"},
		"HTTPPort":                 {cfg.HTTPPort, "8080"},
		"LogLevel":                 {cfg.LogLevel, "info"},
		"Environment":              {cfg.Environment, "development"},
		"EncryptionKeyPath":        {cfg.EncryptionKeyPath, "/opt/rtsprec/.encryption.key"},
		"HLSOutputPath":            {cfg.HLSOutputPath, "/var/cache/rtsprec/hls"},
		"HLSSegmentDurationSecond": {cfg.HLSSegmentDurationSecond, 2},
		"HLSSegmentsPerList":       {cfg.HLSSegmentsPerList, 5},
		"AuthMethod":               {cfg.AuthMethod, "basic"},
		"OIDCProvider":             {cfg.OIDCProvider, "google"},
		"BasicUsername":            {cfg.BasicUsername, "admin"},
		"BasicPassword":            {cfg.BasicPassword, "changeme"},
		"S3Bucket":                 {cfg.S3Bucket, "rtsprec-videos"},
		"S3Region":                 {cfg.S3Region, "us-east-1"},
		"S3PathPrefix":             {cfg.S3PathPrefix, "surveillance"},
	}
	for name, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %v, want default %v", name, c[0], c[1])
		}
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("DB_PATH", "/data/test.db")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ENCRYPTION_KEY_PATH", "/secrets/key")
	t.Setenv("HLS_OUTPUT_PATH", "/tmp/hls")
	t.Setenv("HLS_SEGMENT_DURATION_SECONDS", "4")
	t.Setenv("HLS_SEGMENTS_PER_LIST", "10")
	t.Setenv("AUTH_METHOD", "oidc")
	t.Setenv("AUTH_OIDC_PROVIDER", "yandex")
	t.Setenv("AUTH_OIDC_CLIENT_ID", "client-id")
	t.Setenv("AUTH_OIDC_SECRET", "oidc-secret")
	t.Setenv("AUTH_BASIC_USERNAME", "operator")
	t.Setenv("AUTH_BASIC_PASSWORD", "s3cret")
	t.Setenv("S3_BUCKET", "my-bucket")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("S3_PATH_PREFIX", "cams")
	t.Setenv("S3_ACCESS_KEY_ID", "AKID")
	t.Setenv("S3_SECRET_ACCESS_KEY", "SECRET")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	checks := map[string][2]any{
		"DatabasePath":             {cfg.DatabasePath, "/data/test.db"},
		"HTTPPort":                 {cfg.HTTPPort, "9090"},
		"LogLevel":                 {cfg.LogLevel, "debug"},
		"Environment":              {cfg.Environment, "production"},
		"EncryptionKeyPath":        {cfg.EncryptionKeyPath, "/secrets/key"},
		"HLSOutputPath":            {cfg.HLSOutputPath, "/tmp/hls"},
		"HLSSegmentDurationSecond": {cfg.HLSSegmentDurationSecond, 4},
		"HLSSegmentsPerList":       {cfg.HLSSegmentsPerList, 10},
		"AuthMethod":               {cfg.AuthMethod, "oidc"},
		"OIDCProvider":             {cfg.OIDCProvider, "yandex"},
		"OIDCClientID":             {cfg.OIDCClientID, "client-id"},
		"OIDCSecret":               {cfg.OIDCSecret, "oidc-secret"},
		"BasicUsername":            {cfg.BasicUsername, "operator"},
		"BasicPassword":            {cfg.BasicPassword, "s3cret"},
		"S3Bucket":                 {cfg.S3Bucket, "my-bucket"},
		"S3Region":                 {cfg.S3Region, "eu-west-1"},
		"S3PathPrefix":             {cfg.S3PathPrefix, "cams"},
		"S3AccessKeyID":            {cfg.S3AccessKeyID, "AKID"},
		"S3SecretKey":              {cfg.S3SecretKey, "SECRET"},
	}
	for name, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %v, want %v", name, c[0], c[1])
		}
	}
}

func TestLoadConfigInvalidIntegerEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("HLS_SEGMENTS_PER_LIST", "many")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() error = nil, want error for non-numeric HLS_SEGMENTS_PER_LIST")
	}
}
