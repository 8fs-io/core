package container

import (
	"fmt"

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
			aiCfg := &ai.OllamaConfig{
				BaseURL:    cfg.AI.BaseURL,
				EmbedModel: cfg.AI.Ollama.EmbedModel,
				ChatModel:  cfg.AI.Ollama.ChatModel,
				Timeout:    cfg.AI.Timeout,
			}
			c.AIService = ai.NewService(aiCfg, vecStore, appLogger)

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
				DefaultTopK:        5,
				DefaultMaxTokens:   4000,
				DefaultTemperature: 0.7,
				ContextWindowSize:  8000,
				MinRelevanceScore:  0.1,
				SystemPrompt:       "You are a helpful AI assistant. Use the provided context to answer questions accurately. If the context doesn't contain relevant information, say so clearly.",
			}
			c.RAGService = rag.NewService(c.AIService, vecStore, appLogger, ragConfig)
		}
	}

	return c, nil
}
