package vectors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Error types for better error handling
var (
	ErrDimensionMismatch    = errors.New("dimension mismatch")
	ErrInvalidVector        = errors.New("invalid vector")
	ErrExtensionUnavailable = errors.New("sqlite-vec extension unavailable")
)

// DimensionMismatchError provides detailed dimension mismatch information
type DimensionMismatchError struct {
	Expected int
	Actual   int
}

func (e *DimensionMismatchError) Error() string {
	return fmt.Sprintf("dimension mismatch: vector has %d dimensions, table configured for %d",
		e.Actual, e.Expected)
}

func (e *DimensionMismatchError) Is(target error) bool {
	return target == ErrDimensionMismatch
}

// Logger is a minimal interface to allow structured logging without
// importing a concrete logging package here. The real application logger
// should satisfy this.
type Logger interface {
	Info(msg string, kv ...interface{})
	Warn(msg string, kv ...interface{})
	Error(msg string, kv ...interface{})
}

// noopLogger is used when no logger is provided.
type noopLogger struct{}

func (n *noopLogger) Info(string, ...interface{})  {}
func (n *noopLogger) Warn(string, ...interface{})  {}
func (n *noopLogger) Error(string, ...interface{}) {}

// SQLiteVecConfig carries construction parameters.
type SQLiteVecConfig struct {
	Path      string
	Dimension int // requested / primary dimension (API-level)
}

// SQLiteVecStorage provides sqlite-vec backed vector storage
type SQLiteVecStorage struct {
	db     *sql.DB
	cfg    SQLiteVecConfig
	logger Logger
}

// NewSQLiteVecStorage creates a new sqlite-vec storage instance
func NewSQLiteVecStorage(cfg SQLiteVecConfig, logger Logger) (*SQLiteVecStorage, error) {
	if cfg.Path == "" {
		return nil, errors.New("sqlite vec storage path required")
	}
	if cfg.Dimension == 0 {
		cfg.Dimension = DefaultEmbeddingDim
	}
	if cfg.Dimension < MinEmbeddingDim || cfg.Dimension > MaxEmbeddingDim {
		return nil, fmt.Errorf("configured dimension %d invalid: expected between %d and %d", cfg.Dimension, MinEmbeddingDim, MaxEmbeddingDim)
	}
	if logger == nil {
		logger = &noopLogger{}
	}

	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteVecStorage{db: db, cfg: cfg, logger: logger}

	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary tables and loads sqlite-vec extension
func (s *SQLiteVecStorage) initSchema() error {
	// Register the sqlite-vec extension using Go bindings
	// Note: vec.Auto() doesn't return an error, registration is best-effort
	vec.Auto()

	// Test if sqlite-vec is available by trying to create a virtual table
	testTable := fmt.Sprintf(`
	CREATE VIRTUAL TABLE IF NOT EXISTS vec_test USING vec0(
		embedding FLOAT[%d]
	)`, s.cfg.Dimension)

	if _, err := s.db.Exec(testTable); err != nil {
		return fmt.Errorf("sqlite-vec extension not available: %w", err)
	}

	// Clean up test table
	s.db.Exec("DROP TABLE IF EXISTS vec_test")
	s.logger.Info("sqlite-vec extension available and working")

	// Create sqlite-vec virtual table
	createVecTable := fmt.Sprintf(`
	CREATE VIRTUAL TABLE IF NOT EXISTS embeddings USING vec0(
		id TEXT PRIMARY KEY,
		embedding FLOAT[%d],
		metadata TEXT
	)`, s.cfg.Dimension)

	if _, err := s.db.Exec(createVecTable); err != nil {
		return fmt.Errorf("failed to create sqlite-vec table: %w", err)
	}

	s.logger.Info("sqlite-vec virtual table created successfully", "dimensions", s.cfg.Dimension)
	return nil
}

// Store saves a vector to the database
func (s *SQLiteVecStorage) Store(vector *Vector) error {
	// Validate the vector first
	vm := NewVectorMath()
	if err := vm.ValidateVector(vector); err != nil {
		return fmt.Errorf("invalid vector: %w", err)
	}

	// Check dimension matches the table configuration specifically
	if len(vector.Embedding) != s.cfg.Dimension {
		return &DimensionMismatchError{
			Expected: s.cfg.Dimension,
			Actual:   len(vector.Embedding),
		}
	}

	// Use sqlite-vec binary format
	embeddingData, err := serializeEmbeddingBinary(vector.Embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	// Serialize metadata as JSON
	metadataJSON, err := json.Marshal(vector.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Use sqlite-vec optimized storage
	query := `
	INSERT OR REPLACE INTO embeddings (id, embedding, metadata)
	VALUES (?, ?, ?)`

	_, err = s.db.Exec(query, vector.ID, embeddingData, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to store vector: %w", err)
	}

	s.logger.Info("vector stored successfully", "id", vector.ID, "dimensions", len(vector.Embedding))
	return nil
}

// Search performs vector similarity search using sqlite-vec
func (s *SQLiteVecStorage) Search(query []float64, topK int) ([]SearchResult, error) {
	// Validate query vector
	vm := NewVectorMath()
	if err := vm.ValidateDimensions(query); err != nil {
		return nil, fmt.Errorf("invalid query vector: %w", err)
	}

	if topK <= 0 {
		topK = 10 // Default
	}

	return s.vectorSearch(query, topK)
}

// vectorSearch uses sqlite-vec extension for optimized search
func (s *SQLiteVecStorage) vectorSearch(query []float64, topK int) ([]SearchResult, error) {
	// Serialize query vector for sqlite-vec using binary format
	queryData, err := serializeEmbeddingBinary(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query vector: %w", err)
	}

	// sqlite-vec query syntax - we don't need to retrieve the embedding data
	sqlQuery := `
	SELECT id, metadata,
		   vec_distance_cosine(embedding, ?) as distance
	FROM embeddings
	ORDER BY distance ASC
	LIMIT ?`

	rows, err := s.db.Query(sqlQuery, queryData, topK)
	if err != nil {
		return nil, fmt.Errorf("sqlite-vec query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult

	for rows.Next() {
		var id, metadataJSON string
		var distance float64

		err := rows.Scan(&id, &metadataJSON, &distance)
		if err != nil {
			s.logger.Warn("failed to scan sqlite-vec result", "error", err)
			continue
		}

		// Parse metadata
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			s.logger.Warn("failed to parse metadata", "id", id, "error", err)
			metadata = map[string]interface{}{"raw": metadataJSON}
		}

		// Convert distance to similarity score (1 - cosine_distance)
		similarity := 1.0 - distance
		if similarity < 0 {
			similarity = 0
		}

		vector := &Vector{
			ID:        id,
			Embedding: nil, // Not fetched in sqlite-vec mode for efficiency
			Metadata:  metadata,
		}

		results = append(results, SearchResult{
			Vector: vector,
			Score:  similarity,
		})
	}

	s.logger.Info("sqlite-vec search completed", "query_dims", len(query), "results", len(results), "top_k", topK)
	return results, nil
}

// Close closes the database connection
func (s *SQLiteVecStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Helper functions for embedding serialization

// serializeEmbeddingBinary converts float64 slice to binary format for sqlite-vec
func serializeEmbeddingBinary(embedding []float64) ([]byte, error) {
	// Convert float64 to float32 for sqlite-vec
	float32Vec := make([]float32, len(embedding))
	for i, v := range embedding {
		float32Vec[i] = float32(v)
	}

	// Use sqlite-vec-go-bindings serialization
	data, err := vec.SerializeFloat32(float32Vec)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize embedding to binary: %w", err)
	}
	return data, nil
}
