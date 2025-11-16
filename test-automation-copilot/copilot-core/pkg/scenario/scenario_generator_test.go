package scenario

import (
	"encoding/json"
	"testing"
	"time"
)

// Sample recording for login page
func getSampleLoginRecording() *Recording {
	return &Recording{
		SessionID: "rec_login_001",
		URL:       "https://app.example.com/login",
		Title:     "Login - MyApp",
		Elements: []RecordedElement{
			{
				ID:          "email",
				Type:        "email",
				Name:        "email",
				Placeholder: "Enter your email",
				Required:    true,
				Label:       "Email Address",
			},
			{
				ID:          "password",
				Type:        "password",
				Name:        "password",
				Placeholder: "Password",
				Required:    true,
				MinLength:   8,
				Label:       "Password",
			},
			{
				ID:    "remember-me",
				Type:  "checkbox",
				Name:  "rememberMe",
				Label: "Remember me",
			},
			{
				ID:   "login-btn",
				Type: "submit",
				Text: "Sign In",
			},
			{
				ID:   "forgot-password",
				Type: "link",
				Text: "Forgot Password?",
				Href: "/reset-password",
			},
			{
				ID:   "signup-link",
				Type: "link",
				Text: "Create Account",
				Href: "/register",
			},
		},
		Interactions: []RecordedInteraction{
			{
				Action:  "input",
				Element: "email",
				Value:   "test@example.com",
			},
			{
				Action:  "input",
				Element: "password",
				Value:   "Test1234!",
			},
			{
				Action:  "click",
				Element: "login-btn",
			},
			{
				Action: "navigate",
				URL:    "https://app.example.com/dashboard",
			},
		},
		Timestamp: time.Now(),
	}
}

// Sample recording for checkout page
func getSampleCheckoutRecording() *Recording {
	return &Recording{
		SessionID: "rec_checkout_001",
		URL:       "https://shop.example.com/checkout",
		Title:     "Checkout - Shop",
		Elements: []RecordedElement{
			{
				ID:          "shipping-address",
				Type:        "textarea",
				Name:        "address",
				Placeholder: "Enter shipping address",
				Required:    true,
				Label:       "Shipping Address",
			},
			{
				ID:       "city",
				Type:     "text",
				Name:     "city",
				Required: true,
				Label:    "City",
			},
			{
				ID:       "zipcode",
				Type:     "text",
				Name:     "zip",
				Pattern:  "[0-9]{5}",
				Required: true,
				Label:    "ZIP Code",
			},
			{
				ID:        "card-number",
				Type:      "text",
				Name:      "cardNumber",
				Pattern:   "[0-9]{16}",
				MinLength: 16,
				MaxLength: 16,
				Required:  true,
				Label:     "Card Number",
			},
			{
				ID:        "cvv",
				Type:      "password",
				Name:      "cvv",
				Pattern:   "[0-9]{3}",
				MinLength: 3,
				MaxLength: 3,
				Required:  true,
				Label:     "CVV",
			},
			{
				ID:   "place-order-btn",
				Type: "submit",
				Text: "Place Order",
			},
		},
		Interactions: []RecordedInteraction{
			{
				Action:  "input",
				Element: "shipping-address",
				Value:   "123 Main Street",
			},
			{
				Action:  "input",
				Element: "city",
				Value:   "New York",
			},
			{
				Action:  "input",
				Element: "zipcode",
				Value:   "10001",
			},
			{
				Action:  "input",
				Element: "card-number",
				Value:   "4111111111111111",
			},
			{
				Action:  "input",
				Element: "cvv",
				Value:   "123",
			},
			{
				Action:  "click",
				Element: "place-order-btn",
			},
			{
				Action: "navigate",
				URL:    "https://shop.example.com/order-confirmation",
			},
		},
		Timestamp: time.Now(),
	}
}

// Sample recording for registration page
func getSampleRegistrationRecording() *Recording {
	return &Recording{
		SessionID: "rec_registration_001",
		URL:       "https://app.example.com/register",
		Title:     "Create Account - MyApp",
		Elements: []RecordedElement{
			{
				ID:          "username",
				Type:        "text",
				Name:        "username",
				Placeholder: "Choose a username",
				Required:    true,
				MinLength:   3,
				MaxLength:   20,
				Label:       "Username",
			},
			{
				ID:          "email",
				Type:        "email",
				Name:        "email",
				Placeholder: "Your email address",
				Required:    true,
				Label:       "Email",
			},
			{
				ID:          "password",
				Type:        "password",
				Name:        "password",
				Placeholder: "Create a password",
				Required:    true,
				MinLength:   8,
				Label:       "Password",
			},
			{
				ID:          "confirm-password",
				Type:        "password",
				Name:        "confirmPassword",
				Placeholder: "Confirm your password",
				Required:    true,
				MinLength:   8,
				Label:       "Confirm Password",
			},
			{
				ID:       "terms",
				Type:     "checkbox",
				Name:     "acceptTerms",
				Required: true,
				Label:    "I accept the Terms and Conditions",
			},
			{
				ID:   "register-btn",
				Type: "submit",
				Text: "Create Account",
			},
		},
		Interactions: []RecordedInteraction{
			{
				Action:  "input",
				Element: "username",
				Value:   "testuser",
			},
			{
				Action:  "input",
				Element: "email",
				Value:   "test@example.com",
			},
			{
				Action:  "input",
				Element: "password",
				Value:   "SecurePass123!",
			},
			{
				Action:  "input",
				Element: "confirm-password",
				Value:   "SecurePass123!",
			},
			{
				Action:  "click",
				Element: "terms",
			},
			{
				Action:  "click",
				Element: "register-btn",
			},
			{
				Action: "navigate",
				URL:    "https://app.example.com/welcome",
			},
		},
		Timestamp: time.Now(),
	}
}

// TestGenerateScenariosPrompt tests that the prompt is generated correctly
func TestGenerateScenariosPrompt(t *testing.T) {
	// This test doesn't need actual LLM client
	generator := &ScenarioGenerator{}

	recording := getSampleLoginRecording()
	req := GenerateRequest{
		Recording:  recording,
		UserIntent: "I want to test login functionality",
		AppContext: "web application",
	}

	prompt := generator.buildAnalysisPrompt(req)

	// Verify prompt contains key elements
	if prompt == "" {
		t.Error("Prompt should not be empty")
	}

	// Check for recording data
	if !containsString(prompt, recording.URL) {
		t.Error("Prompt should contain recording URL")
	}

	if !containsString(prompt, recording.Title) {
		t.Error("Prompt should contain recording title")
	}

	// Check for user intent
	if !containsString(prompt, req.UserIntent) {
		t.Error("Prompt should contain user intent")
	}

	// Check for elements
	if !containsString(prompt, "email") {
		t.Error("Prompt should contain email element")
	}

	if !containsString(prompt, "password") {
		t.Error("Prompt should contain password element")
	}

	// Check for interactions
	if !containsString(prompt, "test@example.com") {
		t.Error("Prompt should contain interaction values")
	}

	t.Logf("Generated prompt length: %d characters", len(prompt))
}

// TestFilterScenarios tests scenario filtering
func TestFilterScenarios(t *testing.T) {
	generator := &ScenarioGenerator{}

	scenarios := []TestScenario{
		{Category: "positive", Priority: "high", Name: "Test 1"},
		{Category: "negative", Priority: "high", Name: "Test 2"},
		{Category: "positive", Priority: "medium", Name: "Test 3"},
		{Category: "security", Priority: "high", Name: "Test 4"},
		{Category: "edgeCase", Priority: "low", Name: "Test 5"},
	}

	// Test category filter
	filter := ScenarioFilter{
		Categories: []string{"positive"},
	}
	filtered := generator.FilterScenarios(scenarios, filter)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 positive scenarios, got %d", len(filtered))
	}

	// Test priority filter
	filter = ScenarioFilter{
		Priorities: []string{"high"},
	}
	filtered = generator.FilterScenarios(scenarios, filter)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 high priority scenarios, got %d", len(filtered))
	}

	// Test max count
	filter = ScenarioFilter{
		MaxCount: 2,
	}
	filtered = generator.FilterScenarios(scenarios, filter)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 scenarios (max count), got %d", len(filtered))
	}
}

// TestLimitScenarios tests scenario limiting with priority
func TestLimitScenarios(t *testing.T) {
	generator := &ScenarioGenerator{}

	scenarios := []TestScenario{
		{Priority: "high", Name: "High 1"},
		{Priority: "medium", Name: "Medium 1"},
		{Priority: "low", Name: "Low 1"},
		{Priority: "high", Name: "High 2"},
		{Priority: "medium", Name: "Medium 2"},
		{Priority: "low", Name: "Low 2"},
	}

	// Limit to 3 scenarios - should prioritize high priority
	limited := generator.limitScenarios(scenarios, 3)

	if len(limited) != 3 {
		t.Errorf("Expected 3 scenarios, got %d", len(limited))
	}

	// First two should be high priority
	if limited[0].Priority != "high" || limited[1].Priority != "high" {
		t.Error("First scenarios should be high priority")
	}
}

// TestFilterOutSecurity tests security scenario filtering
func TestFilterOutSecurity(t *testing.T) {
	generator := &ScenarioGenerator{}

	scenarios := []TestScenario{
		{Category: "positive", Name: "Test 1"},
		{Category: "security", Name: "SQL Injection"},
		{Category: "negative", Name: "Test 2"},
		{Category: "security", Name: "XSS Test"},
	}

	filtered := generator.filterOutSecurity(scenarios)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 non-security scenarios, got %d", len(filtered))
	}

	for _, s := range filtered {
		if s.Category == "security" {
			t.Error("Security scenarios should be filtered out")
		}
	}
}

// TestCleanGPTResponse tests GPT response cleaning
func TestCleanGPTResponse(t *testing.T) {
	generator := &ScenarioGenerator{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "With markdown code block",
			input:    "```json\n{\"test\": \"value\"}\n```",
			expected: "{\"test\": \"value\"}",
		},
		{
			name:     "Without markdown",
			input:    "{\"test\": \"value\"}",
			expected: "{\"test\": \"value\"}",
		},
		{
			name:     "With extra whitespace",
			input:    "  \n  {\"test\": \"value\"}  \n  ",
			expected: "{\"test\": \"value\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.cleanGPTResponse(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSampleRecordingsStructure tests that sample recordings are valid
func TestSampleRecordingsStructure(t *testing.T) {
	recordings := []*Recording{
		getSampleLoginRecording(),
		getSampleCheckoutRecording(),
		getSampleRegistrationRecording(),
	}

	for i, rec := range recordings {
		if rec.SessionID == "" {
			t.Errorf("Recording %d: SessionID should not be empty", i)
		}
		if rec.URL == "" {
			t.Errorf("Recording %d: URL should not be empty", i)
		}
		if len(rec.Elements) == 0 {
			t.Errorf("Recording %d: Should have elements", i)
		}
		if len(rec.Interactions) == 0 {
			t.Errorf("Recording %d: Should have interactions", i)
		}

		// Test JSON serialization
		data, err := json.Marshal(rec)
		if err != nil {
			t.Errorf("Recording %d: Failed to marshal to JSON: %v", i, err)
		}

		var decoded Recording
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Recording %d: Failed to unmarshal from JSON: %v", i, err)
		}
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
