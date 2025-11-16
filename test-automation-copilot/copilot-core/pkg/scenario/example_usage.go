package scenario

import (
	"fmt"
	"strings"
	"time"
)

// ExampleUsage demonstrates how to use the ScenarioGenerator
func ExampleUsage() {
	// 1. Create LLM client (use real LLM client in production)
	llmClient := &MockLLMClient{}

	// 2. Create scenario generator
	generator := NewScenarioGenerator(llmClient)

	// 3. Create a recording (normally this comes from browser extension)
	recording := &Recording{
		SessionID: "rec_001",
		URL:       "https://app.example.com/login",
		Title:     "Login - MyApp",
		Elements: []RecordedElement{
			{
				ID:          "email",
				Type:        "email",
				Placeholder: "Enter your email",
				Required:    true,
			},
			{
				ID:        "password",
				Type:      "password",
				Required:  true,
				MinLength: 8,
			},
			{
				ID:   "login-btn",
				Type: "submit",
				Text: "Sign In",
			},
		},
		Interactions: []RecordedInteraction{
			{Action: "input", Element: "email", Value: "test@example.com"},
			{Action: "input", Element: "password", Value: "Pass123!"},
			{Action: "click", Element: "login-btn"},
			{Action: "navigate", URL: "https://app.example.com/dashboard"},
		},
		Timestamp: time.Now(),
	}

	// 4. Generate scenarios
	request := GenerateRequest{
		Recording:  recording,
		UserIntent: "I want to test login functionality",
	}

	// Set constraints
	request.Constraints.MaxScenarios = 20
	request.Constraints.IncludeSecurity = true
	request.Constraints.PriorityFilter = "all"

	response, err := generator.GenerateScenarios(request)
	if err != nil {
		fmt.Printf("Error generating scenarios: %v\n", err)
		return
	}

	// 5. Display results
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("SCENARIO GENERATION RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Show summary
	summary := generator.GetScenarioSummary(response)
	fmt.Println(summary)

	// Show scenarios by category
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("POSITIVE SCENARIOS")
	fmt.Println(strings.Repeat("=", 60))
	for _, scenario := range response.Analysis.TestScenarios {
		if scenario.Category == "positive" {
			printScenario(scenario)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("NEGATIVE SCENARIOS")
	fmt.Println(strings.Repeat("=", 60))
	for _, scenario := range response.Analysis.TestScenarios {
		if scenario.Category == "negative" {
			printScenario(scenario)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("SECURITY SCENARIOS")
	fmt.Println(strings.Repeat("=", 60))
	for _, scenario := range response.Analysis.TestScenarios {
		if scenario.Category == "security" {
			printScenario(scenario)
		}
	}

	// 6. Filter scenarios (optional)
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("HIGH PRIORITY SCENARIOS ONLY")
	fmt.Println(strings.Repeat("=", 60))

	filter := ScenarioFilter{
		Priorities: []string{"high"},
	}
	highPriority := generator.FilterScenarios(response.Analysis.TestScenarios, filter)

	fmt.Printf("Found %d high priority scenarios:\n\n", len(highPriority))
	for _, scenario := range highPriority {
		printScenario(scenario)
	}
}

// printScenario prints a scenario in a readable format
func printScenario(scenario TestScenario) {
	fmt.Printf("📋 %s [%s]\n", scenario.Name, scenario.Priority)
	fmt.Printf("   Category: %s\n", scenario.Category)
	fmt.Printf("   Expected: %s\n", scenario.ExpectedResult)
	fmt.Printf("   Test Data:\n")
	for key, value := range scenario.TestData {
		fmt.Printf("     - %s: %v\n", key, value)
	}
	fmt.Printf("   Reasoning: %s\n", scenario.Reasoning)
	fmt.Println()
}

// ExampleWithFiltering demonstrates scenario filtering
func ExampleWithFiltering() {
	llmClient := &MockLLMClient{}
	generator := NewScenarioGenerator(llmClient)

	// Create sample recording
	recording := &Recording{
		SessionID: "rec_login_001",
		URL:       "https://app.example.com/login",
		Title:     "Login - MyApp",
		Elements: []RecordedElement{
			{ID: "email", Type: "email", Required: true},
			{ID: "password", Type: "password", Required: true, MinLength: 8},
			{ID: "login-btn", Type: "submit"},
		},
		Interactions: []RecordedInteraction{
			{Action: "input", Element: "email", Value: "test@example.com"},
			{Action: "input", Element: "password", Value: "Test1234!"},
			{Action: "click", Element: "login-btn"},
			{Action: "navigate", URL: "https://app.example.com/dashboard"},
		},
		Timestamp: time.Now(),
	}

	// Generate all scenarios
	request := GenerateRequest{
		Recording: recording,
	}

	response, err := generator.GenerateScenarios(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total scenarios generated: %d\n\n", response.TotalScenarios)

	// Example 1: Get only positive scenarios
	positiveFilter := ScenarioFilter{
		Categories: []string{"positive"},
	}
	positive := generator.FilterScenarios(response.Analysis.TestScenarios, positiveFilter)
	fmt.Printf("Positive scenarios: %d\n", len(positive))

	// Example 2: Get high priority scenarios only
	highPriorityFilter := ScenarioFilter{
		Priorities: []string{"high"},
	}
	highPriority := generator.FilterScenarios(response.Analysis.TestScenarios, highPriorityFilter)
	fmt.Printf("High priority scenarios: %d\n", len(highPriority))

	// Example 3: Get top 5 scenarios (prioritized)
	topFilter := ScenarioFilter{
		MaxCount: 5,
	}
	top5 := generator.FilterScenarios(response.Analysis.TestScenarios, topFilter)
	fmt.Printf("Top 5 scenarios: %d\n", len(top5))

	// Example 4: Get security tests only
	securityFilter := ScenarioFilter{
		Categories: []string{"security"},
		Priorities: []string{"high"},
	}
	securityTests := generator.FilterScenarios(response.Analysis.TestScenarios, securityFilter)
	fmt.Printf("High priority security tests: %d\n", len(securityTests))
}
