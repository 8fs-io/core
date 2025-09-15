package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server  ServerConfig
	Storage StorageConfig
	Auth    AuthConfig
	Metrics MetricsConfig
	Audit   AuditConfig
	Logger  LoggerConfig
	Vector  VectorConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Mode         string // gin.ReleaseMode, gin.DebugMode, gin.TestMode
}

type StorageConfig struct {
	Driver   string // filesystem, s3, memory
	BasePath string // for filesystem driver
	S3Config S3Config
}

type S3Config struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	Region         string
	Bucket         string
	UseSSL         bool
	ForcePathStyle bool
}

type AuthConfig struct {
	Enabled    bool
	Driver     string // signature, jwt, none
	JWTSecret  string
	TokenTTL   time.Duration
	DefaultKey struct {
		AccessKey string
		SecretKey string
	}
}

type MetricsConfig struct {
	Enabled              bool
	Path                 string
	UpdateInterval       time.Duration
	EnableGoMetrics      bool
	EnableProcessMetrics bool
}

type AuditConfig struct {
	Enabled    bool
	Driver     string // stdout, file, database
	FilePath   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

type LoggerConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	Output     string // stdout, file
	FilePath   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// VectorConfig holds vector storage related configuration
type VectorConfig struct {
	Enabled         bool
	DBPath          string
	EnableExtension bool
	Dimension       int
}

// Load reads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvOrDefaultInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvOrDefaultDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvOrDefaultDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvOrDefaultDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			Mode:         getEnvOrDefault("SERVER_MODE", "release"),
		},
		Storage: StorageConfig{
			Driver:   getEnvOrDefault("STORAGE_DRIVER", "filesystem"),
			BasePath: getEnvOrDefault("STORAGE_BASE_PATH", "./data"),
			S3Config: S3Config{
				Endpoint:       getEnvOrDefault("S3_ENDPOINT", ""),
				AccessKey:      getEnvOrDefault("S3_ACCESS_KEY", ""),
				SecretKey:      getEnvOrDefault("S3_SECRET_KEY", ""),
				Region:         getEnvOrDefault("S3_REGION", "us-east-1"),
				Bucket:         getEnvOrDefault("S3_BUCKET", "8fs-storage"),
				UseSSL:         getEnvOrDefaultBool("S3_USE_SSL", true),
				ForcePathStyle: getEnvOrDefaultBool("S3_FORCE_PATH_STYLE", false),
			},
		},
		Auth: AuthConfig{
			Enabled:   getEnvOrDefaultBool("AUTH_ENABLED", true),
			Driver:    getEnvOrDefault("AUTH_DRIVER", "signature"),
			JWTSecret: getEnvOrDefault("JWT_SECRET", "change-me-in-production"),
			TokenTTL:  getEnvOrDefaultDuration("JWT_TOKEN_TTL", 24*time.Hour),
			DefaultKey: struct {
				AccessKey string
				SecretKey string
			}{
				AccessKey: getEnvOrDefault("DEFAULT_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE"),
				SecretKey: getEnvOrDefault("DEFAULT_SECRET_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
			},
		},
		Metrics: MetricsConfig{
			Enabled:              getEnvOrDefaultBool("METRICS_ENABLED", true),
			Path:                 getEnvOrDefault("METRICS_PATH", "/metrics"),
			UpdateInterval:       getEnvOrDefaultDuration("METRICS_UPDATE_INTERVAL", 15*time.Second),
			EnableGoMetrics:      getEnvOrDefaultBool("METRICS_ENABLE_GO", true),
			EnableProcessMetrics: getEnvOrDefaultBool("METRICS_ENABLE_PROCESS", true),
		},
		Audit: AuditConfig{
			Enabled:    getEnvOrDefaultBool("AUDIT_ENABLED", true),
			Driver:     getEnvOrDefault("AUDIT_DRIVER", "stdout"),
			FilePath:   getEnvOrDefault("AUDIT_FILE_PATH", "./logs/audit.log"),
			MaxSize:    getEnvOrDefaultInt("AUDIT_MAX_SIZE", 100),
			MaxBackups: getEnvOrDefaultInt("AUDIT_MAX_BACKUPS", 3),
			MaxAge:     getEnvOrDefaultInt("AUDIT_MAX_AGE", 30),
			Compress:   getEnvOrDefaultBool("AUDIT_COMPRESS", true),
		},
		Logger: LoggerConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			Output:     getEnvOrDefault("LOG_OUTPUT", "stdout"),
			FilePath:   getEnvOrDefault("LOG_FILE_PATH", "./logs/app.log"),
			MaxSize:    getEnvOrDefaultInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvOrDefaultInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvOrDefaultInt("LOG_MAX_AGE", 30),
			Compress:   getEnvOrDefaultBool("LOG_COMPRESS", true),
		},
		Vector: VectorConfig{
			Enabled:         getEnvOrDefaultBool("VECTOR_ENABLED", true),
			DBPath:          getEnvOrDefault("VECTOR_DB_PATH", "data/vectors.db"),
			EnableExtension: getEnvOrDefaultBool("VECTOR_ENABLE_EXTENSION", false),
			Dimension:       getEnvOrDefaultInt("VECTOR_DIMENSION", 384),
		},
	}

	// Validation
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks if configuration values are valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Storage.Driver != "filesystem" && c.Storage.Driver != "s3" && c.Storage.Driver != "memory" {
		return fmt.Errorf("unsupported storage driver: %s", c.Storage.Driver)
	}

	if c.Auth.Driver != "signature" && c.Auth.Driver != "jwt" && c.Auth.Driver != "none" {
		return fmt.Errorf("unsupported auth driver: %s", c.Auth.Driver)
	}

	if c.Logger.Level != "debug" && c.Logger.Level != "info" && c.Logger.Level != "warn" && c.Logger.Level != "error" {
		return fmt.Errorf("unsupported log level: %s", c.Logger.Level)
	}

	if c.Logger.Format != "json" && c.Logger.Format != "text" {
		return fmt.Errorf("unsupported log format: %s", c.Logger.Format)
	}

	return nil
}

// Address returns the server address string
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
