package graph

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryEngine handles natural language queries against the knowledge graph
type QueryEngine struct {
	kg *KnowledgeGraph
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(kg *KnowledgeGraph) *QueryEngine {
	return &QueryEngine{kg: kg}
}

// QueryResult represents the result of a query
type QueryResult struct {
	Success bool
	Message string
	Data    interface{}
}

// Execute processes a natural language query
func (qe *QueryEngine) Execute(query string) (*QueryResult, error) {
	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Pattern: "where is <element>" or "where is <element> defined"
	if matches := qe.matchPattern(queryLower, `where\s+(?:is\s+)?(\w+)(?:\s+defined)?`); matches != nil {
		return qe.findElementLocation(matches[1])
	}

	// Pattern: "what tests use <page>" or "which tests use <page>"
	if matches := qe.matchPattern(queryLower, `(?:what|which)\s+tests?\s+use\s+(\w+)`); matches != nil {
		return qe.findTestsUsingPage(matches[1])
	}

	// Pattern: "show me all methods in <class>" or "methods in <class>"
	if matches := qe.matchPattern(queryLower, `(?:show\s+(?:me\s+)?)?(?:all\s+)?methods?\s+in\s+(\w+)`); matches != nil {
		return qe.getClassMethods(matches[1])
	}

	// Pattern: "what elements are in <page>" or "elements in <page>"
	if matches := qe.matchPattern(queryLower, `(?:what\s+)?elements?\s+(?:are\s+)?in\s+(\w+)`); matches != nil {
		return qe.getPageElements(matches[1])
	}

	// Pattern: "who uses <field>" or "what uses <field>"
	if matches := qe.matchPattern(queryLower, `(?:who|what)\s+uses?\s+(\w+)`); matches != nil {
		return qe.findMethodsUsingField(matches[1])
	}

	// Pattern: "call chain for <method>" or "what does <method> call"
	if matches := qe.matchPattern(queryLower, `(?:call\s+chain\s+for|what\s+does)\s+(\w+)(?:\s+call)?`); matches != nil {
		return qe.getCallChain(matches[1])
	}

	// Pattern: "list all page objects" or "show page objects"
	if qe.matchPattern(queryLower, `(?:list|show)\s+(?:all\s+)?page\s+objects?`) != nil {
		return qe.listAllPageObjects()
	}

	// Pattern: "show all tests" or "list tests"
	if qe.matchPattern(queryLower, `(?:list|show)\s+(?:all\s+)?tests?`) != nil {
		return qe.listAllTests()
	}

	return &QueryResult{
		Success: false,
		Message: "Query not understood. Try: 'where is <element>', 'what tests use <page>', 'elements in <page>', etc.",
	}, nil
}

// matchPattern matches a regex pattern and returns captures
func (qe *QueryEngine) matchPattern(query, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(query)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// findElementLocation finds where an element is defined
func (qe *QueryEngine) findElementLocation(elementName string) (*QueryResult, error) {
	location, err := qe.kg.FindFieldDefinition(elementName)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: fmt.Sprintf("Element '%s' not found", elementName),
		}, nil
	}

	message := fmt.Sprintf("Element '%s' found:\n", elementName)
	message += fmt.Sprintf("  File: %s:%d\n", location.FilePath, location.LineNumber)
	message += fmt.Sprintf("  Class: %s\n", location.ClassName)
	if location.LocatorStrategy != "" {
		message += fmt.Sprintf("  Locator: %s = \"%s\"\n", location.LocatorStrategy, location.LocatorValue)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    location,
	}, nil
}

// findTestsUsingPage finds all tests using a page object
func (qe *QueryEngine) findTestsUsingPage(pageName string) (*QueryResult, error) {
	tests, err := qe.kg.FindTestsUsingPageObject(pageName)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: fmt.Sprintf("Page object '%s' not found or no tests use it", pageName),
		}, nil
	}

	if len(tests) == 0 {
		return &QueryResult{
			Success: true,
			Message: fmt.Sprintf("No tests found using '%s'", pageName),
		}, nil
	}

	message := fmt.Sprintf("Found %d test(s) using '%s':\n", len(tests), pageName)
	for _, test := range tests {
		message += fmt.Sprintf("  - %s (%s)\n", test.ClassName, test.FilePath)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    tests,
	}, nil
}

// getPageElements gets all elements in a page object
func (qe *QueryEngine) getPageElements(pageName string) (*QueryResult, error) {
	// Query database directly for elements
	rows, err := qe.kg.db.Query(`
		SELECT f.field_name, f.locator_strategy, f.locator_value, f.line_number
		FROM fields f
		JOIN classes c ON f.class_id = c.id
		WHERE c.class_name = ? AND f.is_web_element = 1
		ORDER BY f.line_number
	`, pageName)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: fmt.Sprintf("Failed to query elements for '%s'", pageName),
		}, nil
	}
	defer rows.Close()

	type Element struct {
		Name     string
		Strategy string
		Value    string
		Line     int
	}

	var elements []Element
	for rows.Next() {
		var elem Element
		if err := rows.Scan(&elem.Name, &elem.Strategy, &elem.Value, &elem.Line); err != nil {
			continue
		}
		elements = append(elements, elem)
	}

	if len(elements) == 0 {
		return &QueryResult{
			Success: true,
			Message: fmt.Sprintf("No WebElements found in '%s'", pageName),
		}, nil
	}

	message := fmt.Sprintf("Found %d WebElement(s) in '%s':\n", len(elements), pageName)
	for _, elem := range elements {
		message += fmt.Sprintf("  Line %d: %s (%s = \"%s\")\n", elem.Line, elem.Name, elem.Strategy, elem.Value)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    elements,
	}, nil
}

// findMethodsUsingField finds all methods using a specific field
func (qe *QueryEngine) findMethodsUsingField(fieldName string) (*QueryResult, error) {
	methods, err := qe.kg.GetMethodsUsingField(fieldName)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: fmt.Sprintf("Field '%s' not found or not used", fieldName),
		}, nil
	}

	if len(methods) == 0 {
		return &QueryResult{
			Success: true,
			Message: fmt.Sprintf("No methods found using '%s'", fieldName),
		}, nil
	}

	message := fmt.Sprintf("Found %d method(s) using '%s':\n", len(methods), fieldName)
	for _, method := range methods {
		message += fmt.Sprintf("  - %s.%s() at line %d\n", method.ClassName, method.MethodName, method.LineNumber)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    methods,
	}, nil
}

// getCallChain gets the call chain for a method
func (qe *QueryEngine) getCallChain(methodName string) (*QueryResult, error) {
	chain, err := qe.kg.GetCompleteCallChain(methodName)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: fmt.Sprintf("Method '%s' not found or is not a test method", methodName),
		}, nil
	}

	if len(chain) == 0 {
		return &QueryResult{
			Success: true,
			Message: fmt.Sprintf("No call chain found for '%s'", methodName),
		}, nil
	}

	message := fmt.Sprintf("Call chain for '%s' (%d methods):\n", methodName, len(chain))
	for i, method := range chain {
		indent := strings.Repeat("  ", i)
		message += fmt.Sprintf("%s%d. %s() -> %s\n", indent, i+1, method.Name, method.ReturnType)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    chain,
	}, nil
}

// getClassMethods gets all methods in a class
func (qe *QueryEngine) getClassMethods(className string) (*QueryResult, error) {
	rows, err := qe.kg.db.Query(`
		SELECT m.method_name, m.return_type, m.is_test, m.line_start
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		WHERE c.class_name = ?
		ORDER BY m.line_start
	`, className)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: fmt.Sprintf("Failed to query methods for '%s'", className),
		}, nil
	}
	defer rows.Close()

	type Method struct {
		Name       string
		ReturnType string
		IsTest     bool
		Line       int
	}

	var methods []Method
	for rows.Next() {
		var method Method
		if err := rows.Scan(&method.Name, &method.ReturnType, &method.IsTest, &method.Line); err != nil {
			continue
		}
		methods = append(methods, method)
	}

	if len(methods) == 0 {
		return &QueryResult{
			Success: true,
			Message: fmt.Sprintf("No methods found in '%s'", className),
		}, nil
	}

	message := fmt.Sprintf("Found %d method(s) in '%s':\n", len(methods), className)
	for _, method := range methods {
		testMarker := ""
		if method.IsTest {
			testMarker = " [@Test]"
		}
		message += fmt.Sprintf("  Line %d: %s %s()%s\n", method.Line, method.ReturnType, method.Name, testMarker)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    methods,
	}, nil
}

// listAllPageObjects lists all page objects in the project
func (qe *QueryEngine) listAllPageObjects() (*QueryResult, error) {
	pageObjects, err := qe.kg.GetAllPageObjects()
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: "Failed to retrieve page objects",
		}, nil
	}

	if len(pageObjects) == 0 {
		return &QueryResult{
			Success: true,
			Message: "No page objects found in the project",
		}, nil
	}

	message := fmt.Sprintf("Found %d page object(s):\n", len(pageObjects))
	for _, po := range pageObjects {
		message += fmt.Sprintf("  - %s (%d elements) - %s\n", po.ClassName, po.ElementCount, po.FilePath)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    pageObjects,
	}, nil
}

// listAllTests lists all test methods in the project
func (qe *QueryEngine) listAllTests() (*QueryResult, error) {
	rows, err := qe.kg.db.Query(`
		SELECT m.method_name, c.class_name, m.test_type, f.file_path
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		JOIN files f ON c.file_id = f.id
		WHERE m.is_test = 1
		ORDER BY c.class_name, m.method_name
	`)
	if err != nil {
		return &QueryResult{
			Success: false,
			Message: "Failed to retrieve tests",
		}, nil
	}
	defer rows.Close()

	type Test struct {
		MethodName string
		ClassName  string
		TestType   string
		FilePath   string
	}

	var tests []Test
	for rows.Next() {
		var test Test
		if err := rows.Scan(&test.MethodName, &test.ClassName, &test.TestType, &test.FilePath); err != nil {
			continue
		}
		tests = append(tests, test)
	}

	if len(tests) == 0 {
		return &QueryResult{
			Success: true,
			Message: "No test methods found in the project",
		}, nil
	}

	message := fmt.Sprintf("Found %d test method(s):\n", len(tests))
	currentClass := ""
	for _, test := range tests {
		if test.ClassName != currentClass {
			message += fmt.Sprintf("\n  %s:\n", test.ClassName)
			currentClass = test.ClassName
		}
		message += fmt.Sprintf("    - %s [%s]\n", test.MethodName, test.TestType)
	}

	return &QueryResult{
		Success: true,
		Message: message,
		Data:    tests,
	}, nil
}
