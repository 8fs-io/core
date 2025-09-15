package vectors

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteVecStorage provides sqlite-vec backed vector storage
type SQLiteVecStorage struct {
	db *sql.DB
}

// NewSQLiteVecStorage creates a new sqlite-vec storage instance
func NewSQLiteVecStorage(dbPath string) (*SQLiteVecStorage, error) {
	// Note: This requires sqlite-vec extension to be compiled with SQLite
	// For now, we'll create the connection and defer extension loading
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteVecStorage{db: db}

	// Initialize the schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary tables and loads sqlite-vec extension
func (s *SQLiteVecStorage) initSchema() error {
	// First, try to load the sqlite-vec extension
	// Note: This will fail without proper sqlite-vec installation
	_, err := s.db.Exec("SELECT load_extension('vec0')")
	if err != nil {
		// For development, we'll continue without the extension
		// but log the error
		fmt.Printf("Warning: sqlite-vec extension not available: %v\n", err)
		fmt.Println("Falling back to regular SQLite tables for development")
	}

	// Create embeddings table with sqlite-vec virtual table syntax
	// If extension is not available, this will fail but we'll create a fallback
	createVecTable := `
	CREATE VIRTUAL TABLE IF NOT EXISTS embeddings USING vec0(
		id TEXT PRIMARY KEY,
		embedding FLOAT[384],
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = s.db.Exec(createVecTable)
	if err != nil {
		// Fallback to regular table for development
		fmt.Println("Creating fallback table (without vec0 optimization)")
		createFallbackTable := `
		CREATE TABLE IF NOT EXISTS embeddings (
			id TEXT PRIMARY KEY,
			embedding BLOB,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		_, err = s.db.Exec(createFallbackTable)
		if err != nil {
			return fmt.Errorf("failed to create fallback embeddings table: %w", err)
		}
	}

	return nil
}

// Store saves a vector to the database
func (s *SQLiteVecStorage) Store(vector *Vector) error {
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

	query := `
	INSERT OR REPLACE INTO embeddings (id, embedding, metadata, created_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)`

	_, err = s.db.Exec(query, vector.ID, embeddingData, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to store vector: %w", err)
	}

	return nil
}

// Search performs vector similarity search
func (s *SQLiteVecStorage) Search(query []float64, topK int) ([]SearchResult, error) {
	// Try sqlite-vec optimized search first
	results, err := s.vectorSearch(query, topK)
	if err != nil {
		// Fall back to linear search if vec extension is not available
		fmt.Println("Vector search failed, falling back to linear search")
		return s.linearSearch(query, topK)
	}

	return results, nil
}

// vectorSearch uses sqlite-vec extension for optimized search
func (s *SQLiteVecStorage) vectorSearch(query []float64, topK int) ([]SearchResult, error) {
	// This requires sqlite-vec extension to be loaded
	queryData, err := serializeEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	sqlQuery := `
	SELECT id, embedding, metadata, distance 
	FROM embeddings 
	WHERE embedding MATCH ? 
	ORDER BY distance 
	LIMIT ?`

	rows, err := s.db.Query(sqlQuery, queryData, topK)
	if err != nil {
		return nil, fmt.Errorf("vector search query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult

	for rows.Next() {
		var id, metadataJSON string
		var embeddingData []byte
		var distance float64

		err := rows.Scan(&id, &embeddingData, &metadataJSON, &distance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		embedding, err := deserializeEmbedding(embeddingData)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
		}

		// Deserialize metadata from JSON
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
		}

		vector := &Vector{
			ID:        id,
			Embedding: embedding,
			Metadata:  metadata,
		}

		// Convert distance to similarity score (assuming cosine distance)
		similarity := 1.0 - distance

		results = append(results, SearchResult{
			Vector: vector,
			Score:  similarity,
		})
	}

	return results, nil
}

// linearSearch provides fallback linear search when vec extension is not available
func (s *SQLiteVecStorage) linearSearch(query []float64, topK int) ([]SearchResult, error) {
	// Get all vectors
	rows, err := s.db.Query("SELECT id, embedding, metadata FROM embeddings")
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}
	defer rows.Close()

	vm := NewVectorMath()
	var allResults []SearchResult

	for rows.Next() {
		var id, metadataJSON string
		var embeddingData []byte

		err := rows.Scan(&id, &embeddingData, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vector: %w", err)
		}

		embedding, err := deserializeEmbedding(embeddingData)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
		}

		// Deserialize metadata from JSON
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			// Handle legacy metadata or corrupted data gracefully
			metadata = map[string]interface{}{"raw": metadataJSON}
		}

		similarity, err := vm.CosineSimilarity(query, embedding)
		if err != nil {
			continue // Skip invalid vectors
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
	for i := 0; i < len(allResults)-1; i++ {
		for j := i + 1; j < len(allResults); j++ {
			if allResults[i].Score < allResults[j].Score {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}

	// Return top K results
	if topK > len(allResults) {
		topK = len(allResults)
	}

	return allResults[:topK], nil
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
