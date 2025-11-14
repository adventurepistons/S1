package llm

import (
	"context"
	"fmt"
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
