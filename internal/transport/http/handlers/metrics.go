package handlers

import (
	"context"
	"time"

	"github.com/8fs/8fs/internal/container"
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

// updateStorageMetrics updates storage-related metrics
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
