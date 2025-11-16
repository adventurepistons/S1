# 🏗️ Implementation Plan - Complete Framework Understanding System

**Goal**: Build a system that knows EVERYTHING about the user's test automation codebase

**Principle**: Extract ALL information upfront, let LLM query it later (NO intent parsing yet)

---

## 🎯 Core Objective

From user message: **"edit login test"**

System must be able to answer:
- ✅ What framework? (TestNG + Selenium)
- ✅ Where is LoginTest.java? (exact path)
- ✅ What methods are in LoginTest? (all test methods)
- ✅ What page objects does it use? (LoginPage, DashboardPage)
- ✅ Where are elements defined? (@FindBy in LoginPage.java line 15, 18, 21...)
- ✅ What's the login() method? (in LoginPage or LoginTest?)
- ✅ What assertions are used? (Assert.assertTrue with messages)
- ✅ What's the coding style? (AAA pattern, camelCase, explicit waits)
- ✅ Show me ALL related code (full files with line numbers)

**How**: Deep AST parsing + Complete database + Query system

---

## 📦 Phase 1: Enhanced AST Parsing (Week 1)

### What We Extract from Each File Type

#### From pom.xml
- Build tool: Maven/Gradle
- Java version
- ALL dependencies (group, artifact, version, scope)
- Framework detection markers (selenium, testng, cucumber, etc.)
- Plugins and configuration

#### From Java - Page Objects (LoginPage.java)
```
Package: pages
Extends: BasePage
Imports: [all imports]

Fields (elements):
- usernameField: @FindBy(id="username") line 15
- passwordField: @FindBy(id="password") line 18
- loginButton: @FindBy(css="button[type='submit']") line 21
- errorMessage: @FindBy(className="error-message") line 24

Methods:
- login(String, String) → void, lines 56-60
  Calls: enterUsername(), enterPassword(), clickLoginButton()
  
- enterUsername(String) → void, lines 39-43
  Uses: wait, usernameField
  Calls: wait.until(), usernameField.clear(), usernameField.sendKeys()
  Wait type: explicit (visibilityOf)
  Action: input
  
- clickLoginButton() → DashboardPage, lines 50-54
  Returns: new DashboardPage(driver)
  Navigation: goes to Dashboard
```

#### From Java - Test Classes (LoginTest.java)
```
Package: tests
Extends: BaseTest
Uses: LoginPage, DashboardPage

Test Methods:
- testLoginWithValidCredentials() lines 12-28
  Type: positive test
  Groups: [smoke, regression]
  Creates: LoginPage, DashboardPage
  Calls: login(), isWelcomeMessageDisplayed(), getUsername()
  Assertions: 2 (both with messages)
  Pattern: AAA (Arrange, Act, Assert)
  Test data: hardcoded
  
Data Providers:
- loginData: 5 test cases (valid/invalid/edge cases)

Coding Style Detected:
- Naming: testActionWithCondition
- Assertions: TestNG Assert with messages (100%)
- Comments: Yes (// Arrange, // Act, // Assert)
```

#### From Gherkin (.feature files)
```
Feature: User Authentication
Tags: @regression, @smoke

Scenarios:
1. "Successful login" - @positive
   Steps: Given/When/Then...
   
2. "Login fails" - @negative
   Steps: ...
   
3. Scenario Outline with Examples table
   Parameters: <username>, <password>, <result>
   3 data rows

Step Definitions Needed:
- "Given the user is on the login page"
- "When the user enters username {string}"
- ...
```

### Database Schema (Complete)

```sql
-- CORE TABLES
CREATE TABLE projects (...);
CREATE TABLE files (file_path, package, file_type, hash, ...);
CREATE TABLE classes (name, extends, implements, class_type, ...);
CREATE TABLE fields (name, type, @FindBy details, locator_type, locator_value, ...);
CREATE TABLE methods (name, params, return_type, line_start, line_end, ...);
CREATE TABLE method_details (calls_methods, uses_fields, wait_type, action_type, ...);

-- TEST-SPECIFIC
CREATE TABLE test_methods (test_type, groups, assertions, follows_AAA, ...);
CREATE TABLE data_providers (name, test_data, ...);
CREATE TABLE assertions (type, condition, message, ...);

-- BDD/CUCUMBER
CREATE TABLE features (name, tags, ...);
CREATE TABLE scenarios (name, type, tags, ...);
CREATE TABLE scenario_steps (keyword, text, parameters, ...);
CREATE TABLE step_definitions (annotation, pattern, method_name, ...);

-- RELATIONSHIPS
CREATE TABLE class_relationships (from_class, to_class, type);
CREATE TABLE method_calls (from_method, to_method, ...);

-- PATTERNS
CREATE TABLE coding_patterns (category, pattern_name, pattern_value, confidence, ...);

-- SEARCH
CREATE TABLE search_index (keyword, entity_type, entity_name, file_path, line_number, ...);
CREATE TABLE embeddings (entity_type, entity_id, embedding, ...);

-- APPLICATION MAP
CREATE TABLE application_pages (page_name, url, ...);
CREATE TABLE application_elements (element_name, locator_type, locator_value, ...);
```

### Go Implementation Structure

```
copilot-core/pkg/
├── parser/
│   ├── java_parser.go          ⭐ Parse with Tree-sitter, extract ALL
│   ├── gherkin_parser.go       Parse .feature files
│   ├── pom_parser.go           Parse pom.xml
│   └── testng_parser.go        Parse testng.xml
│
├── extractor/
│   ├── class_extractor.go      Extract class details
│   ├── field_extractor.go      ⭐ Extract @FindBy with locators
│   ├── method_extractor.go     ⭐ Extract methods + body analysis
│   ├── test_extractor.go       Extract test-specific info
│   └── relationship_extractor.go ⭐ Build all relationships
│
├── framework/
│   ├── detector.go             ⭐ Detect all 6 combinations
│   ├── testng.go
│   ├── junit.go
│   └── cucumber.go
│
├── pattern/
│   ├── detector.go             ⭐ Detect coding patterns
│   ├── naming_analyzer.go      Analyze naming conventions
│   ├── assertion_analyzer.go   Analyze assertion styles
│   └── wait_analyzer.go        Analyze wait strategies
│
├── storage/
│   ├── database.go             SQLite operations
│   ├── schema.go               Create tables
│   └── queries.go              SQL queries
│
├── knowledge/
│   ├── graph.go                ⭐ Knowledge graph
│   ├── search_index.go         Fast keyword search
│   ├── query_engine.go         ⭐ Answer ANY question
│   └── embeddings.go           Vector embeddings
│
└── context/
    ├── builder.go              ⭐ Build context for LLM
    ├── assembler.go            Assemble relevant code
    └── optimizer.go            Optimize tokens
```

### Implementation Tasks

**Week 1, Day 1-2: Enhanced JavaParser**
```go
// Extract EVERYTHING from Java files
func (jp *JavaParser) ParseClass(filePath string) (*ClassData, error) {
    // Parse to AST
    // Extract: package, imports, class metadata
    // Extract: ALL fields with @FindBy details
    // Extract: ALL methods with body analysis
    // Store in database
}
```

**Week 1, Day 3-4: Field & Method Extractors**
```go
// Extract @FindBy with locator details
func (fe *FieldExtractor) ExtractField(node *sitter.Node) *Field {
    // Extract: name, type, modifiers
    // IF @FindBy: extract locator_type and locator_value
    // Infer element_type from field name
}

// Analyze method body
func (me *MethodExtractor) AnalyzeMethodBody(node *sitter.Node) *MethodDetails {
    // Find: field references
    // Find: method calls (what calls what)
    // Find: object creations
    // Detect: waits (explicit/implicit)
    // Detect: control structures
    // Infer: method_type (action/validation/getter)
}
```

**Week 1, Day 5-7: Test, Storage, Testing**
```go
// Extract test-specific info
func (te *TestExtractor) ExtractTestMethod(method *Method) *TestMethod {
    // Extract: groups, priority, description
    // Detect: test_type (positive/negative/data_driven)
    // Extract: ALL assertions
    // Check: AAA pattern
    // Find: page object usage
}

// Store everything in SQLite
func (db *Database) StoreClass(class *ClassData) error {
    // Insert into: classes, fields, methods, method_details
    // Insert into: test_methods (if test class)
    // Build: relationships
}
```

---

## 📊 Phase 2: Framework Detection & Patterns (Week 2)

### Framework Detector

**Goal**: Know EXACTLY which of 6 combinations

```go
func (fd *FrameworkDetector) Detect() *FrameworkConfig {
    deps := fd.parsePOM()
    
    hasSelenium := deps.Has("selenium-java")
    hasTestNG := deps.Has("testng")
    hasJUnit := deps.Has("junit-jupiter")
    hasCucumber := deps.Has("cucumber")
    hasSerenity := deps.Has("serenity")
    hasRestAssured := deps.Has("rest-assured")
    
    hasFeatureFiles := fd.scan("**/*.feature")
    hasStepDefs := fd.scan("**/*Steps.java")
    
    // Determine combination
    if hasSelenium && hasTestNG && hasCucumber {
        return "selenium-testng-cucumber"
    } else if hasSelenium && hasTestNG {
        return "selenium-testng"
    } else if hasSelenium && hasJUnit && hasCucumber {
        return "selenium-junit-cucumber"
    }
    // ... all 6 combinations
}
```

### Pattern Detector

**Goal**: Learn project coding style

```go
func (pd *PatternDetector) DetectPatterns(projectID int) *CodingPatterns {
    // 1. Naming conventions
    testMethods := db.GetAllTestMethods()
    if matchPattern(testMethods, `^test[A-Z].*With[A-Z]`) {
        patterns.Naming.TestMethods = "testActionWithCondition"
    }
    
    pageObjects := db.GetAllPageObjects()
    if allEndWith(pageObjects, "Page") {
        patterns.Naming.PageObjects = "PascalCasePage"
    }
    
    // 2. Assertion style
    assertions := db.GetAllAssertions()
    testngCount := count(assertions, "Assert.")
    junitCount := count(assertions, "Assertions.")
    patterns.Assertions.Library = testngCount > junitCount ? "testng" : "junit"
    
    withMessages := count(assertions, assertion => assertion.Message != "")
    patterns.Assertions.AlwaysHasMessage = withMessages / len(assertions) > 0.9
    
    // 3. Wait strategy
    methods := db.GetAllMethods()
    explicitWaits := count(methods, m => m.WaitType == "explicit")
    patterns.Waits.Type = "explicit"
    patterns.Waits.Timeout = extractCommonTimeout(methods)
    
    // 4. Structure
    testsWithAAA := count(testMethods, t => t.FollowsAAAPattern)
    patterns.Structure.UsesAAA = testsWithAAA / len(testMethods) > 0.8
    
    return patterns
}
```

**Store in database**:
```sql
INSERT INTO coding_patterns VALUES
    ('naming', 'test_methods', '{"pattern":"testActionWithCondition"}', 0.95, 22),
    ('naming', 'page_objects', '{"pattern":"PascalCasePage"}', 1.0, 15),
    ('assertions', 'style', '{"library":"testng","always_has_message":true}', 1.0, 45),
    ('waits', 'strategy', '{"type":"explicit","timeout":10,"unit":"seconds"}', 0.92, 38),
    ('structure', 'pattern', '{"AAA":true,"comments":true}', 0.88, 22);
```

---

## 🕸️ Phase 3: Knowledge Graph & Query System (Week 3)

### Build Relationship Graph

```go
type KnowledgeGraph struct {
    Classes  map[string]*ClassNode
    Methods  map[string]*MethodNode
    Fields   map[string]*FieldNode
    Relationships []Relationship
}

func (kg *KnowledgeGraph) BuildRelationships() {
    // For each class:
    //   - extends → relationship
    //   - uses (imports) → relationship
    //   - creates (new XxxPage) → relationship
    //   - tests (test class → page object) → relationship
    
    // For each method:
    //   - calls → method_calls table
    //   - uses_field → method_details
    //   - creates_object → method_details
    
    // For step definitions:
    //   - maps_to_feature_step → step_mappings
}
```

**Example queries the graph can answer**:
```go
// Find all tests using LoginPage
kg.FindTestsUsingPageObject("LoginPage")
// Returns: ["LoginTest"]

// Find all methods in LoginPage
kg.FindMethodsInClass("LoginPage")
// Returns: [login, enterUsername, enterPassword, ...]

// Find all elements in LoginPage
kg.FindElementsInPageObject("LoginPage")
// Returns: [usernameField, passwordField, loginButton, errorMessage]

// What does login() call?
kg.GetMethodCalls("LoginPage", "login")
// Returns: [enterUsername, enterPassword, clickLoginButton]

// What field does enterUsername() use?
kg.GetMethodFieldUsage("LoginPage", "enterUsername")
// Returns: [usernameField, wait]

// Where is username element defined?
kg.FindFieldDefinition("usernameField")
// Returns: {file: "LoginPage.java", line: 15, locator: "id=username"}
```

### Query Engine

**Goal**: Answer ANY question about the codebase

```go
type QueryEngine struct {
    db *Database
    kg *KnowledgeGraph
}

// Example: User asks "show me all login tests"
func (qe *QueryEngine) FindTests(keyword string) []TestInfo {
    // 1. Search index for keyword
    matches := qe.db.SearchIndex(keyword)
    
    // 2. Filter for test methods
    tests := filterByType(matches, "test")
    
    // 3. Get full info for each
    var results []TestInfo
    for _, match := range tests {
        test := qe.db.GetTestMethod(match.ID)
        test.PageObjects = qe.kg.GetUsedPageObjects(test.ID)
        test.Elements = qe.kg.GetUsedElements(test.ID)
        test.FullCode = qe.db.GetFileContent(test.FilePath)
        results = append(results, test)
    }
    
    return results
}

// Example: "where is the username element?"
func (qe *QueryEngine) FindElement(elementKeyword string) *ElementInfo {
    // 1. Search for element fields
    matches := qe.db.SearchFields(elementKeyword)
    
    // 2. Filter for @FindBy fields
    elements := filterByAnnotation(matches, "FindBy")
    
    // 3. Get full info
    element := elements[0]
    info := &ElementInfo{
        Name: element.Name,
        LocatorType: element.LocatorStrategy,
        LocatorValue: element.LocatorValue,
        FilePath: element.FilePath,
        Line: element.Line,
        UsedInMethods: qe.kg.FindMethodsUsingField(element.ID),
        UsedInTests: qe.kg.FindTestsUsingField(element.ID),
    }
    
    return info
}
```

### Search Index Builder

```go
func (si *SearchIndexBuilder) BuildIndex(projectID int) {
    // Index all classes
    classes := db.GetAllClasses(projectID)
    for _, class := range classes {
        keywords := extractKeywords(class.Name)
        for _, keyword := range keywords {
            si.AddToIndex(keyword, "class", class.ID, class.Name, class.FilePath)
        }
    }
    
    // Index all methods
    methods := db.GetAllMethods(projectID)
    for _, method := range methods {
        keywords := extractKeywords(method.Name)
        for _, keyword := range keywords {
            si.AddToIndex(keyword, "method", method.ID, method.Name, method.FilePath)
        }
    }
    
    // Index all fields
    fields := db.GetAllFields(projectID)
    for _, field := range fields {
        keywords := extractKeywords(field.Name)
        for _, keyword := range keywords {
            si.AddToIndex(keyword, "field", field.ID, field.Name, field.FilePath)
        }
    }
}

// Example: Search for "login"
func (si *SearchIndex) Search(keyword string) []SearchResult {
    results := db.Query(`
        SELECT entity_type, entity_name, file_path, line_number
        FROM search_index
        WHERE keyword LIKE ?
        ORDER BY relevance_score DESC
    `, "%"+keyword+"%")
    
    return results
}
```

---

## 🧩 Phase 4: Context Builder (Week 4)

### Goal: Assemble PERFECT context for ANY user query

```go
type ContextBuilder struct {
    db *Database
    kg *KnowledgeGraph
    qe *QueryEngine
}

func (cb *ContextBuilder) BuildContext(userMessage string, selectedTarget *Target) *Context {
    ctx := &Context{
        UserMessage: userMessage,
        Target: selectedTarget,
    }
    
    // 1. Main file (the target)
    ctx.MainFile = &FileContext{
        Path: selectedTarget.FilePath,
        Content: cb.db.GetFileContent(selectedTarget.FilePath),
        Structure: cb.db.GetClassStructure(selectedTarget.ClassName),
    }
    
    // 2. Dependencies
    ctx.Dependencies = cb.loadAllDependencies(selectedTarget)
    
    // 3. Framework info
    ctx.Framework = cb.db.GetFrameworkConfig()
    
    // 4. Coding patterns
    ctx.CodingStyle = cb.db.GetCodingPatterns()
    
    // 5. Similar code
    ctx.SimilarCode = cb.findSimilarCode(selectedTarget)
    
    // 6. Application Map (if needed)
    ctx.ApplicationMap = cb.loadApplicationMap(selectedTarget)
    
    return ctx
}

func (cb *ContextBuilder) loadAllDependencies(target *Target) []*Dependency {
    var deps []*Dependency
    
    // If target is a test class:
    if target.Type == "test_class" {
        // Find all page objects it uses
        pageObjects := cb.kg.GetUsedPageObjects(target.ID)
        for _, po := range pageObjects {
            deps = append(deps, &Dependency{
                Name: po.Name,
                Type: "page_object",
                FilePath: po.FilePath,
                FullCode: cb.db.GetFileContent(po.FilePath),
                Elements: cb.db.GetFieldsForClass(po.ID),
                Methods: cb.db.GetMethodsForClass(po.ID),
            })
        }
        
        // Include BaseTest
        if target.ExtendsClass != "" {
            baseTest := cb.db.GetClass(target.ExtendsClass)
            deps = append(deps, &Dependency{
                Name: baseTest.Name,
                Type: "base_test",
                FilePath: baseTest.FilePath,
                FullCode: cb.db.GetFileContent(baseTest.FilePath),
            })
        }
    }
    
    // If target is a page object:
    if target.Type == "page_object" {
        // Include BasePage
        if target.ExtendsClass != "" {
            basePage := cb.db.GetClass(target.ExtendsClass)
            deps = append(deps, &Dependency{
                Name: basePage.Name,
                Type: "base_page",
                FilePath: basePage.FilePath,
                FullCode: cb.db.GetFileContent(basePage.FilePath),
            })
        }
        
        // Include tests that use this page object
        tests := cb.kg.FindTestsUsingPageObject(target.Name)
        for _, test := range tests {
            deps = append(deps, &Dependency{
                Name: test.Name,
                Type: "related_test",
                FilePath: test.FilePath,
                Note: "Will be affected by changes to this page object",
            })
        }
    }
    
    return deps
}
```

**Final Context Package** (JSON sent to Cloud Backend):
```json
{
  "user_message": "edit login test",
  
  "target": {
    "type": "test_class",
    "name": "LoginTest",
    "file_path": "src/test/java/tests/LoginTest.java"
  },
  
  "main_file": {
    "path": "src/test/java/tests/LoginTest.java",
    "content": "... FULL CODE ...",
    "structure": {
      "class_name": "LoginTest",
      "extends": "BaseTest",
      "test_methods": [
        {
          "name": "testLoginWithValidCredentials",
          "lines": "12-28",
          "type": "positive",
          "groups": ["smoke", "regression"],
          "assertions": 2
        },
        {
          "name": "testLoginWithInvalidCredentials",
          "lines": "30-46",
          "type": "negative",
          "assertions": 2
        }
      ]
    }
  },
  
  "dependencies": [
    {
      "name": "LoginPage",
      "type": "page_object",
      "file_path": "src/test/java/pages/LoginPage.java",
      "content": "... FULL LoginPage CODE ...",
      "elements": [
        {
          "name": "usernameField",
          "locator": "id=username",
          "line": 15
        },
        {
          "name": "passwordField",
          "locator": "id=password",
          "line": 18
        }
      ],
      "methods": [
        {
          "name": "login",
          "params": ["String username", "String password"],
          "returns": "void",
          "lines": "56-60",
          "calls": ["enterUsername", "enterPassword", "clickLoginButton"]
        },
        {
          "name": "enterUsername",
          "params": ["String username"],
          "lines": "39-43",
          "uses_field": "usernameField",
          "wait_type": "explicit"
        }
      ]
    },
    {
      "name": "DashboardPage",
      "type": "page_object",
      "content": "... FULL CODE ..."
    },
    {
      "name": "BaseTest",
      "type": "base_test",
      "content": "... setup/teardown CODE ..."
    }
  ],
  
  "framework": {
    "type": "selenium-testng",
    "test_runner": "testng",
    "pattern": "page_object_model",
    "uses_bdd": false
  },
  
  "coding_style": {
    "naming": {
      "test_methods": "testActionWithCondition",
      "page_objects": "PascalCasePage",
      "variables": "camelCase"
    },
    "assertions": {
      "library": "testng",
      "always_has_message": true,
      "pattern": "Assert.assertTrue(condition, message)"
    },
    "waits": {
      "type": "explicit",
      "timeout": 10,
      "unit": "seconds"
    },
    "structure": {
      "pattern": "AAA",
      "has_comments": true
    }
  },
  
  "similar_code": [
    {
      "name": "DashboardTest.java",
      "relevance": 0.85,
      "reason": "Similar test structure"
    }
  ],
  
  "application_map": null
}
```

---

## 🎯 Success Criteria

### After Phase 1 (Week 1)
- ✅ Parse LoginPage.java → Extract ALL 5 @FindBy elements with locators
- ✅ Parse LoginTest.java → Extract ALL test methods with assertions
- ✅ Store in database → Query: "SELECT * FROM fields WHERE is_page_element = 1"
- ✅ Get results with locator_type and locator_value

### After Phase 2 (Week 2)
- ✅ Detect framework: "selenium-testng"
- ✅ Detect patterns: naming="testActionWithCondition", assertions="testng with messages"
- ✅ Query: "What's the naming convention?" → Get answer from database

### After Phase 3 (Week 3)
- ✅ Query: "Find all tests using LoginPage" → Return ["LoginTest"]
- ✅ Query: "Where is username element?" → Return "LoginPage.java line 15, locator: id=username"
- ✅ Query: "What does login() call?" → Return ["enterUsername", "enterPassword", "clickLoginButton"]

### After Phase 4 (Week 4)
- ✅ User: "edit login test"
- ✅ System builds COMPLETE context package
- ✅ Includes: LoginTest + LoginPage + DashboardPage + BaseTest (all full code)
- ✅ Includes: Framework info + Coding style + All relationships
- ✅ Ready to send to LLM

---

## 🚀 Next Step

**Start Phase 1, Day 1: Enhanced JavaParser**

Should I:
1. **Start coding the JavaParser** with Tree-sitter integration?
2. **Set up the database schema** first?
3. **Create a small test project** to parse?

What would you like to begin with?
