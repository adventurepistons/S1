package ai

import (
	"context"
	"time"
)

// LLMClient is the interface for interacting with LLM providers
// Supports both Claude (Anthropic) and OpenAI
type LLMClient interface {
	// Generate sends a prompt to the LLM and returns the response
	Generate(ctx context.Context, request *GenerateRequest) (*GenerateResponse, error)

	// Chat supports multi-turn conversations
	Chat(ctx context.Context, messages []Message) (*GenerateResponse, error)

	// GetProviderName returns the name of the LLM provider
	GetProviderName() string

	// EstimateCost estimates the cost for a request
	EstimateCost(totalTokens, cachedTokens int) float64
}

// GenerateRequest represents a request to generate code
type GenerateRequest struct {
	SystemPrompt     string
	UserPrompt       string
	CacheBreakpoints []string // For prompt caching (Claude only)
	MaxTokens        int
	Temperature      float64
	Model            string
}

// GenerateResponse represents the LLM's response
type GenerateResponse struct {
	Content          string
	StopReason       string
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
	Cost             float64
	Latency          time.Duration
	Model            string
}

// Message represents a chat message
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// LLMConfig holds configuration for LLM clients
type LLMConfig struct {
	Provider    string  // "claude" or "openai"
	APIKey      string
	Model       string  // e.g., "claude-sonnet-4.5-20250929"
	MaxTokens   int     // Default: 4000
	Temperature float64 // Default: 0.0 (deterministic)
	BaseURL     string  // Optional custom endpoint
}

// DefaultLLMConfig returns default configuration for Claude
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:    "claude",
		Model:       "claude-sonnet-4.5-20250929",
		MaxTokens:   4000,
		Temperature: 0.0, // Deterministic for code generation
	}
}

// NewLLMClient creates a new LLM client based on configuration
func NewLLMClient(config LLMConfig) (LLMClient, error) {
	switch config.Provider {
	case "claude":
		return NewClaudeClient(config)
	case "openai":
		// Future: return NewOpenAIClient(config)
		return nil, &LLMError{
			Code:    "UNSUPPORTED_PROVIDER",
			Message: "OpenAI provider not yet implemented",
		}
	default:
		return nil, &LLMError{
			Code:    "INVALID_PROVIDER",
			Message: "Provider must be 'claude' or 'openai'",
		}
	}
}

// LLMError represents an error from the LLM provider
type LLMError struct {
	Code       string
	Message    string
	StatusCode int
	Retryable  bool
}

func (e *LLMError) Error() string {
	return e.Message
}
