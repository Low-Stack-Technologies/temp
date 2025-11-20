package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Storage
	StorageDirectory string
	
	// Server
	ServerPort int
	BaseURL    string
	
	// Upload Limits
	MaxUploadSizeMB   int64
	MaxFilesPerUpload int
	
	// Security
	RSAKeyPath string
	
	// Cleanup
	CleanupIntervalSeconds int
	MinTTLSeconds          int
	MaxTTLSeconds          int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()
	
	cfg := &Config{
		StorageDirectory:       getEnv("STORAGE_DIRECTORY", "./storage"),
		ServerPort:             getEnvAsInt("SERVER_PORT", 8080),
		BaseURL:                getEnv("BASE_URL", "http://localhost:8080"),
		MaxUploadSizeMB:        int64(getEnvAsInt("MAX_UPLOAD_SIZE_MB", 1024)),
		MaxFilesPerUpload:      getEnvAsInt("MAX_FILES_PER_UPLOAD", 10),
		RSAKeyPath:             getEnv("RSA_KEY_PATH", "./storage/private_key.pem"),
		CleanupIntervalSeconds: getEnvAsInt("CLEANUP_INTERVAL_SECONDS", 60),
		MinTTLSeconds:          getEnvAsInt("MIN_TTL_SECONDS", 60),
		MaxTTLSeconds:          getEnvAsInt("MAX_TTL_SECONDS", 2592000),
	}
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.StorageDirectory == "" {
		return fmt.Errorf("STORAGE_DIRECTORY cannot be empty")
	}
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("BASE_URL cannot be empty")
	}
	if c.MaxUploadSizeMB <= 0 {
		return fmt.Errorf("MAX_UPLOAD_SIZE_MB must be positive")
	}
	if c.MaxFilesPerUpload <= 0 {
		return fmt.Errorf("MAX_FILES_PER_UPLOAD must be positive")
	}
	if c.MinTTLSeconds <= 0 {
		return fmt.Errorf("MIN_TTL_SECONDS must be positive")
	}
	if c.MaxTTLSeconds <= c.MinTTLSeconds {
		return fmt.Errorf("MAX_TTL_SECONDS must be greater than MIN_TTL_SECONDS")
	}
	return nil
}

// MaxUploadSizeBytes returns the max upload size in bytes
func (c *Config) MaxUploadSizeBytes() int64 {
	return c.MaxUploadSizeMB * 1024 * 1024
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
