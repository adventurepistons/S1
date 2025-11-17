package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClaudeClient implements LLMClient for Anthropic's Claude
type ClaudeClient struct {
	config     LLMConfig
	httpClient *http.Client
	baseURL    string
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(config LLMConfig) (*ClaudeClient, error) {
	if config.APIKey == "" {
		return nil, &LLMError{
			Code:    "MISSING_API_KEY",
			Message: "ANTHROPIC_API_KEY environment variable not set",
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	return &ClaudeClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // 2 minutes for generation
		},
		baseURL: baseURL,
	}, nil
}

// Generate sends a prompt to Claude and returns the response
func (c *ClaudeClient) Generate(ctx context.Context, request *GenerateRequest) (*GenerateResponse, error) {
	startTime := time.Now()

	// Build request body with prompt caching
	reqBody := c.buildRequestBody(request)

	// Send request
	resp, err := c.sendRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &GenerateResponse{
		Content:          resp.Content[0].Text,
		StopReason:       resp.StopReason,
		InputTokens:      resp.Usage.InputTokens,
		OutputTokens:     resp.Usage.OutputTokens,
		CacheReadTokens:  resp.Usage.CacheReadInputTokens,
		CacheWriteTokens: resp.Usage.CacheCreationInputTokens,
		Model:            c.config.Model,
		Latency:          time.Since(startTime),
	}

	// Calculate cost
	response.Cost = c.calculateCost(response)

	return response, nil
}

// Chat supports multi-turn conversations
func (c *ClaudeClient) Chat(ctx context.Context, messages []Message) (*GenerateResponse, error) {
	startTime := time.Now()

	// Convert messages to Claude format
	claudeMessages := make([]claudeMessage, len(messages))
	for i, msg := range messages {
		claudeMessages[i] = claudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request
	reqBody := map[string]interface{}{
		"model":      c.config.Model,
		"messages":   claudeMessages,
		"max_tokens": c.config.MaxTokens,
	}

	if c.config.Temperature > 0 {
		reqBody["temperature"] = c.config.Temperature
	}

	// Send request
	resp, err := c.sendRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &GenerateResponse{
		Content:          resp.Content[0].Text,
		StopReason:       resp.StopReason,
		InputTokens:      resp.Usage.InputTokens,
		OutputTokens:     resp.Usage.OutputTokens,
		CacheReadTokens:  resp.Usage.CacheReadInputTokens,
		CacheWriteTokens: resp.Usage.CacheCreationInputTokens,
		Model:            c.config.Model,
		Latency:          time.Since(startTime),
	}

	response.Cost = c.calculateCost(response)

	return response, nil
}

// GetProviderName returns "claude"
func (c *ClaudeClient) GetProviderName() string {
	return "claude"
}

// EstimateCost estimates the cost for a request
func (c *ClaudeClient) EstimateCost(totalTokens, cachedTokens int) float64 {
	// Claude Sonnet 4.5 pricing (as of 2025)
	inputCost := float64(totalTokens-cachedTokens) * 3.00 / 1_000_000  // $3.00 per 1M tokens
	cacheCost := float64(cachedTokens) * 0.30 / 1_000_000              // $0.30 per 1M cached tokens
	outputCost := float64(c.config.MaxTokens) * 15.00 / 1_000_000      // $15.00 per 1M tokens (estimated)

	return inputCost + cacheCost + outputCost
}

// buildRequestBody builds the Claude API request with prompt caching
func (c *ClaudeClient) buildRequestBody(request *GenerateRequest) map[string]interface{} {
	// System prompt with caching
	systemContent := []map[string]interface{}{}

	if len(request.CacheBreakpoints) > 0 {
		// Split system prompt by cache breakpoints
		parts := splitByBreakpoints(request.SystemPrompt, request.CacheBreakpoints)

		for i, part := range parts {
			content := map[string]interface{}{
				"type": "text",
				"text": part,
			}

			// Add cache control to last part (most likely to be reused)
			if i == len(parts)-1 {
				content["cache_control"] = map[string]string{"type": "ephemeral"}
			}

			systemContent = append(systemContent, content)
		}
	} else {
		// No caching
		systemContent = append(systemContent, map[string]interface{}{
			"type": "text",
			"text": request.SystemPrompt,
		})
	}

	// Build request
	reqBody := map[string]interface{}{
		"model":   c.config.Model,
		"system":  systemContent,
		"messages": []claudeMessage{
			{Role: "user", Content: request.UserPrompt},
		},
		"max_tokens": request.MaxTokens,
	}

	if request.Temperature > 0 {
		reqBody["temperature"] = request.Temperature
	}

	return reqBody
}

// sendRequest sends HTTP request to Claude API
func (c *ClaudeClient) sendRequest(ctx context.Context, reqBody map[string]interface{}) (*claudeResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &LLMError{
			Code:    "JSON_MARSHAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal request: %v", err),
		}
	}

	// Create HTTP request
	url := c.baseURL + "/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &LLMError{
			Code:    "REQUEST_CREATE_ERROR",
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &LLMError{
			Code:      "HTTP_ERROR",
			Message:   fmt.Sprintf("HTTP request failed: %v", err),
			Retryable: true,
		}
	}
	defer httpResp.Body.Close()

	// Read response
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &LLMError{
			Code:    "RESPONSE_READ_ERROR",
			Message: fmt.Sprintf("Failed to read response: %v", err),
		}
	}

	// Check status code
	if httpResp.StatusCode != 200 {
		var errorResp claudeErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, &LLMError{
				Code:       errorResp.Error.Type,
				Message:    errorResp.Error.Message,
				StatusCode: httpResp.StatusCode,
				Retryable:  httpResp.StatusCode == 429 || httpResp.StatusCode >= 500,
			}
		}

		return nil, &LLMError{
			Code:       "API_ERROR",
			Message:    fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)),
			StatusCode: httpResp.StatusCode,
		}
	}

	// Parse response
	var resp claudeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &LLMError{
			Code:    "JSON_UNMARSHAL_ERROR",
			Message: fmt.Sprintf("Failed to parse response: %v", err),
		}
	}

	return &resp, nil
}

// calculateCost calculates the actual cost based on usage
func (c *ClaudeClient) calculateCost(resp *GenerateResponse) float64 {
	// Claude Sonnet 4.5 pricing
	inputCost := float64(resp.InputTokens-resp.CacheReadTokens) * 3.00 / 1_000_000      // $3.00 per 1M tokens
	cacheCost := float64(resp.CacheReadTokens) * 0.30 / 1_000_000                       // $0.30 per 1M cached tokens
	cacheWriteCost := float64(resp.CacheWriteTokens) * 3.75 / 1_000_000                 // $3.75 per 1M tokens (25% markup)
	outputCost := float64(resp.OutputTokens) * 15.00 / 1_000_000                        // $15.00 per 1M tokens

	return inputCost + cacheCost + cacheWriteCost + outputCost
}

// splitByBreakpoints splits text by cache breakpoint markers
func splitByBreakpoints(text string, breakpoints []string) []string {
	parts := []string{text}

	for _, bp := range breakpoints {
		newParts := []string{}
		for _, part := range parts {
			// Find breakpoint
			idx := indexOf(part, bp)
			if idx >= 0 {
				// Split at breakpoint
				before := part[:idx]
				after := part[idx+len(bp):]

				if before != "" {
					newParts = append(newParts, before)
				}
				if after != "" {
					newParts = append(newParts, after)
				}
			} else {
				newParts = append(newParts, part)
			}
		}
		parts = newParts
	}

	return parts
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Claude API types
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Role       string `json:"role"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

type claudeErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
