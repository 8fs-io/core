package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/8fs-io/core/internal/domain/vectors"
)

const (
	defaultServerURL = "http://localhost:8080"
	defaultLlamaAPI  = "http://localhost:11434" // Ollama default
)

// LlamaEmbeddingRequest represents a request to Llama for embeddings
type LlamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// LlamaEmbeddingResponse represents a response from Llama API
type LlamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// DocumentChunk represents a processed document chunk
type DocumentChunk struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"text"`
	Metadata map[string]interface{} `json:"metadata"`
}

// LlamaDemo handles the integration demo
type LlamaDemo struct {
	serverURL string
	llamaAPI  string
	model     string
	client    *http.Client
}

func main() {
	var (
		serverURL = flag.String("server", defaultServerURL, "8fs server URL")
		llamaAPI  = flag.String("llama", defaultLlamaAPI, "Llama API URL (Ollama)")
		model     = flag.String("model", "llama2", "Llama model name")
		command   = flag.String("cmd", "help", "Command: ingest, search, rag, help")
		input     = flag.String("input", "", "Input file or text")
		query     = flag.String("query", "", "Search query")
		topK      = flag.Int("topk", 5, "Number of search results")
		chunkSize = flag.Int("chunk", 500, "Chunk size for document processing")
		verbose   = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	demo := &LlamaDemo{
		serverURL: *serverURL,
		llamaAPI:  *llamaAPI,
		model:     *model,
		client:    &http.Client{Timeout: 30 * time.Second},
	}

	switch *command {
	case "ingest":
		if *input == "" {
			fmt.Println("Error: -input required for ingest command")
			os.Exit(1)
		}
		err := demo.ingestDocument(*input, *chunkSize, *verbose)
		if err != nil {
			fmt.Printf("Error ingesting document: %v\n", err)
			os.Exit(1)
		}
	case "search":
		if *query == "" {
			fmt.Println("Error: -query required for search command")
			os.Exit(1)
		}
		err := demo.searchVectors(*query, *topK, *verbose)
		if err != nil {
			fmt.Printf("Error searching vectors: %v\n", err)
			os.Exit(1)
		}
	case "rag":
		if *query == "" {
			fmt.Println("Error: -query required for RAG command")
			os.Exit(1)
		}
		err := demo.performRAG(*query, *topK, *verbose)
		if err != nil {
			fmt.Printf("Error performing RAG: %v\n", err)
			os.Exit(1)
		}
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		printHelp()
		os.Exit(1)
	}
}

// ingestDocument processes a document and stores it as vectors
func (demo *LlamaDemo) ingestDocument(filePath string, chunkSize int, verbose bool) error {
	if verbose {
		fmt.Printf("🔍 Ingesting document: %s\n", filePath)
	}

	// Read the document
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Split into chunks
	chunks := demo.chunkText(string(content), chunkSize)
	if verbose {
		fmt.Printf("📄 Split document into %d chunks\n", len(chunks))
	}

	// Process each chunk
	successCount := 0
	for i, chunk := range chunks {
		if verbose {
			fmt.Printf("⚙️  Processing chunk %d/%d...\n", i+1, len(chunks))
		}

		// Generate embedding using Llama
		embedding, err := demo.generateEmbedding(chunk.Text)
		if err != nil {
			fmt.Printf("Warning: Failed to generate embedding for chunk %d: %v\n", i+1, err)
			continue
		}

		// Store in 8fs vector storage
		vector := vectors.Vector{
			ID:        chunk.ID,
			Embedding: embedding,
			Metadata:  chunk.Metadata,
		}

		if err := demo.storeVector(&vector); err != nil {
			fmt.Printf("Warning: Failed to store chunk %d: %v\n", i+1, err)
			continue
		}

		successCount++
	}

	fmt.Printf("✅ Successfully ingested %d/%d chunks\n", successCount, len(chunks))
	return nil
}

// searchVectors performs semantic search using vector similarity
func (demo *LlamaDemo) searchVectors(query string, topK int, verbose bool) error {
	if verbose {
		fmt.Printf("🔍 Searching for: %s\n", query)
	}

	// Generate query embedding
	queryEmbedding, err := demo.generateEmbedding(query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search vectors
	results, err := demo.searchSimilarVectors(queryEmbedding, topK)
	if err != nil {
		return fmt.Errorf("failed to search vectors: %w", err)
	}

	// Display results
	fmt.Printf("📊 Found %d results:\n\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. Score: %.4f\n", i+1, result.Score)
		fmt.Printf("   ID: %s\n", result.Vector.ID)
		if text, ok := result.Vector.Metadata["text"]; ok {
			textStr := text.(string)
			if len(textStr) > 200 {
				textStr = textStr[:200] + "..."
			}
			fmt.Printf("   Text: %s\n", textStr)
		}
		fmt.Println()
	}

	return nil
}

// performRAG performs retrieval-augmented generation
func (demo *LlamaDemo) performRAG(query string, topK int, verbose bool) error {
	if verbose {
		fmt.Printf("🤖 Performing RAG for: %s\n", query)
	}

	// First, search for relevant context
	queryEmbedding, err := demo.generateEmbedding(query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}

	results, err := demo.searchSimilarVectors(queryEmbedding, topK)
	if err != nil {
		return fmt.Errorf("failed to search vectors: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No relevant context found for the query.")
		return nil
	}

	// Build context from search results
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Based on the following context:\n\n")

	for i, result := range results {
		if text, ok := result.Vector.Metadata["text"]; ok {
			contextBuilder.WriteString(fmt.Sprintf("Context %d:\n%s\n\n", i+1, text))
		}
	}

	contextBuilder.WriteString(fmt.Sprintf("Question: %s\n\nAnswer:", query))

	// Generate answer using Llama
	answer, err := demo.generateLlamaResponse(contextBuilder.String())
	if err != nil {
		fmt.Printf("📖 Context retrieved from %d documents\n", len(results))
		fmt.Printf("💡 RAG Context:\n%s\n", contextBuilder.String())
		fmt.Printf("⚠️  Could not generate response with Llama: %v\n", err)
		fmt.Println("🤖 Note: Start Ollama and ensure your model is available for full RAG functionality.")
	} else {
		fmt.Printf("📖 Context retrieved from %d documents\n", len(results))
		fmt.Printf("🤖 Generated Answer:\n%s\n", answer)
	}

	return nil
}

// generateEmbedding calls the Llama API to generate embeddings
func (demo *LlamaDemo) generateEmbedding(text string) ([]float64, error) {
	// First try real Ollama API
	if embedding, err := demo.generateOllamaEmbedding(text); err == nil {
		return embedding, nil
	}

	// Fall back to mock embedding for demo purposes
	return demo.mockEmbedding(text), nil
}

// generateOllamaEmbedding calls the real Ollama API for embeddings
func (demo *LlamaDemo) generateOllamaEmbedding(text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model":  demo.model,
		"prompt": text,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := demo.client.Post(
		demo.llamaAPI+"/api/embeddings",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var response struct {
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding from Ollama")
	}

	return response.Embedding, nil
}

// generateLlamaResponse calls Ollama for text generation
func (demo *LlamaDemo) generateLlamaResponse(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  demo.model,
		"prompt": prompt,
		"stream": false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := demo.client.Post(
		demo.llamaAPI+"/api/generate",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var response struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	return response.Response, nil
}

// mockEmbedding generates a deterministic mock embedding based on text
func (demo *LlamaDemo) mockEmbedding(text string) []float64 {
	// Simple hash-based embedding for demo (384 dimensions)
	embedding := make([]float64, 384)
	hash := 0
	for _, r := range text {
		hash = hash*31 + int(r)
	}

	for i := range embedding {
		hash = hash*31 + i
		embedding[i] = float64((hash%1000)-500) / 1000.0 // Normalize to [-0.5, 0.5]
	}

	return embedding
}

// chunkText splits text into manageable chunks
func (demo *LlamaDemo) chunkText(text string, chunkSize int) []DocumentChunk {
	words := strings.Fields(text)
	var chunks []DocumentChunk

	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunkText := strings.Join(words[i:end], " ")
		chunk := DocumentChunk{
			ID:   fmt.Sprintf("chunk_%d", len(chunks)+1),
			Text: chunkText,
			Metadata: map[string]interface{}{
				"text":       chunkText,
				"chunk_id":   len(chunks) + 1,
				"word_start": i,
				"word_end":   end,
				"created_at": time.Now().Format(time.RFC3339),
			},
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// storeVector stores a vector in 8fs via HTTP API
func (demo *LlamaDemo) storeVector(vector *vectors.Vector) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"id":        vector.ID,
		"embedding": vector.Embedding,
		"metadata":  vector.Metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := demo.client.Post(
		demo.serverURL+"/api/v1/vectors/embeddings",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to store vector: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// searchSimilarVectors searches for similar vectors via HTTP API
func (demo *LlamaDemo) searchSimilarVectors(query []float64, topK int) ([]vectors.SearchResult, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"query": query,
		"top_k": topK,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := demo.client.Post(
		demo.serverURL+"/api/v1/vectors/search",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Results []vectors.SearchResult `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Results, nil
}

// printHelp displays usage information
func printHelp() {
	fmt.Println(`8fs Llama Integration Demo

USAGE:
  llama-demo -cmd <command> [options]

COMMANDS:
  ingest    Ingest a document and store as vector embeddings
  search    Search for similar vectors using text query
  rag       Perform retrieval-augmented generation
  help      Show this help message

OPTIONS:
  -server   8fs server URL (default: http://localhost:8080)
  -llama    Llama API URL (default: http://localhost:11434)
  -model    Llama model name (default: llama2)
  -input    Input file path (required for ingest)
  -query    Search query (required for search/rag)
  -topk     Number of search results (default: 5)
  -chunk    Document chunk size (default: 500)
  -verbose  Enable verbose output

EXAMPLES:
  # Ingest a document
  llama-demo -cmd ingest -input document.txt -verbose

  # Search for similar content
  llama-demo -cmd search -query "artificial intelligence" -topk 3

  # Perform RAG-style retrieval
  llama-demo -cmd rag -query "What is machine learning?"

PREREQUISITES:
  1. Start 8fs server: ./bin/8fs
  2. (Optional) Start Ollama: ollama serve
  3. (Optional) Pull Llama model: ollama pull llama2

NOTE: This demo uses mock embeddings. For real Llama embeddings,
      ensure Ollama is running with your desired model.`)
}
