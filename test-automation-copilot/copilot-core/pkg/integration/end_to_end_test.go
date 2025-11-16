package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/copilot-core/pkg/generator"
	"github.com/yourusername/copilot-core/pkg/scenario"
)

// TestEndToEndFlow tests the complete flow: Recording → Scenarios → Code
func TestEndToEndFlow(t *testing.T) {
	// Step 1: Create a browser recording (normally from browser extension)
	recording := createSampleLoginRecording()

	// Step 2: Generate scenarios using GPT (mock)
	llmClient := &scenario.MockLLMClient{}
	scenarioGen := scenario.NewScenarioGenerator(llmClient)

	request := scenario.GenerateRequest{
		Recording:  recording,
		UserIntent: "I want to test login functionality",
	}
	request.Constraints.IncludeSecurity = true

	response, err := scenarioGen.GenerateScenarios(request)
	if err != nil {
		t.Fatalf("Failed to generate scenarios: %v", err)
	}

	t.Logf("Generated %d scenarios", response.TotalScenarios)

	// Step 3: Generate code from scenarios
	codeGen := generator.NewScenarioBasedGenerator("")

	genRequest := generator.BuildGenerationRequest(
		recording,
		response.Analysis,
		response.Analysis.TestScenarios,
	)

	result, err := codeGen.GenerateFromScenarios(genRequest)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Verify results
	t.Logf("\n%s", result.Summary)

	// Verify Page Objects
	if len(result.PageObjects) == 0 {
		t.Error("Expected at least one Page Object to be generated")
	} else {
		t.Logf("✅ Generated Page Objects:")
		for _, po := range result.PageObjects {
			t.Logf("   - %s (%d lines)", po.FileName, countLines(po.Content))
			if po.Content == "" {
				t.Errorf("Page Object %s has empty content", po.FileName)
			}
		}
	}

	// Verify Test Classes
	if len(result.TestClasses) == 0 {
		t.Error("Expected at least one Test Class to be generated")
	} else {
		t.Logf("✅ Generated Test Classes:")
		for _, test := range result.TestClasses {
			t.Logf("   - %s (%d lines)", test.FileName, countLines(test.Content))
			if test.Content == "" {
				t.Errorf("Test Class %s has empty content", test.FileName)
			}
			// Verify DataProvider is present
			if !containsString(test.Content, "@DataProvider") {
				t.Error("Test class should contain @DataProvider annotation")
			}
		}
	}

	// Verify Feature Files
	if len(result.FeatureFiles) == 0 {
		t.Error("Expected at least one Feature File to be generated")
	} else {
		t.Logf("✅ Generated Feature Files:")
		for _, feature := range result.FeatureFiles {
			t.Logf("   - %s (%d lines)", feature.FileName, countLines(feature.Content))
			if feature.Content == "" {
				t.Errorf("Feature File %s has empty content", feature.FileName)
			}
			// Verify Gherkin syntax
			if !containsString(feature.Content, "Feature:") {
				t.Error("Feature file should contain 'Feature:' keyword")
			}
			if !containsString(feature.Content, "Scenario") {
				t.Error("Feature file should contain 'Scenario' keyword")
			}
		}
	}

	// Verify Step Definitions
	if len(result.StepDefinitions) == 0 {
		t.Error("Expected at least one Step Definition to be generated")
	} else {
		t.Logf("✅ Generated Step Definitions:")
		for _, step := range result.StepDefinitions {
			t.Logf("   - %s (%d lines)", step.FileName, countLines(step.Content))
			if step.Content == "" {
				t.Errorf("Step Definition %s has empty content", step.FileName)
			}
			// Verify Cucumber annotations
			if !containsString(step.Content, "@Given") {
				t.Error("Step definition should contain @Given annotation")
			}
		}
	}

	// Output generated code for inspection
	t.Log("\n" + separator(60))
	t.Log("GENERATED PAGE OBJECT")
	t.Log(separator(60))
	if len(result.PageObjects) > 0 {
		t.Log(result.PageObjects[0].Content)
	}

	t.Log("\n" + separator(60))
	t.Log("GENERATED TEST CLASS")
	t.Log(separator(60))
	if len(result.TestClasses) > 0 {
		t.Log(result.TestClasses[0].Content)
	}

	t.Log("\n" + separator(60))
	t.Log("GENERATED FEATURE FILE")
	t.Log(separator(60))
	if len(result.FeatureFiles) > 0 {
		t.Log(result.FeatureFiles[0].Content)
	}
}

// TestSelectiveGeneration tests generating only specific components
func TestSelectiveGeneration(t *testing.T) {
	recording := createSampleLoginRecording()
	llmClient := &scenario.MockLLMClient{}
	scenarioGen := scenario.NewScenarioGenerator(llmClient)

	request := scenario.GenerateRequest{
		Recording: recording,
	}

	response, err := scenarioGen.GenerateScenarios(request)
	if err != nil {
		t.Fatalf("Failed to generate scenarios: %v", err)
	}

	codeGen := generator.NewScenarioBasedGenerator("")

	// Test: Generate only Page Objects
	t.Run("OnlyPageObjects", func(t *testing.T) {
		genRequest := generator.BuildGenerationRequest(
			recording,
			response.Analysis,
			response.Analysis.TestScenarios,
		)
		genRequest.GeneratePageObject = true
		genRequest.GenerateTests = false
		genRequest.GenerateFeatures = false
		genRequest.GenerateSteps = false

		result, err := codeGen.GenerateFromScenarios(genRequest)
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}

		if len(result.PageObjects) == 0 {
			t.Error("Expected Page Objects to be generated")
		}
		if len(result.TestClasses) != 0 {
			t.Error("Expected NO Test Classes to be generated")
		}
		if len(result.FeatureFiles) != 0 {
			t.Error("Expected NO Feature Files to be generated")
		}
		if len(result.StepDefinitions) != 0 {
			t.Error("Expected NO Step Definitions to be generated")
		}
	})

	// Test: Generate only Tests
	t.Run("OnlyTests", func(t *testing.T) {
		genRequest := generator.BuildGenerationRequest(
			recording,
			response.Analysis,
			response.Analysis.TestScenarios,
		)
		genRequest.GeneratePageObject = false
		genRequest.GenerateTests = true
		genRequest.GenerateFeatures = false
		genRequest.GenerateSteps = false

		result, err := codeGen.GenerateFromScenarios(genRequest)
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}

		if len(result.PageObjects) != 0 {
			t.Error("Expected NO Page Objects to be generated")
		}
		if len(result.TestClasses) == 0 {
			t.Error("Expected Test Classes to be generated")
		}
	})
}

// TestMultiplePageObjects tests generating multiple page objects
func TestMultiplePageObjects(t *testing.T) {
	recording := createSampleLoginRecording()
	llmClient := &scenario.MockLLMClient{}
	scenarioGen := scenario.NewScenarioGenerator(llmClient)

	request := scenario.GenerateRequest{
		Recording: recording,
	}

	response, err := scenarioGen.GenerateScenarios(request)
	if err != nil {
		t.Fatalf("Failed to generate scenarios: %v", err)
	}

	// Manually add a second page object to the analysis
	response.Analysis.RecommendedPageObjects = append(
		response.Analysis.RecommendedPageObjects,
		scenario.PageObjectSpec{
			Name:     "DashboardPage",
			Elements: []string{"welcomeMessage", "logoutButton"},
			Methods:  []string{"isLoggedIn", "logout"},
		},
	)

	codeGen := generator.NewScenarioBasedGenerator("")
	genRequest := generator.BuildGenerationRequest(
		recording,
		response.Analysis,
		response.Analysis.TestScenarios,
	)

	result, err := codeGen.GenerateFromScenarios(genRequest)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if len(result.PageObjects) < 2 {
		t.Errorf("Expected at least 2 Page Objects, got %d", len(result.PageObjects))
	}

	t.Logf("Generated %d Page Objects", len(result.PageObjects))
	for _, po := range result.PageObjects {
		t.Logf("  - %s", po.FileName)
	}
}

// Helper functions

func createSampleLoginRecording() *scenario.Recording {
	return &scenario.Recording{
		SessionID: "test_rec_001",
		URL:       "https://app.example.com/login",
		Title:     "Login - MyApp",
		Elements: []scenario.RecordedElement{
			{
				ID:          "email",
				Type:        "email",
				Label:       "Email Address",
				Placeholder: "Enter your email",
				Required:    true,
			},
			{
				ID:        "password",
				Type:      "password",
				Label:     "Password",
				Required:  true,
				MinLength: 8,
			},
			{
				ID:    "remember-me",
				Type:  "checkbox",
				Label: "Remember me",
			},
			{
				ID:   "login-btn",
				Type: "submit",
				Text: "Sign In",
			},
		},
		Interactions: []scenario.RecordedInteraction{
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

func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count + 1
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func separator(length int) string {
	sep := ""
	for i := 0; i < length; i++ {
		sep += "="
	}
	return sep
}

// BenchmarkEndToEndFlow benchmarks the complete flow
func BenchmarkEndToEndFlow(b *testing.B) {
	recording := createSampleLoginRecording()
	llmClient := &scenario.MockLLMClient{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scenarioGen := scenario.NewScenarioGenerator(llmClient)
		request := scenario.GenerateRequest{
			Recording: recording,
		}

		response, err := scenarioGen.GenerateScenarios(request)
		if err != nil {
			b.Fatalf("Failed to generate scenarios: %v", err)
		}

		codeGen := generator.NewScenarioBasedGenerator("")
		genRequest := generator.BuildGenerationRequest(
			recording,
			response.Analysis,
			response.Analysis.TestScenarios,
		)

		_, err = codeGen.GenerateFromScenarios(genRequest)
		if err != nil {
			b.Fatalf("Failed to generate code: %v", err)
		}
	}
}

// Example demonstrates the complete usage
func Example() {
	// 1. Create recording
	recording := createSampleLoginRecording()

	// 2. Generate scenarios
	llmClient := &scenario.MockLLMClient{}
	scenarioGen := scenario.NewScenarioGenerator(llmClient)

	request := scenario.GenerateRequest{
		Recording:  recording,
		UserIntent: "I want to test login",
	}

	response, _ := scenarioGen.GenerateScenarios(request)

	// 3. Generate code
	codeGen := generator.NewScenarioBasedGenerator("")
	genRequest := generator.BuildGenerationRequest(
		recording,
		response.Analysis,
		response.Analysis.TestScenarios,
	)

	result, _ := codeGen.GenerateFromScenarios(genRequest)

	// 4. Display results
	fmt.Printf("Generated %d files\n", len(result.PageObjects)+len(result.TestClasses))
}
