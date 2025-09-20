package container

import (
	"fmt"
	"strings"

	"github.com/8fs-io/core/internal/config"
	"github.com/8fs-io/core/internal/domain/ai"
	"github.com/8fs-io/core/internal/domain/indexing"
	"github.com/8fs-io/core/internal/domain/rag"
	"github.com/8fs-io/core/internal/domain/storage"
	"github.com/8fs-io/core/internal/domain/vectors"
	storageInfra "github.com/8fs-io/core/internal/infrastructure/storage"
	"github.com/8fs-io/core/pkg/logger"
)

// Container holds all application dependencies
type Container struct {
	Config          *config.Config
	Logger          logger.Logger
	AuditLogger     *logger.AuditLogger
	StorageRepo     storage.Repository
	StorageService  storage.Service
	Validator       storage.Validator
	VectorStorage   *vectors.SQLiteVecStorage
	AIService       ai.Service
	IndexingService indexing.Service
	RAGService      rag.Service
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		FilePath:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize audit logger
	auditLogger := logger.NewAuditLogger(appLogger)

	// Initialize validator
	validator := storage.NewValidator()

	// Initialize storage repository based on config
	var storageRepo storage.Repository
	switch cfg.Storage.Driver {
	case "filesystem":
		storageRepo, err = storageInfra.NewFilesystemRepository(cfg.Storage.BasePath, appLogger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize filesystem storage: %w", err)
		}
	case "memory":
		// TODO: Implement memory storage
		return nil, fmt.Errorf("memory storage not implemented yet")
	case "s3":
		// TODO: Implement S3 storage
		return nil, fmt.Errorf("S3 storage not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}

	// Initialize storage service
	storageService := storage.NewService(storageRepo, validator, appLogger)

	c := &Container{
		Config:         cfg,
		Logger:         appLogger,
		AuditLogger:    auditLogger,
		StorageRepo:    storageRepo,
		StorageService: storageService,
		Validator:      validator,
	}

	// Initialize vector storage if enabled
	if cfg.Vector.Enabled {
		vecCfg := vectors.SQLiteVecConfig{Path: cfg.Vector.DBPath, Dimension: cfg.Vector.Dimension}
		vecStore, err := vectors.NewSQLiteVecStorage(vecCfg, appLogger)
		if err != nil {
			// Log but don't fail container creation
			appLogger.Warn("vector storage initialization failed", "error", err)
		} else {
			c.VectorStorage = vecStore

			// Initialize AI service if vector storage is available
			aiService, err := initAIService(cfg, vecStore, appLogger)
			if err != nil {
				appLogger.Error("AI service initialization failed", "error", err)
			} else {
				c.AIService = aiService
			}

			// Initialize async indexing service
			indexingCfg := &indexing.Config{
				Enabled:       cfg.Indexing.Enabled,
				Workers:       cfg.Indexing.Workers,
				QueueSize:     cfg.Indexing.QueueSize,
				MaxRetries:    cfg.Indexing.MaxRetries,
				RetryDelay:    cfg.Indexing.RetryDelay,
				CleanupAfter:  cfg.Indexing.CleanupAfter,
				StatusEnabled: cfg.Indexing.StatusEnabled,
			}
			c.IndexingService = indexing.NewService(indexingCfg, c.AIService, appLogger)

			// Initialize RAG service
			ragConfig := &rag.Config{
				DefaultTopK:        cfg.RAG.DefaultTopK,
				DefaultMaxTokens:   cfg.RAG.DefaultMaxTokens,
				DefaultTemperature: cfg.RAG.DefaultTemperature,
				ContextWindowSize:  cfg.RAG.ContextWindowSize,
				MinRelevanceScore:  cfg.RAG.MinRelevanceScore,
				SystemPrompt:       cfg.RAG.SystemPrompt,
			}
			c.RAGService = rag.NewService(c.AIService, vecStore, appLogger, ragConfig)
		}
	}

	return c, nil
}

// initAIService creates an AI service based on the configured provider
func initAIService(cfg *config.Config, vectorStorage *vectors.SQLiteVecStorage, logger logger.Logger) (ai.Service, error) {
	if !cfg.AI.Enabled {
		logger.Info("AI service disabled in configuration")
		return nil, nil
	}

	switch strings.ToLower(cfg.AI.Provider) {
	case "ollama":
		return initOllamaService(cfg, vectorStorage, logger)
	case "openai":
		return initOpenAIService(cfg, vectorStorage, logger)
	case "bedrock":
		return initBedrockService(cfg, vectorStorage, logger)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.AI.Provider)
	}
}

// initOllamaService creates an Ollama AI service
func initOllamaService(cfg *config.Config, vectorStorage *vectors.SQLiteVecStorage, logger logger.Logger) (ai.Service, error) {
	ollamaCfg := &ai.OllamaConfig{
		BaseURL:    cfg.AI.BaseURL,
		EmbedModel: cfg.AI.Ollama.EmbedModel,
		ChatModel:  cfg.AI.Ollama.ChatModel,
		Timeout:    cfg.AI.Timeout,
	}

	service := ai.NewService(ollamaCfg, vectorStorage, logger)
	logger.Info("initialized Ollama AI service", "base_url", cfg.AI.BaseURL, "embed_model", cfg.AI.Ollama.EmbedModel, "chat_model", cfg.AI.Ollama.ChatModel)
	return service, nil
}

// initOpenAIService creates an OpenAI AI service (placeholder - needs implementation)
func initOpenAIService(cfg *config.Config, vectorStorage *vectors.SQLiteVecStorage, logger logger.Logger) (ai.Service, error) {
	logger.Warn("OpenAI AI service not yet implemented, falling back to Ollama")
	// TODO: Implement OpenAI service
	// For now, fall back to Ollama with a warning
	return initOllamaService(cfg, vectorStorage, logger)
}

// initBedrockService creates an AWS Bedrock AI service (placeholder - needs implementation)
func initBedrockService(cfg *config.Config, vectorStorage *vectors.SQLiteVecStorage, logger logger.Logger) (ai.Service, error) {
	logger.Warn("AWS Bedrock AI service not yet implemented, falling back to Ollama")
	// TODO: Implement Bedrock service
	// For now, fall back to Ollama with a warning
	return initOllamaService(cfg, vectorStorage, logger)
}
