package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/copilot-core/pkg/llm"
	"github.com/yourusername/copilot-core/pkg/prompts"
)

type CodeGenerator struct {
	llmClient     *llm.Client
	promptBuilder *prompts.PromptBuilder
	projectRules  *prompts.ProjectRules
	workspacePath string
}

type GenerateRequest struct {
	Type          string                   `json:"type"` // "pageObject", "test", "utility", "feature", "stepDefinition"
	Specification string                   `json:"specification"`
	Elements      []ElementData            `json:"elements,omitempty"`
	PageData      *RecordedPageData        `json:"pageData,omitempty"`
	Context       map[string]interface{}   `json:"context,omitempty"`
	Framework     string                   `json:"framework"` // "selenium-java", etc.
	ExistingCode  []map[string]interface{} `json:"existingCode,omitempty"`
	Interactions  []InteractionData        `json:"interactions,omitempty"` // For BDD generation
}

type ElementData struct {
	Name               string            `json:"name"`
	LocatorType        string            `json:"locatorType"`
	LocatorValue       string            `json:"locatorValue"`
	InteractionType    string            `json:"interactionType"`
	Text               string            `json:"text,omitempty"`
	Placeholder        string            `json:"placeholder,omitempty"`
	RecommendedLocator map[string]string `json:"recommendedLocator"`
}

type RecordedPageData struct {
	URL          string          `json:"url"`
	Title        string          `json:"title"`
	ClassName    string          `json:"className"`
	Elements     []ElementData   `json:"elements"`
	Interactions []InteractionData `json:"interactions"`
}

type InteractionData struct {
	Type      string `json:"type"`
	Element   string `json:"element"`
	Value     string `json:"value,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type GeneratedCode struct {
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
	Language string `json:"language"`
	Type     string `json:"type"`
}

func NewCodeGenerator() *CodeGenerator {
	return NewCodeGeneratorWithWorkspace("")
}

func NewCodeGeneratorWithWorkspace(workspacePath string) *CodeGenerator {
	// Load project rules from .testcopilot file
	rules, err := prompts.LoadProjectRules(workspacePath)
	if err != nil {
		// Use default rules if config not found
		rules = prompts.GetDefaultRules()
	}

	return &CodeGenerator{
		llmClient:     llm.NewClient(),
		promptBuilder: prompts.NewPromptBuilder(),
		projectRules:  rules,
		workspacePath: workspacePath,
	}
}

func (g *CodeGenerator) Generate(req GenerateRequest) (*GeneratedCode, error) {
	switch req.Type {
	case "pageObject":
		return g.generatePageObject(req)
	case "test":
		return g.generateTestCase(req)
	case "utility":
		return g.generateUtility(req)
	case "feature":
		return g.generateFeatureFile(req)
	case "stepDefinition":
		return g.generateStepDefinitions(req)
	default:
		return nil, fmt.Errorf("unknown generation type: %s", req.Type)
	}
}

func (g *CodeGenerator) generatePageObject(req GenerateRequest) (*GeneratedCode, error) {
	// Prepare elements for prompt
	var elementsForPrompt []map[string]string
	for _, elem := range req.Elements {
		elementsForPrompt = append(elementsForPrompt, map[string]string{
			"name":         elem.Name,
			"locatorType":  elem.LocatorType,
			"locatorValue": elem.LocatorValue,
		})
	}

	// Set framework context - use detected test runner from project rules
	testRunner := "testng"
	if g.projectRules != nil && g.projectRules.CustomInstructions != "" {
		// Could parse test runner from custom instructions, but for now use default
		testRunner = "testng"
	}
	g.promptBuilder.SetFramework(req.Framework, testRunner)

	// Add existing code for style matching
	if len(req.ExistingCode) > 0 {
		for _, code := range req.ExistingCode {
			g.promptBuilder.AddContext(code)
		}
	}

	// Build prompt
	var spec string
	if req.PageData != nil {
		spec = fmt.Sprintf("Page: %s (%s)", req.PageData.Title, req.PageData.URL)
	} else {
		spec = req.Specification
	}

	basePrompt := g.promptBuilder.BuildPageObjectPrompt(spec, elementsForPrompt)

	// Apply project-specific rules from .testcopilot config
	prompt := g.promptBuilder.ApplyProjectRules(basePrompt, g.projectRules)

	// Call LLM
	response, err := g.llmClient.Generate(prompt, map[string]interface{}{
		"temperature":   0.3, // Lower for code generation
		"max_tokens":    2000,
		"system_prompt": prompts.GetSystemPrompt("code_generator"),
	})

	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Extract code from response
	code := g.extractCode(response)

	// Determine file name
	fileName := g.inferFileName(req, code)
	filePath := g.inferFilePath(req, fileName)

	return &GeneratedCode{
		FileName: fileName,
		FilePath: filePath,
		Content:  code,
		Language: "java",
		Type:     "pageObject",
	}, nil
}

func (g *CodeGenerator) generateTestCase(req GenerateRequest) (*GeneratedCode, error) {
	// Get available page objects from context
	var pageObjects []map[string]interface{}
	if ctx, ok := req.Context["pageObjects"].([]interface{}); ok {
		for _, po := range ctx {
			if poMap, ok := po.(map[string]interface{}); ok {
				pageObjects = append(pageObjects, poMap)
			}
		}
	}

	// Set framework - use detected test runner
	testRunner := "testng"
	if g.projectRules != nil {
		testRunner = "testng" // TODO: get from analysis
	}
	g.promptBuilder.SetFramework(req.Framework, testRunner)

	// Add existing code
	if len(req.ExistingCode) > 0 {
		for _, code := range req.ExistingCode {
			g.promptBuilder.AddContext(code)
		}
	}

	// Build prompt
	basePrompt := g.promptBuilder.BuildTestCasePrompt(req.Specification, pageObjects)

	// Apply project-specific rules
	prompt := g.promptBuilder.ApplyProjectRules(basePrompt, g.projectRules)

	// Call LLM
	response, err := g.llmClient.Generate(prompt, map[string]interface{}{
		"temperature":   0.3,
		"max_tokens":    2000,
		"system_prompt": prompts.GetSystemPrompt("code_generator"),
	})

	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Extract code
	code := g.extractCode(response)

	// Determine file name
	fileName := g.inferFileName(req, code)
	filePath := g.inferFilePath(req, fileName)

	return &GeneratedCode{
		FileName: fileName,
		FilePath: filePath,
		Content:  code,
		Language: "java",
		Type:     "test",
	}, nil
}

func (g *CodeGenerator) generateUtility(req GenerateRequest) (*GeneratedCode, error) {
	// Similar to test case but for utilities
	return nil, fmt.Errorf("utility generation not yet implemented")
}

func (g *CodeGenerator) GenerateFromRecordedSession(sessionData map[string]interface{}) ([]*GeneratedCode, error) {
	return g.GenerateFromRecordedSessionWithOptions(sessionData, map[string]bool{
		"pageObjects":     true,
		"tests":           false,
		"features":        false,
		"stepDefinitions": false,
	})
}

// GenerateFromRecordedSessionWithOptions generates multiple file types from a session (Composer-style)
func (g *CodeGenerator) GenerateFromRecordedSessionWithOptions(
	sessionData map[string]interface{},
	options map[string]bool,
) ([]*GeneratedCode, error) {
	var generatedFiles []*GeneratedCode

	// Extract pages from session
	pages, ok := sessionData["pages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid session data: no pages found")
	}

	framework := "selenium-java" // TODO: detect from workspace
	if fw, ok := sessionData["framework"].(string); ok {
		framework = fw
	}

	var pageObjects []*GeneratedCode
	var allInteractions []InteractionData

	// 1. Generate page objects for each page
	if options["pageObjects"] {
		fmt.Println("📄 Generating Page Objects...")
		for _, pageInterface := range pages {
			pageMap, ok := pageInterface.(map[string]interface{})
			if !ok {
				continue
			}

			// Convert to RecordedPageData
			pageData := g.convertToPageData(pageMap)

			// Generate page object
			req := GenerateRequest{
				Type:      "pageObject",
				PageData:  pageData,
				Framework: framework,
			}

			generated, err := g.generatePageObject(req)
			if err != nil {
				fmt.Printf("  ⚠️  Failed to generate page object for %s: %v\n", pageData.Title, err)
				continue
			}

			fmt.Printf("  ✓ Generated %s\n", generated.FileName)
			pageObjects = append(pageObjects, generated)
			generatedFiles = append(generatedFiles, generated)

			// Collect interactions
			allInteractions = append(allInteractions, pageData.Interactions...)
		}
	}

	// 2. Generate test cases (uses generated page objects)
	if options["tests"] && len(pageObjects) > 0 {
		fmt.Println("\n📝 Generating Test Cases...")
		// Build context with page objects
		var pageObjectsContext []map[string]interface{}
		for _, po := range pageObjects {
			pageObjectsContext = append(pageObjectsContext, map[string]interface{}{
				"className": strings.TrimSuffix(po.FileName, ".java"),
				"methods": []map[string]interface{}{
					{"name": "performAction"},
				},
			})
		}

		// Generate one test that uses all page objects
		testReq := GenerateRequest{
			Type:          "test",
			Specification: "Test user flow through recorded session",
			Framework:     framework,
			Context: map[string]interface{}{
				"pageObjects": pageObjectsContext,
			},
		}

		testCode, err := g.generateTestCase(testReq)
		if err == nil {
			fmt.Printf("  ✓ Generated %s\n", testCode.FileName)
			generatedFiles = append(generatedFiles, testCode)
		} else {
			fmt.Printf("  ⚠️  Failed to generate test: %v\n", err)
		}
	}

	// 3. Generate feature files (Gherkin/BDD)
	if options["features"] && len(allInteractions) > 0 {
		fmt.Println("\n🥒 Generating Feature Files...")
		featureReq := GenerateRequest{
			Type:          "feature",
			Specification: "User interaction flow",
			Interactions:  allInteractions,
			Framework:     framework,
		}

		featureCode, err := g.generateFeatureFile(featureReq)
		if err == nil {
			fmt.Printf("  ✓ Generated %s\n", featureCode.FileName)
			generatedFiles = append(generatedFiles, featureCode)
		} else {
			fmt.Printf("  ⚠️  Failed to generate feature: %v\n", err)
		}
	}

	// 4. Generate step definitions (Cucumber)
	if options["stepDefinitions"] && len(allInteractions) > 0 {
		fmt.Println("\n🎯 Generating Step Definitions...")
		// Use first page data for step definitions
		if len(pages) > 0 {
			pageMap := pages[0].(map[string]interface{})
			pageData := g.convertToPageData(pageMap)

			stepReq := GenerateRequest{
				Type:          "stepDefinition",
				Specification: pageData.Title,
				PageData:      pageData,
				Framework:     framework,
			}

			stepCode, err := g.generateStepDefinitions(stepReq)
			if err == nil {
				fmt.Printf("  ✓ Generated %s\n", stepCode.FileName)
				generatedFiles = append(generatedFiles, stepCode)
			} else {
				fmt.Printf("  ⚠️  Failed to generate step definitions: %v\n", err)
			}
		}
	}

	fmt.Printf("\n🎉 Generated %d files total\n", len(generatedFiles))
	return generatedFiles, nil
}

func (g *CodeGenerator) convertToPageData(pageMap map[string]interface{}) *RecordedPageData {
	pageData := &RecordedPageData{}

	if url, ok := pageMap["url"].(string); ok {
		pageData.URL = url
	}

	if title, ok := pageMap["title"].(string); ok {
		pageData.Title = title
	}

	// Infer class name from title or URL
	pageData.ClassName = g.inferClassName(pageData.Title, pageData.URL)

	// Extract elements
	if elements, ok := pageMap["elements"].([]interface{}); ok {
		for _, elemInterface := range elements {
			if elemMap, ok := elemInterface.(map[string]interface{}); ok {
				element := ElementData{}

				if text, ok := elemMap["text"].(string); ok {
					element.Text = text
				}

				// Get recommended locator
				if recLocator, ok := elemMap["recommendedLocator"].(map[string]interface{}); ok {
					if locType, ok := recLocator["type"].(string); ok {
						element.LocatorType = locType
					}
					if locValue, ok := recLocator["value"].(string); ok {
						element.LocatorValue = locValue
					}
				}

				// Infer element name
				element.Name = g.inferElementName(elemMap)

				// Get interaction type
				if intType, ok := elemMap["interactionType"].(string); ok {
					element.InteractionType = intType
				}

				pageData.Elements = append(pageData.Elements, element)
			}
		}
	}

	// Extract interactions
	if interactions, ok := pageMap["interactions"].([]interface{}); ok {
		for _, intInterface := range interactions {
			if intMap, ok := intInterface.(map[string]interface{}); ok {
				interaction := InteractionData{}

				if intType, ok := intMap["type"].(string); ok {
					interaction.Type = intType
				}

				if timestamp, ok := intMap["timestamp"].(float64); ok {
					interaction.Timestamp = int64(timestamp)
				}

				pageData.Interactions = append(pageData.Interactions, interaction)
			}
		}
	}

	return pageData
}

func (g *CodeGenerator) extractCode(response string) string {
	// Extract code from markdown code blocks
	codeBlocks := prompts.ExtractCodeBlocks(response)

	if len(codeBlocks) > 0 {
		// Return the first (or largest) code block
		return strings.TrimSpace(codeBlocks[0])
	}

	// If no code blocks, return the whole response (LLM might have returned just code)
	return strings.TrimSpace(response)
}

func (g *CodeGenerator) inferFileName(req GenerateRequest, code string) string {
	// Try to extract class name from code
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "public class ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "class" && i+1 < len(parts) {
					className := strings.TrimSuffix(parts[i+1], "{")
					return className + ".java"
				}
			}
		}
	}

	// Fallback: use specification or page title
	if req.PageData != nil && req.PageData.ClassName != "" {
		return req.PageData.ClassName + ".java"
	}

	// Last resort
	return fmt.Sprintf("Generated_%d.java", time.Now().Unix())
}

func (g *CodeGenerator) inferFilePath(req GenerateRequest, fileName string) string {
	// Determine appropriate directory based on type
	switch req.Type {
	case "pageObject":
		return "src/test/java/pages/" + fileName
	case "test":
		return "src/test/java/tests/" + fileName
	case "utility":
		return "src/test/java/utils/" + fileName
	case "stepDefinition":
		return "src/test/java/steps/" + fileName
	case "feature":
		return "src/test/resources/features/" + fileName
	default:
		return "src/test/java/" + fileName
	}
}

func (g *CodeGenerator) inferClassName(title string, url string) string {
	// Try title first
	if title != "" {
		// Remove special characters, convert to PascalCase
		words := strings.FieldsFunc(title, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
		})

		var className string
		for _, word := range words {
			if len(word) > 0 {
				className += strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
			}
		}

		if className != "" {
			return className + "Page"
		}
	}

	// Fallback to URL path
	if url != "" {
		// Extract last segment of path
		parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			// Convert to PascalCase
			words := strings.Split(lastPart, "-")
			var className string
			for _, word := range words {
				if len(word) > 0 {
					className += strings.ToUpper(word[0:1]) + word[1:]
				}
			}
			if className != "" {
				return className + "Page"
			}
		}
	}

	// Last resort
	return fmt.Sprintf("Page%d", time.Now().Unix())
}

func (g *CodeGenerator) inferElementName(elemMap map[string]interface{}) string {
	// Priority: id > name > data-testid > text > placeholder

	if id, ok := elemMap["id"].(string); ok && id != "" {
		return g.toCamelCase(id)
	}

	if name, ok := elemMap["name"].(string); ok && name != "" {
		return g.toCamelCase(name)
	}

	if attrs, ok := elemMap["attributes"].(map[string]interface{}); ok {
		if testId, ok := attrs["data-testid"].(string); ok && testId != "" {
			return g.toCamelCase(testId)
		}
	}

	if text, ok := elemMap["text"].(string); ok && text != "" && len(text) < 30 {
		return g.toCamelCase(text)
	}

	if placeholder, ok := elemMap["placeholder"].(string); ok && placeholder != "" {
		return g.toCamelCase(placeholder)
	}

	// Fallback: element type + random
	tagName := "element"
	if tag, ok := elemMap["tagName"].(string); ok {
		tagName = tag
	}

	return fmt.Sprintf("%s%d", tagName, time.Now().UnixNano()%1000)
}

func (g *CodeGenerator) toCamelCase(str string) string {
	// Remove special characters and convert to camelCase
	words := strings.FieldsFunc(str, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})

	if len(words) == 0 {
		return "element"
	}

	var result string
	for i, word := range words {
		if len(word) == 0 {
			continue
		}

		if i == 0 {
			// First word: lowercase
			result += strings.ToLower(word)
		} else {
			// Subsequent words: capitalize first letter
			result += strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
		}
	}

	return result
}

// ============= BDD/Cucumber Generation =============

// generateFeatureFile generates a Gherkin feature file from recorded interactions
func (g *CodeGenerator) generateFeatureFile(req GenerateRequest) (*GeneratedCode, error) {
	if req.PageData == nil && len(req.Interactions) == 0 {
		return nil, fmt.Errorf("no interaction data provided for feature generation")
	}

	interactions := req.Interactions
	if req.PageData != nil {
		interactions = req.PageData.Interactions
	}

	// Build feature content from interactions
	featureName := req.Specification
	if featureName == "" && req.PageData != nil {
		featureName = req.PageData.Title
	}

	var content strings.Builder

	content.WriteString(fmt.Sprintf("Feature: %s\n", featureName))
	content.WriteString("  As a user\n")
	content.WriteString("  I want to interact with the application\n")
	content.WriteString("  So that I can complete my tasks\n\n")

	// Build scenario from interactions
	content.WriteString("  Scenario: User interaction flow\n")

	// Convert interactions to Gherkin steps
	stepNumber := 1
	for _, interaction := range interactions {
		step := g.convertInteractionToGherkinStep(interaction, stepNumber)
		if step != "" {
			keyword := "Given"
			if stepNumber > 1 {
				keyword = "And"
			}
			if interaction.Type == "click" && stepNumber > 1 {
				keyword = "When"
			}
			content.WriteString(fmt.Sprintf("    %s %s\n", keyword, step))
			stepNumber++
		}
	}

	// Add Then step for validation
	content.WriteString("    Then I should see the expected page\n")

	fileName := g.toSnakeCase(featureName) + ".feature"
	filePath := "src/test/resources/features/" + fileName

	return &GeneratedCode{
		FileName: fileName,
		FilePath: filePath,
		Content:  content.String(),
		Language: "gherkin",
		Type:     "feature",
	}, nil
}

// generateStepDefinitions generates Cucumber step definition class from feature
func (g *CodeGenerator) generateStepDefinitions(req GenerateRequest) (*GeneratedCode, error) {
	// Build step definition class
	className := "StepDefinitions"
	if req.Specification != "" {
		className = g.inferClassName(req.Specification, "") + "Steps"
		className = strings.TrimSuffix(className, "Page") + "Steps"
	}

	var content strings.Builder
	content.WriteString("package steps;\n\n")
	content.WriteString("import io.cucumber.java.en.*;\n")
	content.WriteString("import org.openqa.selenium.WebDriver;\n")
	content.WriteString("import pages.*;\n")
	content.WriteString("import static org.testng.Assert.*;\n\n")

	content.WriteString(fmt.Sprintf("public class %s {\n", className))
	content.WriteString("    private WebDriver driver;\n")

	// Add page object fields
	if req.PageData != nil {
		pageClassName := req.PageData.ClassName
		content.WriteString(fmt.Sprintf("    private %s %sPage;\n\n",
			pageClassName, strings.ToLower(pageClassName[:1])+pageClassName[1:]))
	}

	// Generate step methods from interactions
	if req.PageData != nil {
		for i, interaction := range req.PageData.Interactions {
			stepMethod := g.generateStepMethod(interaction, i)
			if stepMethod != "" {
				content.WriteString(stepMethod)
				content.WriteString("\n")
			}
		}
	}

	content.WriteString("}\n")

	fileName := className + ".java"
	filePath := "src/test/java/steps/" + fileName

	return &GeneratedCode{
		FileName: fileName,
		FilePath: filePath,
		Content:  content.String(),
		Language: "java",
		Type:     "stepDefinition",
	}, nil
}

// convertInteractionToGherkinStep converts an interaction to a Gherkin step
func (g *CodeGenerator) convertInteractionToGherkinStep(interaction InteractionData, stepNum int) string {
	switch interaction.Type {
	case "click":
		elementName := interaction.Element
		if elementName == "" {
			elementName = "the button"
		}
		return fmt.Sprintf("I click on %s", elementName)

	case "input", "type":
		elementName := interaction.Element
		value := interaction.Value
		if elementName == "" {
			elementName = "the field"
		}
		if value != "" {
			return fmt.Sprintf("I enter \"%s\" in %s", value, elementName)
		}
		return fmt.Sprintf("I enter text in %s", elementName)

	case "navigate":
		if interaction.Value != "" {
			return fmt.Sprintf("I navigate to \"%s\"", interaction.Value)
		}
		return "I navigate to the page"

	case "select":
		return fmt.Sprintf("I select an option from %s", interaction.Element)

	default:
		return ""
	}
}

// generateStepMethod generates a step definition method
func (g *CodeGenerator) generateStepMethod(interaction InteractionData, index int) string {
	var method strings.Builder

	switch interaction.Type {
	case "click":
		method.WriteString("    @When(\"I click on {string}\")\n")
		method.WriteString("    public void iClickOn(String element) {\n")
		method.WriteString("        // TODO: Implement click action\n")
		method.WriteString("    }\n")

	case "input", "type":
		method.WriteString("    @When(\"I enter {string} in {string}\")\n")
		method.WriteString("    public void iEnterTextIn(String text, String element) {\n")
		method.WriteString("        // TODO: Implement input action\n")
		method.WriteString("    }\n")

	case "navigate":
		method.WriteString("    @Given(\"I navigate to {string}\")\n")
		method.WriteString("    public void iNavigateTo(String url) {\n")
		method.WriteString("        driver.get(url);\n")
		method.WriteString("    }\n")
	}

	return method.String()
}

// toSnakeCase converts string to snake_case for file names
func (g *CodeGenerator) toSnakeCase(str string) string {
	// Remove special characters and convert to snake_case
	words := strings.FieldsFunc(str, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})

	var result []string
	for _, word := range words {
		if len(word) > 0 {
			result = append(result, strings.ToLower(word))
		}
	}

	return strings.Join(result, "_")
}
