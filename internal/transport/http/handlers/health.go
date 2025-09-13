package handlers

import (
	"net/http"
	"time"

	"github.com/8fs/8fs/internal/container"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	container *container.Container
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(c *container.Container) *HealthHandler {
	return &HealthHandler{container: c}
}

// Handle processes health check requests
func (h *HealthHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Perform health checks
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "0.2.0",
		"storage":   h.container.Config.Storage.Driver,
	}

	// Check storage health
	if err := h.container.StorageService.HealthCheck(ctx); err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	// Get storage stats
	if stats, err := h.container.StorageService.GetStorageStats(ctx); err == nil {
		health["stats"] = stats
	}

	c.JSON(http.StatusOK, health)
}
