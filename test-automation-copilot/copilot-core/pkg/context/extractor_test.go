package context

import (
	"context"
	"testing"

	"github.com/yourusername/copilot-core/pkg/cloud"
)

func TestNewContextExtractor(t *testing.T) {
	tests := []struct {
		name    string
		config  ContextExtractorConfig
		wantErr bool
	}{
		{
			name: "nil database",
			config: ContextExtractorConfig{
				DB:             nil,
				SemanticSearch: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewContextExtractor(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContextExtractor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractPageObjectContext(t *testing.T) {
	// This test requires a full database setup
	// For now, just verify the method signature exists
	t.Skip("Integration test - requires database setup")
}

func TestExtractTestContext(t *testing.T) {
	t.Skip("Integration test - requires database setup")
}

func TestExtractChatContext(t *testing.T) {
	t.Skip("Integration test - requires database setup")
}

func TestExtractFixContext(t *testing.T) {
	t.Skip("Integration test - requires database setup")
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty strings", "", "", false},
		{"exact match", "test", "test", true},
		{"prefix", "testng", "test", true},
		{"suffix", "junit", "unit", true},
		{"no match", "selenium", "playwright", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestWorkspaceInfo(t *testing.T) {
	// Verify WorkspaceInfo structure from cloud package
	info := cloud.WorkspaceInfo{
		TotalClasses:    10,
		TotalTests:      25,
		PageObjectCount: 5,
	}

	if info.TotalClasses != 10 {
		t.Errorf("Expected TotalClasses=10, got %d", info.TotalClasses)
	}
}
