package scenario

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ScenarioGenerator generates test scenarios using GPT analysis
type ScenarioGenerator struct {
	llmClient LLMClient
}

// NewScenarioGenerator creates a new scenario generator
func NewScenarioGenerator(llmClient LLMClient) *ScenarioGenerator {
	return &ScenarioGenerator{
		llmClient: llmClient,
	}
}

// GenerateScenarios analyzes a recording and generates comprehensive test scenarios
func (g *ScenarioGenerator) GenerateScenarios(req GenerateRequest) (*GenerateResponse, error) {
	startTime := time.Now()

	// Build GPT analysis prompt
	prompt := g.buildAnalysisPrompt(req)

	// Call GPT for intelligent analysis
	responseText, err := g.llmClient.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("GPT analysis failed: %w", err)
	}

	// Clean response (remove markdown formatting if present)
	cleanedResponse := g.cleanGPTResponse(responseText)

	// Parse GPT response
	var analysis IntelligentAnalysis
	if err := json.Unmarshal([]byte(cleanedResponse), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse GPT response: %w\nResponse: %s", err, cleanedResponse)
	}

	// Apply constraints/filters
	if req.Constraints.MaxScenarios > 0 {
		analysis.TestScenarios = g.limitScenarios(analysis.TestScenarios, req.Constraints.MaxScenarios)
	}

	if !req.Constraints.IncludeSecurity {
		analysis.TestScenarios = g.filterOutSecurity(analysis.TestScenarios)
	}

	if req.Constraints.PriorityFilter != "" && req.Constraints.PriorityFilter != "all" {
		analysis.TestScenarios = g.filterByPriority(analysis.TestScenarios, req.Constraints.PriorityFilter)
	}

	// Count scenarios by category
	response := &GenerateResponse{
		Analysis:       &analysis,
		TotalScenarios: len(analysis.TestScenarios),
		ProcessingTime: time.Since(startTime),
	}

	for _, scenario := range analysis.TestScenarios {
		switch scenario.Category {
		case "positive":
			response.PositiveCount++
		case "negative":
			response.NegativeCount++
		case "edgeCase":
			response.EdgeCaseCount++
		case "security":
			response.SecurityTestCount++
		}
	}

	return response, nil
}

// FilterScenarios filters scenarios based on criteria
func (g *ScenarioGenerator) FilterScenarios(scenarios []TestScenario, filter ScenarioFilter) []TestScenario {
	filtered := []TestScenario{}

	for _, scenario := range scenarios {
		// Category filter
		if len(filter.Categories) > 0 && !contains(filter.Categories, scenario.Category) {
			continue
		}

		// Priority filter
		if len(filter.Priorities) > 0 && !contains(filter.Priorities, scenario.Priority) {
			continue
		}

		filtered = append(filtered, scenario)
	}

	// Apply count limits
	if filter.MaxCount > 0 && len(filtered) > filter.MaxCount {
		filtered = filtered[:filter.MaxCount]
	}

	return filtered
}

// buildAnalysisPrompt creates the GPT prompt for intelligent analysis
func (g *ScenarioGenerator) buildAnalysisPrompt(req GenerateRequest) string {
	recording := req.Recording

	var prompt strings.Builder

	prompt.WriteString(`You are a senior QA engineer analyzing a recorded web page interaction.

RECORDING DATA:
URL: `)
	prompt.WriteString(recording.URL)
	prompt.WriteString("\nTitle: ")
	prompt.WriteString(recording.Title)
	prompt.WriteString("\n\nElements Found:\n")

	// List all elements with their attributes
	for i, elem := range recording.Elements {
		prompt.WriteString(fmt.Sprintf("%d. ", i+1))

		if elem.Label != "" {
			prompt.WriteString(elem.Label)
		} else if elem.Text != "" {
			prompt.WriteString(elem.Text)
		} else {
			prompt.WriteString(elem.Type)
		}

		prompt.WriteString(fmt.Sprintf(" (id: %s, type: %s", elem.ID, elem.Type))

		if elem.Required {
			prompt.WriteString(", required: true")
		}
		if elem.MinLength > 0 {
			prompt.WriteString(fmt.Sprintf(", minLength: %d", elem.MinLength))
		}
		if elem.MaxLength > 0 {
			prompt.WriteString(fmt.Sprintf(", maxLength: %d", elem.MaxLength))
		}
		if elem.Pattern != "" {
			prompt.WriteString(fmt.Sprintf(", pattern: %s", elem.Pattern))
		}
		if elem.Placeholder != "" {
			prompt.WriteString(fmt.Sprintf(", placeholder: \"%s\"", elem.Placeholder))
		}
		if elem.Href != "" {
			prompt.WriteString(fmt.Sprintf(", href: %s", elem.Href))
		}

		prompt.WriteString(")\n")
	}

	// List user interactions
	prompt.WriteString("\nUser Interaction Flow:\n")
	for i, interaction := range recording.Interactions {
		switch interaction.Action {
		case "input":
			prompt.WriteString(fmt.Sprintf("%d. Entered %s: %s\n", i+1, interaction.Element, interaction.Value))
		case "click":
			prompt.WriteString(fmt.Sprintf("%d. Clicked %s\n", i+1, interaction.Element))
		case "navigate":
			prompt.WriteString(fmt.Sprintf("%d. Navigated to: %s\n", i+1, interaction.URL))
		case "select":
			prompt.WriteString(fmt.Sprintf("%d. Selected %s from %s\n", i+1, interaction.Value, interaction.Element))
		}
	}

	// Add user intent if provided
	if req.UserIntent != "" {
		prompt.WriteString("\nUser Intent: ")
		prompt.WriteString(req.UserIntent)
		prompt.WriteString("\n")
	}

	// Add app context if provided
	if req.AppContext != "" {
		prompt.WriteString("\nApplication Context: ")
		prompt.WriteString(req.AppContext)
		prompt.WriteString("\n")
	}

	// Add the main task instructions
	prompt.WriteString(`
YOUR TASK:
Based ONLY on this recording data, infer and generate a comprehensive test analysis.

1. PAGE TYPE & PURPOSE
   - Identify the page type (login, registration, checkout, search, profile, settings, etc.)
   - Describe the primary functionality
   - Provide confidence level (0.0 to 1.0)

2. VALIDATION RULES
   - Infer validation rules from element attributes (required, type, pattern, minLength, maxLength)
   - Consider HTML5 input types (email, url, number, tel, date, etc.)
   - Generate examples of valid and invalid data

3. SUCCESS CRITERIA
   - Determine how to validate successful completion
   - Identify expected URL changes or navigation patterns
   - List elements that should appear on success

4. FAILURE CRITERIA
   - Determine how to detect failures
   - Identify expected error messages or states
   - List elements that indicate failure

5. TEST SCENARIOS
   Generate at least 12-20 comprehensive test scenarios covering:

   a) POSITIVE scenarios (happy path):
      - Valid inputs that should succeed
      - Different valid combinations
      - Optional field variations

   b) NEGATIVE scenarios (validation):
      - Empty required fields
      - Invalid formats (wrong email, invalid phone, etc.)
      - Boundary violations (too short, too long)
      - Type mismatches

   c) EDGE CASES:
      - Boundary values (min/max lengths)
      - Special characters
      - Unicode characters
      - Very long inputs
      - Whitespace handling
      - Case sensitivity

   d) SECURITY tests:
      - SQL injection attempts
      - XSS (Cross-Site Scripting)
      - HTML injection
      - Script injection
      - Path traversal
      - Command injection

   For EACH scenario provide:
   - category: "positive" | "negative" | "edgeCase" | "security"
   - name: Descriptive name
   - priority: "high" | "medium" | "low"
   - testData: Actual input values for each field
   - expectedResult: "success" | "failure"
   - reasoning: Why this scenario is important

6. RELATED FEATURES
   - Detect links to related functionality (forgot password, registration, help, etc.)
   - Identify potential additional test flows

7. PAGE OBJECTS
   - Recommend Page Object classes needed for this test
   - List elements that should be in each Page Object
   - Suggest methods for each Page Object

OUTPUT FORMAT:
Return ONLY valid JSON matching this exact schema (no markdown, no code blocks):

{
  "pageAnalysis": {
    "type": "string",
    "purpose": "string",
    "confidence": 0.95
  },
  "inferredValidationRules": {
    "fieldName": {
      "required": true,
      "format": "email",
      "minLength": 5,
      "maxLength": 100,
      "pattern": "regex_pattern",
      "reasoning": "Inferred from type='email' and required attribute"
    }
  },
  "successCriteria": {
    "navigation": {
      "expectedUrl": "https://example.com/dashboard",
      "reasoning": "User navigated here after interaction"
    },
    "elementsToCheck": [
      {
        "selector": "id=welcome-message",
        "expectedState": "visible",
        "text": "Welcome",
        "reasoning": "Indicates successful login"
      }
    ]
  },
  "failureCriteria": {
    "navigation": {
      "expectedUrl": "https://example.com/login",
      "reasoning": "Stays on same page on failure"
    },
    "elementsToCheck": [
      {
        "selector": "class=error-message",
        "expectedState": "visible",
        "reasoning": "Error message should appear"
      }
    ]
  },
  "testScenarios": [
    {
      "category": "positive",
      "name": "Valid login with correct credentials",
      "priority": "high",
      "testData": {
        "username": "user@example.com",
        "password": "ValidPass123!"
      },
      "expectedResult": "success",
      "reasoning": "Primary happy path - must work"
    }
  ],
  "relatedFeatures": {
    "forgotPassword": {
      "detected": true,
      "link": "/reset-password",
      "reasoning": "Forgot password link found in page"
    }
  },
  "recommendedPageObjects": [
    {
      "name": "LoginPage",
      "elements": ["usernameField", "passwordField", "loginButton"],
      "methods": ["login", "isLoginButtonVisible", "getErrorMessage"],
      "reasoning": "Main page under test"
    }
  ]
}

IMPORTANT RULES:
- Output ONLY the JSON object, no markdown formatting, no code blocks
- Be thorough - aim for 15-20 test scenarios minimum
- Think like a senior QA engineer with security awareness
- Base ALL inferences on the provided recording data
- Provide realistic, usable test data
- Consider both functional and non-functional testing
`)

	return prompt.String()
}

// cleanGPTResponse removes markdown formatting from GPT response
func (g *ScenarioGenerator) cleanGPTResponse(response string) string {
	// Remove markdown code blocks
	cleaned := strings.TrimSpace(response)

	// Remove ```json and ```
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")

	return strings.TrimSpace(cleaned)
}

// limitScenarios limits the number of scenarios
func (g *ScenarioGenerator) limitScenarios(scenarios []TestScenario, max int) []TestScenario {
	if len(scenarios) <= max {
		return scenarios
	}

	// Prioritize: high priority first, then medium, then low
	high := []TestScenario{}
	medium := []TestScenario{}
	low := []TestScenario{}

	for _, s := range scenarios {
		switch s.Priority {
		case "high":
			high = append(high, s)
		case "medium":
			medium = append(medium, s)
		case "low":
			low = append(low, s)
		}
	}

	result := []TestScenario{}
	result = append(result, high...)

	if len(result) < max {
		remaining := max - len(result)
		if len(medium) <= remaining {
			result = append(result, medium...)
		} else {
			result = append(result, medium[:remaining]...)
		}
	}

	if len(result) < max {
		remaining := max - len(result)
		if len(low) <= remaining {
			result = append(result, low...)
		} else {
			result = append(result, low[:remaining]...)
		}
	}

	return result[:max]
}

// filterOutSecurity removes security test scenarios
func (g *ScenarioGenerator) filterOutSecurity(scenarios []TestScenario) []TestScenario {
	filtered := []TestScenario{}
	for _, s := range scenarios {
		if s.Category != "security" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// filterByPriority filters scenarios by priority level
func (g *ScenarioGenerator) filterByPriority(scenarios []TestScenario, priority string) []TestScenario {
	filtered := []TestScenario{}
	for _, s := range scenarios {
		if s.Priority == priority {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// GetScenarioSummary returns a human-readable summary of scenarios
func (g *ScenarioGenerator) GetScenarioSummary(response *GenerateResponse) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Generated %d test scenarios in %s\n\n",
		response.TotalScenarios, response.ProcessingTime.Round(time.Millisecond)))

	summary.WriteString(fmt.Sprintf("Page Type: %s (%.0f%% confidence)\n",
		response.Analysis.PageAnalysis.Type,
		response.Analysis.PageAnalysis.Confidence*100))

	summary.WriteString(fmt.Sprintf("Purpose: %s\n\n", response.Analysis.PageAnalysis.Purpose))

	summary.WriteString("Scenario Breakdown:\n")
	summary.WriteString(fmt.Sprintf("  ✓ Positive: %d\n", response.PositiveCount))
	summary.WriteString(fmt.Sprintf("  ✗ Negative: %d\n", response.NegativeCount))
	summary.WriteString(fmt.Sprintf("  ⚠ Edge Cases: %d\n", response.EdgeCaseCount))
	summary.WriteString(fmt.Sprintf("  🔒 Security: %d\n\n", response.SecurityTestCount))

	// Show validation rules
	if len(response.Analysis.ValidationRules) > 0 {
		summary.WriteString("Inferred Validation Rules:\n")
		for field, rule := range response.Analysis.ValidationRules {
			summary.WriteString(fmt.Sprintf("  • %s: ", field))
			if rule.Required {
				summary.WriteString("required, ")
			}
			if rule.Format != "" {
				summary.WriteString(fmt.Sprintf("format=%s, ", rule.Format))
			}
			if rule.MinLength > 0 {
				summary.WriteString(fmt.Sprintf("min=%d, ", rule.MinLength))
			}
			if rule.MaxLength > 0 {
				summary.WriteString(fmt.Sprintf("max=%d, ", rule.MaxLength))
			}
			summary.WriteString("\n")
		}
		summary.WriteString("\n")
	}

	// Show recommended page objects
	if len(response.Analysis.RecommendedPageObjects) > 0 {
		summary.WriteString("Recommended Page Objects:\n")
		for _, po := range response.Analysis.RecommendedPageObjects {
			summary.WriteString(fmt.Sprintf("  • %s (%d elements, %d methods)\n",
				po.Name, len(po.Elements), len(po.Methods)))
		}
	}

	return summary.String()
}
