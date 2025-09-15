package router

import (
	"github.com/8fs/8fs/internal/container"
	"github.com/8fs/8fs/internal/domain/vectors"
	"github.com/8fs/8fs/internal/transport/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes
func SetupRoutes(r *gin.Engine, c *container.Container) {
	// Health and metrics endpoints (no auth required)
	r.GET("/healthz", handlers.NewHealthHandler(c).Handle)

	if c.Config.Metrics.Enabled {
		r.GET(c.Config.Metrics.Path, handlers.NewMetricsHandler(c).Handle)
	}

	// API version 1
	v1 := r.Group("/api/v1")
	{
		// Storage endpoints
		storage := v1.Group("/storage")
		{
			storageHandler := handlers.NewStorageHandler(c)
			storage.GET("/buckets", storageHandler.ListBuckets)
			storage.POST("/buckets/:bucket", storageHandler.CreateBucket)
			storage.DELETE("/buckets/:bucket", storageHandler.DeleteBucket)
			storage.GET("/buckets/:bucket", storageHandler.GetBucket)

			storage.PUT("/buckets/:bucket/objects/*key", storageHandler.PutObject)
			storage.GET("/buckets/:bucket/objects/*key", storageHandler.GetObject)
			storage.HEAD("/buckets/:bucket/objects/*key", storageHandler.HeadObject)
			storage.DELETE("/buckets/:bucket/objects/*key", storageHandler.DeleteObject)
			storage.GET("/buckets/:bucket/objects", storageHandler.ListObjects)
		}

		// Vector endpoints (experimental)
		vectorGroup := v1.Group("/vectors")
		{
			// Initialize vector storage (this would be dependency injected in production)
			vectorStorage, err := vectors.NewSQLiteVecStorage("data/vectors.db")
			if err != nil {
				// Log error and continue without vector endpoints
				// In production, this should be handled more gracefully
				return
			}
			
			vectorHandler := handlers.NewVectorHandler(c, vectorStorage)
			
			// Vector CRUD operations
			vectorGroup.POST("/embeddings", vectorHandler.StoreEmbedding)
			vectorGroup.POST("/search", vectorHandler.SearchEmbeddings)
			vectorGroup.GET("/embeddings/:id", vectorHandler.GetEmbedding)
			vectorGroup.GET("/embeddings", vectorHandler.ListEmbeddings)
			vectorGroup.DELETE("/embeddings/:id", vectorHandler.DeleteEmbedding)
		}
	}

	// S3-compatible endpoints (with optional auth)
	if c.Config.Auth.Enabled && c.Config.Auth.Driver == "signature" {
		// Apply AWS signature middleware to S3 endpoints
		s3Group := r.Group("")
		s3Group.Use(handlers.NewAuthHandler(c).AWSSignatureMiddleware())
		setupS3Routes(s3Group, c)
	} else {
		setupS3Routes(r, c)
	}
}

// setupS3Routes configures S3-compatible routes
func setupS3Routes(r gin.IRoutes, c *container.Container) {
	s3Handler := handlers.NewS3Handler(c)

	// Bucket operations
	r.GET("/", s3Handler.ListBuckets)
	r.PUT("/:bucket", s3Handler.CreateBucket)
	r.DELETE("/:bucket", s3Handler.DeleteBucket)
	r.GET("/:bucket", s3Handler.ListObjects)

	// Object operations
	r.PUT("/:bucket/*key", s3Handler.PutObject)
	r.GET("/:bucket/*key", s3Handler.GetObject)
	r.HEAD("/:bucket/*key", s3Handler.HeadObject)
	r.DELETE("/:bucket/*key", s3Handler.DeleteObject)
}
