package handlers

import (
	"net/http"
	"time"

	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/internal/domain/rag"
	"github.com/gin-gonic/gin"
)

// RAGHandler handles RAG-related HTTP requests
type RAGHandler struct {
	container  *container.Container
	ragService rag.Service
}

// NewRAGHandler creates a new RAG handler
func NewRAGHandler(c *container.Container, ragService rag.Service) *RAGHandler {
	return &RAGHandler{
		container:  c,
		ragService: ragService,
	}
}

// ChatCompletions handles POST /chat/completions - OpenAI compatible endpoint
func (h *RAGHandler) ChatCompletions(c *gin.Context) {
	var req rag.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"message": "Invalid request format",
				"type":    "invalid_request_error",
				"param":   nil,
				"code":    "bad_request",
			},
		})
		return
	}

	// Set defaults if not provided
	if req.MaxTokens <= 0 {
		req.MaxTokens = 4000
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}
	if req.TopK <= 0 {
		req.TopK = 5
	}

	// Process the chat request
	response, err := h.ragService.Chat(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"message": "Failed to process chat request",
				"type":    "internal_error",
				"param":   nil,
				"code":    "internal_error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchContext handles POST /search/context - retrieve context for a query
func (h *RAGHandler) SearchContext(c *gin.Context) {
	var req struct {
		Query    string            `json:"query" binding:"required"`
		TopK     int               `json:"top_k,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 10 // Default to 10 context documents
	}

	response, err := h.ragService.SearchContext(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search context",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateWithContext handles POST /generate/context - generate text with provided context
func (h *RAGHandler) GenerateWithContext(c *gin.Context) {
	var req rag.GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if req.MaxTokens <= 0 {
		req.MaxTokens = 4000
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}

	response, err := h.ragService.GenerateWithContext(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate text",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealth handles GET /health - RAG service health check
func (h *RAGHandler) GetHealth(c *gin.Context) {
	// Simple health check - could be expanded to check AI service connectivity
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "rag",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}
