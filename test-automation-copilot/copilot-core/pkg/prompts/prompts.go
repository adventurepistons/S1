package prompts

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 🔒 THIS IS YOUR IP - Prompt engineering is the competitive advantage

type PromptBuilder struct {
	framework    string
	testRunner   string
	codingStyle  map[string]interface{}
	existingCode []map[string]interface{}
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		codingStyle:  make(map[string]interface{}),
		existingCode: []map[string]interface{}{},
	}
}

func (pb *PromptBuilder) SetFramework(framework, testRunner string) *PromptBuilder {
	pb.framework = framework
	pb.testRunner = testRunner
	return pb
}

func (pb *PromptBuilder) SetCodingStyle(style map[string]interface{}) *PromptBuilder {
	pb.codingStyle = style
	return pb
}

func (pb *PromptBuilder) AddContext(code map[string]interface{}) *PromptBuilder {
	pb.existingCode = append(pb.existingCode, code)
	return pb
}

// ===== Page Object Generation =====

func (pb *PromptBuilder) BuildPageObjectPrompt(spec string, elements []map[string]string) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert QA automation engineer specializing in Page Object Model pattern.\n\n")

	// Framework-specific instructions
	prompt.WriteString(fmt.Sprintf("Framework: %s with %s\n", pb.framework, pb.testRunner))
	prompt.WriteString("Follow best practices for maintainable test automation.\n\n")

	// Task
	prompt.WriteString(fmt.Sprintf("Task: Create a Page Object class for: %s\n\n", spec))

	// Elements (if provided)
	if len(elements) > 0 {
		prompt.WriteString("Elements to include:\n")
		for _, elem := range elements {
			prompt.WriteString(fmt.Sprintf("- %s: %s = \"%s\"\n",
				elem["name"], elem["locatorType"], elem["locatorValue"]))
		}
		prompt.WriteString("\n")
	}

	// Coding style from existing code
	if len(pb.existingCode) > 0 {
		prompt.WriteString("Match the coding style of this existing page object:\n")
		prompt.WriteString("```java\n")
		if example := pb.findExamplePageObject(); example != "" {
			prompt.WriteString(example)
		}
		prompt.WriteString("\n```\n\n")
	}

	// Requirements
	prompt.WriteString("Requirements:\n")
	prompt.WriteString("1. Use Page Object Model pattern\n")
	prompt.WriteString("2. Use explicit waits (avoid Thread.sleep)\n")
	prompt.WriteString("3. Create reusable action methods\n")
	prompt.WriteString("4. Add meaningful method names\n")
	prompt.WriteString("5. Follow naming conventions from existing code\n")

	if pb.framework == "selenium-java" {
		prompt.WriteString("6. Use @FindBy annotations for elements\n")
		prompt.WriteString("7. Initialize elements with PageFactory\n")
		prompt.WriteString("8. Prefer stable locators (id > css > xpath)\n")
	}

	prompt.WriteString("\nGenerate ONLY the Java code, no explanations.\n")

	return prompt.String()
}

// ===== Test Case Generation =====

func (pb *PromptBuilder) BuildTestCasePrompt(spec string, pageObjects []map[string]interface{}) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert QA automation engineer.\n\n")

	prompt.WriteString(fmt.Sprintf("Framework: %s with %s\n", pb.framework, pb.testRunner))
	prompt.WriteString(fmt.Sprintf("Task: Create a test case for: %s\n\n", spec))

	// Available page objects
	if len(pageObjects) > 0 {
		prompt.WriteString("Available Page Objects:\n")
		for _, po := range pageObjects {
			prompt.WriteString(fmt.Sprintf("- %s\n", po["className"]))
			if methods, ok := po["methods"].([]interface{}); ok {
				for _, m := range methods {
					if method, ok := m.(map[string]interface{}); ok {
						prompt.WriteString(fmt.Sprintf("  - %s()\n", method["name"]))
					}
				}
			}
		}
		prompt.WriteString("\n")
	}

	// Example test from existing code
	if len(pb.existingCode) > 0 {
		prompt.WriteString("Match the style of this existing test:\n```java\n")
		if example := pb.findExampleTest(); example != "" {
			prompt.WriteString(example)
		}
		prompt.WriteString("\n```\n\n")
	}

	// Test Runner specific instructions
	switch pb.testRunner {
	case "testng":
		prompt.WriteString("Use TestNG annotations (@Test, @BeforeMethod, @AfterMethod)\n")
	case "junit":
		prompt.WriteString("Use JUnit annotations (@Test, @Before, @After)\n")
	case "cucumber":
		prompt.WriteString("Create step definitions with Given/When/Then annotations\n")
	}

	prompt.WriteString("\nRequirements:\n")
	prompt.WriteString("1. Use existing page objects\n")
	prompt.WriteString("2. Add proper assertions\n")
	prompt.WriteString("3. Include setup and teardown\n")
	prompt.WriteString("4. Handle WebDriver lifecycle\n")
	prompt.WriteString("5. Follow AAA pattern (Arrange, Act, Assert)\n")
	prompt.WriteString("\nGenerate ONLY the Java code, no explanations.\n")

	return prompt.String()
}

// ===== Chat/Conversation Prompts =====

func (pb *PromptBuilder) BuildChatPrompt(userMessage string, context map[string]interface{}) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert QA automation consultant specializing in test automation frameworks.\n\n")

	// Add framework context
	if pb.framework != "" {
		prompt.WriteString(fmt.Sprintf("Current project: %s with %s\n\n", pb.framework, pb.testRunner))
	}

	// Add relevant code context
	if relevantCode, ok := context["relevantCode"].([]interface{}); ok && len(relevantCode) > 0 {
		prompt.WriteString("Relevant code from the framework:\n\n")
		for i, code := range relevantCode {
			if codeMap, ok := code.(map[string]interface{}); ok {
				prompt.WriteString(fmt.Sprintf("### File %d: %s\n", i+1, codeMap["className"]))
				if codeStr, ok := codeMap["code"].(string); ok {
					prompt.WriteString("```java\n" + codeStr + "\n```\n\n")
				}
			}
		}
	}

	// Add analysis data if available
	if analysis, ok := context["analysis"].(map[string]interface{}); ok {
		prompt.WriteString(fmt.Sprintf("Framework summary:\n"))
		if pageObjects, ok := analysis["pageObjects"].([]interface{}); ok {
			prompt.WriteString(fmt.Sprintf("- %d page objects\n", len(pageObjects)))
		}
		if tests, ok := analysis["testCases"].([]interface{}); ok {
			prompt.WriteString(fmt.Sprintf("- %d test cases\n", len(tests)))
		}
		prompt.WriteString("\n")
	}

	// User question
	prompt.WriteString(fmt.Sprintf("User question: %s\n\n", userMessage))

	// Instructions
	prompt.WriteString("Provide a helpful, accurate response. If suggesting code:\n")
	prompt.WriteString("- Match the existing framework's style\n")
	prompt.WriteString("- Use best practices\n")
	prompt.WriteString("- Explain your reasoning\n")
	prompt.WriteString("- Provide working, tested code\n")

	return prompt.String()
}

// ===== Code Fix/Refactor Prompts =====

func (pb *PromptBuilder) BuildFixPrompt(brokenCode string, error string, context map[string]interface{}) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert debugging QA automation code.\n\n")

	prompt.WriteString("Broken code:\n```java\n" + brokenCode + "\n```\n\n")
	prompt.WriteString(fmt.Sprintf("Error: %s\n\n", error))

	if workingExample := pb.findSimilarWorkingCode(context); workingExample != "" {
		prompt.WriteString("Similar working code for reference:\n```java\n" + workingExample + "\n```\n\n")
	}

	prompt.WriteString("Fix the code. Requirements:\n")
	prompt.WriteString("1. Fix the error\n")
	prompt.WriteString("2. Maintain the same functionality\n")
	prompt.WriteString("3. Follow framework best practices\n")
	prompt.WriteString("4. Explain what was wrong\n\n")

	prompt.WriteString("Provide the fixed code and explanation.\n")

	return prompt.String()
}

// ===== Data-Driven Test Prompts =====

func (pb *PromptBuilder) BuildDataDrivenPrompt(testSpec string, dataSource string) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert in data-driven test automation.\n\n")

	prompt.WriteString(fmt.Sprintf("Task: Create data-driven test for: %s\n", testSpec))
	prompt.WriteString(fmt.Sprintf("Data source: %s\n\n", dataSource))

	if pb.testRunner == "testng" {
		prompt.WriteString("Use TestNG @DataProvider annotation\n")
		prompt.WriteString("Example:\n")
		prompt.WriteString("```java\n")
		prompt.WriteString("@DataProvider(name = \"loginData\")\n")
		prompt.WriteString("public Object[][] getData() {\n")
		prompt.WriteString("    return new Object[][] {\n")
		prompt.WriteString("        {\"user1\", \"pass1\", true},\n")
		prompt.WriteString("        {\"user2\", \"wrong\", false}\n")
		prompt.WriteString("    };\n")
		prompt.WriteString("}\n\n")
		prompt.WriteString("@Test(dataProvider = \"loginData\")\n")
		prompt.WriteString("public void testLogin(String user, String pass, boolean expected) {...}\n")
		prompt.WriteString("```\n\n")
	}

	if dataSource == "excel" || dataSource == "xlsx" {
		prompt.WriteString("Include Apache POI code to read Excel file\n")
	}

	prompt.WriteString("Generate complete implementation with data provider.\n")

	return prompt.String()
}

// ===== Helper Methods (Find examples from existing code) =====

func (pb *PromptBuilder) findExamplePageObject() string {
	for _, code := range pb.existingCode {
		if code["type"] == "pageObject" {
			if codeStr, ok := code["code"].(string); ok {
				return codeStr
			}
		}
	}
	return ""
}

func (pb *PromptBuilder) findExampleTest() string {
	for _, code := range pb.existingCode {
		if code["type"] == "test" {
			if codeStr, ok := code["code"].(string); ok {
				return codeStr
			}
		}
	}
	return ""
}

func (pb *PromptBuilder) findSimilarWorkingCode(context map[string]interface{}) string {
	// Logic to find similar working code from context
	// This is a simplified version
	if similarCode, ok := context["similarCode"].(string); ok {
		return similarCode
	}
	return ""
}

// ===== System Prompts =====

func GetSystemPrompt(role string) string {
	prompts := map[string]string{
		"code_generator": `You are an expert QA automation engineer. Generate clean, maintainable test automation code following industry best practices. Always use Page Object Model pattern, explicit waits, and meaningful names.`,

		"code_reviewer": `You are a senior QA automation architect reviewing code. Identify issues with stability, maintainability, and best practices. Suggest improvements.`,

		"consultant": `You are an expert QA automation consultant. Provide strategic guidance on test automation frameworks, best practices, and implementation approaches. Be concise and practical.`,

		"debugger": `You are an expert at debugging test automation code. Analyze errors, identify root causes, and provide precise fixes. Explain what went wrong and why.`,
	}

	if prompt, ok := prompts[role]; ok {
		return prompt
	}
	return prompts["consultant"]
}

// ===== Utility Functions =====

func FormatContext(context map[string]interface{}) string {
	formatted, _ := json.MarshalIndent(context, "", "  ")
	return string(formatted)
}

func ExtractCodeBlocks(response string) []string {
	var blocks []string
	lines := strings.Split(response, "\n")

	var inBlock bool
	var currentBlock strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inBlock {
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			} else {
				inBlock = true
			}
		} else if inBlock {
			currentBlock.WriteString(line + "\n")
		}
	}

	return blocks
}
