package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingClient calls the local embedding service
type EmbeddingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewEmbeddingClient creates a new embedding client
func NewEmbeddingClient(baseURL string) *EmbeddingClient {
	return &EmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// embedRequest is the request payload for /embed endpoint
type embedRequest struct {
	Text string `json:"text"`
}

// embedResponse is the response from /embed endpoint
type embedResponse struct {
	Embedding        []float64 `json:"embedding"`
	Dimension        int       `json:"dimension"`
	ProcessingTimeMs float64   `json:"processing_time_ms"`
}

// embedBatchRequest is the request payload for /embed_batch endpoint
type embedBatchRequest struct {
	Texts []string `json:"texts"`
}

// embedBatchResponse is the response from /embed_batch endpoint
type embedBatchResponse struct {
	Embeddings         [][]float64 `json:"embeddings"`
	Count              int         `json:"count"`
	Dimension          int         `json:"dimension"`
	ProcessingTimeMs   float64     `json:"processing_time_ms"`
	AvgTimePerTextMs   float64     `json:"avg_time_per_text_ms"`
}

// healthResponse is the response from /health endpoint
type healthResponse struct {
	Status      string `json:"status"`
	Model       string `json:"model"`
	Dimension   int    `json:"dimension"`
	ModelLoaded bool   `json:"model_loaded"`
}

// Embed generates an embedding for a single text
func (ec *EmbeddingClient) Embed(text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	// Prepare request
	reqBody := embedRequest{Text: text}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := ec.httpClient.Post(
		ec.baseURL+"/embed",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embedResp embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embedResp.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts (much faster!)
func (ec *EmbeddingClient) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty text list provided")
	}

	// Prepare request
	reqBody := embedBatchRequest{Texts: texts}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := ec.httpClient.Post(
		ec.baseURL+"/embed_batch",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embedResp embedBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embedResp.Embeddings, nil
}

// HealthCheck checks if the embedding service is running
func (ec *EmbeddingClient) HealthCheck() error {
	resp, err := ec.httpClient.Get(ec.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("embedding service not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("embedding service unhealthy: status %d", resp.StatusCode)
	}

	var health healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return fmt.Errorf("failed to decode health response: %w", err)
	}

	if health.Status != "healthy" {
		return fmt.Errorf("embedding service reports unhealthy status: %s", health.Status)
	}

	if !health.ModelLoaded {
		return fmt.Errorf("embedding model not loaded")
	}

	return nil
}

// GetModelInfo returns information about the loaded model
func (ec *EmbeddingClient) GetModelInfo() (string, int, error) {
	resp, err := ec.httpClient.Get(ec.baseURL + "/model_info")
	if err != nil {
		return "", 0, fmt.Errorf("failed to get model info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("failed to get model info: status %d", resp.StatusCode)
	}

	var info struct {
		ModelName string `json:"model_name"`
		Dimension int    `json:"dimension"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", 0, fmt.Errorf("failed to decode model info: %w", err)
	}

	return info.ModelName, info.Dimension, nil
}
