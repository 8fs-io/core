package vectors

import (
	"math/rand"
	"path/filepath"
	"testing"
)

func TestSQLiteVecStorage(t *testing.T) {
	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_vectors.db")

	// Initialize storage with config (extension disabled for test)
	cfg := SQLiteVecConfig{Path: dbPath, Dimension: EmbeddingDim, EnableExtension: false}
	storage, err := NewSQLiteVecStorage(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create SQLiteVecStorage: %v", err)
	}
	defer storage.Close()

	// Test data (minimal varied embeddings of required dimension)
	base := make([]float64, EmbeddingDim)
	for i := range base {
		base[i] = 0.001 * float64(i%10)
	}
	alt := make([]float64, EmbeddingDim)
	copy(alt, base)
	for i := 0; i < 10 && i < EmbeddingDim; i++ {
		alt[i] += 0.01
	}

	testVectors := []*Vector{
		{ID: "test1", Embedding: base, Metadata: map[string]interface{}{"description": "Test vector 1"}},
		{ID: "test2", Embedding: alt, Metadata: map[string]interface{}{"description": "Test vector 2"}},
	}

	// Test storing vectors
	for _, vector := range testVectors {
		err := storage.Store(vector)
		if err != nil {
			t.Errorf("Failed to store vector %s: %v", vector.ID, err)
		}
	}

	// Test search (use alt embedding to ensure similarity ordering)
	query := alt
	results, err := storage.Search(query, 2)
	if err != nil {
		t.Fatalf("Failed to search vectors: %v", err)
	}

	// Verify results stronger assertions
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Score < results[1].Score {
		t.Errorf("expected results sorted by descending score")
	}
	ids := []string{results[0].Vector.ID, results[1].Vector.ID}
	if !((ids[0] == "test2" && ids[1] == "test1") || (ids[0] == "test1" && ids[1] == "test2")) {
		t.Errorf("unexpected result IDs order=%v", ids)
	}
}

// Helper function for generating test data
func generateRandomVector(dimensions int) []float64 {
	vec := make([]float64, dimensions)
	for i := 0; i < dimensions; i++ {
		vec[i] = rand.Float64()*2.0 - 1.0 // Range [-1, 1]
	}
	return vec
}
