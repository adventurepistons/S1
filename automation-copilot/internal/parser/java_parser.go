package parser

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/adventurepistons/automation-copilot/pkg/models"
)

// JavaParser parses Java files using Tree-sitter
type JavaParser struct {
	parser *sitter.Parser
}

// NewJavaParser creates a new Java parser
func NewJavaParser() *JavaParser {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	return &JavaParser{
		parser: parser,
	}
}

// ParseFile parses a Java file and extracts COMPLETE information
func (jp *JavaParser) ParseFile(filePath string) (*models.ClassData, error) {
	// Read source code
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse to AST
	tree, err := jp.parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()

	// Extract complete class data
	classData := &models.ClassData{
		FilePath: filePath,
	}

	// Parse package declaration
	classData.Package = jp.extractPackage(rootNode, sourceCode)

	// Parse imports
	classData.Imports = jp.extractImports(rootNode, sourceCode)

	// Parse class declaration
	jp.extractClassDeclaration(rootNode, sourceCode, classData)

	// Parse ALL fields (with special handling for @FindBy)
	classData.Fields = jp.extractFields(rootNode, sourceCode)

	// Parse ALL methods (with complete body analysis)
	classData.Methods = jp.extractMethods(rootNode, sourceCode)

	// Detect framework hints
	jp.detectFrameworkHints(classData)

	return classData, nil
}

// extractPackage extracts package declaration
func (jp *JavaParser) extractPackage(node *sitter.Node, source []byte) string {
	query, err := sitter.NewQuery([]byte(`(package_declaration) @package`), java.GetLanguage())
	if err != nil {
		return ""
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			pkgNode := capture.Node
			// Find the scoped_identifier child
			for i := 0; i < int(pkgNode.ChildCount()); i++ {
				child := pkgNode.Child(i)
				if child.Type() == "scoped_identifier" || child.Type() == "identifier" {
					return child.Content(source)
				}
			}
		}
	}

	return ""
}

// extractImports extracts all import statements
func (jp *JavaParser) extractImports(node *sitter.Node, source []byte) []models.Import {
	imports := []models.Import{}

	query, err := sitter.NewQuery([]byte(`(import_declaration) @import`), java.GetLanguage())
	if err != nil {
		return imports
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			importNode := capture.Node
			importText := importNode.Content(source)

			imp := models.Import{
				Path:       jp.parseImportPath(importNode, source),
				IsStatic:   strings.Contains(importText, "static"),
				IsWildcard: strings.Contains(importText, "*"),
			}
			imports = append(imports, imp)
		}
	}

	return imports
}

// parseImportPath extracts the import path from import node
func (jp *JavaParser) parseImportPath(importNode *sitter.Node, source []byte) string {
	for i := 0; i < int(importNode.ChildCount()); i++ {
		child := importNode.Child(i)
		if child.Type() == "scoped_identifier" || child.Type() == "identifier" {
			return child.Content(source)
		}
		if child.Type() == "asterisk" {
			// Handle wildcard imports
			continue
		}
	}
	return ""
}

// extractClassDeclaration extracts class information
func (jp *JavaParser) extractClassDeclaration(rootNode *sitter.Node, source []byte, classData *models.ClassData) {
	query, err := sitter.NewQuery([]byte(`(class_declaration) @class`), java.GetLanguage())
	if err != nil {
		return
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, rootNode)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			classNode := capture.Node
			classData.LineStart = int(classNode.StartPoint().Row) + 1
			classData.LineEnd = int(classNode.EndPoint().Row) + 1

			// Extract class name
			nameNode := classNode.ChildByFieldName("name")
			if nameNode != nil {
				classData.ClassName = nameNode.Content(source)
			}

			// Extract extends
			superclassNode := classNode.ChildByFieldName("superclass")
			if superclassNode != nil {
				// Find the type_identifier in superclass
				for i := 0; i < int(superclassNode.ChildCount()); i++ {
					child := superclassNode.Child(i)
					if child.Type() == "type_identifier" {
						classData.Extends = child.Content(source)
					}
				}
			}

			// Extract implements
			interfacesNode := classNode.ChildByFieldName("interfaces")
			if interfacesNode != nil {
				classData.Implements = jp.extractInterfaces(interfacesNode, source)
			}

			// Extract modifiers
			for i := 0; i < int(classNode.ChildCount()); i++ {
				child := classNode.Child(i)
				if child.Type() == "modifiers" {
					classData.Modifiers = jp.extractModifiers(child, source)
				}
			}
		}
	}
}

// extractInterfaces extracts implemented interfaces
func (jp *JavaParser) extractInterfaces(interfacesNode *sitter.Node, source []byte) []string {
	interfaces := []string{}

	for i := 0; i < int(interfacesNode.ChildCount()); i++ {
		child := interfacesNode.Child(i)
		if child.Type() == "type_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				typeNode := child.Child(j)
				if typeNode.Type() == "type_identifier" {
					interfaces = append(interfaces, typeNode.Content(source))
				}
			}
		}
	}

	return interfaces
}

// extractModifiers extracts modifiers (public, private, static, etc.)
func (jp *JavaParser) extractModifiers(node *sitter.Node, source []byte) []string {
	modifiers := []string{}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		modType := child.Type()
		if modType != "marker_annotation" && modType != "annotation" {
			modifiers = append(modifiers, child.Content(source))
		}
	}
	return modifiers
}

// extractFields extracts ALL fields with COMPLETE @FindBy details
func (jp *JavaParser) extractFields(node *sitter.Node, source []byte) []models.Field {
	fields := []models.Field{}

	query, err := sitter.NewQuery([]byte(`(field_declaration) @field`), java.GetLanguage())
	if err != nil {
		return fields
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			fieldNode := capture.Node
			field := jp.parseFieldNode(fieldNode, source)
			fields = append(fields, field)
		}
	}

	return fields
}

// parseFieldNode extracts COMPLETE field information including @FindBy
func (jp *JavaParser) parseFieldNode(fieldNode *sitter.Node, source []byte) models.Field {
	field := models.Field{
		LineNumber:  int(fieldNode.StartPoint().Row) + 1,
		Annotations: []models.Annotation{},
	}

	// Extract field components
	for i := 0; i < int(fieldNode.ChildCount()); i++ {
		child := fieldNode.Child(i)

		switch child.Type() {
		case "modifiers":
			field.Modifiers = jp.extractModifiers(child, source)
			// Also extract annotations from modifiers
			field.Annotations = jp.extractAnnotationsFromModifiers(child, source)

		case "type_identifier", "generic_type":
			field.Type = child.Content(source)
			// Check if it's a WebElement
			fieldType := child.Content(source)
			if strings.Contains(fieldType, "WebElement") || strings.Contains(fieldType, "List<WebElement>") {
				field.IsWebElement = true
			}

		case "variable_declarator":
			// Extract field name and initializer
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				field.Name = nameNode.Content(source)
			}

			valueNode := child.ChildByFieldName("value")
			if valueNode != nil {
				field.Initializer = valueNode.Content(source)
			}
		}
	}

	// Extract @FindBy details
	for _, ann := range field.Annotations {
		if ann.Type == "FindBy" {
			field.LocatorStrategy = jp.extractLocatorStrategy(ann)
			field.LocatorValue = jp.extractLocatorValue(ann)
		}
	}

	return field
}

// extractAnnotationsFromModifiers extracts annotations from modifiers node
func (jp *JavaParser) extractAnnotationsFromModifiers(modifiersNode *sitter.Node, source []byte) []models.Annotation {
	annotations := []models.Annotation{}

	for i := 0; i < int(modifiersNode.ChildCount()); i++ {
		child := modifiersNode.Child(i)
		if child.Type() == "marker_annotation" || child.Type() == "annotation" {
			ann := jp.parseAnnotation(child, source)
			annotations = append(annotations, ann)
		}
	}

	return annotations
}

// parseAnnotation extracts COMPLETE annotation with ALL parameters
func (jp *JavaParser) parseAnnotation(annotationNode *sitter.Node, source []byte) models.Annotation {
	annotation := models.Annotation{
		LineNumber: int(annotationNode.StartPoint().Row) + 1,
		Parameters: make(map[string]string),
		Enabled:    true, // Default for @Test
	}

	// Extract annotation type
	nameNode := annotationNode.ChildByFieldName("name")
	if nameNode != nil {
		annotation.Type = nameNode.Content(source)
	}

	// Extract annotation arguments
	argsNode := annotationNode.ChildByFieldName("arguments")
	if argsNode != nil {
		jp.extractAnnotationParameters(argsNode, source, &annotation)
	}

	return annotation
}

// extractAnnotationParameters extracts ALL annotation parameters
func (jp *JavaParser) extractAnnotationParameters(argsNode *sitter.Node, source []byte, annotation *models.Annotation) {
	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)

		if child.Type() == "element_value_pair" {
			// Extract key-value pair
			keyNode := child.ChildByFieldName("key")
			valueNode := child.ChildByFieldName("value")

			if keyNode != nil && valueNode != nil {
				key := keyNode.Content(source)
				value := jp.cleanStringLiteral(valueNode.Content(source))
				annotation.Parameters[key] = value

				// Special handling for @FindBy
				if annotation.Type == "FindBy" {
					jp.handleFindByParameter(key, value, annotation)
				}

				// Special handling for @Test
				if annotation.Type == "Test" {
					jp.handleTestParameter(key, value, annotation)
				}
			}
		} else if child.Type() == "string_literal" {
			// Single value annotation like @Test("description")
			value := jp.cleanStringLiteral(child.Content(source))
			if annotation.Type == "Test" {
				annotation.Description = value
			}
		}
	}
}

// handleFindByParameter processes @FindBy parameters
func (jp *JavaParser) handleFindByParameter(key, value string, annotation *models.Annotation) {
	switch key {
	case "how":
		annotation.How = value
	case "using":
		annotation.Using = value
	case "id":
		annotation.How = "ID"
		annotation.Using = value
	case "css":
		annotation.How = "CSS"
		annotation.Using = value
	case "xpath":
		annotation.How = "XPATH"
		annotation.Using = value
	case "name":
		annotation.How = "NAME"
		annotation.Using = value
	case "className":
		annotation.How = "CLASS_NAME"
		annotation.Using = value
	case "linkText":
		annotation.How = "LINK_TEXT"
		annotation.Using = value
	case "partialLinkText":
		annotation.How = "PARTIAL_LINK_TEXT"
		annotation.Using = value
	}
}

// handleTestParameter processes @Test parameters
func (jp *JavaParser) handleTestParameter(key, value string, annotation *models.Annotation) {
	switch key {
	case "description":
		annotation.Description = value
	case "priority":
		if priority, err := strconv.Atoi(value); err == nil {
			annotation.Priority = priority
		}
	case "enabled":
		annotation.Enabled = value == "true"
	}
}

// extractLocatorStrategy determines the locator strategy from @FindBy
func (jp *JavaParser) extractLocatorStrategy(annotation models.Annotation) string {
	if annotation.How != "" {
		return strings.ToLower(annotation.How)
	}

	// Check for shorthand parameters
	locatorTypes := []string{"id", "css", "xpath", "name", "className", "linkText", "partialLinkText"}
	for _, locType := range locatorTypes {
		if _, ok := annotation.Parameters[locType]; ok {
			return locType
		}
	}

	return "unknown"
}

// extractLocatorValue gets the actual selector value from @FindBy
func (jp *JavaParser) extractLocatorValue(annotation models.Annotation) string {
	if annotation.Using != "" {
		return annotation.Using
	}

	// Check for shorthand values
	locatorTypes := []string{"id", "css", "xpath", "name", "className", "linkText", "partialLinkText"}
	for _, locType := range locatorTypes {
		if value, ok := annotation.Parameters[locType]; ok {
			return value
		}
	}

	return ""
}

// extractMethods extracts ALL methods with COMPLETE body analysis
func (jp *JavaParser) extractMethods(node *sitter.Node, source []byte) []models.Method {
	methods := []models.Method{}

	query, err := sitter.NewQuery([]byte(`(method_declaration) @method`), java.GetLanguage())
	if err != nil {
		return methods
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			methodNode := capture.Node
			method := jp.parseMethodNode(methodNode, source)
			methods = append(methods, method)
		}
	}

	return methods
}

// parseMethodNode extracts COMPLETE method information including body analysis
func (jp *JavaParser) parseMethodNode(methodNode *sitter.Node, source []byte) models.Method {
	method := models.Method{
		LineStart:   int(methodNode.StartPoint().Row) + 1,
		LineEnd:     int(methodNode.EndPoint().Row) + 1,
		BodySource:  methodNode.Content(source),
		Annotations: []models.Annotation{},
		Parameters:  []models.Parameter{},
	}

	// Extract method components
	for i := 0; i < int(methodNode.ChildCount()); i++ {
		child := methodNode.Child(i)

		switch child.Type() {
		case "modifiers":
			method.Modifiers = jp.extractModifiers(child, source)
			method.Annotations = jp.extractAnnotationsFromModifiers(child, source)

		case "type_identifier", "void_type", "generic_type", "integral_type", "boolean_type":
			method.ReturnType = child.Content(source)

		case "identifier":
			method.Name = child.Content(source)

		case "formal_parameters":
			method.Parameters = jp.extractParameters(child, source)

		case "block":
			// CRITICAL: Analyze method body
			jp.analyzeMethodBody(child, source, &method)
		}
	}

	// Check if it's a test method
	for _, ann := range method.Annotations {
		if ann.Type == "Test" || ann.Type == "ParameterizedTest" {
			method.IsTest = true
			break
		}
	}

	// Classify test type
	if method.IsTest {
		method.TestType = jp.classifyTestType(&method)
	}

	return method
}

// extractParameters extracts method parameters
func (jp *JavaParser) extractParameters(paramsNode *sitter.Node, source []byte) []models.Parameter {
	params := []models.Parameter{}

	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		child := paramsNode.Child(i)
		if child.Type() == "formal_parameter" {
			param := models.Parameter{}

			// Extract parameter type and name
			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				param.Type = typeNode.Content(source)
			}

			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				param.Name = nameNode.Content(source)
			}

			params = append(params, param)
		}
	}

	return params
}

// analyzeMethodBody performs DEEP analysis of method body
func (jp *JavaParser) analyzeMethodBody(bodyNode *sitter.Node, source []byte, method *models.Method) {
	// Extract ALL method calls
	method.MethodCalls = jp.extractMethodCalls(bodyNode, source)

	// Extract ALL field accesses
	method.FieldAccess = jp.extractFieldAccess(bodyNode, source)

	// Count control flow structures
	method.IfStatements = jp.countNodeType(bodyNode, "if_statement")
	method.ForLoops = jp.countNodeType(bodyNode, "for_statement") + jp.countNodeType(bodyNode, "enhanced_for_statement")
	method.TryCatchBlocks = jp.countNodeType(bodyNode, "try_statement")

	// Detect wait strategies
	jp.detectWaitStrategies(bodyNode, source, method)

	// Extract assertions
	method.Assertions = jp.extractAssertions(bodyNode, source)
}

// extractMethodCalls finds ALL method invocations in method body
func (jp *JavaParser) extractMethodCalls(bodyNode *sitter.Node, source []byte) []models.MethodCall {
	calls := []models.MethodCall{}

	query, err := sitter.NewQuery([]byte(`(method_invocation) @call`), java.GetLanguage())
	if err != nil {
		return calls
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, bodyNode)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			callNode := capture.Node

			call := models.MethodCall{
				LineNumber: int(callNode.StartPoint().Row) + 1,
			}

			// Extract method name
			nameNode := callNode.ChildByFieldName("name")
			if nameNode != nil {
				call.MethodName = nameNode.Content(source)
			}

			// Extract object (what it's called on)
			objectNode := callNode.ChildByFieldName("object")
			if objectNode != nil {
				call.ObjectName = objectNode.Content(source)
			}

			// Extract arguments
			argsNode := callNode.ChildByFieldName("arguments")
			if argsNode != nil {
				call.Arguments = jp.extractArguments(argsNode, source)
			}

			calls = append(calls, call)
		}
	}

	return calls
}

// extractFieldAccess extracts field accesses
func (jp *JavaParser) extractFieldAccess(bodyNode *sitter.Node, source []byte) []string {
	fields := []string{}
	seen := make(map[string]bool)

	query, err := sitter.NewQuery([]byte(`(field_access) @field`), java.GetLanguage())
	if err != nil {
		return fields
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, bodyNode)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			fieldNode := capture.Node
			fieldNode = fieldNode.ChildByFieldName("field")
			if fieldNode != nil {
				fieldName := fieldNode.Content(source)
				if !seen[fieldName] {
					fields = append(fields, fieldName)
					seen[fieldName] = true
				}
			}
		}
	}

	return fields
}

// extractArguments extracts method call arguments
func (jp *JavaParser) extractArguments(argsNode *sitter.Node, source []byte) []string {
	args := []string{}

	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child.Type() != "(" && child.Type() != ")" && child.Type() != "," {
			args = append(args, child.Content(source))
		}
	}

	return args
}

// countNodeType counts nodes of a specific type recursively
func (jp *JavaParser) countNodeType(node *sitter.Node, nodeType string) int {
	count := 0

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == nodeType {
			count++
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			traverse(n.Child(i))
		}
	}

	traverse(node)
	return count
}

// detectWaitStrategies detects wait usage in method
func (jp *JavaParser) detectWaitStrategies(bodyNode *sitter.Node, source []byte, method *models.Method) {
	bodyText := bodyNode.Content(source)

	// Detect explicit waits
	if strings.Contains(bodyText, "WebDriverWait") || strings.Contains(bodyText, "ExpectedConditions") {
		method.UsesExplicitWait = true
		method.WaitTimeout = jp.extractWaitTimeout(bodyText)
	}

	// Detect implicit waits
	if strings.Contains(bodyText, "implicitlyWait") {
		method.UsesImplicitWait = true
	}
}

// extractWaitTimeout extracts timeout value from wait code
func (jp *JavaParser) extractWaitTimeout(code string) int {
	// Look for patterns like: new WebDriverWait(driver, 10) or Duration.ofSeconds(10)
	patterns := []string{
		`WebDriverWait\([^,]+,\s*(\d+)`,
		`Duration\.ofSeconds\((\d+)`,
		`Duration\.ofMinutes\((\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(code)
		if len(matches) > 1 {
			if timeout, err := strconv.Atoi(matches[1]); err == nil {
				return timeout
			}
		}
	}

	return 10 // Default
}

// extractAssertions finds ALL assertions in method
func (jp *JavaParser) extractAssertions(bodyNode *sitter.Node, source []byte) []models.Assertion {
	assertions := []models.Assertion{}

	// Query for method calls
	query, err := sitter.NewQuery([]byte(`(method_invocation) @call`), java.GetLanguage())
	if err != nil {
		return assertions
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, bodyNode)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			callNode := capture.Node

			// Get method name
			nameNode := callNode.ChildByFieldName("name")
			if nameNode == nil {
				continue
			}

			methodName := nameNode.Content(source)

			// Check if it's an assertion
			if jp.isAssertion(methodName) {
				assertion := models.Assertion{
					Type:       methodName,
					LineNumber: int(callNode.StartPoint().Row) + 1,
				}

				// Extract arguments
				argsNode := callNode.ChildByFieldName("arguments")
				if argsNode != nil {
					args := jp.extractArguments(argsNode, source)
					if len(args) >= 2 {
						assertion.Expected = args[0]
						assertion.Actual = args[1]
					}
					if len(args) >= 3 {
						assertion.Message = args[2]
					}
				}

				assertions = append(assertions, assertion)
			}
		}
	}

	return assertions
}

// isAssertion checks if a method name is an assertion
func (jp *JavaParser) isAssertion(methodName string) bool {
	assertKeywords := []string{
		"assert", "verify", "should", "expect",
		"assertEquals", "assertTrue", "assertFalse", "assertNotNull", "assertNull",
		"assertThat", "assertArrayEquals", "assertSame", "assertNotSame",
	}

	methodLower := strings.ToLower(methodName)
	for _, keyword := range assertKeywords {
		if strings.Contains(methodLower, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// classifyTestType classifies test as positive, negative, or data-driven
func (jp *JavaParser) classifyTestType(method *models.Method) string {
	nameLower := strings.ToLower(method.Name)

	// Check for negative test indicators
	negativeKeywords := []string{"invalid", "error", "fail", "negative", "exception", "bad"}
	for _, keyword := range negativeKeywords {
		if strings.Contains(nameLower, keyword) {
			return "negative"
		}
	}

	// Check for data-driven test indicators
	for _, ann := range method.Annotations {
		if ann.Type == "DataProvider" || ann.Type == "ParameterizedTest" {
			return "data-driven"
		}
	}

	return "positive"
}

// detectFrameworkHints detects framework usage hints
func (jp *JavaParser) detectFrameworkHints(classData *models.ClassData) {
	// Detect if it's a page object (has @FindBy fields)
	for _, field := range classData.Fields {
		if field.IsWebElement {
			classData.PageObjectModel = true
			break
		}
	}

	// Detect if it's a test class (has @Test methods)
	for _, method := range classData.Methods {
		if method.IsTest {
			classData.TestClass = true
			break
		}
	}

	// Detect step definitions (Cucumber)
	cucumberAnnotations := []string{"Given", "When", "Then", "And", "But"}
	for _, method := range classData.Methods {
		for _, ann := range method.Annotations {
			for _, cucAnn := range cucumberAnnotations {
				if ann.Type == cucAnn {
					classData.StepDefinition = true
					return
				}
			}
		}
	}
}

// cleanStringLiteral removes quotes from string literals
func (jp *JavaParser) cleanStringLiteral(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
