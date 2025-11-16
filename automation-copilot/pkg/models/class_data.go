package models

// ClassData represents complete information extracted from a Java class file
type ClassData struct {
	// File metadata
	FilePath string   `json:"file_path"`
	Package  string   `json:"package"`
	Imports  []Import `json:"imports"`

	// Class metadata
	ClassName   string       `json:"class_name"`
	Modifiers   []string     `json:"modifiers"` // public, abstract, final
	Extends     string       `json:"extends"`
	Implements  []string     `json:"implements"`
	Annotations []Annotation `json:"annotations"`

	// Class contents
	Fields       []Field     `json:"fields"`
	Methods      []Method    `json:"methods"`
	InnerClasses []ClassData `json:"inner_classes"`

	// Framework detection hints
	PageObjectModel bool `json:"is_page_object"`
	TestClass       bool `json:"is_test_class"`
	StepDefinition  bool `json:"is_step_definition"`

	// Line numbers
	LineStart int `json:"line_start"`
	LineEnd   int `json:"line_end"`
}

// Import represents an import statement
type Import struct {
	Path       string `json:"path"`
	IsStatic   bool   `json:"is_static"`
	IsWildcard bool   `json:"is_wildcard"`
}

// Field represents a class field with ALL details
type Field struct {
	// Basic info
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Modifiers []string `json:"modifiers"` // private, static, final

	// Annotations (CRITICAL for @FindBy)
	Annotations []Annotation `json:"annotations"`

	// For WebElement fields with @FindBy
	IsWebElement    bool   `json:"is_web_element"`
	LocatorStrategy string `json:"locator_strategy"` // id, css, xpath, name, etc.
	LocatorValue    string `json:"locator_value"`

	// Initialization
	Initializer string `json:"initializer"` // Field value if present

	// Location
	LineNumber int `json:"line_number"`

	// Usage tracking (populated during analysis)
	UsedByMethods []string `json:"used_by_methods"` // Which methods use this field
}

// Annotation represents ANY Java annotation with ALL parameters
type Annotation struct {
	Type       string            `json:"type"` // FindBy, Test, BeforeMethod, etc.
	Parameters map[string]string `json:"parameters"` // Key-value pairs

	// For @FindBy specifically
	How   string `json:"how,omitempty"`   // ID, CSS, XPATH, NAME, etc.
	Using string `json:"using,omitempty"` // The actual selector value

	// For @Test
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Enabled     bool   `json:"enabled"`

	LineNumber int `json:"line_number"`
}

// Method represents a method with COMPLETE body analysis
type Method struct {
	// Basic info
	Name        string       `json:"name"`
	Modifiers   []string     `json:"modifiers"`
	ReturnType  string       `json:"return_type"`
	Parameters  []Parameter  `json:"parameters"`
	Annotations []Annotation `json:"annotations"`

	// Body analysis
	MethodCalls    []MethodCall    `json:"method_calls"` // ALL method calls
	FieldAccess    []string        `json:"field_access"` // ALL fields accessed
	LocalVariables []LocalVariable `json:"local_variables"`

	// Control flow
	IfStatements   int `json:"if_statements"`
	ForLoops       int `json:"for_loops"`
	TryCatchBlocks int `json:"try_catch_blocks"`

	// Test-specific
	IsTest   bool   `json:"is_test"`
	TestType string `json:"test_type"` // positive, negative, data-driven

	// Wait strategies
	UsesExplicitWait bool `json:"uses_explicit_wait"`
	UsesImplicitWait bool `json:"uses_implicit_wait"`
	WaitTimeout      int  `json:"wait_timeout"` // in seconds

	// Assertions
	Assertions []Assertion `json:"assertions"`

	// Location
	LineStart  int    `json:"line_start"`
	LineEnd    int    `json:"line_end"`
	BodySource string `json:"body_source"` // Full method source code
}

// Parameter represents a method parameter
type Parameter struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Annotations []Annotation `json:"annotations"` // For @DataProvider, etc.
}

// MethodCall represents a method invocation with FULL context
type MethodCall struct {
	MethodName string   `json:"method_name"`
	ObjectName string   `json:"object_name"` // What object is it called on
	Arguments  []string `json:"arguments"`
	LineNumber int      `json:"line_number"`

	// For chained calls: loginPage.enterUsername().clickLogin()
	ChainedFrom string `json:"chained_from,omitempty"`
}

// LocalVariable represents a local variable in a method
type LocalVariable struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	InitValue  string `json:"init_value"`
	LineNumber int    `json:"line_number"`
}

// Assertion represents any assertion statement
type Assertion struct {
	Type       string `json:"type"` // assertEquals, assertTrue, assertThat, etc.
	Expected   string `json:"expected"`
	Actual     string `json:"actual"`
	Message    string `json:"message"` // Custom assertion message
	LineNumber int    `json:"line_number"`
}
