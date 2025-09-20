package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/internal/domain/vectors"
	"github.com/gin-gonic/gin"
)

const (
	// Operation constants
	INSERT_OPERATION   = "insert"
	SEARCH_OPERATION   = "search"
	RETRIEVE_OPERATION = "retrieve"
	DELETE_OPERATION   = "delete"

	// Status constants
	STATUS_ERROR   = "error"
	STATUS_SUCCESS = "success"

	// Error types
	NO_ERROR                 = ""
	DIMENSION_MISMATCH_ERROR = "dimension_mismatch"
	INVALID_FORMAT_ERROR     = "invalid_format"
	NOT_FOUND_ERROR          = "not_found"
	STORAGE_ERROR            = "storage_error"
)

// VectorHandler handles vector-related HTTP requests
type VectorHandler struct {
	container *container.Container
	storage   *vectors.SQLiteVecStorage
}

// NewVectorHandler creates a new vector handler
func NewVectorHandler(c *container.Container, vectorStorage *vectors.SQLiteVecStorage) *VectorHandler {
	return &VectorHandler{
		container: c,
		storage:   vectorStorage,
	}
}

// StoreEmbeddingRequest represents the request payload for storing embeddings
type StoreEmbeddingRequest struct {
	ID        string                 `json:"id" binding:"required"`
	Embedding []float64              `json:"embedding" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SearchEmbeddingsRequest represents the request payload for searching embeddings
type SearchEmbeddingsRequest struct {
	Query []float64 `json:"query" binding:"required"`
	TopK  int       `json:"top_k,omitempty"`
}

// SearchTextRequest represents the request payload for text-based searching
type SearchTextRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k,omitempty"`
}

// StoreEmbedding handles POST /vectors/embeddings
func (h *VectorHandler) StoreEmbedding(c *gin.Context) {
	// Tracking the insert operation duration
	status := STATUS_SUCCESS
	errType := NO_ERROR
	start := time.Now()
	dimension := 0

	defer func() {
		trackOperation(start, INSERT_OPERATION, status, errType)
		trackStorage(dimension)
	}()

	if h == nil || h.storage == nil {
		status = STATUS_ERROR
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vector storage not initialized"})
		return
	}

	var req StoreEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status = STATUS_ERROR
		errType = INVALID_FORMAT_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Validate vector dimensions
	vm := vectors.NewVectorMath()
	if err := vm.ValidateDimensions(req.Embedding); err != nil {
		status = STATUS_ERROR
		errType = DIMENSION_MISMATCH_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid embedding dimensions",
			"details": err.Error(),
		})
		return
	}

	// Create vector object
	vector := &vectors.Vector{
		ID:        req.ID,
		Embedding: req.Embedding,
		Metadata:  req.Metadata,
	}

	// Validate the complete vector
	if err := vm.ValidateVector(vector); err != nil {
		status = STATUS_ERROR
		errType = INVALID_FORMAT_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid vector data",
			"details": err.Error(),
		})
		return
	}

	// Store the vector
	if err := h.storage.Store(vector); err != nil {
		status = STATUS_ERROR
		// Check for dimension mismatch errors (client errors)
		var dimErr *vectors.DimensionMismatchError
		if errors.As(err, &dimErr) {
			errType = DIMENSION_MISMATCH_ERROR
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Dimension mismatch",
				"details": dimErr.Error(),
			})
			return
		}

		// Other errors are internal server errors
		errType = STORAGE_ERROR
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to store vector",
			"details": err.Error(),
		})
		return
	}

	dimension = len(req.Embedding)

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Vector stored successfully",
		"id":         req.ID,
		"dimensions": dimension,
	})

}

// SearchEmbeddings handles POST /vectors/search
func (h *VectorHandler) SearchEmbeddings(c *gin.Context) {
	// Tracking the search operation duration
	status := STATUS_SUCCESS
	errType := NO_ERROR
	start := time.Now()
	total := 0

	defer func() {
		trackOperation(start, SEARCH_OPERATION, status, errType)
		trackSearchPerformance(start, total)
	}()

	var req SearchEmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status = STATUS_ERROR
		errType = INVALID_FORMAT_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Set default top_k if not provided
	if req.TopK <= 0 {
		req.TopK = 10
	}

	// Validate query dimensions
	vm := vectors.NewVectorMath()
	if err := vm.ValidateDimensions(req.Query); err != nil {
		status = STATUS_ERROR
		errType = DIMENSION_MISMATCH_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query dimensions",
			"details": err.Error(),
		})
		return
	}

	// Perform the search
	results, err := h.storage.Search(req.Query, req.TopK)
	if err != nil {
		status = STATUS_ERROR
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	// Update total for metrics tracking
	total = len(results)

	c.JSON(http.StatusOK, gin.H{
		"results":          results,
		"query_dimensions": len(req.Query),
		"top_k":            req.TopK,
		"count":            len(results),
	})
}

// SearchText handles POST /vectors/search/text - text-based semantic search
func (h *VectorHandler) SearchText(c *gin.Context) {
	// Tracking the search operation duration
	status := STATUS_SUCCESS
	start := time.Now()
	defer func() {
		trackOperation(start, SEARCH_OPERATION, status)
	}()

	var req SearchTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status = STATUS_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Set default top_k if not provided
	if req.TopK <= 0 {
		req.TopK = 10
	}

	// Check if AI service is available for embedding generation
	if h.container.AIService == nil {
		status = STATUS_ERROR
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "AI service not available",
			"message": "Text search requires AI service for embedding generation",
		})
		return
	}

	// Generate embedding from query text
	queryEmbedding, err := h.container.AIService.GenerateEmbedding(c.Request.Context(), req.Query)
	if err != nil {
		status = STATUS_ERROR
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate query embedding",
			"details": err.Error(),
		})
		return
	}

	// Perform the search with generated embedding
	results, err := h.storage.Search(queryEmbedding, req.TopK)
	if err != nil {
		status = STATUS_ERROR
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":                    results,
		"query_text":                 req.Query,
		"query_embedding_dimensions": len(queryEmbedding),
		"top_k":                      req.TopK,
		"count":                      len(results),
	})
}

// GetEmbedding handles GET /vectors/embeddings/:id
func (h *VectorHandler) GetEmbedding(c *gin.Context) {
	// Tracking the retrieve operation duration
	status := STATUS_SUCCESS
	errType := NO_ERROR
	start := time.Now()

	defer func() {
		trackOperation(start, RETRIEVE_OPERATION, status, errType)
	}()

	id := c.Param("id")
	if id == "" {
		status = STATUS_ERROR
		errType = INVALID_FORMAT_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing vector ID",
		})
		return
	}

	// For now, we'll implement this as a search-by-ID approach
	// In a full implementation, we'd have a direct lookup method
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Direct vector lookup not yet implemented",
		"message": "Use search endpoint instead",
	})
}

// ListEmbeddings handles GET /vectors/embeddings
func (h *VectorHandler) ListEmbeddings(c *gin.Context) {
	// Tracking the retrieve operation duration
	status := STATUS_SUCCESS
	errType := NO_ERROR
	start := time.Now()

	defer func() {
		trackOperation(start, RETRIEVE_OPERATION, status, errType)
	}()

	// Parse query parameters
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if limit > 100 {
		limit = 100 // Cap at 100 for performance
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Vector listing not yet implemented",
		"message": "Use search endpoint instead",
		"limit":   limit,
	})
}

// DeleteEmbedding handles DELETE /vectors/embeddings/:id
func (h *VectorHandler) DeleteEmbedding(c *gin.Context) {
	// Tracking the delete operation duration
	status := STATUS_SUCCESS
	errType := NO_ERROR
	start := time.Now()

	defer func() {
		trackOperation(start, DELETE_OPERATION, status, errType)
	}()

	id := c.Param("id")
	if id == "" {
		status = STATUS_ERROR
		errType = INVALID_FORMAT_ERROR
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing vector ID",
		})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Vector deletion not yet implemented",
		"message": "Will be implemented in future version",
		"id":      id,
	})
}
