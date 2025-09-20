package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Storage  StorageConfig  `yaml:"storage"`
	Auth     AuthConfig     `yaml:"auth"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Audit    AuditConfig    `yaml:"audit"`
	Logger   LoggerConfig   `yaml:"logger"`
	Vector   VectorConfig   `yaml:"vector"`
	AI       AIConfig       `yaml:"ai"`
	Indexing IndexingConfig `yaml:"indexing"`
	RAG      RAGConfig      `yaml:"rag"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	Mode         string        `yaml:"mode"` // gin.ReleaseMode, gin.DebugMode, gin.TestMode
}

type StorageConfig struct {
	Driver   string   `yaml:"driver"`    // filesystem, s3, memory
	BasePath string   `yaml:"base_path"` // for filesystem driver
	S3Config S3Config `yaml:"s3"`
}

type S3Config struct {
	Endpoint       string `yaml:"endpoint"`
	AccessKey      string `yaml:"access_key"`
	SecretKey      string `yaml:"secret_key"`
	Region         string `yaml:"region"`
	Bucket         string `yaml:"bucket"`
	UseSSL         bool   `yaml:"use_ssl"`
	ForcePathStyle bool   `yaml:"force_path_style"`
}

type AuthConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Driver     string        `yaml:"driver"` // signature, jwt, none
	JWTSecret  string        `yaml:"jwt_secret"`
	TokenTTL   time.Duration `yaml:"token_ttl"`
	DefaultKey struct {
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"default_key"`
}

type MetricsConfig struct {
	Enabled              bool          `yaml:"enabled"`
	Path                 string        `yaml:"path"`
	UpdateInterval       time.Duration `yaml:"update_interval"`
	EnableGoMetrics      bool          `yaml:"enable_go_metrics"`
	EnableProcessMetrics bool          `yaml:"enable_process_metrics"`
}

type AuditConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Driver     string `yaml:"driver"` // stdout, file, database
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"` // MB
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"` // days
	Compress   bool   `yaml:"compress"`
}

type LoggerConfig struct {
	Level      string `yaml:"level"`  // debug, info, warn, error
	Format     string `yaml:"format"` // json, text
	Output     string `yaml:"output"` // stdout, file
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"` // MB
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"` // days
	Compress   bool   `yaml:"compress"`
}

// VectorConfig holds vector storage related configuration
type VectorConfig struct {
	Enabled   bool   `yaml:"enabled"`
	DBPath    string `yaml:"db_path"`
	Dimension int    `yaml:"dimension"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Provider   string        `yaml:"provider"` // ollama, openai, bedrock
	BaseURL    string        `yaml:"base_url"`
	Model      string        `yaml:"model"`
	Timeout    time.Duration `yaml:"timeout"`
	ChunkSize  int           `yaml:"chunk_size"`
	MaxRetries int           `yaml:"max_retries"`

	// Provider-specific configurations
	OpenAI  OpenAIConfig  `yaml:"openai"`
	Bedrock BedrockConfig `yaml:"bedrock"`
	Ollama  OllamaConfig  `yaml:"ollama"`
}

// OpenAIConfig holds OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	OrgID       string  `yaml:"org_id"`
	EmbedModel  string  `yaml:"embed_model"` // text-embedding-3-small, text-embedding-ada-002
	ChatModel   string  `yaml:"chat_model"`  // gpt-4, gpt-3.5-turbo
	MaxTokens   int     `yaml:"max_tokens"`  // Max tokens for generation
	Temperature float64 `yaml:"temperature"` // Generation temperature
}

// BedrockConfig holds AWS Bedrock-specific configuration
type BedrockConfig struct {
	Region          string  `yaml:"region"`
	AccessKeyID     string  `yaml:"access_key_id"`
	SecretAccessKey string  `yaml:"secret_access_key"`
	EmbedModel      string  `yaml:"embed_model"` // amazon.titan-embed-text-v1, etc.
	ChatModel       string  `yaml:"chat_model"`  // anthropic.claude-v2, etc.
	MaxTokens       int     `yaml:"max_tokens"`  // Max tokens for generation
	Temperature     float64 `yaml:"temperature"` // Generation temperature
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	EmbedModel  string  `yaml:"embed_model"` // all-minilm:latest
	ChatModel   string  `yaml:"chat_model"`  // llama2, codellama, etc.
	MaxTokens   int     `yaml:"max_tokens"`  // Max tokens for generation
	Temperature float64 `yaml:"temperature"` // Generation temperature
}

// IndexingConfig holds async indexing configuration
type IndexingConfig struct {
	Enabled       bool          `yaml:"enabled"`
	Workers       int           `yaml:"workers"`
	QueueSize     int           `yaml:"queue_size"`
	MaxRetries    int           `yaml:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	CleanupAfter  time.Duration `yaml:"cleanup_after"`
	StatusEnabled bool          `yaml:"status_enabled"`
}

type RAGConfig struct {
	DefaultTopK        int     `yaml:"default_top_k"`       // Default number of documents to retrieve
	DefaultMaxTokens   int     `yaml:"default_max_tokens"`  // Default max tokens for generation
	DefaultTemperature float64 `yaml:"default_temperature"` // Default generation temperature
	ContextWindowSize  int     `yaml:"context_window_size"` // Max context window size
	MinRelevanceScore  float64 `yaml:"min_relevance_score"` // Minimum relevance score for documents
	SystemPrompt       string  `yaml:"system_prompt"`       // Default system prompt
}

// Load reads configuration from YAML file first, then environment variables with defaults
func Load() (*Config, error) {
	// Try to load from YAML file first
	cfg, err := loadFromYAML()
	if err != nil {
		// If YAML loading fails, fall back to environment variables
		cfg = loadFromEnv()
	} else {
		// Merge with environment variables (env vars take precedence)
		mergeWithEnv(cfg)
	}

	// Validation
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// loadFromYAML loads configuration from YAML file
func loadFromYAML() (*Config, error) {
	configPath := getEnvOrDefault("CONFIG_FILE", "config.yml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables with defaults
func loadFromEnv() *Config {
	return &Config{
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
			Enabled:   determineAuthEnabled(),
			Driver:    getEnvOrDefault("AUTH_DRIVER", "signature"),
			JWTSecret: getEnvOrDefault("JWT_SECRET", "change-me-in-production"),
			TokenTTL:  getEnvOrDefaultDuration("JWT_TOKEN_TTL", 24*time.Hour),
			DefaultKey: struct {
				AccessKey string `yaml:"access_key"`
				SecretKey string `yaml:"secret_key"`
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
			Enabled:   getEnvOrDefaultBool("VECTOR_ENABLED", true),
			DBPath:    getEnvOrDefault("VECTOR_DB_PATH", "data/vectors.db"),
			Dimension: getEnvOrDefaultInt("VECTOR_DIMENSION", 384),
		},
		AI: AIConfig{
			Enabled:    getEnvOrDefaultBool("AI_ENABLED", true),
			Provider:   getEnvOrDefault("AI_PROVIDER", "ollama"),
			BaseURL:    getEnvOrDefault("AI_BASE_URL", "http://localhost:11434"),
			Model:      getEnvOrDefault("AI_MODEL", "all-minilm:latest"),
			Timeout:    getEnvOrDefaultDuration("AI_TIMEOUT", 30*time.Second),
			ChunkSize:  getEnvOrDefaultInt("AI_CHUNK_SIZE", 500),
			MaxRetries: getEnvOrDefaultInt("AI_MAX_RETRIES", 3),
			OpenAI: OpenAIConfig{
				APIKey:      getEnvOrDefault("OPENAI_API_KEY", ""),
				OrgID:       getEnvOrDefault("OPENAI_ORG_ID", ""),
				EmbedModel:  getEnvOrDefault("OPENAI_EMBED_MODEL", "text-embedding-3-small"),
				ChatModel:   getEnvOrDefault("OPENAI_CHAT_MODEL", "gpt-4o-mini"),
				MaxTokens:   getEnvOrDefaultInt("OPENAI_MAX_TOKENS", 4000),
				Temperature: getEnvOrDefaultFloat("OPENAI_TEMPERATURE", 0.7),
			},
			Bedrock: BedrockConfig{
				Region:          getEnvOrDefault("AWS_BEDROCK_REGION", "us-east-1"),
				AccessKeyID:     getEnvOrDefault("AWS_BEDROCK_ACCESS_KEY_ID", ""),
				SecretAccessKey: getEnvOrDefault("AWS_BEDROCK_SECRET_ACCESS_KEY", ""),
				EmbedModel:      getEnvOrDefault("AWS_BEDROCK_EMBED_MODEL", "amazon.titan-embed-text-v1"),
				ChatModel:       getEnvOrDefault("AWS_BEDROCK_CHAT_MODEL", "anthropic.claude-3-sonnet-20240229-v1:0"),
				MaxTokens:       getEnvOrDefaultInt("AWS_BEDROCK_MAX_TOKENS", 4000),
				Temperature:     getEnvOrDefaultFloat("AWS_BEDROCK_TEMPERATURE", 0.7),
			},
			Ollama: OllamaConfig{
				EmbedModel:  getEnvOrDefault("OLLAMA_EMBED_MODEL", "all-minilm:latest"),
				ChatModel:   getEnvOrDefault("OLLAMA_CHAT_MODEL", "llama3.2:1b"),
				MaxTokens:   getEnvOrDefaultInt("OLLAMA_MAX_TOKENS", 4000),
				Temperature: getEnvOrDefaultFloat("OLLAMA_TEMPERATURE", 0.7),
			},
		},
		Indexing: IndexingConfig{
			Enabled:       getEnvOrDefaultBool("INDEXING_ENABLED", true),
			Workers:       getEnvOrDefaultInt("INDEXING_WORKERS", 3),
			QueueSize:     getEnvOrDefaultInt("INDEXING_QUEUE_SIZE", 1000),
			MaxRetries:    getEnvOrDefaultInt("INDEXING_MAX_RETRIES", 3),
			RetryDelay:    getEnvOrDefaultDuration("INDEXING_RETRY_DELAY", 5*time.Second),
			CleanupAfter:  getEnvOrDefaultDuration("INDEXING_CLEANUP_AFTER", 24*time.Hour),
			StatusEnabled: getEnvOrDefaultBool("INDEXING_STATUS_ENABLED", true),
		},
		RAG: RAGConfig{
			DefaultTopK:        getEnvOrDefaultInt("RAG_DEFAULT_TOP_K", 5),
			DefaultMaxTokens:   getEnvOrDefaultInt("RAG_DEFAULT_MAX_TOKENS", 4000),
			DefaultTemperature: getEnvOrDefaultFloat("RAG_DEFAULT_TEMPERATURE", 0.7),
			ContextWindowSize:  getEnvOrDefaultInt("RAG_CONTEXT_WINDOW_SIZE", 8000),
			MinRelevanceScore:  getEnvOrDefaultFloat("RAG_MIN_RELEVANCE_SCORE", 0.1),
			SystemPrompt:       getEnvOrDefault("RAG_SYSTEM_PROMPT", "You are a helpful AI assistant. Use the provided context to answer questions accurately. If the context doesn't contain relevant information, say so clearly."),
		},
	}
}

// mergeWithEnv merges environment variables into the YAML config (env vars take precedence)
func mergeWithEnv(cfg *Config) {
	// Server config
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if portInt, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = portInt
		}
	}
	if timeout := os.Getenv("SERVER_READ_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			cfg.Server.ReadTimeout = duration
		}
	}
	if timeout := os.Getenv("SERVER_WRITE_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			cfg.Server.WriteTimeout = duration
		}
	}
	if timeout := os.Getenv("SERVER_IDLE_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			cfg.Server.IdleTimeout = duration
		}
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}

	// Storage config
	if driver := os.Getenv("STORAGE_DRIVER"); driver != "" {
		cfg.Storage.Driver = driver
	}
	if basePath := os.Getenv("STORAGE_BASE_PATH"); basePath != "" {
		cfg.Storage.BasePath = basePath
	}

	// S3 config
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		cfg.Storage.S3Config.Endpoint = endpoint
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY"); accessKey != "" {
		cfg.Storage.S3Config.AccessKey = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_KEY"); secretKey != "" {
		cfg.Storage.S3Config.SecretKey = secretKey
	}
	if region := os.Getenv("S3_REGION"); region != "" {
		cfg.Storage.S3Config.Region = region
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		cfg.Storage.S3Config.Bucket = bucket
	}
	if useSSL := os.Getenv("S3_USE_SSL"); useSSL != "" {
		if sslBool, err := strconv.ParseBool(useSSL); err == nil {
			cfg.Storage.S3Config.UseSSL = sslBool
		}
	}
	if forcePathStyle := os.Getenv("S3_FORCE_PATH_STYLE"); forcePathStyle != "" {
		if pathBool, err := strconv.ParseBool(forcePathStyle); err == nil {
			cfg.Storage.S3Config.ForcePathStyle = pathBool
		}
	}

	// Auth config - use our smart auth detection
	cfg.Auth.Enabled = determineAuthEnabled()
	if driver := os.Getenv("AUTH_DRIVER"); driver != "" {
		cfg.Auth.Driver = driver
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
	}
	if ttl := os.Getenv("JWT_TOKEN_TTL"); ttl != "" {
		if duration, err := time.ParseDuration(ttl); err == nil {
			cfg.Auth.TokenTTL = duration
		}
	}
	if accessKey := os.Getenv("DEFAULT_ACCESS_KEY"); accessKey != "" {
		cfg.Auth.DefaultKey.AccessKey = accessKey
	}
	if secretKey := os.Getenv("DEFAULT_SECRET_KEY"); secretKey != "" {
		cfg.Auth.DefaultKey.SecretKey = secretKey
	}

	// AI config
	if enabled := os.Getenv("AI_ENABLED"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			cfg.AI.Enabled = enabledBool
		}
	}
	if provider := os.Getenv("AI_PROVIDER"); provider != "" {
		cfg.AI.Provider = provider
	}
	if baseURL := os.Getenv("AI_BASE_URL"); baseURL != "" {
		cfg.AI.BaseURL = baseURL
	}
	if model := os.Getenv("AI_MODEL"); model != "" {
		cfg.AI.Model = model
	}
	if timeout := os.Getenv("AI_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			cfg.AI.Timeout = duration
		}
	}
	if chunkSize := os.Getenv("AI_CHUNK_SIZE"); chunkSize != "" {
		if size, err := strconv.Atoi(chunkSize); err == nil {
			cfg.AI.ChunkSize = size
		}
	}
	if retries := os.Getenv("AI_MAX_RETRIES"); retries != "" {
		if retriesInt, err := strconv.Atoi(retries); err == nil {
			cfg.AI.MaxRetries = retriesInt
		}
	}

	// OpenAI config
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		cfg.AI.OpenAI.APIKey = apiKey
	}
	if orgID := os.Getenv("OPENAI_ORG_ID"); orgID != "" {
		cfg.AI.OpenAI.OrgID = orgID
	}
	if embedModel := os.Getenv("OPENAI_EMBED_MODEL"); embedModel != "" {
		cfg.AI.OpenAI.EmbedModel = embedModel
	}
	if chatModel := os.Getenv("OPENAI_CHAT_MODEL"); chatModel != "" {
		cfg.AI.OpenAI.ChatModel = chatModel
	}
	if maxTokens := os.Getenv("OPENAI_MAX_TOKENS"); maxTokens != "" {
		if tokens, err := strconv.Atoi(maxTokens); err == nil {
			cfg.AI.OpenAI.MaxTokens = tokens
		}
	}
	if temperature := os.Getenv("OPENAI_TEMPERATURE"); temperature != "" {
		if temp, err := strconv.ParseFloat(temperature, 64); err == nil {
			cfg.AI.OpenAI.Temperature = temp
		}
	}

	// Bedrock config
	if region := os.Getenv("AWS_BEDROCK_REGION"); region != "" {
		cfg.AI.Bedrock.Region = region
	}
	if accessKeyID := os.Getenv("AWS_BEDROCK_ACCESS_KEY_ID"); accessKeyID != "" {
		cfg.AI.Bedrock.AccessKeyID = accessKeyID
	}
	if secretAccessKey := os.Getenv("AWS_BEDROCK_SECRET_ACCESS_KEY"); secretAccessKey != "" {
		cfg.AI.Bedrock.SecretAccessKey = secretAccessKey
	}
	if embedModel := os.Getenv("AWS_BEDROCK_EMBED_MODEL"); embedModel != "" {
		cfg.AI.Bedrock.EmbedModel = embedModel
	}
	if chatModel := os.Getenv("AWS_BEDROCK_CHAT_MODEL"); chatModel != "" {
		cfg.AI.Bedrock.ChatModel = chatModel
	}
	if maxTokens := os.Getenv("AWS_BEDROCK_MAX_TOKENS"); maxTokens != "" {
		if tokens, err := strconv.Atoi(maxTokens); err == nil {
			cfg.AI.Bedrock.MaxTokens = tokens
		}
	}
	if temperature := os.Getenv("AWS_BEDROCK_TEMPERATURE"); temperature != "" {
		if temp, err := strconv.ParseFloat(temperature, 64); err == nil {
			cfg.AI.Bedrock.Temperature = temp
		}
	}

	// Ollama config
	if embedModel := os.Getenv("OLLAMA_EMBED_MODEL"); embedModel != "" {
		cfg.AI.Ollama.EmbedModel = embedModel
	}
	if chatModel := os.Getenv("OLLAMA_CHAT_MODEL"); chatModel != "" {
		cfg.AI.Ollama.ChatModel = chatModel
	}
	if maxTokens := os.Getenv("OLLAMA_MAX_TOKENS"); maxTokens != "" {
		if tokens, err := strconv.Atoi(maxTokens); err == nil {
			cfg.AI.Ollama.MaxTokens = tokens
		}
	}
	if temperature := os.Getenv("OLLAMA_TEMPERATURE"); temperature != "" {
		if temp, err := strconv.ParseFloat(temperature, 64); err == nil {
			cfg.AI.Ollama.Temperature = temp
		}
	}

	// Vector config
	if enabled := os.Getenv("VECTOR_ENABLED"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			cfg.Vector.Enabled = enabledBool
		}
	}
	if dbPath := os.Getenv("VECTOR_DB_PATH"); dbPath != "" {
		cfg.Vector.DBPath = dbPath
	}
	if dimension := os.Getenv("VECTOR_DIMENSION"); dimension != "" {
		if dim, err := strconv.Atoi(dimension); err == nil {
			cfg.Vector.Dimension = dim
		}
	}

	// Indexing config
	if enabled := os.Getenv("INDEXING_ENABLED"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			cfg.Indexing.Enabled = enabledBool
		}
	}
	if workers := os.Getenv("INDEXING_WORKERS"); workers != "" {
		if workersInt, err := strconv.Atoi(workers); err == nil {
			cfg.Indexing.Workers = workersInt
		}
	}
	if queueSize := os.Getenv("INDEXING_QUEUE_SIZE"); queueSize != "" {
		if size, err := strconv.Atoi(queueSize); err == nil {
			cfg.Indexing.QueueSize = size
		}
	}
	if maxRetries := os.Getenv("INDEXING_MAX_RETRIES"); maxRetries != "" {
		if retries, err := strconv.Atoi(maxRetries); err == nil {
			cfg.Indexing.MaxRetries = retries
		}
	}
	if retryDelay := os.Getenv("INDEXING_RETRY_DELAY"); retryDelay != "" {
		if duration, err := time.ParseDuration(retryDelay); err == nil {
			cfg.Indexing.RetryDelay = duration
		}
	}
	if cleanupAfter := os.Getenv("INDEXING_CLEANUP_AFTER"); cleanupAfter != "" {
		if duration, err := time.ParseDuration(cleanupAfter); err == nil {
			cfg.Indexing.CleanupAfter = duration
		}
	}
	if statusEnabled := os.Getenv("INDEXING_STATUS_ENABLED"); statusEnabled != "" {
		if enabled, err := strconv.ParseBool(statusEnabled); err == nil {
			cfg.Indexing.StatusEnabled = enabled
		}
	}

	// RAG config
	if topK := os.Getenv("RAG_DEFAULT_TOP_K"); topK != "" {
		if topKInt, err := strconv.Atoi(topK); err == nil {
			cfg.RAG.DefaultTopK = topKInt
		}
	}
	if maxTokens := os.Getenv("RAG_DEFAULT_MAX_TOKENS"); maxTokens != "" {
		if maxTokensInt, err := strconv.Atoi(maxTokens); err == nil {
			cfg.RAG.DefaultMaxTokens = maxTokensInt
		}
	}
	if temperature := os.Getenv("RAG_DEFAULT_TEMPERATURE"); temperature != "" {
		if tempFloat, err := strconv.ParseFloat(temperature, 64); err == nil {
			cfg.RAG.DefaultTemperature = tempFloat
		}
	}
	if contextWindowSize := os.Getenv("RAG_CONTEXT_WINDOW_SIZE"); contextWindowSize != "" {
		if contextSizeInt, err := strconv.Atoi(contextWindowSize); err == nil {
			cfg.RAG.ContextWindowSize = contextSizeInt
		}
	}
	if minRelevance := os.Getenv("RAG_MIN_RELEVANCE_SCORE"); minRelevance != "" {
		if relevanceFloat, err := strconv.ParseFloat(minRelevance, 64); err == nil {
			cfg.RAG.MinRelevanceScore = relevanceFloat
		}
	}
	if systemPrompt := os.Getenv("RAG_SYSTEM_PROMPT"); systemPrompt != "" {
		cfg.RAG.SystemPrompt = systemPrompt
	}
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

func getEnvOrDefaultFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// determineAuthEnabled decides if authentication should be enabled based on credential configuration
func determineAuthEnabled() bool {
	// If AUTH_ENABLED is explicitly set, respect that setting
	if authEnv := os.Getenv("AUTH_ENABLED"); authEnv != "" {
		if boolValue, err := strconv.ParseBool(authEnv); err == nil {
			return boolValue
		}
	}

	// Check if custom credentials are provided
	hasCustomAccessKey := os.Getenv("DEFAULT_ACCESS_KEY") != ""
	hasCustomSecretKey := os.Getenv("DEFAULT_SECRET_KEY") != ""

	// If no custom credentials are provided, default to no auth
	if !hasCustomAccessKey && !hasCustomSecretKey {
		return false
	}

	// If custom credentials are provided, enable auth by default
	return true
}
