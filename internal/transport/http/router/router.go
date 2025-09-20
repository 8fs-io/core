package router

import (
	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/internal/transport/http/handlers"
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
		if c.Config.Vector.Enabled && c.VectorStorage != nil {
			vectorGroup := v1.Group("/vectors")
			vectorHandler := handlers.NewVectorHandler(c, c.VectorStorage)
			vectorGroup.POST("/embeddings", vectorHandler.StoreEmbedding)
			vectorGroup.POST("/search", vectorHandler.SearchEmbeddings)
			vectorGroup.POST("/search/text", vectorHandler.SearchText) // Text-based semantic search
			vectorGroup.GET("/embeddings/:id", vectorHandler.GetEmbedding)
			vectorGroup.GET("/embeddings", vectorHandler.ListEmbeddings)
			vectorGroup.DELETE("/embeddings/:id", vectorHandler.DeleteEmbedding)
		}

		// Async indexing endpoints
		if c.Config.Indexing.Enabled && c.Config.Indexing.StatusEnabled && c.IndexingService != nil {
			indexingGroup := v1.Group("/indexing")
			indexingHandler := handlers.NewIndexingHandler(c)
			indexingGroup.GET("/jobs/:jobId", indexingHandler.GetJobStatus)
			indexingGroup.GET("/jobs", indexingHandler.GetJobsByObject)
			indexingGroup.GET("/stats", indexingHandler.GetIndexingStats)
			indexingGroup.GET("/health", indexingHandler.HealthCheck)
		}

		// RAG endpoints
		if c.RAGService != nil {
			chatGroup := v1.Group("/chat")
			ragHandler := handlers.NewRAGHandler(c, c.RAGService)
			chatGroup.POST("/completions", ragHandler.ChatCompletions)          // OpenAI compatible
			chatGroup.POST("/search/context", ragHandler.SearchContext)         // Context retrieval
			chatGroup.POST("/generate/context", ragHandler.GenerateWithContext) // Direct generation with context
			chatGroup.GET("/health", ragHandler.GetHealth)                      // RAG health check
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
	r.POST("/:bucket", s3Handler.DeleteObjects) // S3 delete objects API

	// Object operations
	r.PUT("/:bucket/*key", s3Handler.PutObject)
	r.GET("/:bucket/*key", s3Handler.GetObject)
	r.HEAD("/:bucket/*key", s3Handler.HeadObject)
	r.DELETE("/:bucket/*key", s3Handler.DeleteObject)
}
