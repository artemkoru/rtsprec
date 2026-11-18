package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port     int    `yaml:"port"`
		Host     string `yaml:"host"`
		TLSCert  string `yaml:"tls_cert"`
		TLSKey   string `yaml:"tls_key"`
	} `yaml:"server"`

	AWS struct {
		Region          string `yaml:"region"`
		S3Bucket        string `yaml:"s3_bucket"`
		S3PathPrefix    string `yaml:"s3_path_prefix"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
	} `yaml:"aws"`

	Auth struct {
		Method       string `yaml:"method"` // "oidc" or "basic"
		OIDCProvider string `yaml:"oidc_provider"`
		OIDCClientID string `yaml:"oidc_client_id"`
		OIDCSecret   string `yaml:"oidc_secret"`
		BasicUsername string `yaml:"basic_username"`
		BasicPassword string `yaml:"basic_password"`
	} `yaml:"auth"`

	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`

	HLS struct {
		SegmentDuration int    `yaml:"segment_duration_seconds"`
		SegmentsPerList int    `yaml:"segments_per_list"`
		OutputPath      string `yaml:"output_path"`
	} `yaml:"hls"`
}

type ConfigStore struct {
	db    *sql.DB
	mu    sync.RWMutex
	cache map[string]string
	key   [32]byte // 256-bit encryption key
}

func NewConfigStore(db *sql.DB, encryptionKeyFile string) (*ConfigStore, error) {
	cs := &ConfigStore{
		db:    db,
		cache: make(map[string]string),
	}

	// Load or create encryption key
	key, err := loadOrCreateEncryptionKey(encryptionKeyFile)
	if err != nil {
		return nil, err
	}
	copy(cs.key[:], key)

	return cs, nil
}

func (cs *ConfigStore) Get(key string) (string, error) {
	cs.mu.RLock()
	if val, exists := cs.cache[key]; exists {
		cs.mu.RUnlock()
		return val, nil
	}
	cs.mu.RUnlock()

	var encrypted string
	err := cs.db.QueryRow("SELECT value FROM configuration WHERE key = ?", key).Scan(&encrypted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	// Decrypt if it starts with encrypted marker
	if len(encrypted) > 11 && encrypted[:11] == "encrypted:" {
		decrypted, err := cs.decrypt(encrypted[11:])
		if err != nil {
			return "", err
		}
		encrypted = decrypted
	}

	cs.mu.Lock()
	cs.cache[key] = encrypted
	cs.mu.Unlock()

	return encrypted, nil
}

func (cs *ConfigStore) Set(key, value string, encrypted bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var stored string
	if encrypted {
		encrypted, err := cs.encrypt(value)
		if err != nil {
			return err
		}
		stored = "encrypted:" + encrypted
	} else {
		stored = value
	}

	_, err := cs.db.Exec(
		"INSERT INTO configuration (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?",
		key, stored, stored,
	)
	if err != nil {
		return err
	}

	cs.cache[key] = value
	return nil
}

func (cs *ConfigStore) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(cs.key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (cs *ConfigStore) decrypt(ciphertext string) (string, error) {
	block, err := aes.NewCipher(cs.key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonce, ciphertext2 := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext2, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func loadOrCreateEncryptionKey(keyFile string) ([]byte, error) {
	// If key file exists, load it
	if data, err := os.ReadFile(keyFile); err == nil && len(data) == 32 {
		return data, nil
	}

	// Create new key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	// Save key
	if err := os.WriteFile(keyFile, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encryption key: %w", err)
	}

	return key, nil
}

func LoadFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "/opt/rtsprec/data/rtsprec.db"
	}
	if cfg.HLS.SegmentDuration == 0 {
		cfg.HLS.SegmentDuration = 2
	}
	if cfg.HLS.SegmentsPerList == 0 {
		cfg.HLS.SegmentsPerList = 5
	}
	if cfg.HLS.OutputPath == "" {
		cfg.HLS.OutputPath = "/tmp/hls"
	}

	return cfg, nil
}
