package vectors_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/8fs/8fs/internal/container"
	"github.com/8fs/8fs/internal/domain/vectors"
	"github.com/8fs/8fs/internal/transport/http/handlers"
	"github.com/gin-gonic/gin"
)

// TestVectorHandlerIntegration tests the complete HTTP flow for vector operations
func TestVectorHandlerIntegration(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_integration.db")

	// Initialize vector storage (extension disabled for integration test)
	cfg := vectors.SQLiteVecConfig{Path: dbPath, Dimension: vectors.EmbeddingDim, EnableExtension: false}
	storage, err := vectors.NewSQLiteVecStorage(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create vector storage: %v", err)
	}
	defer storage.Close()

	// Create a minimal container (mock)
	container := &container.Container{}

	// Create handler
	handler := handlers.NewVectorHandler(container, storage)

	// Setup gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	v1 := router.Group("/api/v1/vectors")
	{
		v1.POST("/embeddings", handler.StoreEmbedding)
		v1.POST("/search", handler.SearchEmbeddings)
		v1.GET("/embeddings/:id", handler.GetEmbedding)
	}

	t.Run("Store and Search Vector", func(t *testing.T) {
		// Test data
		// Build a valid 384-dim embedding (sparse pattern for brevity)
		emb := make([]float64, vectors.EmbeddingDim)
		for i := 0; i < vectors.EmbeddingDim; i++ {
			if i < 4 {
				emb[i] = 0.1 * float64(i+1)
			}
		}
		storeReq := map[string]interface{}{
			"id":        "test-doc-1",
			"embedding": emb,
			"metadata": map[string]interface{}{
				"document": "test-document.txt",
				"chunk":    1,
			},
		}

		// Store vector
		reqBody, _ := json.Marshal(storeReq)
		req := httptest.NewRequest("POST", "/api/v1/vectors/embeddings", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		// Verify response
		var storeResp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &storeResp); err != nil {
			t.Fatalf("Failed to unmarshal store response: %v", err)
		}

		if storeResp["id"] != "test-doc-1" {
			t.Errorf("Expected ID 'test-doc-1', got %v", storeResp["id"])
		}

		// Search for similar vectors
		searchReq := map[string]interface{}{
			"query": emb, // Same vector
			"top_k": 5,
		}

		searchBody, _ := json.Marshal(searchReq)
		searchReqHTTP := httptest.NewRequest("POST", "/api/v1/vectors/search", bytes.NewBuffer(searchBody))
		searchReqHTTP.Header.Set("Content-Type", "application/json")

		searchW := httptest.NewRecorder()
		router.ServeHTTP(searchW, searchReqHTTP)

		if searchW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", searchW.Code, searchW.Body.String())
		}

		// Verify search response
		var searchResp map[string]interface{}
		if err := json.Unmarshal(searchW.Body.Bytes(), &searchResp); err != nil {
			t.Fatalf("Failed to unmarshal search response: %v", err)
		}

		results, ok := searchResp["results"].([]interface{})
		if !ok || len(results) == 0 {
			t.Errorf("Expected search results, got: %v", searchResp["results"])
		}

		// Verify the result contains our stored vector
		result := results[0].(map[string]interface{})
		vector := result["vector"].(map[string]interface{})

		if vector["id"] != "test-doc-1" {
			t.Errorf("Expected vector ID 'test-doc-1', got %v", vector["id"])
		}

		// Score should be very high (close to 1.0) for identical vectors
		score := result["score"].(float64)
		if score < 0.99 {
			t.Errorf("Expected high similarity score, got %.4f", score)
		}

		t.Logf("Search successful: Found vector %s with score %.4f", vector["id"], score)
	})

	t.Run("Invalid Requests", func(t *testing.T) {
		// Test invalid embedding dimensions
		invalidReq := map[string]interface{}{
			"id":        "invalid-test",
			"embedding": []float64{0.1, 0.2}, // Too few dimensions
		}

		reqBody, _ := json.Marshal(invalidReq)
		req := httptest.NewRequest("POST", "/api/v1/vectors/embeddings", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid dimensions, got %d", w.Code)
		}

		// Test empty request
		emptyReq := httptest.NewRequest("POST", "/api/v1/vectors/embeddings", bytes.NewBuffer([]byte("{}")))
		emptyReq.Header.Set("Content-Type", "application/json")

		emptyW := httptest.NewRecorder()
		router.ServeHTTP(emptyW, emptyReq)

		if emptyW.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty request, got %d", emptyW.Code)
		}
	})

	t.Run("Multiple Vectors", func(t *testing.T) {
		// Store multiple vectors
		// Create three orthogonal-like sparse vectors in 384-dim space
		mk := func(id string, idx int) map[string]interface{} {
			emb := make([]float64, vectors.EmbeddingDim)
			emb[idx] = 1.0
			return map[string]interface{}{"id": id, "embedding": emb, "metadata": map[string]interface{}{"type": id}}
		}
		docs := []map[string]interface{}{mk("doc-1", 0), mk("doc-2", 1), mk("doc-3", 2)}

		// Store each vector
		for _, vector := range docs {
			reqBody, _ := json.Marshal(vector)
			req := httptest.NewRequest("POST", "/api/v1/vectors/embeddings", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("Failed to store vector %s: status %d", vector["id"], w.Code)
			}
		}

		// Search with a query similar to doc-1
		// Query close to doc-1
		qq := make([]float64, vectors.EmbeddingDim)
		qq[0] = 0.9
		qq[1] = 0.1
		searchReq := map[string]interface{}{
			"query": qq,
			"top_k": 2,
		}

		searchBody, _ := json.Marshal(searchReq)
		searchReqHTTP := httptest.NewRequest("POST", "/api/v1/vectors/search", bytes.NewBuffer(searchBody))
		searchReqHTTP.Header.Set("Content-Type", "application/json")

		searchW := httptest.NewRecorder()
		router.ServeHTTP(searchW, searchReqHTTP)

		if searchW.Code != http.StatusOK {
			t.Errorf("Search failed: status %d", searchW.Code)
			return
		}

		// Parse results
		var searchResp map[string]interface{}
		json.Unmarshal(searchW.Body.Bytes(), &searchResp)

		results := searchResp["results"].([]interface{})
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		// First result should be doc-1 (most similar)
		firstResult := results[0].(map[string]interface{})
		firstVector := firstResult["vector"].(map[string]interface{})

		if firstVector["id"] != "doc-1" {
			t.Errorf("Expected first result to be doc-1, got %v", firstVector["id"])
		}

		t.Logf("Multi-vector search successful: Top result is %s with score %.4f",
			firstVector["id"], firstResult["score"].(float64))
	})
}

// TestVectorHandlerErrors tests error handling
func TestVectorHandlerErrors(t *testing.T) {
	// Test with invalid database path to trigger initialization error
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// This should fail gracefully
	v1 := router.Group("/api/v1/vectors")
	{
		// This would normally fail in the router setup
		// but we'll test the handler directly with nil storage
		handler := &handlers.VectorHandler{}
		v1.POST("/embeddings", handler.StoreEmbedding)
	}

	// Test request to handler with no storage
	req := httptest.NewRequest("POST", "/api/v1/vectors/embeddings",
		bytes.NewBuffer([]byte(`{"id":"test","embedding":[0.1,0.2,0.3,0.4]}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 due to nil storage
	if w.Code == http.StatusOK {
		t.Error("Expected error with nil storage, but got success")
	}
}
