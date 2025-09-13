package handlers

import (
	"net/http"
	"strconv"

	"github.com/8fs/8fs/internal/container"
	"github.com/8fs/8fs/internal/domain/storage"
	"github.com/8fs/8fs/pkg/errors"
	"github.com/gin-gonic/gin"
)

// StorageHandler handles RESTful storage API requests
type StorageHandler struct {
	container *container.Container
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(c *container.Container) *StorageHandler {
	return &StorageHandler{container: c}
}

// ListBuckets lists all buckets
func (h *StorageHandler) ListBuckets(c *gin.Context) {
	ctx := c.Request.Context()

	buckets, err := h.container.StorageService.ListBuckets(ctx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"buckets": buckets,
		"count":   len(buckets),
	})
}

// CreateBucket creates a new bucket
func (h *StorageHandler) CreateBucket(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	var req struct {
		Metadata map[string]string `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	bucket, err := h.container.StorageService.CreateBucket(ctx, bucketName, req.Metadata)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, bucket)
}

// DeleteBucket deletes a bucket
func (h *StorageHandler) DeleteBucket(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	err := h.container.StorageService.DeleteBucket(ctx, bucketName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetBucket retrieves bucket information
func (h *StorageHandler) GetBucket(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	bucket, err := h.container.StorageService.GetBucket(ctx, bucketName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, bucket)
}

// PutObject stores an object
func (h *StorageHandler) PutObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")[1:] // Remove leading slash

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Extract metadata from headers (x-amz-meta-*)
	metadata := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(key) > 11 && key[:11] == "X-Amz-Meta-" {
			if len(values) > 0 {
				metadata[key[11:]] = values[0]
			}
		}
	}

	object, err := h.container.StorageService.PutObject(ctx, bucketName, objectKey, data, contentType, metadata)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Header("ETag", object.ETag)
	c.JSON(http.StatusOK, gin.H{
		"key":           object.Key,
		"size":          object.Size,
		"etag":          object.ETag,
		"last_modified": object.LastModified,
	})
}

// GetObject retrieves an object
func (h *StorageHandler) GetObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")[1:] // Remove leading slash

	object, err := h.container.StorageService.GetObject(ctx, bucketName, objectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Header("Content-Type", object.ContentType)
	c.Header("Content-Length", strconv.FormatInt(object.Size, 10))
	c.Header("ETag", object.ETag)
	c.Header("Last-Modified", object.LastModified.Format(http.TimeFormat))

	// Set metadata headers
	for key, value := range object.Metadata {
		c.Header("X-Amz-Meta-"+key, value)
	}

	c.Data(http.StatusOK, object.ContentType, object.Data)
}

// HeadObject retrieves object metadata only
func (h *StorageHandler) HeadObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")[1:] // Remove leading slash

	objectInfo, err := h.container.StorageService.GetObjectInfo(ctx, bucketName, objectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Header("Content-Type", objectInfo.ContentType)
	c.Header("Content-Length", strconv.FormatInt(objectInfo.Size, 10))
	c.Header("ETag", objectInfo.ETag)
	c.Header("Last-Modified", objectInfo.LastModified.Format(http.TimeFormat))

	// Set metadata headers
	for key, value := range objectInfo.Metadata {
		c.Header("X-Amz-Meta-"+key, value)
	}

	c.Status(http.StatusOK)
}

// DeleteObject deletes an object
func (h *StorageHandler) DeleteObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")[1:] // Remove leading slash

	err := h.container.StorageService.DeleteObject(ctx, bucketName, objectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListObjects lists objects in a bucket
func (h *StorageHandler) ListObjects(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	// Parse query parameters
	opts := storage.ListOptions{
		Prefix:    c.Query("prefix"),
		Delimiter: c.Query("delimiter"),
		Marker:    c.Query("marker"),
	}

	if maxKeys := c.Query("max-keys"); maxKeys != "" {
		if keys, err := strconv.Atoi(maxKeys); err == nil {
			opts.MaxKeys = keys
		}
	}

	result, err := h.container.StorageService.ListObjects(ctx, bucketName, opts)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleError converts domain errors to appropriate HTTP responses
func (h *StorageHandler) handleError(c *gin.Context, err error) {
	var appErr *errors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"context": appErr.Context,
			},
		})
		return
	}

	// Fallback for unexpected errors
	h.container.Logger.Error("Unexpected error in storage handler", "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}
