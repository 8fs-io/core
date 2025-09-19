package handlers

import (
	"net/http"

	"github.com/8fs-io/core/internal/container"
	"github.com/gin-gonic/gin"
)

// IndexingHandler handles indexing-related HTTP requests
type IndexingHandler struct {
	container *container.Container
}

// NewIndexingHandler creates a new indexing handler
func NewIndexingHandler(c *container.Container) *IndexingHandler {
	return &IndexingHandler{
		container: c,
	}
}

// GetJobStatus returns the status of a specific indexing job
func (h *IndexingHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job ID is required",
		})
		return
	}

	if h.container.IndexingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "indexing service not available",
		})
		return
	}

	job, err := h.container.IndexingService.GetJobStatus(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "job not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job": job,
	})
}

// GetJobsByObject returns all indexing jobs for a specific object
func (h *IndexingHandler) GetJobsByObject(c *gin.Context) {
	objectID := c.Query("object_id")
	if objectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "object_id query parameter is required",
		})
		return
	}

	if h.container.IndexingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "indexing service not available",
		})
		return
	}

	jobs, err := h.container.IndexingService.GetJobsByObjectID(c.Request.Context(), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to retrieve jobs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object_id": objectID,
		"jobs":      jobs,
		"count":     len(jobs),
	})
}

// GetIndexingStats returns indexing service statistics
func (h *IndexingHandler) GetIndexingStats(c *gin.Context) {
	if h.container.IndexingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "indexing service not available",
		})
		return
	}

	stats := h.container.IndexingService.Stats()

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// HealthCheck returns the health status of the indexing service
func (h *IndexingHandler) HealthCheck(c *gin.Context) {
	if h.container.IndexingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "indexing service not initialized",
		})
		return
	}

	stats := h.container.IndexingService.Stats()

	// Simple health check based on queue length
	status := "healthy"
	if stats.QueueLength > 500 { // If queue is more than half full
		status = "degraded"
	}
	if stats.QueueLength >= 1000 { // If queue is full
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         status,
		"queue_length":   stats.QueueLength,
		"workers_active": stats.WorkersActive,
		"total_jobs":     stats.TotalJobs,
		"completed_jobs": stats.CompletedJobs,
		"failed_jobs":    stats.FailedJobs,
		"last_processed": stats.LastProcessed,
	})
}
