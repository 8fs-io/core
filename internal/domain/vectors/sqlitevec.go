package vectors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

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
	Path            string
	Dimension       int  // requested / primary dimension (API-level)
	EnableExtension bool // attempt to load vec0 extension
}

// SQLiteVecStorage provides sqlite-vec backed vector storage
type SQLiteVecStorage struct {
	db       *sql.DB
	cfg      SQLiteVecConfig
	logger   Logger
	fallback bool // true if vec extension not loaded
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
	// Try to load sqlite-vec extension if enabled
	if s.cfg.EnableExtension {
		// Try different common paths for the sqlite-vec extension
		extensionPaths := []string{
			"vec0",                // System installed
			"./vec0",              // Local
			"/usr/lib/vec0",       // Common system path
			"/usr/local/lib/vec0", // Another common path
		}

		var extensionLoaded bool
		for _, path := range extensionPaths {
			if _, err := s.db.Exec(fmt.Sprintf("SELECT load_extension('%s')", path)); err == nil {
				s.logger.Info("sqlite-vec extension loaded successfully", "path", path)
				extensionLoaded = true
				break
			}
		}

		if !extensionLoaded {
			s.logger.Warn("sqlite-vec extension failed to load from all paths; using fallback",
				"paths", extensionPaths)
			s.fallback = true
		}
	} else {
		s.logger.Info("sqlite-vec extension disabled via config")
		s.fallback = true
	}

	// Create appropriate table based on extension availability
	if !s.fallback {
		// Try to create sqlite-vec virtual table
		// Note: sqlite-vec supports dynamic dimensions, but we'll use a sensible default
		createVecTable := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS embeddings USING vec0(
			id TEXT PRIMARY KEY,
			embedding FLOAT[%d],
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`, s.cfg.Dimension)

		if _, err := s.db.Exec(createVecTable); err != nil {
			s.logger.Warn("failed to create sqlite-vec table, falling back to standard SQLite", "error", err)
			s.fallback = true
		} else {
			s.logger.Info("sqlite-vec virtual table created successfully", "dimensions", s.cfg.Dimension)
		}
	}

	// Create fallback table if extension failed or disabled
	if s.fallback {
		createFallbackTable := `
		CREATE TABLE IF NOT EXISTS embeddings (
			id TEXT PRIMARY KEY,
			embedding BLOB,
			metadata TEXT,
			dimensions INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		if _, err := s.db.Exec(createFallbackTable); err != nil {
			return fmt.Errorf("failed to create fallback embeddings table: %w", err)
		}

		// Create indices for better performance
		indices := []string{
			"CREATE INDEX IF NOT EXISTS idx_embeddings_created_at ON embeddings(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_embeddings_dimensions ON embeddings(dimensions)",
		}

		for _, idx := range indices {
			if _, err := s.db.Exec(idx); err != nil {
				// Log but don't fail - some indices might not be applicable to current schema
				s.logger.Warn("failed to create index", "error", err, "index", idx)
			}
		}

		s.logger.Info("fallback SQLite table created successfully")
	}

	return nil
}

// Store saves a vector to the database
func (s *SQLiteVecStorage) Store(vector *Vector) error {
	// Validate the vector first
	vm := NewVectorMath()
	if err := vm.ValidateVector(vector); err != nil {
		return fmt.Errorf("invalid vector: %w", err)
	}

	// Convert embedding to suitable format
	embeddingData, err := serializeEmbedding(vector.Embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	// Serialize metadata as JSON
	metadataJSON, err := json.Marshal(vector.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	var query string
	var args []interface{}

	if !s.fallback {
		// Use sqlite-vec optimized storage
		query = `
		INSERT OR REPLACE INTO embeddings (id, embedding, metadata, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`
		args = []interface{}{vector.ID, embeddingData, string(metadataJSON)}
	} else {
		// Use fallback storage with dimension tracking
		query = `
		INSERT OR REPLACE INTO embeddings (id, embedding, metadata, dimensions, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`
		args = []interface{}{vector.ID, embeddingData, string(metadataJSON), len(vector.Embedding)}
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to store vector: %w", err)
	}

	s.logger.Info("vector stored successfully", "id", vector.ID, "dimensions", len(vector.Embedding), "fallback", s.fallback)
	return nil
}

// Search performs vector similarity search
func (s *SQLiteVecStorage) Search(query []float64, topK int) ([]SearchResult, error) {
	// Validate query vector
	vm := NewVectorMath()
	if err := vm.ValidateDimensions(query); err != nil {
		return nil, fmt.Errorf("invalid query vector: %w", err)
	}

	if topK <= 0 {
		topK = 10 // Default
	}

	// Try optimized sqlite-vec search first
	if !s.fallback {
		results, err := s.vectorSearch(query, topK)
		if err == nil {
			return results, nil
		}
		s.logger.Warn("sqlite-vec search failed, falling back to linear search", "error", err)
	}

	// Fallback to linear search
	return s.linearSearch(query, topK)
}

// vectorSearch uses sqlite-vec extension for optimized search
func (s *SQLiteVecStorage) vectorSearch(query []float64, topK int) ([]SearchResult, error) {
	// Serialize query vector for sqlite-vec
	queryData, err := serializeEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query vector: %w", err)
	}

	// sqlite-vec query syntax - this uses the vec0 extension's similarity search
	// The exact syntax depends on the sqlite-vec version, but typically uses MATCH
	sqlQuery := `
	SELECT id, embedding, metadata,
		   vec_distance_cosine(embedding, ?) as distance
	FROM embeddings
	ORDER BY distance ASC
	LIMIT ?`

	rows, err := s.db.Query(sqlQuery, queryData, topK)
	if err != nil {
		// If the vec_distance_cosine function doesn't exist, sqlite-vec isn't fully loaded
		return nil, fmt.Errorf("sqlite-vec query failed (extension may not be loaded): %w", err)
	}
	defer rows.Close()

	var results []SearchResult

	for rows.Next() {
		var id, metadataJSON string
		var embeddingData []byte
		var distance float64

		err := rows.Scan(&id, &embeddingData, &metadataJSON, &distance)
		if err != nil {
			s.logger.Warn("failed to scan sqlite-vec result", "error", err)
			continue
		}

		embedding, err := deserializeEmbedding(embeddingData)
		if err != nil {
			s.logger.Warn("failed to deserialize embedding", "id", id, "error", err)
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
			Embedding: embedding,
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

// linearSearch provides fallback linear search when vec extension is not available
func (s *SQLiteVecStorage) linearSearch(query []float64, topK int) ([]SearchResult, error) {
	queryDim := len(query)

	// Get all vectors, optionally filtering by dimension for better performance
	var rows *sql.Rows
	var err error

	if s.fallback {
		// In fallback mode, we can filter by dimension to skip incompatible vectors
		rows, err = s.db.Query("SELECT id, embedding, metadata, dimensions FROM embeddings WHERE dimensions = ?", queryDim)
	} else {
		// In non-fallback mode, try all vectors (dimensions tracked separately)
		rows, err = s.db.Query("SELECT id, embedding, metadata FROM embeddings")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}
	defer rows.Close()

	vm := NewVectorMath()
	var allResults []SearchResult

	for rows.Next() {
		var id, metadataJSON string
		var embeddingData []byte
		var storedDim *int

		// Scan with or without dimensions column
		if s.fallback {
			err = rows.Scan(&id, &embeddingData, &metadataJSON, &storedDim)
		} else {
			err = rows.Scan(&id, &embeddingData, &metadataJSON)
		}

		if err != nil {
			s.logger.Warn("failed to scan vector row", "id", id, "error", err)
			continue
		}

		embedding, err := deserializeEmbedding(embeddingData)
		if err != nil {
			s.logger.Warn("failed to deserialize embedding", "id", id, "error", err)
			continue
		}

		// Skip vectors with incompatible dimensions
		if len(embedding) != queryDim {
			// Note: vectors with different dimensions are automatically skipped
			continue
		}

		// Parse metadata
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			s.logger.Warn("corrupted metadata, using fallback", "id", id, "error", err)
			metadata = map[string]interface{}{"raw": metadataJSON}
		}

		// Calculate similarity
		similarity, err := vm.CosineSimilarity(query, embedding)
		if err != nil {
			s.logger.Warn("failed to calculate similarity", "id", id, "error", err)
			continue
		}

		vector := &Vector{
			ID:        id,
			Embedding: embedding,
			Metadata:  metadata,
		}

		allResults = append(allResults, SearchResult{
			Vector: vector,
			Score:  similarity,
		})
	}

	// Sort by similarity (highest first)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// Return top K results
	if topK > len(allResults) {
		topK = len(allResults)
	}

	results := allResults[:topK]
	s.logger.Info("linear search completed",
		"query_dims", queryDim,
		"total_candidates", len(allResults),
		"returned", len(results))

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

// serializeEmbedding converts float64 slice to bytes for storage
func serializeEmbedding(embedding []float64) ([]byte, error) {
	// For sqlite-vec compatibility, we'll use JSON for now
	// In production with actual sqlite-vec, this would be optimized binary format
	return json.Marshal(embedding)
}

// deserializeEmbedding converts bytes back to float64 slice
func deserializeEmbedding(data []byte) ([]float64, error) {
	var embedding []float64
	err := json.Unmarshal(data, &embedding)
	return embedding, err
}
