package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/8fs-io/core/internal/domain/vectors"
	"github.com/8fs-io/core/pkg/logger"
)

// Service provides AI integration capabilities
type Service interface {
	// GenerateEmbedding generates vector embeddings from text using Ollama
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)

	// ProcessAndStoreDocument processes text content and stores vectors
	ProcessAndStoreDocument(ctx context.Context, objectID, text string, metadata map[string]interface{}) error

	// DeleteDocument removes all vectors associated with a document
	DeleteDocument(ctx context.Context, documentID string) error

	// IsTextContent determines if content should be processed for embeddings
	IsTextContent(contentType string) bool

	// ChunkText splits text into manageable chunks for embedding
	ChunkText(text string, chunkSize int) []string

	// GenerateText generates text using the configured AI provider
	GenerateText(ctx context.Context, req TextGenerationRequest) (*TextGenerationResponse, error)
}

// TextGenerationRequest represents a text generation request
type TextGenerationRequest struct {
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// TextGenerationResponse represents the response from text generation
type TextGenerationResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Created int64      `json:"created"`
	Model   string     `json:"model"`
	Choices []Choice   `json:"choices"`
	Usage   TokenUsage `json:"usage"`
}

// Choice represents a generation choice
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	BaseURL    string        `json:"base_url"`
	EmbedModel string        `json:"embed_model"`
	ChatModel  string        `json:"chat_model"`
	Timeout    time.Duration `json:"timeout"`
}

// DefaultOllamaConfig returns default Ollama configuration
func DefaultOllamaConfig() *OllamaConfig {
	return &OllamaConfig{
		BaseURL:    "http://localhost:11434",
		EmbedModel: "all-minilm:latest", // Default embedding model for 384-dim embeddings
		ChatModel:  "llama3.2:1b",       // Default chat model
		Timeout:    30 * time.Second,
	}
}

// service implements Service interface
type service struct {
	config        *OllamaConfig
	vectorStorage *vectors.SQLiteVecStorage
	client        *http.Client
	logger        logger.Logger
}

// NewService creates a new AI service
func NewService(config *OllamaConfig, vectorStorage *vectors.SQLiteVecStorage, logger logger.Logger) Service {
	if config == nil {
		config = DefaultOllamaConfig()
	}

	return &service{
		config:        config,
		vectorStorage: vectorStorage,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// GenerateEmbedding generates vector embeddings using Ollama API
func (s *service) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	// Prepare request for Ollama embeddings API
	reqBody := map[string]interface{}{
		"model":  s.config.EmbedModel,
		"prompt": text,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Make request to Ollama
	url := s.config.BaseURL + "/api/embeddings"
	s.logger.Info("Making Ollama embedding request", "url", url, "model", s.config.EmbedModel)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Ollama API error", "status", resp.StatusCode, "url", url, "model", s.config.EmbedModel)
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding from Ollama")
	}

	return result.Embedding, nil
}

// ProcessAndStoreDocument processes text and stores embeddings
func (s *service) ProcessAndStoreDocument(ctx context.Context, objectID, text string, metadata map[string]interface{}) error {
	if text == "" {
		s.logger.Debug("Skipping empty text content", "object_id", objectID)
		return nil
	}

	// Add AI processing metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["ai_processed_at"] = time.Now().UTC()
	metadata["ai_model"] = s.config.EmbedModel

	// For small texts, process as single chunk
	if len(text) <= 1000 {
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		metadata["text"] = text
		metadata["chunk_index"] = 0
		metadata["total_chunks"] = 1

		vector := &vectors.Vector{
			ID:        objectID,
			Embedding: embedding,
			Metadata:  metadata,
		}

		err = s.vectorStorage.Store(vector)
		if err != nil {
			return fmt.Errorf("failed to store vector: %w", err)
		}

		s.logger.Info("Stored vector embedding", "object_id", objectID, "embedding_dim", len(embedding))
		return nil
	}

	// For larger texts, chunk and process each chunk
	chunks := s.ChunkText(text, 500) // 500 word chunks

	for i, chunk := range chunks {
		chunkID := fmt.Sprintf("%s_chunk_%d", objectID, i)

		embedding, err := s.GenerateEmbedding(ctx, chunk)
		if err != nil {
			s.logger.Warn("Failed to generate embedding for chunk", "object_id", objectID, "chunk", i, "error", err)
			continue // Continue with other chunks
		}

		chunkMetadata := make(map[string]interface{})
		for k, v := range metadata {
			chunkMetadata[k] = v
		}
		chunkMetadata["text"] = chunk
		chunkMetadata["chunk_index"] = i
		chunkMetadata["total_chunks"] = len(chunks)
		chunkMetadata["parent_object"] = objectID

		vector := &vectors.Vector{
			ID:        chunkID,
			Embedding: embedding,
			Metadata:  chunkMetadata,
		}

		err = s.vectorStorage.Store(vector)
		if err != nil {
			s.logger.Warn("Failed to store vector for chunk", "object_id", objectID, "chunk", i, "error", err)
			continue
		}
	}

	s.logger.Info("Processed document chunks", "object_id", objectID, "total_chunks", len(chunks))
	return nil
}

// DeleteDocument removes all vectors associated with a document
func (s *service) DeleteDocument(ctx context.Context, documentID string) error {
	if s.vectorStorage == nil {
		return fmt.Errorf("vector storage not available")
	}

	s.logger.Info("deleting document vectors", "document_id", documentID)

	err := s.vectorStorage.Delete(documentID)
	if err != nil {
		s.logger.Error("failed to delete document vectors", "document_id", documentID, "error", err)
		return fmt.Errorf("failed to delete document vectors: %w", err)
	}

	s.logger.Info("successfully deleted document vectors", "document_id", documentID)
	return nil
}

// IsTextContent determines if content should be processed for embeddings
func (s *service) IsTextContent(contentType string) bool {
	textTypes := []string{
		"text/plain",
		"text/markdown",
		"text/html",
		"text/csv",
		"application/json",
		"application/xml",
		"text/xml",
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))

	for _, textType := range textTypes {
		if strings.HasPrefix(contentType, textType) {
			return true
		}
	}

	return false
}

// ChunkText splits text into chunks for embedding processing
func (s *service) ChunkText(text string, chunkSize int) []string {
	if text == "" {
		return []string{}
	}

	words := strings.Fields(text)
	if len(words) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
	}

	return chunks
}

// GenerateText generates text using Ollama chat API
func (s *service) GenerateText(ctx context.Context, req TextGenerationRequest) (*TextGenerationResponse, error) {
	// Convert messages to Ollama format
	var prompt strings.Builder
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			prompt.WriteString(fmt.Sprintf("System: %s\n\n", msg.Content))
		case "user":
			prompt.WriteString(fmt.Sprintf("User: %s\n\n", msg.Content))
		case "assistant":
			prompt.WriteString(fmt.Sprintf("Assistant: %s\n\n", msg.Content))
		}
	}
	prompt.WriteString("Assistant: ")

	// Prepare Ollama request
	ollamaReq := map[string]interface{}{
		"model":  s.config.ChatModel, // Use configurable chat model
		"prompt": prompt.String(),
		"stream": false,
	}

	if req.MaxTokens > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"num_predict": req.MaxTokens,
			"temperature": req.Temperature,
		}
	}

	bodyBytes, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Make request to Ollama
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL+"/api/generate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Convert to OpenAI-compatible format
	return &TextGenerationResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "llama3.2:1b",
		Choices: []Choice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: strings.TrimSpace(ollamaResp.Response),
				},
				FinishReason: "stop",
			},
		},
		Usage: TokenUsage{
			PromptTokens:     len(prompt.String()) / 4, // Rough estimation
			CompletionTokens: len(ollamaResp.Response) / 4,
			TotalTokens:      (len(prompt.String()) + len(ollamaResp.Response)) / 4,
		},
	}, nil
}
