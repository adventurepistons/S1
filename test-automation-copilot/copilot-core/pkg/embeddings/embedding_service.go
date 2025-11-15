package embeddings

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// EmbeddingService generates embeddings using OpenAI
type EmbeddingService struct {
	client *openai.Client
	model  string
}

// EmbeddingConfig configures the embedding service
type EmbeddingConfig struct {
	APIKey string // OpenAI API key
	Model  string // Embedding model (default: text-embedding-3-small)
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(config EmbeddingConfig) (*EmbeddingService, error) {
	// Get API key from config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not provided (set OPENAI_API_KEY env var)")
	}

	// Default to text-embedding-3-small (cheaper and faster)
	model := config.Model
	if model == "" {
		model = openai.SmallEmbedding3
	}

	client := openai.NewClient(apiKey)

	return &EmbeddingService{
		client: client,
		model:  model,
	}, nil
}

// Embed generates an embedding for a single text
func (s *EmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Create embedding request
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(s.model),
	}

	// Call OpenAI API
	resp, err := s.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	// Convert []float64 to []float32 for chromem-go compatibility
	embedding64 := resp.Data[0].Embedding
	embedding32 := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		embedding32[i] = float32(v)
	}

	return embedding32, nil
}

// EmbedBatch generates embeddings for multiple texts
func (s *EmbeddingService) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// OpenAI supports batch embeddings (up to 2048 inputs)
	// We'll batch in chunks of 100 to be safe
	const batchSize = 100
	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]

		// Create embedding request
		req := openai.EmbeddingRequest{
			Input: batch,
			Model: openai.EmbeddingModel(s.model),
		}

		// Call OpenAI API
		resp, err := s.client.CreateEmbeddings(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch embeddings: %w", err)
		}

		// Convert embeddings
		for _, data := range resp.Data {
			embedding64 := data.Embedding
			embedding32 := make([]float32, len(embedding64))
			for j, v := range embedding64 {
				embedding32[j] = float32(v)
			}
			allEmbeddings = append(allEmbeddings, embedding32)
		}
	}

	return allEmbeddings, nil
}

// GetDimensions returns the dimension size of the embeddings
func (s *EmbeddingService) GetDimensions() int {
	// text-embedding-3-small: 1536 dimensions
	// text-embedding-3-large: 3072 dimensions
	// text-embedding-ada-002: 1536 dimensions
	switch s.model {
	case openai.SmallEmbedding3:
		return 1536
	case openai.LargeEmbedding3:
		return 3072
	case openai.AdaEmbeddingV2:
		return 1536
	default:
		return 1536 // Default
	}
}

// GetModel returns the current embedding model
func (s *EmbeddingService) GetModel() string {
	return s.model
}

// EstimateCost estimates the cost for embedding N tokens
func (s *EmbeddingService) EstimateCost(tokenCount int) float64 {
	// Pricing as of 2024 (USD per 1M tokens):
	// text-embedding-3-small: $0.02
	// text-embedding-3-large: $0.13
	// text-embedding-ada-002: $0.10

	var pricePerMillion float64
	switch s.model {
	case openai.SmallEmbedding3:
		pricePerMillion = 0.02
	case openai.LargeEmbedding3:
		pricePerMillion = 0.13
	case openai.AdaEmbeddingV2:
		pricePerMillion = 0.10
	default:
		pricePerMillion = 0.02
	}

	return (float64(tokenCount) / 1_000_000) * pricePerMillion
}

// ApproximateTokenCount estimates token count from text
// Rule of thumb: ~4 characters per token for English
func ApproximateTokenCount(text string) int {
	return len(text) / 4
}
