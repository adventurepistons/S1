package embeddings

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewEmbeddingService(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	service, err := NewEmbeddingService(EmbeddingConfig{
		APIKey: apiKey,
	})

	if err != nil {
		t.Fatalf("Failed to create embedding service: %v", err)
	}

	if service == nil {
		t.Fatal("Service is nil")
	}

	if service.model != "text-embedding-3-small" {
		t.Errorf("Expected model text-embedding-3-small, got %s", service.model)
	}
}

func TestNewEmbeddingService_NoAPIKey(t *testing.T) {
	// Temporarily unset API key
	oldKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", oldKey)

	_, err := NewEmbeddingService(EmbeddingConfig{})

	if err == nil {
		t.Error("Expected error when API key not provided")
	}
}

func TestEmbed(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	service, err := NewEmbeddingService(EmbeddingConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	text := "This is a test for login functionality"

	embedding, err := service.Embed(ctx, text)
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	// Check embedding dimensions
	expectedDim := service.GetDimensions()
	if len(embedding) != expectedDim {
		t.Errorf("Expected %d dimensions, got %d", expectedDim, len(embedding))
	}

	// Check that embedding is not all zeros
	sum := float32(0)
	for _, v := range embedding {
		sum += v
	}
	if sum == 0 {
		t.Error("Embedding is all zeros")
	}
}

func TestEmbed_EmptyText(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	service, err := NewEmbeddingService(EmbeddingConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	_, err = service.Embed(ctx, "")

	if err == nil {
		t.Error("Expected error for empty text")
	}
}

func TestEmbedBatch(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	service, err := NewEmbeddingService(EmbeddingConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	texts := []string{
		"Login page with username and password",
		"Dashboard showing user statistics",
		"Profile page with user information",
	}

	embeddings, err := service.EmbedBatch(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to generate batch embeddings: %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}

	// Check each embedding
	expectedDim := service.GetDimensions()
	for i, embedding := range embeddings {
		if len(embedding) != expectedDim {
			t.Errorf("Embedding %d: expected %d dimensions, got %d", i, expectedDim, len(embedding))
		}

		// Check not all zeros
		sum := float32(0)
		for _, v := range embedding {
			sum += v
		}
		if sum == 0 {
			t.Errorf("Embedding %d is all zeros", i)
		}
	}
}

func TestGetDimensions(t *testing.T) {
	tests := []struct {
		model    string
		expected int
	}{
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"text-embedding-ada-002", 1536},
		{"unknown-model", 1536}, // Default
	}

	for _, tt := range tests {
		service := &EmbeddingService{
			model: tt.model,
		}

		dim := service.GetDimensions()
		if dim != tt.expected {
			t.Errorf("Model %s: expected %d dimensions, got %d", tt.model, tt.expected, dim)
		}
	}
}

func TestEstimateCost(t *testing.T) {
	service := &EmbeddingService{
		model: "text-embedding-3-small",
	}

	// Test different token counts
	tests := []struct {
		tokens   int
		expected float64
	}{
		{1000, 0.00002},      // 1K tokens
		{1000000, 0.02},      // 1M tokens
		{10000000, 0.2},      // 10M tokens
	}

	for _, tt := range tests {
		cost := service.EstimateCost(tt.tokens)
		if cost != tt.expected {
			t.Errorf("For %d tokens: expected cost $%.5f, got $%.5f", tt.tokens, tt.expected, cost)
		}
	}
}

func TestApproximateTokenCount(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"Hello", 1},               // 5 chars / 4 = 1
		{"Hello World", 2},         // 11 chars / 4 = 2
		{"This is a test", 3},      // 15 chars / 4 = 3
		{strings.Repeat("a", 100), 25}, // 100 chars / 4 = 25
	}

	for _, tt := range tests {
		count := ApproximateTokenCount(tt.text)
		if count != tt.expected {
			t.Errorf("Text %q: expected %d tokens, got %d", tt.text, tt.expected, count)
		}
	}
}

// Benchmark embedding generation
func BenchmarkEmbed(b *testing.B) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set")
	}

	service, err := NewEmbeddingService(EmbeddingConfig{
		APIKey: apiKey,
	})
	if err != nil {
		b.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	text := "This is a test text for benchmarking embedding generation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Embed(ctx, text)
		if err != nil {
			b.Fatalf("Failed to embed: %v", err)
		}
	}
}
