package ai

import (
	"fmt"
	"regexp"
	"strings"
)

// ResponseParser extracts code from LLM responses
type ResponseParser struct {
	extractThinking bool // Whether to extract planning/verification sections
}

// NewResponseParser creates a new response parser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{
		extractThinking: true,
	}
}

// ParseResponse parses the LLM response and extracts generated code
func (rp *ResponseParser) ParseResponse(response string) (*ParsedResponse, error) {
	parsed := &ParsedResponse{
		RawResponse: response,
	}

	// Extract sections
	parsed.Planning = rp.extractSection(response, "Step 1: Plan", "Step 2: Generate")
	parsed.Generation = rp.extractSection(response, "Step 2: Generate", "Step 3: Verify")
	parsed.Verification = rp.extractSection(response, "Step 3: Verify", "Final Code")

	// Extract final code (most important!)
	finalCode, err := rp.extractFinalCode(response)
	if err != nil {
		return nil, err
	}

	parsed.GeneratedCode = finalCode

	return parsed, nil
}

// extractFinalCode extracts the final verified code from the response
func (rp *ResponseParser) extractFinalCode(response string) (*GeneratedCode, error) {
	// Look for "Final Code" section
	finalCodeIdx := strings.Index(response, "**Final Code**")
	if finalCodeIdx < 0 {
		// Try alternative markers
		finalCodeIdx = strings.Index(response, "## Final Code")
		if finalCodeIdx < 0 {
			// If no "Final Code" section, try to extract any Java code block
			return rp.extractFirstJavaCode(response)
		}
	}

	// Extract content after "Final Code"
	afterFinal := response[finalCodeIdx:]

	// Find Java code block
	code, err := rp.extractJavaCodeBlock(afterFinal)
	if err != nil {
		return nil, fmt.Errorf("failed to extract final code: %w", err)
	}

	// Parse code metadata
	generatedCode := &GeneratedCode{
		Code:     code,
		Language: "java",
	}

	// Detect code type and file path
	rp.detectCodeMetadata(generatedCode)

	return generatedCode, nil
}

// extractJavaCodeBlock extracts Java code from a markdown code block
func (rp *ResponseParser) extractJavaCodeBlock(text string) (string, error) {
	// Regex to match ```java ... ```
	re := regexp.MustCompile("(?s)```java\\s*\n(.*?)\n```")
	matches := re.FindStringSubmatch(text)

	if len(matches) < 2 {
		return "", fmt.Errorf("no Java code block found")
	}

	return strings.TrimSpace(matches[1]), nil
}

// extractFirstJavaCode extracts the first Java code block (fallback)
func (rp *ResponseParser) extractFirstJavaCode(response string) (*GeneratedCode, error) {
	code, err := rp.extractJavaCodeBlock(response)
	if err != nil {
		return nil, fmt.Errorf("no code found in response")
	}

	return &GeneratedCode{
		Code:     code,
		Language: "java",
	}, nil
}

// extractSection extracts content between two markers
func (rp *ResponseParser) extractSection(text, startMarker, endMarker string) string {
	startIdx := strings.Index(text, startMarker)
	if startIdx < 0 {
		return ""
	}

	endIdx := strings.Index(text[startIdx:], endMarker)
	if endIdx < 0 {
		// No end marker, take rest of text
		return strings.TrimSpace(text[startIdx+len(startMarker):])
	}

	section := text[startIdx+len(startMarker) : startIdx+endIdx]
	return strings.TrimSpace(section)
}

// detectCodeMetadata detects code type and suggested file path
func (rp *ResponseParser) detectCodeMetadata(code *GeneratedCode) {
	// Extract package name
	packageRe := regexp.MustCompile(`package\s+([\w.]+)\s*;`)
	if matches := packageRe.FindStringSubmatch(code.Code); len(matches) > 1 {
		code.PackageName = matches[1]
	}

	// Extract class name
	classRe := regexp.MustCompile(`public\s+class\s+(\w+)`)
	if matches := classRe.FindStringSubmatch(code.Code); len(matches) > 1 {
		code.ClassName = matches[1]
	}

	// Detect code type
	code.CodeType = rp.detectCodeType(code.Code)

	// Build suggested file path
	if code.ClassName != "" {
		if code.PackageName != "" {
			// Convert package to path: com.example.pages → com/example/pages
			packagePath := strings.ReplaceAll(code.PackageName, ".", "/")
			code.SuggestedPath = fmt.Sprintf("%s/%s.java", packagePath, code.ClassName)
		} else {
			code.SuggestedPath = fmt.Sprintf("%s.java", code.ClassName)
		}
	}
}

// detectCodeType detects whether this is a test, page object, util, etc.
func (rp *ResponseParser) detectCodeType(code string) string {
	lowerCode := strings.ToLower(code)

	// Check for test
	if strings.Contains(lowerCode, "@test") ||
		strings.Contains(lowerCode, "extends testbase") ||
		strings.Contains(lowerCode, "import org.testng") ||
		strings.Contains(lowerCode, "import org.junit") {
		return "test"
	}

	// Check for page object
	if strings.Contains(lowerCode, "page") ||
		strings.Contains(lowerCode, "@findby") ||
		strings.Contains(lowerCode, "pagefactory") {
		return "page_object"
	}

	// Check for step definition (BDD)
	if strings.Contains(lowerCode, "@given") ||
		strings.Contains(lowerCode, "@when") ||
		strings.Contains(lowerCode, "@then") {
		return "step_definition"
	}

	// Check for utility
	if strings.Contains(lowerCode, "util") ||
		strings.Contains(lowerCode, "helper") {
		return "utility"
	}

	return "unknown"
}

// ValidateCode performs basic validation on generated code
func (rp *ResponseParser) ValidateCode(code *GeneratedCode) []ValidationError {
	errors := []ValidationError{}

	// Check for empty code
	if strings.TrimSpace(code.Code) == "" {
		errors = append(errors, ValidationError{
			Severity: "error",
			Message:  "Generated code is empty",
		})
		return errors
	}

	// Check for package declaration
	if !strings.Contains(code.Code, "package ") {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  "Missing package declaration",
		})
	}

	// Check for class declaration
	if !strings.Contains(code.Code, "class ") {
		errors = append(errors, ValidationError{
			Severity: "error",
			Message:  "No class declaration found",
		})
	}

	// Check for common anti-patterns
	if strings.Contains(code.Code, "Thread.sleep") {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  "Code contains Thread.sleep() - consider using explicit waits",
			Line:     rp.findLineNumber(code.Code, "Thread.sleep"),
		})
	}

	// Check for missing imports (rough heuristic)
	hasImports := strings.Contains(code.Code, "import ")
	if !hasImports && (strings.Contains(code.Code, "WebDriver") ||
		strings.Contains(code.Code, "@FindBy") ||
		strings.Contains(code.Code, "@Test")) {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  "Code may be missing import statements",
		})
	}

	// Check for test without assertions
	if code.CodeType == "test" {
		if !strings.Contains(code.Code, "assert") &&
			!strings.Contains(code.Code, "Assert") &&
			!strings.Contains(code.Code, "verify") {
			errors = append(errors, ValidationError{
				Severity: "warning",
				Message:  "Test code may be missing assertions",
			})
		}
	}

	return errors
}

// findLineNumber finds the line number of a substring
func (rp *ResponseParser) findLineNumber(code, substr string) int {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i + 1 // 1-indexed
		}
	}
	return 0
}

// ParsedResponse represents a parsed LLM response
type ParsedResponse struct {
	RawResponse   string
	Planning      string
	Generation    string
	Verification  string
	GeneratedCode *GeneratedCode
}

// GeneratedCode represents extracted code from LLM response
type GeneratedCode struct {
	Code          string
	Language      string
	CodeType      string // "test", "page_object", "step_definition", "utility"
	ClassName     string
	PackageName   string
	SuggestedPath string
}

// ValidationError represents a code validation issue
type ValidationError struct {
	Severity string // "error", "warning", "info"
	Message  string
	Line     int
}
