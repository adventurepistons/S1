package llm

import (
	"context"
	"fmt"
	"io"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	openaiClient   *openai.Client
	anthropicKey   string
	provider       string // "openai" or "anthropic"
	defaultModel   string
}

type GenerateOptions struct {
	Temperature  float32
	MaxTokens    int
	SystemPrompt string
}

// CompletionRequest represents a request for code generation
type CompletionRequest struct {
	SystemPrompt string   // System instructions
	UserPrompt   string   // User request
	Context      []string // Additional context (code snippets, docs, etc.)
	Temperature  float32  // Creativity (0.0-2.0, default 0.7)
	MaxTokens    int      // Maximum response tokens (default 2000)
}

// CompletionResponse represents the response from the LLM
type CompletionResponse struct {
	Content      string // Generated content
	FinishReason string // Why the completion stopped
	TokensUsed   int    // Total tokens used
	Model        string // Model used
}

// StreamCallback is called for each chunk of the streaming response
type StreamCallback func(chunk string) error

func NewClient() *Client {
	// Check for API keys
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	provider := "openai" // Default
	if anthropicKey != "" && openaiKey == "" {
		provider = "anthropic"
	}

	var openaiClient *openai.Client
	if openaiKey != "" {
		openaiClient = openai.NewClient(openaiKey)
	}

	return &Client{
		openaiClient: openaiClient,
		anthropicKey: anthropicKey,
		provider:     provider,
		defaultModel: "gpt-4", // or "claude-3-sonnet"
	}
}

func (c *Client) Generate(prompt string, options map[string]interface{}) (string, error) {
	if c.provider == "openai" && c.openaiClient != nil {
		return c.generateOpenAI(prompt, options)
	} else if c.provider == "anthropic" && c.anthropicKey != "" {
		return c.generateAnthropic(prompt, options)
	}

	return "", fmt.Errorf("no LLM provider configured. Set OPENAI_API_KEY or ANTHROPIC_API_KEY")
}

func (c *Client) generateOpenAI(prompt string, options map[string]interface{}) (string, error) {
	// Extract options
	temperature := float32(0.7)
	if temp, ok := options["temperature"].(float64); ok {
		temperature = float32(temp)
	}

	maxTokens := 2000
	if tokens, ok := options["max_tokens"].(int); ok {
		maxTokens = tokens
	}

	systemPrompt := "You are a helpful AI assistant."
	if sys, ok := options["system_prompt"].(string); ok {
		systemPrompt = sys
	}

	model := c.defaultModel
	if m, ok := options["model"].(string); ok {
		model = m
	}

	// Create messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// Call OpenAI API
	resp, err := c.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   maxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *Client) generateAnthropic(prompt string, options map[string]interface{}) (string, error) {
	// TODO: Implement Anthropic API client
	// For now, return error
	return "", fmt.Errorf("Anthropic integration not yet implemented")
}

func (c *Client) Chat(message string, context map[string]interface{}) (string, error) {
	// Build conversation prompt
	prompt := c.buildConversationPrompt(message, context)

	// Generate response
	return c.Generate(prompt, map[string]interface{}{
		"temperature":   0.7,
		"max_tokens":    1000,
		"system_prompt": "You are a helpful AI assistant for test automation.",
	})
}

func (c *Client) buildConversationPrompt(message string, context map[string]interface{}) string {
	var prompt string

	// Add context if available
	if ctx, ok := context["relevantCode"]; ok {
		prompt += fmt.Sprintf("Relevant code context:\n%v\n\n", ctx)
	}

	// Add user message
	prompt += fmt.Sprintf("User: %s\n\nAssistant:", message)

	return prompt
}

func (c *Client) SetProvider(provider string) {
	c.provider = provider
}

func (c *Client) SetModel(model string) {
	c.defaultModel = model
}

func (c *Client) IsConfigured() bool {
	return (c.provider == "openai" && c.openaiClient != nil) ||
		(c.provider == "anthropic" && c.anthropicKey != "")
}

// Complete generates a completion without streaming (new API)
func (c *Client) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if c.provider != "openai" || c.openaiClient == nil {
		return nil, fmt.Errorf("OpenAI client not configured")
	}

	// Build messages
	messages := c.buildMessages(req)

	// Set defaults
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	// Create chat completion request
	chatReq := openai.ChatCompletionRequest{
		Model:       c.defaultModel,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	// Call OpenAI API
	resp, err := c.openaiClient.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	return &CompletionResponse{
		Content:      resp.Choices[0].Message.Content,
		FinishReason: string(resp.Choices[0].FinishReason),
		TokensUsed:   resp.Usage.TotalTokens,
		Model:        resp.Model,
	}, nil
}

// CompleteStream generates a completion with streaming
func (c *Client) CompleteStream(ctx context.Context, req CompletionRequest, callback StreamCallback) (*CompletionResponse, error) {
	if c.provider != "openai" || c.openaiClient == nil {
		return nil, fmt.Errorf("OpenAI client not configured")
	}

	// Build messages
	messages := c.buildMessages(req)

	// Set defaults
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	// Create streaming chat completion request
	chatReq := openai.ChatCompletionRequest{
		Model:       c.defaultModel,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      true,
	}

	// Call OpenAI API with streaming
	stream, err := c.openaiClient.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create completion stream: %w", err)
	}
	defer stream.Close()

	// Collect the full response
	var fullContent string
	var finishReason string

	// Read streaming response
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				fullContent += delta

				// Call the callback with the chunk
				if callback != nil {
					if err := callback(delta); err != nil {
						return nil, fmt.Errorf("callback error: %w", err)
					}
				}
			}

			// Check finish reason
			if response.Choices[0].FinishReason != "" {
				finishReason = string(response.Choices[0].FinishReason)
			}
		}
	}

	return &CompletionResponse{
		Content:      fullContent,
		FinishReason: finishReason,
		TokensUsed:   0, // Token count not available in streaming mode
		Model:        c.defaultModel,
	}, nil
}

// buildMessages constructs the message array for the API request
func (c *Client) buildMessages(req CompletionRequest) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0)

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	// Add context as separate messages if provided
	for _, ctx := range req.Context {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: ctx,
		})
	}

	// Add user prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.UserPrompt,
	})

	return messages
}

// EstimateCost estimates the cost for a completion
func (c *Client) EstimateCost(inputTokens, outputTokens int) float64 {
	// GPT-4 Turbo pricing (as of 2024):
	// Input: $0.01 per 1K tokens
	// Output: $0.03 per 1K tokens

	inputCost := (float64(inputTokens) / 1000.0) * 0.01
	outputCost := (float64(outputTokens) / 1000.0) * 0.03

	return inputCost + outputCost
}

// ApproximateTokenCount estimates token count from text
// Rule of thumb: ~4 characters per token for English
func ApproximateTokenCount(text string) int {
	return len(text) / 4
}
