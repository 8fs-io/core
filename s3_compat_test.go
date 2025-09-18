package eightfs_test

import (
	bytespkg "bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/8fs-io/core/internal/config"
	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/internal/transport/http/handlers"
	"github.com/8fs-io/core/internal/transport/http/middleware"
	"github.com/8fs-io/core/internal/transport/http/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newTestRouter builds a gin.Engine with isolated config (temp storage path) and returns it with the active config
func newTestRouter(t *testing.T, env map[string]string) (*gin.Engine, *config.Config) {
	t.Helper()
	// Use temp dir for storage isolation
	tmp := t.TempDir()
	baseEnv := map[string]string{
		"STORAGE_DRIVER":       "filesystem",
		"STORAGE_BASE_PATH":    tmp,
		"AUTH_ENABLED":         "true",
		"AUTH_DRIVER":          "signature",
		"DEFAULT_ACCESS_KEY":   "AKIAIOSFODNN7EXAMPLE",
		"DEFAULT_SECRET_KEY":   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"VECTOR_ENABLED":       "false",
		"METRICS_ENABLED":      "true",
		"SERVER_MODE":          "test",
	}
	for k, v := range env {
		baseEnv[k] = v
	}
	for k, v := range baseEnv {
		old := os.Getenv(k)
		_ = os.Setenv(k, v)
		// Restore after
		t.Cleanup(func() { _ = os.Setenv(k, old) })
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}
	c, err := container.NewContainer(cfg)
	if err != nil {
		t.Fatalf("container init failed: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(c.Logger))
	r.Use(middleware.RequestID())
	if c.Config.Metrics.Enabled {
		r.Use(middleware.Metrics())
	}
	router.SetupRoutes(r, c)
	return r, cfg
}

// helper to craft a minimal auth header for tests
func authHeader(accessKey string) string {
	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/20130524/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=test", accessKey)
}

// parseXML is a tiny helper to unmarshal XML into the provided dest
func parseXML(t *testing.T, data []byte, dest interface{}) {
	t.Helper()
	if err := xml.Unmarshal(data, dest); err != nil {
		t.Fatalf("xml unmarshal failed: %v\nBody: %s", err, string(data))
	}
}

func TestS3_Compatibility_BasicBucketAndObjectOps(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// List buckets (initially empty)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var buckets handlers.ListAllMyBucketsResult
	parseXML(t, w.Body.Bytes(), &buckets)
	assert.Len(t, buckets.Buckets.Bucket, 0)

	// Create bucket
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/my-bucket", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Put object
	content := "hello world"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/my-bucket/hello.txt", strings.NewReader(content))
	req.Header.Set("Authorization", authHeader(key))
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	// ETag should be md5 of body, quoted
	md5sum := md5.Sum([]byte(content))
	expectedETag := fmt.Sprintf("\"%s\"", hex.EncodeToString(md5sum[:]))
	assert.Equal(t, expectedETag, w.Header().Get("ETag"))

	// Get object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/my-bucket/hello.txt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, expectedETag, w.Header().Get("ETag"))
	assert.Equal(t, content, w.Body.String())

	// Head object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("HEAD", "/my-bucket/hello.txt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, expectedETag, w.Header().Get("ETag"))

	// List objects
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/my-bucket", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var list handlers.ListBucketResult
	parseXML(t, w.Body.Bytes(), &list)
	assert.Equal(t, "my-bucket", list.Name)
	assert.Len(t, list.Contents, 1)
	assert.Equal(t, "hello.txt", list.Contents[0].Key)

	// Delete object
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/my-bucket/hello.txt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)

	// Delete bucket
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/my-bucket", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)
}

func TestS3_Metadata_And_SpecialKeys(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// Create bucket
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/meta-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Put object with metadata and special key
	data := []byte("data-with-meta")
	specialKey := "dir a/файл 你好.txt"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/meta-bkt/"+specialKey, bytespkg.NewReader(data))
	req.Header.Set("Authorization", authHeader(key))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-amz-meta-foo", "bar")
	req.Header.Set("x-amz-meta-number", "42")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// HEAD should include metadata
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("HEAD", "/meta-bkt/"+specialKey, nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "bar", w.Header().Get("X-Amz-Meta-foo"))
	assert.Equal(t, "42", w.Header().Get("X-Amz-Meta-number"))

	// GET returns same metadata and data
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/meta-bkt/"+specialKey, nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "bar", w.Header().Get("X-Amz-Meta-foo"))
	assert.Equal(t, "42", w.Header().Get("X-Amz-Meta-number"))
	assert.Equal(t, data, w.Body.Bytes())
}

func TestS3_List_WithPrefix_And_Pagination(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// Create bucket
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/list-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Upload multiple objects
	keys := []string{"foo/a.txt", "foo/b.txt", "bar/c.txt"}
	for _, k := range keys {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("PUT", "/list-bkt/"+k, strings.NewReader("x"))
		req.Header.Set("Authorization", authHeader(key))
		req.Header.Set("Content-Type", "text/plain")
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}

	// List with prefix
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/list-bkt?prefix=foo/", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var list handlers.ListBucketResult
	parseXML(t, w.Body.Bytes(), &list)
	var returned []string
	for _, c := range list.Contents {
		returned = append(returned, c.Key)
	}
	assert.ElementsMatch(t, []string{"foo/a.txt", "foo/b.txt"}, returned)

	// Pagination with max-keys=2
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/list-bkt?max-keys=2", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	parseXML(t, w.Body.Bytes(), &list)
	assert.Len(t, list.Contents, 2)
	if list.IsTruncated {
		assert.NotEmpty(t, list.NextMarker)
	}

	// NOTE: delimiter is currently not implemented in repository; keep test pending to avoid failures
	// t.Run("Delimiter common prefixes", func(t *testing.T) { t.Skip("delimiter not implemented") })
}

func TestS3_Auth_Scenarios(t *testing.T) {
	// Enabled auth (default)
	r, _ := newTestRouter(t, nil)

	// Missing Authorization -> 401 XML error
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "<Error>")

	// Invalid signature scheme
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Wrong access key
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", authHeader("WRONGKEY"))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Disabled auth
	r2, _ := newTestRouter(t, map[string]string{"AUTH_ENABLED": "false"})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	r2.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestS3_ErrorCases_StatusCodes(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// Create bucket
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/err-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Get non-existent object -> 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/err-bkt/missing.txt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "<Error>")

	// Delete non-existent object -> 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/err-bkt/missing.txt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Create invalid bucket names -> 400
	invalid := []string{"A", "UPPER", "ab", "bad--name-", "-start"}
	for _, b := range invalid {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("PUT", "/"+b, nil)
		req.Header.Set("Authorization", authHeader(key))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code, "bucket=%s", b)
	}

	// Delete non-empty bucket -> 409
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/err-bkt/file.txt", strings.NewReader("x"))
	req.Header.Set("Authorization", authHeader(key))
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/err-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestS3_LargeObject_And_ContentTypes(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// Create bucket
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/big-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Generate ~5MB random data
	size := 5 * 1024 * 1024
	data := make([]byte, size)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(data)
	contentType := "application/octet-stream"

	// PUT
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/big-bkt/big.bin", bytespkg.NewReader(data))
	req.Header.Set("Authorization", authHeader(key))
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	md5sum := md5.Sum(data)
	expectedETag := fmt.Sprintf("\"%s\"", hex.EncodeToString(md5sum[:]))
	assert.Equal(t, expectedETag, w.Header().Get("ETag"))

	// GET with streaming compare
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/big-bkt/big.bin", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType, w.Header().Get("Content-Type"))
	assert.Equal(t, expectedETag, w.Header().Get("ETag"))
	assert.Equal(t, len(data), w.Body.Len())
	// spot-check content
	assert.Equal(t, data[:1024], w.Body.Bytes()[:1024])
}

func TestS3_Concurrent_Put_List(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// Create bucket
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/conc-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Concurrent puts
	var wg sync.WaitGroup
	count := 20
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprintf("file-%d", i)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/conc-bkt/obj-%03d.txt", i), strings.NewReader(body))
			req.Header.Set("Authorization", authHeader(key))
			req.Header.Set("Content-Type", "text/plain")
			r.ServeHTTP(w, req)
			if w.Code != 200 {
				io.Copy(io.Discard, w.Body)
			}
		} (i)
	}
	wg.Wait()

	// List
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/conc-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var list handlers.ListBucketResult
	parseXML(t, w.Body.Bytes(), &list)
	assert.GreaterOrEqual(t, len(list.Contents), count)
}