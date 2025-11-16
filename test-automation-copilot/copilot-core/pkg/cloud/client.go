package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CloudClient handles communication with the cloud AI backend
type CloudClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// CloudConfig configures the cloud client
type CloudConfig struct {
	BaseURL string // e.g., "https://api.testcopilot.ai"
	APIKey  string // User's API key for your service
}

// ContextPayload contains preprocessed context sent to cloud
type ContextPayload struct {
	// User intent
	Action      string `json:"action"` // "pageobject", "test", "chat", "fix"
	UserQuery   string `json:"userQuery"`
	Spec        string `json:"spec,omitempty"`

	// Code context
	RelevantCode []CodeSnippet `json:"relevantCode,omitempty"`
	PageObjects  []PageObject  `json:"pageObjects,omitempty"`
	TestMethods  []TestMethod  `json:"testMethods,omitempty"`

	// Workspace info
	Framework   string        `json:"framework,omitempty"`
	TestRunner  string        `json:"testRunner,omitempty"`
	Workspace   WorkspaceInfo `json:"workspace,omitempty"`

	// Additional context
	Elements    []Element     `json:"elements,omitempty"`    // For page object generation
	ErrorInfo   *ErrorContext `json:"errorInfo,omitempty"`   // For code fixing
}

// CodeSnippet represents a piece of relevant code
type CodeSnippet struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"` // "class", "method", "field"
	Content    string  `json:"content"`
	FilePath   string  `json:"filePath"`
	Similarity float32 `json:"similarity,omitempty"`
}

// PageObject represents a page object class
type PageObject struct {
	Name     string   `json:"name"`
	FilePath string   `json:"filePath"`
	Methods  []string `json:"methods"`
}

// TestMethod represents a test method
type TestMethod struct {
	Name        string   `json:"name"`
	ClassName   string   `json:"className"`
	Annotations []string `json:"annotations"`
}

// WorkspaceInfo contains workspace metadata
type WorkspaceInfo struct {
	TotalClasses    int `json:"totalClasses"`
	TotalTests      int `json:"totalTests"`
	PageObjectCount int `json:"pageObjectCount"`
}

// Element represents a UI element for page object generation
type Element struct {
	Name         string `json:"name"`
	LocatorType  string `json:"locatorType"`
	LocatorValue string `json:"locatorValue"`
}

// ErrorContext contains information about code errors
type ErrorContext struct {
	BrokenCode   string `json:"brokenCode"`
	ErrorMessage string `json:"errorMessage"`
	ErrorType    string `json:"errorType,omitempty"`
}

// GenerationResult contains the response from cloud
type GenerationResult struct {
	Code         string `json:"code"`
	TokensUsed   int    `json:"tokensUsed"`
	Model        string `json:"model"`
	FinishReason string `json:"finishReason"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Type    string      `json:"type"` // "chunk", "complete", "error"
	Content string      `json:"content,omitempty"`
	Result  *GenerationResult `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewCloudClient creates a new cloud API client
func NewCloudClient(config CloudConfig) (*CloudClient, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		// Try environment variable
		baseURL = os.Getenv("TESTCOPILOT_CLOUD_URL")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("cloud API URL not configured")
	}

	apiKey := config.APIKey
	if apiKey == "" {
		// Try environment variable
		apiKey = os.Getenv("TESTCOPILOT_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("cloud API key not configured")
	}

	return &CloudClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

// Generate sends context to cloud and gets generated code
func (c *CloudClient) Generate(ctx context.Context, payload ContextPayload) (*GenerationResult, error) {
	url := fmt.Sprintf("%s/v1/generate", c.baseURL)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("User-Agent", "TestCopilot-Local/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloud API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result GenerationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GenerateStream sends context to cloud and streams the response
func (c *CloudClient) GenerateStream(
	ctx context.Context,
	payload ContextPayload,
	onChunk func(chunk string) error,
) (*GenerationResult, error) {
	url := fmt.Sprintf("%s/v1/generate/stream", c.baseURL)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloud API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read streaming response
	decoder := json.NewDecoder(resp.Body)
	var finalResult *GenerationResult

	for {
		var chunk StreamChunk
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode stream: %w", err)
		}

		switch chunk.Type {
		case "chunk":
			if onChunk != nil {
				if err := onChunk(chunk.Content); err != nil {
					return nil, err
				}
			}

		case "complete":
			finalResult = chunk.Result

		case "error":
			return nil, fmt.Errorf("cloud error: %s", chunk.Error)
		}
	}

	if finalResult == nil {
		return nil, fmt.Errorf("no result received from cloud")
	}

	return finalResult, nil
}

// CheckHealth checks if cloud API is available
func (c *CloudClient) CheckHealth(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloud API unhealthy (status %d)", resp.StatusCode)
	}

	return nil
}

// GetUsageStats gets user's usage statistics from cloud
func (c *CloudClient) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/usage", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get usage stats (status %d)", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}
