package llm

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	client := NewClient()

	if client == nil {
		t.Fatal("Client is nil")
	}

	if !client.IsConfigured() {
		t.Error("Client should be configured with OPENAI_API_KEY")
	}
}

func TestComplete(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	client := NewClient()
	ctx := context.Background()

	req := CompletionRequest{
		SystemPrompt: "You are a helpful assistant that responds concisely.",
		UserPrompt:   "What is 2+2? Answer with just the number.",
		Temperature:  0.1,
		MaxTokens:    50,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Failed to complete: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Content == "" {
		t.Error("Response content is empty")
	}

	// Response should contain "4"
	if !strings.Contains(resp.Content, "4") {
		t.Errorf("Expected response to contain '4', got: %s", resp.Content)
	}

	if resp.TokensUsed == 0 {
		t.Error("Expected token count > 0")
	}
}

func TestCompleteWithContext(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	client := NewClient()
	ctx := context.Background()

	req := CompletionRequest{
		SystemPrompt: "You are a Java test automation expert.",
		UserPrompt:   "Based on the context provided, what is the purpose of this class?",
		Context: []string{
			"public class LoginPage { private WebDriver driver; }",
		},
		Temperature: 0.3,
		MaxTokens:   100,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Failed to complete with context: %v", err)
	}

	if resp.Content == "" {
		t.Error("Response content is empty")
	}

	// Response should mention login or page object
	lower := strings.ToLower(resp.Content)
	if !strings.Contains(lower, "login") && !strings.Contains(lower, "page") {
		t.Errorf("Expected response to mention login or page, got: %s", resp.Content)
	}
}

func TestCompleteStream(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	client := NewClient()
	ctx := context.Background()

	req := CompletionRequest{
		SystemPrompt: "You are a helpful assistant.",
		UserPrompt:   "Count from 1 to 3. Output just the numbers separated by commas.",
		Temperature:  0.1,
		MaxTokens:    50,
	}

	var chunks []string
	callback := func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	}

	resp, err := client.CompleteStream(ctx, req, callback)
	if err != nil {
		t.Fatalf("Failed to complete stream: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Content == "" {
		t.Error("Response content is empty")
	}

	// Check that we received multiple chunks
	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	// Verify that concatenating chunks equals full content
	var concatenated string
	for _, chunk := range chunks {
		concatenated += chunk
	}

	if concatenated != resp.Content {
		t.Errorf("Concatenated chunks don't match full content")
	}
}

func TestEstimateCost(t *testing.T) {
	client := NewClient()

	tests := []struct {
		inputTokens  int
		outputTokens int
		expectedCost float64
	}{
		{1000, 1000, 0.04},    // (1000/1000)*0.01 + (1000/1000)*0.03 = 0.01 + 0.03
		{2000, 1000, 0.05},    // (2000/1000)*0.01 + (1000/1000)*0.03 = 0.02 + 0.03
		{10000, 5000, 0.25},   // (10000/1000)*0.01 + (5000/1000)*0.03 = 0.10 + 0.15
	}

	for _, tt := range tests {
		cost := client.EstimateCost(tt.inputTokens, tt.outputTokens)
		if cost != tt.expectedCost {
			t.Errorf("EstimateCost(%d, %d) = %.3f, want %.3f",
				tt.inputTokens, tt.outputTokens, cost, tt.expectedCost)
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
			t.Errorf("ApproximateTokenCount(%q) = %d, want %d",
				tt.text, count, tt.expected)
		}
	}
}

func TestBuildMessages(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name          string
		req           CompletionRequest
		expectedCount int
		checkSystem   bool
		checkContext  bool
	}{
		{
			name: "Only user prompt",
			req: CompletionRequest{
				UserPrompt: "Hello",
			},
			expectedCount: 1,
			checkSystem:   false,
			checkContext:  false,
		},
		{
			name: "System and user prompt",
			req: CompletionRequest{
				SystemPrompt: "You are helpful",
				UserPrompt:   "Hello",
			},
			expectedCount: 2,
			checkSystem:   true,
			checkContext:  false,
		},
		{
			name: "System, context, and user",
			req: CompletionRequest{
				SystemPrompt: "You are helpful",
				UserPrompt:   "Hello",
				Context:      []string{"context1", "context2"},
			},
			expectedCount: 4, // system + 2 context + user
			checkSystem:   true,
			checkContext:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := client.buildMessages(tt.req)

			if len(messages) != tt.expectedCount {
				t.Errorf("Expected %d messages, got %d", tt.expectedCount, len(messages))
			}

			if tt.checkSystem && messages[0].Content != tt.req.SystemPrompt {
				t.Error("First message should be system prompt")
			}

			// User prompt should be last
			lastMsg := messages[len(messages)-1]
			if lastMsg.Content != tt.req.UserPrompt {
				t.Error("Last message should be user prompt")
			}
		})
	}
}

// Benchmark completion
func BenchmarkComplete(b *testing.B) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set")
	}

	client := NewClient()
	ctx := context.Background()

	req := CompletionRequest{
		SystemPrompt: "You are helpful.",
		UserPrompt:   "Say hello",
		Temperature:  0.1,
		MaxTokens:    10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Complete(ctx, req)
		if err != nil {
			b.Fatalf("Failed to complete: %v", err)
		}
	}
}
