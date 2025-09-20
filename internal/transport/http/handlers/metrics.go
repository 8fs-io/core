package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/8fs-io/core/internal/container"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Storage metrics
	bucketsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "buckets_total",
			Help: "Total number of buckets",
		},
	)

	objectsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "objects_total",
			Help: "Total number of objects per bucket",
		},
		[]string{"bucket"},
	)

	storageBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "storage_bytes_total",
			Help: "Total storage used in bytes per bucket",
		},
		[]string{"bucket"},
	)

	// S3 operation metrics
	s3OperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "s3_operations_total",
			Help: "Total number of S3 operations",
		},
		[]string{"operation", "bucket", "status"},
	)

	// Authentication metrics
	authRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_requests_total",
			Help: "Total number of authentication requests",
		},
		[]string{"access_key", "status"},
	)

	// Operation (insert/search) counter
	vectorOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vector_operations_total",
			Help: "Total number of vector operations",
		},
		[]string{"operation", "status"},
	)

	// Operation (insert/search) duration
	operationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "vector_operation_duration_seconds",
			Help:    "Duration of the vector operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	vectorEmbeddingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vector_embeddings_total",
			Help: "Total number of vector embeddings",
		},
		[]string{"dimension"},
	)

	vectorStorageBytesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vector_storage_bytes_total",
			Help: "Total storage used in bytes for all vectors",
		},
	)

	vectorDimensionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vector_dimensions_total",
			Help: "Sum of dimensions of all stored vectors.",
		},
	)

	vectorSearchDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "vector_search_duration_seconds",
			Help:    "Duration of the vector search in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	vectorSearchResultsCount = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "vector_search_results_count",
			Help:    "Total number of vector search results",
			Buckets: prometheus.LinearBuckets(0, 10, 10),
		},
	)

	vectorErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vector_errors_total",
			Help: "Total errors",
		},
		[]string{"error_type"},
	)
)

// MetricsHandler handles Prometheus metrics requests
type MetricsHandler struct {
	container *container.Container
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(c *container.Container) *MetricsHandler {
	return &MetricsHandler{container: c}
}

// Handle processes metrics requests
func (h *MetricsHandler) Handle(c *gin.Context) {
	// Update storage metrics before serving
	h.updateStorageMetrics()

	// Serve Prometheus metrics
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// Tracking the vector operation
func trackOperation(start time.Time, operation, status, errorType string) {
	vectorOperationsTotal.WithLabelValues(operation, status).Inc()
	operationDuration.WithLabelValues(operation).Observe(time.Since(start).Seconds())

	if status == STATUS_ERROR {
		vectorErrorsTotal.WithLabelValues(errorType).Inc()
	}
}

// Tracking storage
func trackStorage(dimension int) {
	dimensionStr := strconv.Itoa(dimension)

	vectorEmbeddingsTotal.WithLabelValues(dimensionStr).Inc()
	vectorStorageBytesTotal.Add(float64(dimension * 8))
	vectorDimensionsTotal.Add(float64(dimension))
}

// Tracking the vector search operation performance
func trackSearchPerformance(start time.Time, count int) {
	vectorSearchResultsCount.Observe(float64(count))
	vectorSearchDuration.Observe(time.Since(start).Seconds())
}

// UpdateStorageMetrics updates storage-related metrics
func (h *MetricsHandler) updateStorageMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	buckets, err := h.container.StorageService.ListBuckets(ctx)
	if err != nil {
		h.container.Logger.Error("Failed to list buckets for metrics", "error", err)
		return
	}

	bucketsTotal.Set(float64(len(buckets)))

	if len(buckets) == 0 {
		// Ensure metric family appears even with no buckets
		storageBytes.WithLabelValues("_none").Set(0)
		return
	}

	for _, bucket := range buckets {
		objectsTotal.WithLabelValues(bucket.Name).Set(float64(bucket.ObjectCount))
		storageBytes.WithLabelValues(bucket.Name).Set(float64(bucket.Size))
	}
}
