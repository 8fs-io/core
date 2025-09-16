package handlers

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/internal/domain/storage"
	"github.com/8fs-io/core/pkg/errors"
	"github.com/gin-gonic/gin"
)

// S3Handler handles S3-compatible API requests
type S3Handler struct {
	container *container.Container
}

// NewS3Handler creates a new S3 handler
func NewS3Handler(c *container.Container) *S3Handler {
	return &S3Handler{container: c}
}

// XML response structures for S3 compatibility
type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Owner   Owner    `xml:"Owner"`
	Buckets Buckets  `xml:"Buckets"`
}

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type Buckets struct {
	Bucket []BucketXML `xml:"Bucket"`
}

type BucketXML struct {
	Name         string    `xml:"Name"`
	CreationDate time.Time `xml:"CreationDate"`
}

type ListBucketResult struct {
	XMLName        xml.Name    `xml:"ListBucketResult"`
	Name           string      `xml:"Name"`
	Prefix         string      `xml:"Prefix,omitempty"`
	Marker         string      `xml:"Marker,omitempty"`
	MaxKeys        int         `xml:"MaxKeys"`
	IsTruncated    bool        `xml:"IsTruncated"`
	NextMarker     string      `xml:"NextMarker,omitempty"`
	Contents       []ObjectXML `xml:"Contents"`
	CommonPrefixes []PrefixXML `xml:"CommonPrefixes,omitempty"`
}

type ObjectXML struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         int64     `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
	Owner        Owner     `xml:"Owner"`
}

type PrefixXML struct {
	Prefix string `xml:"Prefix"`
}

type ErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource,omitempty"`
	RequestID string   `xml:"RequestId,omitempty"`
}

// ListBuckets handles S3 list buckets request
func (h *S3Handler) ListBuckets(c *gin.Context) {
	ctx := c.Request.Context()

	buckets, err := h.container.StorageService.ListBuckets(ctx)
	if err != nil {
		h.handleS3Error(c, err, c.Request.URL.Path)
		return
	}

	// Convert to S3 XML format
	var bucketXMLs []BucketXML
	for _, bucket := range buckets {
		bucketXMLs = append(bucketXMLs, BucketXML{
			Name:         bucket.Name,
			CreationDate: bucket.CreatedAt,
		})
	}

	response := ListAllMyBucketsResult{
		Owner: Owner{
			ID:          "8fs-owner",
			DisplayName: "8fs",
		},
		Buckets: Buckets{Bucket: bucketXMLs},
	}

	c.XML(http.StatusOK, response)
}

// CreateBucket handles S3 create bucket request
func (h *S3Handler) CreateBucket(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	// Extract metadata from headers
	metadata := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(key) > 11 && strings.ToLower(key[:11]) == "x-amz-meta-" {
			if len(values) > 0 {
				metadata[key[11:]] = values[0]
			}
		}
	}

	_, err := h.container.StorageService.CreateBucket(ctx, bucketName, metadata)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName)
		return
	}

	// Record S3 operation metric
	s3OperationsTotal.WithLabelValues("CreateBucket", bucketName, "success").Inc()

	c.Status(http.StatusOK)
}

// DeleteBucket handles S3 delete bucket request
func (h *S3Handler) DeleteBucket(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	err := h.container.StorageService.DeleteBucket(ctx, bucketName)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName)
		s3OperationsTotal.WithLabelValues("DeleteBucket", bucketName, "error").Inc()
		return
	}

	s3OperationsTotal.WithLabelValues("DeleteBucket", bucketName, "success").Inc()
	c.Status(http.StatusNoContent)
}

// ListObjects handles S3 list objects request
func (h *S3Handler) ListObjects(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")

	// Parse query parameters
	opts := storage.ListOptions{
		Prefix:    c.Query("prefix"),
		Delimiter: c.Query("delimiter"),
		Marker:    c.Query("marker"),
		MaxKeys:   1000, // Default
	}

	if maxKeys := c.Query("max-keys"); maxKeys != "" {
		if keys, err := strconv.Atoi(maxKeys); err == nil && keys > 0 {
			opts.MaxKeys = keys
		}
	}

	result, err := h.container.StorageService.ListObjects(ctx, bucketName, opts)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName)
		return
	}

	// Convert to S3 XML format
	var objectXMLs []ObjectXML
	for _, obj := range result.Objects {
		objectXMLs = append(objectXMLs, ObjectXML{
			Key:          obj.Key,
			LastModified: obj.LastModified,
			ETag:         obj.ETag,
			Size:         obj.Size,
			StorageClass: "STANDARD",
			Owner: Owner{
				ID:          "8fs-owner",
				DisplayName: "8fs",
			},
		})
	}

	var commonPrefixes []PrefixXML
	for _, prefix := range result.CommonPrefixes {
		commonPrefixes = append(commonPrefixes, PrefixXML{Prefix: prefix})
	}

	response := ListBucketResult{
		Name:           bucketName,
		Prefix:         opts.Prefix,
		Marker:         opts.Marker,
		MaxKeys:        opts.MaxKeys,
		IsTruncated:    result.IsTruncated,
		NextMarker:     result.NextMarker,
		Contents:       objectXMLs,
		CommonPrefixes: commonPrefixes,
	}

	c.XML(http.StatusOK, response)
}

// PutObject handles S3 put object request
func (h *S3Handler) PutObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := strings.TrimPrefix(c.Param("key"), "/")

	data, err := c.GetRawData()
	if err != nil {
		h.handleS3Error(c, errors.Wrap(errors.ErrCodeInvalidRequest, "Failed to read request body", err), "/"+bucketName+"/"+objectKey)
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "binary/octet-stream"
	}

	// Extract metadata from headers
	metadata := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(key) > 11 && strings.ToLower(key[:11]) == "x-amz-meta-" {
			if len(values) > 0 {
				metadata[key[11:]] = values[0]
			}
		}
	}

	object, err := h.container.StorageService.PutObject(ctx, bucketName, objectKey, data, contentType, metadata)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName+"/"+objectKey)
		s3OperationsTotal.WithLabelValues("PutObject", bucketName, "error").Inc()
		return
	}

	s3OperationsTotal.WithLabelValues("PutObject", bucketName, "success").Inc()

	c.Header("ETag", object.ETag)
	c.Status(http.StatusOK)
}

// GetObject handles S3 get object request
func (h *S3Handler) GetObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := strings.TrimPrefix(c.Param("key"), "/")

	object, err := h.container.StorageService.GetObject(ctx, bucketName, objectKey)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName+"/"+objectKey)
		s3OperationsTotal.WithLabelValues("GetObject", bucketName, "error").Inc()
		return
	}

	s3OperationsTotal.WithLabelValues("GetObject", bucketName, "success").Inc()

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

// HeadObject handles S3 head object request
func (h *S3Handler) HeadObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := strings.TrimPrefix(c.Param("key"), "/")

	objectInfo, err := h.container.StorageService.GetObjectInfo(ctx, bucketName, objectKey)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName+"/"+objectKey)
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

// DeleteObject handles S3 delete object request
func (h *S3Handler) DeleteObject(c *gin.Context) {
	ctx := c.Request.Context()
	bucketName := c.Param("bucket")
	objectKey := strings.TrimPrefix(c.Param("key"), "/")

	err := h.container.StorageService.DeleteObject(ctx, bucketName, objectKey)
	if err != nil {
		h.handleS3Error(c, err, "/"+bucketName+"/"+objectKey)
		s3OperationsTotal.WithLabelValues("DeleteObject", bucketName, "error").Inc()
		return
	}

	s3OperationsTotal.WithLabelValues("DeleteObject", bucketName, "success").Inc()
	c.Status(http.StatusNoContent)
}

// handleS3Error converts domain errors to S3-compatible XML error responses
func (h *S3Handler) handleS3Error(c *gin.Context, err error, resource string) {
	var appErr *errors.AppError
	if !errors.As(err, &appErr) {
		// Fallback for unexpected errors
		h.container.Logger.Error("Unexpected error in S3 handler", "error", err)
		appErr = errors.ErrInternalError
	}

	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	errorResponse := ErrorResponse{
		Code:      string(appErr.Code),
		Message:   appErr.Message,
		Resource:  resource,
		RequestID: requestID,
	}

	c.XML(appErr.HTTPStatus, errorResponse)
}
