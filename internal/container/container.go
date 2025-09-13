package container

import (
	"fmt"

	"github.com/8fs/8fs/internal/config"
	"github.com/8fs/8fs/internal/domain/storage"
	storageInfra "github.com/8fs/8fs/internal/infrastructure/storage"
	"github.com/8fs/8fs/pkg/logger"
)

// Container holds all application dependencies
type Container struct {
	Config         *config.Config
	Logger         logger.Logger
	AuditLogger    *logger.AuditLogger
	StorageRepo    storage.Repository
	StorageService storage.Service
	Validator      storage.Validator
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

	return &Container{
		Config:         cfg,
		Logger:         appLogger,
		AuditLogger:    auditLogger,
		StorageRepo:    storageRepo,
		StorageService: storageService,
		Validator:      validator,
	}, nil
}
