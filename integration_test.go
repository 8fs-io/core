package eightfs_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/8fs-io/core/internal/config"
	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/internal/transport/http/middleware"
	"github.com/8fs-io/core/internal/transport/http/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationNewArchitecture(t *testing.T) {
	// Load config
	cfg, err := config.Load()
	assert.NoError(t, err)

	// Create container
	c, err := container.NewContainer(cfg)
	assert.NoError(t, err)

	// Create router manually like in cmd/server/main.go
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(c.Logger))
	r.Use(middleware.RequestID())

	if c.Config.Audit.Enabled {
		r.Use(middleware.Audit(c.AuditLogger))
	}

	if c.Config.Metrics.Enabled {
		r.Use(middleware.Metrics())
	}

	// Setup routes
	router.SetupRoutes(r, c)

	// Test health endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthz", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var health map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &health)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, "0.2.0", health["version"])

	// Test metrics endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "buckets_total")

	// Test S3 endpoints with authentication
	authHeader := "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=test"

	// Test list buckets (should be empty initially)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "<ListAllMyBucketsResult>")

	// Test create bucket
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/testbucket", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test put object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/testbucket/testfile.txt", bytes.NewBufferString("Hello New Architecture!"))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test get object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/testbucket/testfile.txt", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "Hello New Architecture!", w.Body.String())

	// Test head object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("HEAD", "/testbucket/testfile.txt", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))

	// Test list objects
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/testbucket", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "<Key>testfile.txt</Key>")

	// Test delete object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/testbucket/testfile.txt", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)

	// Test delete bucket
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/testbucket", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)
}
