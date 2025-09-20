package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/8fs-io/core/internal/domain/ai"
	"github.com/8fs-io/core/internal/domain/vectors"
	"github.com/8fs-io/core/pkg/logger"
)

// Service provides RAG (Retrieval-Augmented Generation) capabilities
type Service interface {
	// Chat performs RAG-based completion with retrieval and generation
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// SearchContext retrieves relevant context for a query
	SearchContext(ctx context.Context, query string, limit int) (*ContextResponse, error)

	// GenerateWithContext generates text using provided context
	GenerateWithContext(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Query       string            `json:"query" binding:"required"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	TopK        int               `json:"top_k,omitempty"`    // Number of documents to retrieve
	Metadata    map[string]string `json:"metadata,omitempty"` // Filter metadata for retrieval
	Stream      bool              `json:"stream,omitempty"`   // Stream response (future)
}

// ChatResponse represents the chat completion response
type ChatResponse struct {
	ID          string           `json:"id"`
	Object      string           `json:"object"`
	Created     int64            `json:"created"`
	Model       string           `json:"model"`
	Choices     []ChatChoice     `json:"choices"`
	Usage       TokenUsage       `json:"usage"`
	Context     *ContextResponse `json:"context,omitempty"`
	ProcessTime time.Duration    `json:"process_time"`
}

// ChatChoice represents a single completion choice
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// ContextResponse represents retrieved context
type ContextResponse struct {
	Documents []ContextDocument `json:"documents"`
	Query     string            `json:"query"`
	TopK      int               `json:"top_k"`
	TotalDocs int               `json:"total_docs"`
}

// ContextDocument represents a retrieved document with context
type ContextDocument struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Score    float64                `json:"score"`
	Source   string                 `json:"source,omitempty"`
	Chunk    int                    `json:"chunk,omitempty"`
}

// GenerationRequest represents a text generation request
type GenerationRequest struct {
	Prompt      string  `json:"prompt" binding:"required"`
	Context     string  `json:"context,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	SystemMsg   string  `json:"system_message,omitempty"`
}

// GenerationResponse represents the generation response
type GenerationResponse struct {
	Text     string        `json:"text"`
	Model    string        `json:"model"`
	Usage    TokenUsage    `json:"usage"`
	Duration time.Duration `json:"duration"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Config holds RAG service configuration
type Config struct {
	DefaultTopK        int     `json:"default_top_k"`       // Default number of documents to retrieve
	DefaultMaxTokens   int     `json:"default_max_tokens"`  // Default max tokens for generation
	DefaultTemperature float64 `json:"default_temperature"` // Default generation temperature
	ContextWindowSize  int     `json:"context_window_size"` // Max context window size
	SystemPrompt       string  `json:"system_prompt"`       // Default system prompt
	MinRelevanceScore  float64 `json:"min_relevance_score"` // Minimum relevance score for documents
}

// service implements the RAG Service interface
type service struct {
	aiService     ai.Service
	vectorStorage *vectors.SQLiteVecStorage
	logger        logger.Logger
	config        *Config
}

// NewService creates a new RAG service
func NewService(aiService ai.Service, vectorStorage *vectors.SQLiteVecStorage, logger logger.Logger, config *Config) Service {
	if config == nil {
		config = &Config{
			DefaultTopK:        5,
			DefaultMaxTokens:   4000,
			DefaultTemperature: 0.7,
			ContextWindowSize:  8000,
			MinRelevanceScore:  0.1,
			SystemPrompt:       "You are a helpful AI assistant. Use the provided context to answer questions accurately. If the context doesn't contain relevant information, say so clearly.",
		}
	}

	return &service{
		aiService:     aiService,
		vectorStorage: vectorStorage,
		logger:        logger,
		config:        config,
	}
}

// Chat performs RAG-based completion
func (s *service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	start := time.Now()
	requestID := fmt.Sprintf("rag_%d", time.Now().UnixNano())

	s.logger.Info("RAG chat request", "request_id", requestID, "query_length", len(req.Query))

	// Set defaults
	if req.TopK <= 0 {
		req.TopK = s.config.DefaultTopK
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = s.config.DefaultMaxTokens
	}
	if req.Temperature <= 0 {
		req.Temperature = s.config.DefaultTemperature
	}

	// Step 1: Retrieve relevant context
	contextResp, err := s.SearchContext(ctx, req.Query, req.TopK)
	if err != nil {
		s.logger.Error("failed to retrieve context", "request_id", requestID, "error", err)
		return nil, fmt.Errorf("context retrieval failed: %w", err)
	}

	// Step 2: Build context string
	contextText := s.buildContextString(contextResp.Documents)
	s.logger.Info("context built", "request_id", requestID, "docs_count", len(contextResp.Documents), "context_length", len(contextText))

	// Step 3: Generate response with context
	genReq := GenerationRequest{
		Prompt:      req.Query,
		Context:     contextText,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		SystemMsg:   s.config.SystemPrompt,
	}

	genResp, err := s.GenerateWithContext(ctx, genReq)
	if err != nil {
		s.logger.Error("failed to generate response", "request_id", requestID, "error", err)
		return nil, fmt.Errorf("text generation failed: %w", err)
	}

	// Build response
	response := &ChatResponse{
		ID:      requestID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   genResp.Model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: genResp.Text,
				},
				FinishReason: "stop",
			},
		},
		Usage:       genResp.Usage,
		Context:     contextResp,
		ProcessTime: time.Since(start),
	}

	s.logger.Info("RAG chat completed", "request_id", requestID, "duration", time.Since(start), "response_length", len(genResp.Text))
	return response, nil
}

// SearchContext retrieves relevant context for a query
func (s *service) SearchContext(ctx context.Context, query string, limit int) (*ContextResponse, error) {
	if limit <= 0 {
		limit = s.config.DefaultTopK
	}

	// Generate query embedding
	queryEmbedding, err := s.aiService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search for similar vectors
	results, err := s.vectorStorage.Search(queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert to context documents
	documents := make([]ContextDocument, 0, len(results))
	for _, result := range results {
		// Filter by minimum relevance score
		if result.Score < s.config.MinRelevanceScore {
			continue
		}

		// Extract content from metadata
		content := ""
		source := ""
		chunk := 0

		if result.Vector.Metadata != nil {
			if textContent, ok := result.Vector.Metadata["content"].(string); ok {
				content = textContent
			}
			if sourceFile, ok := result.Vector.Metadata["source"].(string); ok {
				source = sourceFile
			}
			if chunkNum, ok := result.Vector.Metadata["chunk"].(float64); ok {
				chunk = int(chunkNum)
			}
		}

		documents = append(documents, ContextDocument{
			ID:       result.Vector.ID,
			Content:  content,
			Metadata: result.Vector.Metadata,
			Score:    result.Score,
			Source:   source,
			Chunk:    chunk,
		})
	}

	return &ContextResponse{
		Documents: documents,
		Query:     query,
		TopK:      limit,
		TotalDocs: len(documents),
	}, nil
}

// GenerateWithContext generates text using provided context
func (s *service) GenerateWithContext(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	start := time.Now()

	// Build the full prompt with context
	prompt := s.buildRAGPrompt(req.SystemMsg, req.Context, req.Prompt)

	// Use the actual AI service for generation
	aiResponse, err := s.aiService.GenerateText(ctx, ai.TextGenerationRequest{
		Messages: []ai.ChatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	})
	if err != nil {
		return nil, fmt.Errorf("text generation failed: %w", err)
	}

	// Extract response
	if len(aiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received from AI service")
	}

	generatedText := aiResponse.Choices[0].Message.Content

	return &GenerationResponse{
		Text:     generatedText,
		Model:    aiResponse.Model,
		Duration: time.Since(start),
		Usage: TokenUsage{
			PromptTokens:     aiResponse.Usage.PromptTokens,
			CompletionTokens: aiResponse.Usage.CompletionTokens,
			TotalTokens:      aiResponse.Usage.TotalTokens,
		},
	}, nil
}

// buildContextString creates a formatted context string from documents
func (s *service) buildContextString(documents []ContextDocument) string {
	if len(documents) == 0 {
		return ""
	}

	var contextParts []string
	contextParts = append(contextParts, "## Context Information")

	for i, doc := range documents {
		contextParts = append(contextParts, fmt.Sprintf("\n### Document %d", i+1))
		if doc.Source != "" {
			contextParts = append(contextParts, fmt.Sprintf("Source: %s", doc.Source))
		}
		contextParts = append(contextParts, doc.Content)
	}

	return strings.Join(contextParts, "\n")
}

// buildRAGPrompt constructs the full RAG prompt
func (s *service) buildRAGPrompt(systemMsg, context, query string) string {
	if systemMsg == "" {
		systemMsg = s.config.SystemPrompt
	}

	var promptParts []string
	promptParts = append(promptParts, systemMsg)

	if context != "" {
		promptParts = append(promptParts, "\n"+context)
	}

	promptParts = append(promptParts, fmt.Sprintf("\n\n## Question\n%s\n\n## Answer", query))

	return strings.Join(promptParts, "\n")
}
