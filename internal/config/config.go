package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Server      ServerConfig
	Model       ModelConfig
	Upload      UploadConfig
	CORS        CORSConfig
	Logging     LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port           int
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int
	MaxHeaderBytes int
	RateLimit      float64
	RateBurst      int
}

// ModelConfig holds model-related configuration
type ModelConfig struct {
	Path         string
	Version      string
	UpdateURL    string
	CachePath    string
	MaxModels    int
	LoadTimeout  int
}

// UploadConfig holds upload-related configuration
type UploadConfig struct {
	MaxFileSize   int64
	AllowedTypes  []string
	UploadDir     string
	TempDir       string
	CleanupAfter  int
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level  string
	Output string
	File   string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Debug("No .env file found, using environment variables")
	}

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:           getEnvAsInt("PORT", 8080),
			ReadTimeout:    getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout:   getEnvAsInt("WRITE_TIMEOUT", 30),
			IdleTimeout:    getEnvAsInt("IDLE_TIMEOUT", 120),
			MaxHeaderBytes: getEnvAsInt("MAX_HEADER_BYTES", 1048576), // 1MB
			RateLimit:      getEnvAsFloat64("RATE_LIMIT", 10.0),
			RateBurst:      getEnvAsInt("RATE_BURST", 20),
		},
		Model: ModelConfig{
			Path:        getEnv("MODEL_PATH", "./models"),
			Version:     getEnv("MODEL_VERSION", "latest"),
			UpdateURL:   getEnv("MODEL_UPDATE_URL", ""),
			CachePath:   getEnv("MODEL_CACHE_PATH", "./cache/models"),
			MaxModels:   getEnvAsInt("MAX_MODELS", 3),
			LoadTimeout: getEnvAsInt("MODEL_LOAD_TIMEOUT", 60),
		},
		Upload: UploadConfig{
			MaxFileSize:  getEnvAsInt64("MAX_FILE_SIZE", 10485760), // 10MB
			AllowedTypes: getEnvAsSlice("ALLOWED_TYPES", []string{"image/jpeg", "image/png", "image/webp"}),
			UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
			TempDir:      getEnv("TEMP_DIR", "./temp"),
			CleanupAfter: getEnvAsInt("CLEANUP_AFTER", 3600), // 1 hour
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"*"}),
			ExposedHeaders:   getEnvAsSlice("CORS_EXPOSED_HEADERS", []string{}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 86400), // 24 hours
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
			File:   getEnv("LOG_FILE", ""),
		},
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Upload.MaxFileSize <= 0 {
		return fmt.Errorf("invalid max file size: %d", config.Upload.MaxFileSize)
	}

	if len(config.Upload.AllowedTypes) == 0 {
		return fmt.Errorf("no allowed file types specified")
	}

	// Create necessary directories
	dirs := []string{
		config.Upload.UploadDir,
		config.Upload.TempDir,
		config.Model.Path,
		config.Model.CachePath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}