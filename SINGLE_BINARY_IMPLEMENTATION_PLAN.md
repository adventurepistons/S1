# Single Binary Implementation Plan - Deep Codebase Understanding First

**Goal**: Build a single Go binary that achieves the deepest possible understanding of Selenium + Java test automation codebases BEFORE implementing queries or AI features.

**Philosophy**: Extract EVERYTHING from code first, store ALL relationships, THEN build query capabilities.

---

## Executive Summary

**Priority Order**:
1. **Weeks 1-2**: Deep Code Extraction (Tree-sitter AST parsing)
2. **Week 3**: Complete Storage (SQLite with 25+ tables)
3. **Week 4**: Query Engine (Knowledge graph & search)
4. **Week 5**: AI Integration (Context building & LLM)

**Single Binary Structure**:
```
automation-copilot/
├── cmd/
│   └── copilot/
│       └── main.go                 # Single entry point
├── internal/
│   ├── parser/                     # Tree-sitter AST parsing
│   ├── extractor/                  # Information extraction
│   ├── storage/                    # SQLite database
│   ├── graph/                      # Knowledge graph
│   ├── query/                      # Query engine
│   └── ai/                         # LLM integration
├── pkg/
│   └── models/                     # Data models
└── go.mod
```

---

## Phase 1: Deep Code Extraction (Weeks 1-2)

### Goal
Extract EVERY detail from Java files with 100% completeness:
- Every annotation parameter
- Every method call
- Every field usage
- Every line number
- Every relationship

### 1.1 What to Extract - Complete Data Models

#### File: `pkg/models/class_data.go`

```go
package models

// ClassData represents complete information extracted from a Java class file
type ClassData struct {
    // File metadata
    FilePath        string   `json:"file_path"`
    Package         string   `json:"package"`
    Imports         []Import `json:"imports"`

    // Class metadata
    ClassName       string              `json:"class_name"`
    Modifiers       []string            `json:"modifiers"` // public, abstract, final
    Extends         string              `json:"extends"`
    Implements      []string            `json:"implements"`
    Annotations     []Annotation        `json:"annotations"`

    // Class contents
    Fields          []Field             `json:"fields"`
    Methods         []Method            `json:"methods"`
    InnerClasses    []ClassData         `json:"inner_classes"`

    // Framework detection hints
    PageObjectModel bool                `json:"is_page_object"`
    TestClass       bool                `json:"is_test_class"`
    StepDefinition  bool                `json:"is_step_definition"`

    // Line numbers
    LineStart       int                 `json:"line_start"`
    LineEnd         int                 `json:"line_end"`
}

type Import struct {
    Path            string   `json:"path"`
    IsStatic        bool     `json:"is_static"`
    IsWildcard      bool     `json:"is_wildcard"`
}

// Field represents a class field with ALL details
type Field struct {
    // Basic info
    Name            string              `json:"name"`
    Type            string              `json:"type"`
    Modifiers       []string            `json:"modifiers"` // private, static, final

    // Annotations (CRITICAL for @FindBy)
    Annotations     []Annotation        `json:"annotations"`

    // For WebElement fields with @FindBy
    IsWebElement    bool                `json:"is_web_element"`
    LocatorStrategy string              `json:"locator_strategy"` // id, css, xpath, name, etc.
    LocatorValue    string              `json:"locator_value"`

    // Initialization
    Initializer     string              `json:"initializer"` // Field value if present

    // Location
    LineNumber      int                 `json:"line_number"`

    // Usage tracking (populated during analysis)
    UsedByMethods   []string            `json:"used_by_methods"` // Which methods use this field
}

// Annotation represents ANY Java annotation with ALL parameters
type Annotation struct {
    Type            string              `json:"type"` // FindBy, Test, BeforeMethod, etc.
    Parameters      map[string]string   `json:"parameters"` // Key-value pairs

    // For @FindBy specifically
    How             string              `json:"how,omitempty"` // ID, CSS, XPATH, NAME, etc.
    Using           string              `json:"using,omitempty"` // The actual selector value

    // For @Test
    Description     string              `json:"description,omitempty"`
    Priority        int                 `json:"priority,omitempty"`
    Enabled         bool                `json:"enabled,omitempty"`

    LineNumber      int                 `json:"line_number"`
}

// Method represents a method with COMPLETE body analysis
type Method struct {
    // Basic info
    Name            string              `json:"name"`
    Modifiers       []string            `json:"modifiers"`
    ReturnType      string              `json:"return_type"`
    Parameters      []Parameter         `json:"parameters"`
    Annotations     []Annotation        `json:"annotations"`

    // Body analysis
    MethodCalls     []MethodCall        `json:"method_calls"` // ALL method calls
    FieldAccess     []string            `json:"field_access"` // ALL fields accessed
    LocalVariables  []LocalVariable     `json:"local_variables"`

    // Control flow
    IfStatements    int                 `json:"if_statements"`
    ForLoops        int                 `json:"for_loops"`
    TryCatchBlocks  int                 `json:"try_catch_blocks"`

    // Test-specific
    IsTest          bool                `json:"is_test"`
    TestType        string              `json:"test_type"` // positive, negative, data-driven

    // Wait strategies
    UsesExplicitWait bool               `json:"uses_explicit_wait"`
    UsesImplicitWait bool               `json:"uses_implicit_wait"`
    WaitTimeout      int                `json:"wait_timeout"` // in seconds

    // Assertions
    Assertions      []Assertion         `json:"assertions"`

    // Location
    LineStart       int                 `json:"line_start"`
    LineEnd         int                 `json:"line_end"`
    BodySource      string              `json:"body_source"` // Full method source code
}

type Parameter struct {
    Name            string   `json:"name"`
    Type            string   `json:"type"`
    Annotations     []Annotation `json:"annotations"` // For @DataProvider, etc.
}

// MethodCall represents a method invocation with FULL context
type MethodCall struct {
    MethodName      string              `json:"method_name"`
    ObjectName      string              `json:"object_name"` // What object is it called on
    Arguments       []string            `json:"arguments"`
    LineNumber      int                 `json:"line_number"`

    // For chained calls: loginPage.enterUsername().clickLogin()
    ChainedFrom     string              `json:"chained_from,omitempty"`
}

type LocalVariable struct {
    Name            string   `json:"name"`
    Type            string   `json:"type"`
    InitValue       string   `json:"init_value"`
    LineNumber      int      `json:"line_number"`
}

// Assertion represents any assertion statement
type Assertion struct {
    Type            string   `json:"type"` // assertEquals, assertTrue, assertThat, etc.
    Expected        string   `json:"expected"`
    Actual          string   `json:"actual"`
    Message         string   `json:"message"` // Custom assertion message
    LineNumber      int      `json:"line_number"`
}
```

#### File: `pkg/models/framework_config.go`

```go
package models

// FrameworkConfig represents detected framework combination
type FrameworkConfig struct {
    // Primary framework
    TestFramework   string   `json:"test_framework"` // TestNG, JUnit5

    // Additional frameworks
    BDDFramework    string   `json:"bdd_framework"` // Cucumber, Serenity, none
    APIFramework    string   `json:"api_framework"` // RestAssured, none

    // Detected from dependencies
    SeleniumVersion string   `json:"selenium_version"`
    JavaVersion     string   `json:"java_version"`

    // Architecture patterns
    UsesPageObjectModel bool `json:"uses_pom"`
    UsesPageFactory     bool `json:"uses_page_factory"`
    UsesScreenplay      bool `json:"uses_screenplay"`

    // Configuration files
    TestNGXMLPath   string   `json:"testng_xml_path"`
    POMFilePath     string   `json:"pom_file_path"`
    CucumberFeatures string  `json:"cucumber_features_path"`
}

// CodingPatterns represents project coding style
type CodingPatterns struct {
    // Naming conventions
    TestMethodNaming    string   `json:"test_method_naming"` // testActionWithCondition, givenWhenThen
    PageObjectNaming    string   `json:"page_object_naming"` // LoginPage, LoginPageObject

    // Test structure
    UsesAAAPattern      bool     `json:"uses_aaa_pattern"` // Arrange-Act-Assert
    UsesGivenWhenThen   bool     `json:"uses_given_when_then"`

    // Waits
    PreferredWaitType   string   `json:"preferred_wait_type"` // explicit, implicit, fluent
    DefaultWaitTimeout  int      `json:"default_wait_timeout"` // in seconds

    // Assertions
    AssertionLibrary    string   `json:"assertion_library"` // TestNG, JUnit, AssertJ, Hamcrest
    UsesAssertMessages  bool     `json:"uses_assert_messages"`

    // Page objects
    WebElementNaming    string   `json:"web_element_naming"` // usernameField, username_field
    MethodNaming        string   `json:"method_naming"` // enterUsername, setUsername, typeUsername

    // Data patterns
    UsesDataProviders   bool     `json:"uses_data_providers"`
    UsesCSVFiles        bool     `json:"uses_csv_files"`
    UsesExcelFiles      bool     `json:"uses_excel_files"`
}
```

### 1.2 How to Extract - Tree-sitter Implementation

#### File: `internal/parser/java_parser.go`

```go
package parser

import (
    "context"
    "fmt"
    "os"

    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/java"
    "automation-copilot/pkg/models"
)

type JavaParser struct {
    parser *sitter.Parser
}

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
    // Query: (package_declaration (scoped_identifier) @package)
    query, _ := sitter.NewQuery([]byte(`(package_declaration (scoped_identifier) @package)`), java.GetLanguage())
    qc := sitter.NewQueryCursor()
    qc.Exec(query, node)

    for {
        match, ok := qc.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            return capture.Node.Content(source)
        }
    }

    return ""
}

// extractImports extracts all import statements
func (jp *JavaParser) extractImports(node *sitter.Node, source []byte) []models.Import {
    imports := []models.Import{}

    query, _ := sitter.NewQuery([]byte(`(import_declaration) @import`), java.GetLanguage())
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
                Path: jp.parseImportPath(importText),
                IsStatic: jp.isStaticImport(importText),
                IsWildcard: jp.isWildcardImport(importText),
            }
            imports = append(imports, imp)
        }
    }

    return imports
}

// extractFields extracts ALL fields with COMPLETE @FindBy details
func (jp *JavaParser) extractFields(node *sitter.Node, source []byte) []models.Field {
    fields := []models.Field{}

    // Query for field declarations
    query, _ := sitter.NewQuery([]byte(`(field_declaration) @field`), java.GetLanguage())
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
        LineNumber: int(fieldNode.StartPoint().Row) + 1,
    }

    // Extract modifiers
    for i := 0; i < int(fieldNode.ChildCount()); i++ {
        child := fieldNode.Child(i)

        switch child.Type() {
        case "modifiers":
            field.Modifiers = jp.extractModifiers(child, source)

        case "type_identifier", "generic_type":
            field.Type = child.Content(source)

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

        case "marker_annotation", "annotation":
            // CRITICAL: Extract @FindBy and other annotations
            annotation := jp.parseAnnotation(child, source)
            field.Annotations = append(field.Annotations, annotation)

            // Special handling for @FindBy
            if annotation.Type == "FindBy" {
                field.IsWebElement = true
                field.LocatorStrategy = jp.extractLocatorStrategy(annotation)
                field.LocatorValue = jp.extractLocatorValue(annotation)
            }
        }
    }

    return field
}

// parseAnnotation extracts COMPLETE annotation with ALL parameters
func (jp *JavaParser) parseAnnotation(annotationNode *sitter.Node, source []byte) models.Annotation {
    annotation := models.Annotation{
        LineNumber: int(annotationNode.StartPoint().Row) + 1,
        Parameters: make(map[string]string),
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
// Examples: @FindBy(id = "username"), @Test(priority = 1, enabled = true)
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
                    }
                }

                // Special handling for @Test
                if annotation.Type == "Test" {
                    switch key {
                    case "description":
                        annotation.Description = value
                    case "priority":
                        // Parse int from value
                        annotation.Priority = jp.parseInt(value)
                    case "enabled":
                        annotation.Enabled = value == "true"
                    }
                }
            }
        }
    }
}

// extractLocatorStrategy determines the locator strategy from @FindBy
func (jp *JavaParser) extractLocatorStrategy(annotation models.Annotation) string {
    if annotation.How != "" {
        return annotation.How
    }

    // Check for shorthand: @FindBy(id = "username")
    if _, ok := annotation.Parameters["id"]; ok {
        return "id"
    }
    if _, ok := annotation.Parameters["css"]; ok {
        return "css"
    }
    if _, ok := annotation.Parameters["xpath"]; ok {
        return "xpath"
    }
    if _, ok := annotation.Parameters["name"]; ok {
        return "name"
    }
    if _, ok := annotation.Parameters["className"]; ok {
        return "className"
    }
    if _, ok := annotation.Parameters["linkText"]; ok {
        return "linkText"
    }

    return "unknown"
}

// extractLocatorValue gets the actual selector value from @FindBy
func (jp *JavaParser) extractLocatorValue(annotation models.Annotation) string {
    if annotation.Using != "" {
        return annotation.Using
    }

    // Check for shorthand values
    for key, value := range annotation.Parameters {
        if key == "id" || key == "css" || key == "xpath" || key == "name" ||
           key == "className" || key == "linkText" {
            return value
        }
    }

    return ""
}

// extractMethods extracts ALL methods with COMPLETE body analysis
func (jp *JavaParser) extractMethods(node *sitter.Node, source []byte) []models.Method {
    methods := []models.Method{}

    query, _ := sitter.NewQuery([]byte(`(method_declaration) @method`), java.GetLanguage())
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
        LineStart: int(methodNode.StartPoint().Row) + 1,
        LineEnd:   int(methodNode.EndPoint().Row) + 1,
        BodySource: methodNode.Content(source),
    }

    // Extract method components
    for i := 0; i < int(methodNode.ChildCount()); i++ {
        child := methodNode.Child(i)

        switch child.Type() {
        case "modifiers":
            method.Modifiers = jp.extractModifiers(child, source)

        case "type_identifier", "void_type", "generic_type":
            method.ReturnType = child.Content(source)

        case "identifier":
            method.Name = child.Content(source)

        case "formal_parameters":
            method.Parameters = jp.extractParameters(child, source)

        case "marker_annotation", "annotation":
            annotation := jp.parseAnnotation(child, source)
            method.Annotations = append(method.Annotations, annotation)

            if annotation.Type == "Test" {
                method.IsTest = true
            }

        case "block":
            // CRITICAL: Analyze method body
            jp.analyzeMethodBody(child, source, &method)
        }
    }

    return method
}

// analyzeMethodBody performs DEEP analysis of method body
func (jp *JavaParser) analyzeMethodBody(bodyNode *sitter.Node, source []byte, method *models.Method) {
    // Extract ALL method calls
    method.MethodCalls = jp.extractMethodCalls(bodyNode, source)

    // Extract ALL field accesses
    method.FieldAccess = jp.extractFieldAccess(bodyNode, source)

    // Extract local variables
    method.LocalVariables = jp.extractLocalVariables(bodyNode, source)

    // Count control flow structures
    method.IfStatements = jp.countNodes(bodyNode, "if_statement")
    method.ForLoops = jp.countNodes(bodyNode, "for_statement") + jp.countNodes(bodyNode, "enhanced_for_statement")
    method.TryCatchBlocks = jp.countNodes(bodyNode, "try_statement")

    // Detect wait strategies
    jp.detectWaitStrategies(bodyNode, source, method)

    // Extract assertions
    method.Assertions = jp.extractAssertions(bodyNode, source)

    // Classify test type
    method.TestType = jp.classifyTestType(method)
}

// extractMethodCalls finds ALL method invocations in method body
func (jp *JavaParser) extractMethodCalls(bodyNode *sitter.Node, source []byte) []models.MethodCall {
    calls := []models.MethodCall{}

    query, _ := sitter.NewQuery([]byte(`(method_invocation) @call`), java.GetLanguage())
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

// detectWaitStrategies detects wait usage in method
func (jp *JavaParser) detectWaitStrategies(bodyNode *sitter.Node, source []byte, method *models.Method) {
    bodyText := bodyNode.Content(source)

    // Detect explicit waits
    if jp.contains(bodyText, "WebDriverWait") || jp.contains(bodyText, "ExpectedConditions") {
        method.UsesExplicitWait = true
        method.WaitTimeout = jp.extractWaitTimeout(bodyText)
    }

    // Detect implicit waits
    if jp.contains(bodyText, "implicitlyWait") {
        method.UsesImplicitWait = true
    }
}

// extractAssertions finds ALL assertions in method
func (jp *JavaParser) extractAssertions(bodyNode *sitter.Node, source []byte) []models.Assertion {
    assertions := []models.Assertion{}

    // Query for method calls that look like assertions
    query, _ := sitter.NewQuery([]byte(`(method_invocation) @call`), java.GetLanguage())
    qc := sitter.NewQueryCursor()
    qc.Exec(query, bodyNode)

    for {
        match, ok := qc.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            callNode := capture.Node
            callText := callNode.Content(source)

            // Check if it's an assertion
            if jp.isAssertion(callText) {
                assertion := models.Assertion{
                    Type: jp.extractAssertionType(callText),
                    LineNumber: int(callNode.StartPoint().Row) + 1,
                }

                // Extract expected, actual, message
                jp.parseAssertionArguments(callNode, source, &assertion)

                assertions = append(assertions, assertion)
            }
        }
    }

    return assertions
}

// Helper functions
func (jp *JavaParser) extractModifiers(node *sitter.Node, source []byte) []string {
    modifiers := []string{}
    for i := 0; i < int(node.ChildCount()); i++ {
        child := node.Child(i)
        modifiers = append(modifiers, child.Content(source))
    }
    return modifiers
}

func (jp *JavaParser) extractParameters(node *sitter.Node, source []byte) []models.Parameter {
    params := []models.Parameter{}
    // Implementation: parse formal_parameter nodes
    return params
}

func (jp *JavaParser) extractFieldAccess(node *sitter.Node, source []byte) []string {
    fields := []string{}
    // Implementation: find field_access nodes
    return fields
}

func (jp *JavaParser) extractLocalVariables(node *sitter.Node, source []byte) []models.LocalVariable {
    vars := []models.LocalVariable{}
    // Implementation: find local_variable_declaration nodes
    return vars
}

func (jp *JavaParser) countNodes(node *sitter.Node, nodeType string) int {
    count := 0
    // Recursive count of nodes matching type
    return count
}

func (jp *JavaParser) classifyTestType(method *models.Method) string {
    // Classify based on method name and structure
    if jp.containsAny(method.Name, []string{"invalid", "error", "fail", "negative"}) {
        return "negative"
    }
    if jp.containsAny(method.Name, []string{"dataProvider", "parameterized"}) {
        return "data-driven"
    }
    return "positive"
}

func (jp *JavaParser) cleanStringLiteral(s string) string {
    // Remove quotes from string literals
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
        return s[1:len(s)-1]
    }
    return s
}

func (jp *JavaParser) parseInt(s string) int {
    // Parse integer from string
    return 0
}

func (jp *JavaParser) contains(text, substr string) bool {
    // String contains check
    return false
}

func (jp *JavaParser) containsAny(text string, substrs []string) bool {
    // Check if text contains any of substrs
    return false
}

func (jp *JavaParser) isAssertion(text string) bool {
    // Check if method call is an assertion
    assertKeywords := []string{"assert", "verify", "should", "expect"}
    for _, keyword := range assertKeywords {
        if jp.contains(text, keyword) {
            return true
        }
    }
    return false
}

func (jp *JavaParser) extractAssertionType(text string) string {
    // Extract assertion type (assertEquals, assertTrue, etc.)
    return ""
}

func (jp *JavaParser) parseAssertionArguments(node *sitter.Node, source []byte, assertion *models.Assertion) {
    // Parse assertion arguments
}

func (jp *JavaParser) extractWaitTimeout(text string) int {
    // Extract timeout value from wait statement
    return 10 // Default
}

func (jp *JavaParser) parseImportPath(importText string) string {
    // Extract import path
    return ""
}

func (jp *JavaParser) isStaticImport(importText string) bool {
    return jp.contains(importText, "static")
}

func (jp *JavaParser) isWildcardImport(importText string) bool {
    return jp.contains(importText, "*")
}

func (jp *JavaParser) extractArguments(node *sitter.Node, source []byte) []string {
    args := []string{}
    // Extract argument expressions
    return args
}

func (jp *JavaParser) detectFrameworkHints(classData *models.ClassData) {
    // Detect if it's a page object
    for _, field := range classData.Fields {
        if field.IsWebElement {
            classData.PageObjectModel = true
            break
        }
    }

    // Detect if it's a test class
    for _, method := range classData.Methods {
        if method.IsTest {
            classData.TestClass = true
            break
        }
    }

    // Detect step definitions (Cucumber)
    for _, method := range classData.Methods {
        for _, ann := range method.Annotations {
            if ann.Type == "Given" || ann.Type == "When" || ann.Type == "Then" {
                classData.StepDefinition = true
                break
            }
        }
    }
}

func (jp *JavaParser) extractClassDeclaration(rootNode *sitter.Node, source []byte, classData *models.ClassData) {
    // Extract class name, extends, implements, annotations
    query, _ := sitter.NewQuery([]byte(`(class_declaration) @class`), java.GetLanguage())
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

            nameNode := classNode.ChildByFieldName("name")
            if nameNode != nil {
                classData.ClassName = nameNode.Content(source)
            }

            // Extract extends
            extendsNode := classNode.ChildByFieldName("superclass")
            if extendsNode != nil {
                classData.Extends = extendsNode.Content(source)
            }

            // Extract implements
            implementsNode := classNode.ChildByFieldName("interfaces")
            if implementsNode != nil {
                // Parse interfaces
            }
        }
    }
}
```

### 1.3 Complete SQLite Schema

#### File: `internal/storage/schema.sql`

```sql
-- ============================================================================
-- COMPLETE DATABASE SCHEMA FOR DEEP CODEBASE UNDERSTANDING
-- ============================================================================

-- File metadata
CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT UNIQUE NOT NULL,
    package TEXT,
    last_modified TIMESTAMP,
    last_indexed TIMESTAMP,
    file_hash TEXT,  -- For change detection
    INDEX idx_package (package)
);

-- Class metadata
CREATE TABLE classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    class_name TEXT NOT NULL,
    fully_qualified_name TEXT UNIQUE NOT NULL,
    extends TEXT,
    modifiers TEXT,  -- JSON array: ["public", "abstract"]
    is_page_object BOOLEAN DEFAULT 0,
    is_test_class BOOLEAN DEFAULT 0,
    is_step_definition BOOLEAN DEFAULT 0,
    line_start INTEGER,
    line_end INTEGER,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    INDEX idx_class_name (class_name),
    INDEX idx_is_page_object (is_page_object),
    INDEX idx_is_test_class (is_test_class)
);

-- Interfaces implemented by classes
CREATE TABLE class_implements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    interface_name TEXT NOT NULL,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    INDEX idx_class_id (class_id)
);

-- Import statements
CREATE TABLE imports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    import_path TEXT NOT NULL,
    is_static BOOLEAN DEFAULT 0,
    is_wildcard BOOLEAN DEFAULT 0,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    INDEX idx_file_id (file_id),
    INDEX idx_import_path (import_path)
);

-- Fields (with SPECIAL focus on @FindBy WebElements)
CREATE TABLE fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    field_name TEXT NOT NULL,
    field_type TEXT NOT NULL,
    modifiers TEXT,  -- JSON array
    initializer TEXT,  -- Initial value if present

    -- WebElement specifics (CRITICAL for page objects)
    is_web_element BOOLEAN DEFAULT 0,
    locator_strategy TEXT,  -- 'id', 'css', 'xpath', 'name', etc.
    locator_value TEXT,     -- The actual selector

    line_number INTEGER,

    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    INDEX idx_class_id (class_id),
    INDEX idx_field_name (field_name),
    INDEX idx_is_web_element (is_web_element),
    INDEX idx_locator_strategy (locator_strategy)
);

-- Annotations on fields (especially @FindBy)
CREATE TABLE field_annotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    field_id INTEGER NOT NULL,
    annotation_type TEXT NOT NULL,  -- 'FindBy', 'CacheLookup', etc.
    parameters TEXT,  -- JSON object of all parameters

    -- @FindBy specifics
    how TEXT,  -- ID, CSS, XPATH, NAME, etc.
    using TEXT,  -- The selector value

    line_number INTEGER,

    FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE,
    INDEX idx_field_id (field_id),
    INDEX idx_annotation_type (annotation_type)
);

-- Methods
CREATE TABLE methods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    method_name TEXT NOT NULL,
    return_type TEXT NOT NULL,
    modifiers TEXT,  -- JSON array

    -- Test method specifics
    is_test BOOLEAN DEFAULT 0,
    test_type TEXT,  -- 'positive', 'negative', 'data-driven'

    -- Wait strategies
    uses_explicit_wait BOOLEAN DEFAULT 0,
    uses_implicit_wait BOOLEAN DEFAULT 0,
    wait_timeout INTEGER,  -- in seconds

    -- Control flow complexity
    if_statements INTEGER DEFAULT 0,
    for_loops INTEGER DEFAULT 0,
    try_catch_blocks INTEGER DEFAULT 0,

    -- Location and source
    line_start INTEGER,
    line_end INTEGER,
    body_source TEXT,  -- Full method source code

    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    INDEX idx_class_id (class_id),
    INDEX idx_method_name (method_name),
    INDEX idx_is_test (is_test)
);

-- Method parameters
CREATE TABLE method_parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    param_name TEXT NOT NULL,
    param_type TEXT NOT NULL,
    param_order INTEGER NOT NULL,  -- Position in parameter list
    annotations TEXT,  -- JSON array of annotations

    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE,
    INDEX idx_method_id (method_id)
);

-- Annotations on methods (especially @Test, @BeforeMethod, etc.)
CREATE TABLE method_annotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    annotation_type TEXT NOT NULL,  -- 'Test', 'BeforeMethod', 'Given', etc.
    parameters TEXT,  -- JSON object of all parameters

    -- @Test specifics
    description TEXT,
    priority INTEGER,
    enabled BOOLEAN,

    line_number INTEGER,

    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE,
    INDEX idx_method_id (method_id),
    INDEX idx_annotation_type (annotation_type)
);

-- Method calls (WHO calls WHAT - critical for understanding flow)
CREATE TABLE method_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    caller_method_id INTEGER NOT NULL,  -- Method making the call
    method_name TEXT NOT NULL,  -- Name of method being called
    object_name TEXT,  -- Object it's called on (e.g., 'loginPage')
    arguments TEXT,  -- JSON array of arguments
    chained_from TEXT,  -- For chained calls
    line_number INTEGER,

    -- Resolved call (if we can determine which method it is)
    callee_method_id INTEGER,  -- Method being called (if found)

    FOREIGN KEY (caller_method_id) REFERENCES methods(id) ON DELETE CASCADE,
    FOREIGN KEY (callee_method_id) REFERENCES methods(id) ON DELETE CASCADE,
    INDEX idx_caller_method_id (caller_method_id),
    INDEX idx_method_name (method_name),
    INDEX idx_callee_method_id (callee_method_id)
);

-- Field access (WHO uses WHAT field - critical for page object usage)
CREATE TABLE field_access (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,  -- Method accessing the field
    field_name TEXT NOT NULL,  -- Name of field being accessed
    field_id INTEGER,  -- Resolved field (if found)
    access_type TEXT,  -- 'read', 'write'
    line_number INTEGER,

    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE,
    FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE,
    INDEX idx_method_id (method_id),
    INDEX idx_field_id (field_id),
    INDEX idx_field_name (field_name)
);

-- Local variables in methods
CREATE TABLE local_variables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    var_name TEXT NOT NULL,
    var_type TEXT NOT NULL,
    init_value TEXT,
    line_number INTEGER,

    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE,
    INDEX idx_method_id (method_id)
);

-- Assertions
CREATE TABLE assertions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    assertion_type TEXT NOT NULL,  -- 'assertEquals', 'assertTrue', etc.
    expected TEXT,
    actual TEXT,
    message TEXT,  -- Custom assertion message
    line_number INTEGER,

    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE,
    INDEX idx_method_id (method_id),
    INDEX idx_assertion_type (assertion_type)
);

-- ============================================================================
-- FRAMEWORK DETECTION TABLES
-- ============================================================================

-- Detected framework configuration
CREATE TABLE framework_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_root TEXT UNIQUE NOT NULL,
    test_framework TEXT,  -- 'TestNG', 'JUnit5'
    bdd_framework TEXT,  -- 'Cucumber', 'Serenity', NULL
    api_framework TEXT,  -- 'RestAssured', NULL
    selenium_version TEXT,
    java_version TEXT,
    uses_page_object_model BOOLEAN DEFAULT 0,
    uses_page_factory BOOLEAN DEFAULT 0,
    uses_screenplay BOOLEAN DEFAULT 0,
    testng_xml_path TEXT,
    pom_file_path TEXT,
    cucumber_features_path TEXT,
    last_detected TIMESTAMP
);

-- Coding patterns detected in the project
CREATE TABLE coding_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_root TEXT UNIQUE NOT NULL,

    -- Naming conventions
    test_method_naming TEXT,  -- Pattern detected
    page_object_naming TEXT,
    web_element_naming TEXT,
    method_naming TEXT,

    -- Test structure
    uses_aaa_pattern BOOLEAN DEFAULT 0,
    uses_given_when_then BOOLEAN DEFAULT 0,

    -- Waits
    preferred_wait_type TEXT,  -- 'explicit', 'implicit', 'fluent'
    default_wait_timeout INTEGER,

    -- Assertions
    assertion_library TEXT,  -- 'TestNG', 'JUnit', 'AssertJ', 'Hamcrest'
    uses_assert_messages BOOLEAN DEFAULT 0,

    -- Data patterns
    uses_data_providers BOOLEAN DEFAULT 0,
    uses_csv_files BOOLEAN DEFAULT 0,
    uses_excel_files BOOLEAN DEFAULT 0,

    last_analyzed TIMESTAMP
);

-- Pattern examples (for learning by example when generating code)
CREATE TABLE pattern_examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern_type TEXT NOT NULL,  -- 'test_method', 'assertion', 'wait', etc.
    example_code TEXT NOT NULL,
    file_path TEXT,
    line_start INTEGER,
    line_end INTEGER,
    frequency INTEGER DEFAULT 1,  -- How often this pattern appears

    INDEX idx_pattern_type (pattern_type)
);

-- ============================================================================
-- KNOWLEDGE GRAPH TABLES
-- ============================================================================

-- Relationships between entities (for building knowledge graph)
CREATE TABLE relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    relationship_type TEXT NOT NULL,  -- 'EXTENDS', 'IMPLEMENTS', 'CALLS', 'USES', etc.
    from_entity_type TEXT NOT NULL,  -- 'class', 'method', 'field'
    from_entity_id INTEGER NOT NULL,
    to_entity_type TEXT NOT NULL,
    to_entity_id INTEGER NOT NULL,
    metadata TEXT,  -- JSON object with additional info

    INDEX idx_relationship_type (relationship_type),
    INDEX idx_from_entity (from_entity_type, from_entity_id),
    INDEX idx_to_entity (to_entity_type, to_entity_id)
);

-- ============================================================================
-- SEARCH AND INDEXING TABLES
-- ============================================================================

-- Full-text search index for code
CREATE VIRTUAL TABLE code_search USING fts5(
    entity_type,  -- 'class', 'method', 'field'
    entity_id,
    name,
    content,
    tokenize='porter'
);

-- ============================================================================
-- METADATA AND TRACKING
-- ============================================================================

-- Indexing progress tracking
CREATE TABLE indexing_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_root TEXT UNIQUE NOT NULL,
    total_files INTEGER,
    indexed_files INTEGER,
    failed_files INTEGER,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    status TEXT  -- 'in_progress', 'completed', 'failed'
);

-- Errors encountered during parsing
CREATE TABLE parsing_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    error_message TEXT NOT NULL,
    error_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_file_path (file_path)
);
```

### 1.4 Execution Plan for Phase 1

**Week 1: Foundation**

Day 1-2: Project Setup
```bash
# Create project structure
mkdir -p automation-copilot/{cmd/copilot,internal/{parser,extractor,storage,graph,query,ai},pkg/models}
cd automation-copilot

# Initialize Go module
go mod init automation-copilot

# Install dependencies
go get github.com/smacker/go-tree-sitter
go get github.com/smacker/go-tree-sitter/java
go get github.com/mattn/go-sqlite3
go get github.com/gorilla/websocket

# Create initial files
touch pkg/models/class_data.go
touch pkg/models/framework_config.go
touch internal/parser/java_parser.go
touch internal/storage/schema.sql
touch internal/storage/database.go
touch cmd/copilot/main.go
```

Day 3-4: Implement Data Models
- Complete `pkg/models/class_data.go` (ALL structs)
- Complete `pkg/models/framework_config.go`
- Add JSON marshaling tags
- Add validation methods

Day 5-7: Implement Java Parser (Core)
- Implement `NewJavaParser()`
- Implement `ParseFile()` - entry point
- Implement `extractPackage()`
- Implement `extractImports()`
- Implement `extractFields()` - with @FindBy parsing
- Implement `extractMethods()` - basic structure

**Week 2: Deep Method Analysis**

Day 8-10: Complete Method Body Parsing
- Implement `analyzeMethodBody()`
- Implement `extractMethodCalls()` - ALL calls
- Implement `extractFieldAccess()` - ALL field usage
- Implement `extractLocalVariables()`
- Implement control flow counting

Day 11-12: Wait & Assertion Detection
- Implement `detectWaitStrategies()`
- Implement `extractAssertions()`
- Implement `classifyTestType()`

Day 13-14: Database Integration
- Create `internal/storage/database.go`
- Implement `InitDatabase(schema.sql)`
- Implement `SaveClassData(classData)`
- Implement transactions for batch inserts

**Success Criteria for Phase 1:**
- [ ] Can parse LoginPage.java with 100% accuracy
- [ ] Extracts ALL @FindBy annotations with correct locator strategies
- [ ] Extracts ALL method calls with line numbers
- [ ] Stores everything in SQLite database
- [ ] Can query: "Show me all WebElements in LoginPage"
- [ ] Can query: "Which methods use the usernameField?"

---

## Phase 2: Framework Detection & Pattern Learning (Week 3)

### Goal
Detect which of 6 framework combinations the project uses and learn project-specific coding patterns.

### 2.1 Framework Detection Implementation

#### File: `internal/detector/framework_detector.go`

```go
package detector

import (
    "os"
    "path/filepath"
    "automation-copilot/pkg/models"
)

type FrameworkDetector struct {
    projectRoot string
    db          *Database
}

func NewFrameworkDetector(projectRoot string, db *Database) *FrameworkDetector {
    return &FrameworkDetector{
        projectRoot: projectRoot,
        db:          db,
    }
}

// Detect identifies framework combination from 6 possibilities
func (fd *FrameworkDetector) Detect() (*models.FrameworkConfig, error) {
    config := &models.FrameworkConfig{}

    // 1. Check pom.xml for dependencies
    pomPath := filepath.Join(fd.projectRoot, "pom.xml")
    if _, err := os.Stat(pomPath); err == nil {
        config.POMFilePath = pomPath
        fd.detectFromPOM(pomPath, config)
    }

    // 2. Check imports in Java files (from database)
    fd.detectFromImports(config)

    // 3. Check for TestNG XML
    testNGXML := fd.findTestNGXML(fd.projectRoot)
    if testNGXML != "" {
        config.TestNGXMLPath = testNGXML
        config.TestFramework = "TestNG"
    }

    // 4. Check for Cucumber features
    featuresPath := fd.findCucumberFeatures(fd.projectRoot)
    if featuresPath != "" {
        config.CucumberFeatures = featuresPath
        config.BDDFramework = "Cucumber"
    }

    // 5. Detect architecture patterns
    fd.detectArchitecturePatterns(config)

    return config, nil
}

func (fd *FrameworkDetector) detectFromPOM(pomPath string, config *models.FrameworkConfig) {
    // Parse pom.xml and look for dependencies
    // - org.testng:testng -> TestNG
    // - org.junit.jupiter:junit-jupiter -> JUnit5
    // - io.cucumber:cucumber-java -> Cucumber
    // - io.rest-assured:rest-assured -> RestAssured
    // - net.serenity-bdd -> Serenity
    // - org.seleniumhq.selenium:selenium-java -> Selenium version
}

func (fd *FrameworkDetector) detectFromImports(config *models.FrameworkConfig) {
    // Query imports table for framework indicators
    imports := fd.db.GetAllImports()

    for _, imp := range imports {
        switch {
        case contains(imp.Path, "org.testng"):
            config.TestFramework = "TestNG"
        case contains(imp.Path, "org.junit.jupiter"):
            config.TestFramework = "JUnit5"
        case contains(imp.Path, "io.cucumber"):
            config.BDDFramework = "Cucumber"
        case contains(imp.Path, "net.serenity"):
            config.BDDFramework = "Serenity"
        case contains(imp.Path, "io.restassured"):
            config.APIFramework = "RestAssured"
        case contains(imp.Path, "org.openqa.selenium"):
            // Extract version from imports if possible
        }
    }
}

func (fd *FrameworkDetector) detectArchitecturePatterns(config *models.FrameworkConfig) {
    // Check if Page Object Model is used
    pageObjects := fd.db.GetPageObjects()  // Classes with @FindBy fields
    if len(pageObjects) > 0 {
        config.UsesPageObjectModel = true
    }

    // Check if PageFactory is used (PageFactory.initElements)
    if fd.db.HasMethodCalls("initElements") {
        config.UsesPageFactory = true
    }

    // Check for Screenplay pattern (Serenity)
    if fd.db.HasImports("net.serenitybdd.screenplay") {
        config.UsesScreenplay = true
    }
}

func (fd *FrameworkDetector) findTestNGXML(root string) string {
    // Search for testng.xml files
    return ""
}

func (fd *FrameworkDetector) findCucumberFeatures(root string) string {
    // Search for .feature files
    return ""
}
```

### 2.2 Pattern Detection Implementation

#### File: `internal/detector/pattern_detector.go`

```go
package detector

import (
    "automation-copilot/pkg/models"
)

type PatternDetector struct {
    db *Database
}

func NewPatternDetector(db *Database) *PatternDetector {
    return &PatternDetector{db: db}
}

// DetectPatterns analyzes all code to learn project coding style
func (pd *PatternDetector) DetectPatterns() (*models.CodingPatterns, error) {
    patterns := &models.CodingPatterns{}

    // 1. Analyze test method naming
    patterns.TestMethodNaming = pd.detectTestNaming()

    // 2. Analyze page object naming
    patterns.PageObjectNaming = pd.detectPageObjectNaming()

    // 3. Detect test structure (AAA vs Given-When-Then)
    patterns.UsesAAAPattern = pd.detectAAAPattern()
    patterns.UsesGivenWhenThen = pd.detectGivenWhenThen()

    // 4. Detect wait strategies
    patterns.PreferredWaitType = pd.detectPreferredWait()
    patterns.DefaultWaitTimeout = pd.detectDefaultTimeout()

    // 5. Detect assertion style
    patterns.AssertionLibrary = pd.detectAssertionLibrary()
    patterns.UsesAssertMessages = pd.detectAssertMessages()

    // 6. Detect data patterns
    patterns.UsesDataProviders = pd.detectDataProviders()

    // 7. Extract pattern examples
    pd.extractPatternExamples()

    return patterns, nil
}

func (pd *PatternDetector) detectTestNaming() string {
    // Analyze test method names to find pattern
    testMethods := pd.db.GetTestMethods()

    // Count naming patterns:
    // - testActionWithCondition
    // - givenConditionWhenActionThenResult
    // - shouldDoSomethingWhenCondition
    // - verify_something_with_underscores

    return "testActionWithCondition"  // Most common pattern
}

func (pd *PatternDetector) detectPreferredWait() string {
    explicitWaits := pd.db.CountMethodsWith("uses_explicit_wait = 1")
    implicitWaits := pd.db.CountMethodsWith("uses_implicit_wait = 1")

    if explicitWaits > implicitWaits {
        return "explicit"
    }
    return "implicit"
}

func (pd *PatternDetector) detectDefaultTimeout() int {
    // Get most common wait timeout value
    timeouts := pd.db.GetAllWaitTimeouts()
    return mostCommon(timeouts)
}

func (pd *PatternDetector) detectAssertionLibrary() string {
    // Check which assertion methods are used most
    assertions := pd.db.GetAllAssertions()

    // Count by type
    testNGCount := 0
    junitCount := 0
    assertJCount := 0

    for _, assertion := range assertions {
        if hasPrefix(assertion.Type, "Assert.") {
            testNGCount++
        } else if hasPrefix(assertion.Type, "Assertions.") {
            junitCount++
        } else if hasPrefix(assertion.Type, "assertThat") {
            assertJCount++
        }
    }

    // Return most common
    if testNGCount > junitCount && testNGCount > assertJCount {
        return "TestNG"
    } else if junitCount > assertJCount {
        return "JUnit"
    }
    return "AssertJ"
}

func (pd *PatternDetector) extractPatternExamples() {
    // Extract REAL examples from codebase to use as templates

    // Example: Find best test method structure
    testMethods := pd.db.GetTestMethods()
    for _, method := range testMethods {
        pd.db.SavePatternExample("test_method", method.BodySource, method.FilePath, method.LineStart, method.LineEnd)
    }

    // Example: Find assertion patterns
    methodsWithAssertions := pd.db.GetMethodsWithAssertions()
    for _, method := range methodsWithAssertions {
        // Extract just the assertion lines
        assertionCode := extractAssertionCode(method)
        pd.db.SavePatternExample("assertion", assertionCode, method.FilePath, 0, 0)
    }

    // Example: Find wait patterns
    methodsWithWaits := pd.db.GetMethodsWithExplicitWait()
    for _, method := range methodsWithWaits {
        waitCode := extractWaitCode(method)
        pd.db.SavePatternExample("wait", waitCode, method.FilePath, 0, 0)
    }
}
```

### 2.3 Execution Plan for Phase 2

**Day 15-17: Framework Detection**
- Implement POM.xml parser
- Implement import analysis
- Implement file structure analysis (testng.xml, .feature files)
- Test with all 6 framework combinations

**Day 18-19: Pattern Detection**
- Implement naming pattern analysis
- Implement assertion pattern analysis
- Implement wait pattern analysis

**Day 20-21: Pattern Example Extraction**
- Extract real code examples
- Store in pattern_examples table
- Build pattern frequency counts

**Success Criteria for Phase 2:**
- [ ] Correctly identifies framework: "Selenium + TestNG + Cucumber"
- [ ] Detects naming pattern: "testActionWithCondition"
- [ ] Detects wait strategy: "explicit, 10 seconds"
- [ ] Detects assertion style: "TestNG with messages"
- [ ] Has 50+ pattern examples stored

---

## Phase 3: Knowledge Graph & Query Engine (Week 4)

### Goal
Build a knowledge graph connecting all entities and implement query engine for answering questions.

### 3.1 Knowledge Graph Builder

#### File: `internal/graph/knowledge_graph.go`

```go
package graph

import (
    "automation-copilot/pkg/models"
)

type KnowledgeGraph struct {
    db *Database
}

func NewKnowledgeGraph(db *Database) *KnowledgeGraph {
    return &KnowledgeGraph{db: db}
}

// BuildGraph creates ALL relationships between entities
func (kg *KnowledgeGraph) BuildGraph() error {
    // 1. Build class inheritance relationships
    kg.buildClassRelationships()

    // 2. Build method call graph
    kg.buildMethodCallGraph()

    // 3. Build field usage graph
    kg.buildFieldUsageGraph()

    // 4. Build test-to-page-object relationships
    kg.buildTestPageObjectLinks()

    return nil
}

// buildClassRelationships creates EXTENDS and IMPLEMENTS relationships
func (kg *KnowledgeGraph) buildClassRelationships() {
    classes := kg.db.GetAllClasses()

    for _, class := range classes {
        // Create EXTENDS relationship
        if class.Extends != "" {
            parentClass := kg.db.FindClassByName(class.Extends)
            if parentClass != nil {
                kg.db.CreateRelationship("EXTENDS", "class", class.ID, "class", parentClass.ID, nil)
            }
        }

        // Create IMPLEMENTS relationships
        implements := kg.db.GetClassImplements(class.ID)
        for _, iface := range implements {
            interfaceClass := kg.db.FindClassByName(iface)
            if interfaceClass != nil {
                kg.db.CreateRelationship("IMPLEMENTS", "class", class.ID, "class", interfaceClass.ID, nil)
            }
        }
    }
}

// buildMethodCallGraph creates CALLS relationships
func (kg *KnowledgeGraph) buildMethodCallGraph() {
    methods := kg.db.GetAllMethods()

    for _, method := range methods {
        calls := kg.db.GetMethodCalls(method.ID)

        for _, call := range calls {
            // Try to resolve which method is being called
            calledMethod := kg.resolveMethodCall(call, method)

            if calledMethod != nil {
                metadata := map[string]interface{}{
                    "line_number": call.LineNumber,
                    "arguments": call.Arguments,
                }
                kg.db.CreateRelationship("CALLS", "method", method.ID, "method", calledMethod.ID, metadata)

                // Update method_calls table with resolved callee
                kg.db.UpdateMethodCallResolution(call.ID, calledMethod.ID)
            }
        }
    }
}

// resolveMethodCall determines which method is being called
func (kg *KnowledgeGraph) resolveMethodCall(call *MethodCall, callerMethod *Method) *Method {
    // Strategy 1: If object name matches a field, get field type and search for method in that class
    if call.ObjectName != "" {
        field := kg.db.FindFieldInClass(callerMethod.ClassID, call.ObjectName)
        if field != nil {
            targetClass := kg.db.FindClassByName(field.Type)
            if targetClass != nil {
                return kg.db.FindMethodInClass(targetClass.ID, call.MethodName)
            }
        }
    }

    // Strategy 2: Search in same class (this.methodName)
    sameClassMethod := kg.db.FindMethodInClass(callerMethod.ClassID, call.MethodName)
    if sameClassMethod != nil {
        return sameClassMethod
    }

    // Strategy 3: Search in parent class
    // (Implementation omitted for brevity)

    return nil
}

// buildFieldUsageGraph creates USES relationships
func (kg *KnowledgeGraph) buildFieldUsageGraph() {
    methods := kg.db.GetAllMethods()

    for _, method := range methods {
        fieldAccesses := kg.db.GetFieldAccesses(method.ID)

        for _, access := range fieldAccesses {
            // Resolve field
            field := kg.db.FindFieldInClass(method.ClassID, access.FieldName)

            if field != nil {
                metadata := map[string]interface{}{
                    "access_type": access.AccessType,
                    "line_number": access.LineNumber,
                }
                kg.db.CreateRelationship("USES", "method", method.ID, "field", field.ID, metadata)

                // Update field_access table with resolved field
                kg.db.UpdateFieldAccessResolution(access.ID, field.ID)
            }
        }
    }
}

// buildTestPageObjectLinks creates TEST_USES_PAGE relationships
func (kg *KnowledgeGraph) buildTestPageObjectLinks() {
    testClasses := kg.db.GetTestClasses()

    for _, testClass := range testClasses {
        // Find all page object fields in test class
        fields := kg.db.GetFieldsForClass(testClass.ID)

        for _, field := range fields {
            // Check if field type is a page object
            pageClass := kg.db.FindClassByName(field.Type)
            if pageClass != nil && pageClass.IsPageObject {
                kg.db.CreateRelationship("TEST_USES_PAGE", "class", testClass.ID, "class", pageClass.ID, nil)
            }
        }
    }
}

// QUERY METHODS

// FindTestsUsingPageObject finds all test classes using a specific page object
func (kg *KnowledgeGraph) FindTestsUsingPageObject(pageObjectName string) []Class {
    pageClass := kg.db.FindClassByName(pageObjectName)
    if pageClass == nil {
        return nil
    }

    relationships := kg.db.GetRelationships("TEST_USES_PAGE", "", 0, "class", pageClass.ID)

    testClasses := []Class{}
    for _, rel := range relationships {
        testClass := kg.db.GetClass(rel.FromEntityID)
        testClasses = append(testClasses, testClass)
    }

    return testClasses
}

// FindFieldDefinition finds where a field is defined
func (kg *KnowledgeGraph) FindFieldDefinition(fieldName string) *FieldLocation {
    fields := kg.db.FindFieldsByName(fieldName)

    if len(fields) == 0 {
        return nil
    }

    field := fields[0]
    class := kg.db.GetClass(field.ClassID)
    file := kg.db.GetFile(class.FileID)

    return &FieldLocation{
        File: file.FilePath,
        Line: field.LineNumber,
        LocatorStrategy: field.LocatorStrategy,
        LocatorValue: field.LocatorValue,
    }
}

// GetCompleteCallChain gets full call chain for a test method
func (kg *KnowledgeGraph) GetCompleteCallChain(testMethodID int) []Method {
    chain := []Method{}
    kg.traverseCallGraph(testMethodID, &chain, make(map[int]bool))
    return chain
}

func (kg *KnowledgeGraph) traverseCallGraph(methodID int, chain *[]Method, visited map[int]bool) {
    if visited[methodID] {
        return // Prevent infinite loops
    }
    visited[methodID] = true

    method := kg.db.GetMethod(methodID)
    *chain = append(*chain, method)

    // Get all methods called by this method
    relationships := kg.db.GetRelationships("CALLS", "method", methodID, "", 0)
    for _, rel := range relationships {
        kg.traverseCallGraph(rel.ToEntityID, chain, visited)
    }
}
```

### 3.2 Query Engine

#### File: `internal/query/query_engine.go`

```go
package query

import (
    "automation-copilot/internal/graph"
)

type QueryEngine struct {
    db *Database
    kg *graph.KnowledgeGraph
}

func NewQueryEngine(db *Database, kg *graph.KnowledgeGraph) *QueryEngine {
    return &QueryEngine{
        db: db,
        kg: kg,
    }
}

// ExecuteQuery handles natural language queries (simple matching for now)
func (qe *QueryEngine) ExecuteQuery(query string) (*QueryResult, error) {
    // Pattern matching for common queries

    // Pattern: "Where is <element> defined?"
    if matches := regexp.MustCompile(`where is (\w+)`).FindStringSubmatch(query); matches != nil {
        elementName := matches[1]
        return qe.findElementLocation(elementName)
    }

    // Pattern: "What tests use <page>?"
    if matches := regexp.MustCompile(`what tests use (\w+)`).FindStringSubmatch(query); matches != nil {
        pageName := matches[1]
        return qe.findTestsUsingPage(pageName)
    }

    // Pattern: "Show me all methods in <class>"
    if matches := regexp.MustCompile(`methods in (\w+)`).FindStringSubmatch(query); matches != nil {
        className := matches[1]
        return qe.getClassMethods(className)
    }

    // Pattern: "What elements are in <page>?"
    if matches := regexp.MustCompile(`elements in (\w+)`).FindStringSubmatch(query); matches != nil {
        pageName := matches[1]
        return qe.getPageElements(pageName)
    }

    return nil, fmt.Errorf("query not understood")
}

func (qe *QueryEngine) findElementLocation(elementName string) (*QueryResult, error) {
    location := qe.kg.FindFieldDefinition(elementName)

    if location == nil {
        return &QueryResult{
            Success: false,
            Message: fmt.Sprintf("Element '%s' not found", elementName),
        }, nil
    }

    return &QueryResult{
        Success: true,
        Message: fmt.Sprintf("Element '%s' found in %s at line %d", elementName, location.File, location.Line),
        Data: map[string]interface{}{
            "file": location.File,
            "line": location.Line,
            "locator_strategy": location.LocatorStrategy,
            "locator_value": location.LocatorValue,
        },
    }, nil
}

func (qe *QueryEngine) getPageElements(pageName string) (*QueryResult, error) {
    class := qe.db.FindClassByName(pageName)
    if class == nil {
        return &QueryResult{Success: false, Message: "Page not found"}, nil
    }

    fields := qe.db.GetWebElementFields(class.ID)

    elements := []map[string]interface{}{}
    for _, field := range fields {
        elements = append(elements, map[string]interface{}{
            "name": field.Name,
            "locator_strategy": field.LocatorStrategy,
            "locator_value": field.LocatorValue,
            "line": field.LineNumber,
        })
    }

    return &QueryResult{
        Success: true,
        Message: fmt.Sprintf("Found %d elements in %s", len(elements), pageName),
        Data: map[string]interface{}{
            "elements": elements,
        },
    }, nil
}
```

### 3.3 Execution Plan for Phase 3

**Day 22-24: Knowledge Graph**
- Implement relationship building
- Implement call graph resolution
- Test graph traversal

**Day 25-27: Query Engine**
- Implement query pattern matching
- Implement result formatting
- Add full-text search integration

**Day 28: Testing & Validation**
- Test with real Selenium project
- Validate all queries work
- Performance optimization

**Success Criteria for Phase 3:**
- [ ] Query "Where is usernameField?" returns correct file and line
- [ ] Query "What tests use LoginPage?" returns all test classes
- [ ] Query "Show elements in LoginPage" returns all @FindBy fields
- [ ] Call graph correctly resolves method calls
- [ ] Can traverse test -> page object -> elements

---

## Phase 4: AI Integration (Week 5)

### Goal
Build context from queries and integrate with LLM for code generation.

### 4.1 Context Builder

#### File: `internal/ai/context_builder.go`

```go
package ai

type ContextBuilder struct {
    db *Database
    kg *graph.KnowledgeGraph
    qe *query.QueryEngine
}

// BuildContext assembles COMPLETE context for AI
func (cb *ContextBuilder) BuildContext(userMessage string) (*Context, error) {
    context := &Context{
        UserMessage: userMessage,
    }

    // 1. Get framework config
    context.Framework = cb.db.GetFrameworkConfig()

    // 2. Get coding patterns
    context.CodingPatterns = cb.db.GetCodingPatterns()

    // 3. Identify relevant files from user message
    relevantFiles := cb.identifyRelevantFiles(userMessage)

    // 4. Get complete class data for relevant files
    for _, file := range relevantFiles {
        classData := cb.db.GetClassForFile(file)
        context.RelevantClasses = append(context.RelevantClasses, classData)
    }

    // 5. Get pattern examples
    context.PatternExamples = cb.db.GetPatternExamples()

    // 6. Build final prompt
    context.Prompt = cb.buildPrompt(context)

    return context, nil
}
```

### 4.2 Execution Plan for Phase 4

**Day 29-31: Context Building**
- Implement context assembly
- Implement prompt building
- Test context completeness

**Day 32-33: LLM Integration**
- Implement OpenAI API calls
- Implement response parsing
- Implement code validation

**Day 34-35: End-to-End Testing**
- Test complete flow: user message -> context -> LLM -> code
- Validate generated code quality
- Performance tuning

**Success Criteria for Phase 4:**
- [ ] Can generate new test method matching project style
- [ ] Can generate new page object matching project patterns
- [ ] Generated code compiles
- [ ] Generated code uses correct framework APIs

---

## Summary: Build Order

**EXACT EXECUTION ORDER:**

1. **Week 1-2**: Parse EVERYTHING from Java files
   - Extract ALL @FindBy annotations
   - Extract ALL method calls
   - Extract ALL field usage
   - Store in SQLite with 25+ tables

2. **Week 3**: Detect framework and learn patterns
   - Identify which of 6 framework combinations
   - Learn project coding style
   - Extract pattern examples

3. **Week 4**: Build knowledge graph and queries
   - Connect all entities
   - Implement query engine
   - Test with real queries

4. **Week 5**: Integrate AI
   - Build context from knowledge graph
   - Call LLM with complete context
   - Validate generated code

**Philosophy**: Deep understanding FIRST, then AI. The better we understand the codebase, the better code we can generate.
