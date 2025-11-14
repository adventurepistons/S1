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
}

type GenerateRequest struct {
	Type          string                   `json:"type"` // "pageObject", "test", "utility"
	Specification string                   `json:"specification"`
	Elements      []ElementData            `json:"elements,omitempty"`
	PageData      *RecordedPageData        `json:"pageData,omitempty"`
	Context       map[string]interface{}   `json:"context,omitempty"`
	Framework     string                   `json:"framework"` // "selenium-java", etc.
	ExistingCode  []map[string]interface{} `json:"existingCode,omitempty"`
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
	return &CodeGenerator{
		llmClient:     llm.NewClient(),
		promptBuilder: prompts.NewPromptBuilder(),
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

	// Set framework context
	g.promptBuilder.SetFramework(req.Framework, "testng")

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

	prompt := g.promptBuilder.BuildPageObjectPrompt(spec, elementsForPrompt)

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

	// Set framework
	g.promptBuilder.SetFramework(req.Framework, "testng")

	// Add existing code
	if len(req.ExistingCode) > 0 {
		for _, code := range req.ExistingCode {
			g.promptBuilder.AddContext(code)
		}
	}

	// Build prompt
	prompt := g.promptBuilder.BuildTestCasePrompt(req.Specification, pageObjects)

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

	// Generate page object for each page
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
			fmt.Printf("Failed to generate page object for %s: %v\n", pageData.Title, err)
			continue
		}

		generatedFiles = append(generatedFiles, generated)
	}

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
