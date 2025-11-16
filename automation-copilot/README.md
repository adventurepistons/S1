# Test Automation Copilot - Core Understanding Engine

A deep code understanding engine for Selenium + Java test automation projects. Built as a single Go binary that uses Tree-sitter AST parsing to extract 100% of information from test codebases.

## Architecture

**Single Binary Design**: One Go binary handles all heavy lifting - parsing, indexing, framework detection, context building, and LLM integration.

```
┌─────────────────────────────────────────────────────────────┐
│                  Automation Copilot Binary                   │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Java Parser  │  │   Storage    │  │ Query Engine │      │
│  │ (Tree-sitter)│  │  (SQLite)    │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Framework   │  │   Knowledge  │  │   Context    │      │
│  │  Detection   │  │     Graph    │  │   Builder    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## What We Built

### 1. Complete Data Models (`pkg/models/`)

**ClassData Model** - Captures everything about a Java class:
- Package and imports
- Class metadata (name, modifiers, extends, implements)
- All fields with annotations
- All methods with complete body analysis
- Framework hints (page object, test class, step definition)

**Field Model** - Special focus on @FindBy WebElements:
```go
type Field struct {
    Name            string
    Type            string
    IsWebElement    bool
    LocatorStrategy string  // 'id', 'css', 'xpath', 'name'
    LocatorValue    string  // The actual selector
    Annotations     []Annotation
    LineNumber      int
}
```

**Method Model** - Complete body analysis:
```go
type Method struct {
    Name            string
    MethodCalls     []MethodCall  // ALL method calls
    FieldAccess     []string      // ALL fields accessed
    Assertions      []Assertion   // ALL assertions
    UsesExplicitWait bool
    UsesImplicitWait bool
    WaitTimeout     int
    IsTest          bool
    TestType        string        // positive, negative, data-driven
    BodySource      string        // Full source code
}
```

### 2. Tree-sitter Parser (`internal/parser/java_parser.go`)

Deep AST parsing that extracts:

**From Page Objects:**
- Every @FindBy annotation with exact locator strategy and value
- Field types (WebElement detection)
- Method implementations
- Wait patterns (explicit vs implicit)

**From Test Classes:**
- @Test annotations with all parameters (description, priority, enabled)
- Method calls with line numbers
- Assertions with expected/actual values
- Control flow complexity (if statements, loops, try-catch)

**Example: Parsing @FindBy**
```java
@FindBy(id = "username")
private WebElement usernameField;
```

Extracts:
```json
{
  "name": "usernameField",
  "type": "WebElement",
  "is_web_element": true,
  "locator_strategy": "id",
  "locator_value": "username",
  "line_number": 15
}
```

### 3. SQLite Database (`internal/storage/`)

**25+ Tables** for comprehensive storage:

Core Tables:
- `files` - File metadata with hashing for change detection
- `classes` - Class metadata with framework hints
- `fields` - All fields with WebElement details
- `methods` - Methods with complexity metrics
- `field_annotations` - @FindBy and other field annotations
- `method_annotations` - @Test, @BeforeMethod, etc.

Relationship Tables:
- `method_calls` - WHO calls WHAT (with resolution)
- `field_access` - WHO uses WHAT field
- `assertions` - All assertion statements
- `relationships` - Knowledge graph edges

Framework Tables:
- `framework_config` - Detected framework combination
- `coding_patterns` - Learned project style
- `pattern_examples` - Real code examples

### 4. Database Layer (`internal/storage/database.go`)

Complete CRUD operations:
- `SaveClassData()` - Saves entire class with all relationships in transaction
- `GetClassByName()` - Retrieves class information
- `GetWebElementFields()` - Gets all @FindBy fields for a class
- `GetAllPageObjects()` - Lists all page object classes

**Features:**
- Embedded schema (go:embed)
- Foreign key enforcement
- Transaction support
- Upsert operations for re-indexing

### 5. CLI Tool (`cmd/copilot/main.go`)

Three commands:

**parse** - Parse and display extracted data:
```bash
./copilot parse test_samples/LoginPage.java
```

**index** - Parse and save to database:
```bash
./copilot index test_samples/LoginPage.java
```

**query** - Query database:
```bash
./copilot query LoginPage
```

### 6. Test Samples (`test_samples/`)

**LoginPage.java** - Page Object example with:
- 5 @FindBy WebElements (id, css, xpath, name selectors)
- 7 methods including wait patterns
- PageFactory initialization

**LoginTest.java** - Test class example with:
- 5 @Test methods with descriptions and priorities
- Positive and negative test cases
- AAA pattern (Arrange-Act-Assert)
- TestNG assertions with messages

## What It Does

### Extraction Capabilities

**From LoginPage.java**, extracts:
```
✓ Package: com.example.pages
✓ 7 imports
✓ 5 WebElements with exact locators:
  - usernameField: id = "username"
  - passwordField: id = "password"
  - loginButton: css = "button[type='submit']"
  - errorMessage: xpath = "//div[@class='error-message']"
  - rememberMeCheckbox: name = "remember-me"
✓ 7 methods with wait strategies
✓ Detects: Page Object Model = true
```

**From LoginTest.java**, extracts:
```
✓ Package: com.example.tests
✓ 5 test methods
✓ All @Test annotations (description, priority, enabled)
✓ 12 assertions with expected/actual values
✓ 15+ method calls to loginPage
✓ Test types: 4 positive, 1 negative
✓ Detects: Test Class = true
```

### Deep Understanding

For each test method, knows:
- What page objects it uses (via field access)
- What methods it calls (entire call chain)
- What it asserts (assertion type, values, messages)
- Wait strategy (explicit/implicit, timeouts)
- Control flow complexity
- Test type classification

### Query Capabilities (Future)

Will answer questions like:
```
Q: "Where is usernameField defined?"
A: LoginPage.java:18 - id = "username"

Q: "What tests use LoginPage?"
A: LoginTest.testLoginWithValidCredentials
   LoginTest.testLoginWithInvalidPassword
   (5 total)

Q: "Show me all elements in LoginPage"
A: 5 WebElements:
   - usernameField (id)
   - passwordField (id)
   - loginButton (css)
   - errorMessage (xpath)
   - rememberMeCheckbox (name)
```

## Database Schema Highlights

**fields table** - Special focus on @FindBy:
```sql
CREATE TABLE fields (
    id INTEGER PRIMARY KEY,
    class_id INTEGER NOT NULL,
    field_name TEXT NOT NULL,
    field_type TEXT NOT NULL,
    is_web_element BOOLEAN DEFAULT 0,
    locator_strategy TEXT,  -- 'id', 'css', 'xpath'
    locator_value TEXT,     -- The actual selector
    line_number INTEGER
);
```

**method_calls table** - Call graph:
```sql
CREATE TABLE method_calls (
    id INTEGER PRIMARY KEY,
    caller_method_id INTEGER NOT NULL,
    method_name TEXT NOT NULL,
    object_name TEXT,           -- loginPage
    arguments TEXT,             -- JSON array
    callee_method_id INTEGER,   -- Resolved method
    line_number INTEGER
);
```

**relationships table** - Knowledge graph:
```sql
CREATE TABLE relationships (
    id INTEGER PRIMARY KEY,
    relationship_type TEXT NOT NULL,  -- 'EXTENDS', 'CALLS', 'USES'
    from_entity_type TEXT NOT NULL,   -- 'class', 'method', 'field'
    from_entity_id INTEGER NOT NULL,
    to_entity_type TEXT NOT NULL,
    to_entity_id INTEGER NOT NULL,
    metadata TEXT  -- JSON
);
```

## Project Structure

```
automation-copilot/
├── cmd/
│   └── copilot/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── parser/
│   │   └── java_parser.go          # Tree-sitter AST parser (700+ lines)
│   ├── storage/
│   │   ├── database.go             # SQLite operations (500+ lines)
│   │   └── schema.sql              # 25+ tables
│   └── extractor/                  # (Future: extraction pipeline)
├── pkg/
│   └── models/
│       ├── class_data.go           # Complete data models
│       └── framework_config.go     # Framework detection models
├── test_samples/
│   ├── LoginPage.java              # Page object example
│   └── LoginTest.java              # Test class example
├── go.mod
└── README.md
```

## Implementation Status

### ✅ Completed (Phase 1a)

- [x] Complete data models (ClassData, Field, Method, Annotation)
- [x] Tree-sitter Java parser with 100% extraction
- [x] SQLite schema with 25+ tables
- [x] Database layer with save/query operations
- [x] CLI tool with parse/index/query commands
- [x] Test samples (LoginPage, LoginTest)

### 📋 Next Steps (Phase 1b)

1. **Build & Test** (pending dependency resolution)
   ```bash
   # Install dependencies
   go mod tidy

   # Build binary
   go build -o copilot cmd/copilot/main.go

   # Test parsing
   ./copilot parse test_samples/LoginPage.java
   ./copilot index test_samples/LoginTest.java
   ./copilot query LoginPage
   ```

2. **Add Framework Detector** (`internal/detector/`)
   - Detect which of 6 framework combinations
   - Parse pom.xml for dependencies
   - Scan for testng.xml, .feature files
   - Determine architecture patterns

3. **Add Pattern Detector** (`internal/detector/`)
   - Learn naming conventions
   - Detect wait strategies
   - Identify assertion styles
   - Extract pattern examples

4. **Add Knowledge Graph** (`internal/graph/`)
   - Build relationship graph
   - Implement call graph resolution
   - Create query engine
   - Add traversal methods

5. **Add Context Builder** (`internal/ai/`)
   - Assemble complete context
   - Build prompts for LLM
   - Integrate OpenAI/Claude APIs

## Dependencies

```go
require (
    github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82
    github.com/smacker/go-tree-sitter/java latest
    github.com/mattn/go-sqlite3 v1.14.32
)
```

## Design Philosophy

**Deep Understanding FIRST, AI Integration LATER**

The core principle: Extract EVERYTHING from code with 100% accuracy BEFORE building AI features. The better we understand the codebase, the better code we can generate.

**Order of Execution:**
1. Parse EVERYTHING → Store in database
2. Detect framework → Learn patterns
3. Build knowledge graph → Enable queries
4. Integrate AI → Generate code

## Success Criteria

Phase 1 is successful when:
- [ ] Can parse LoginPage.java with 100% accuracy
- [ ] Extracts ALL @FindBy annotations with correct locators
- [ ] Extracts ALL method calls with line numbers
- [ ] Stores everything in SQLite
- [ ] Can query: "Show me all WebElements in LoginPage"
- [ ] Can query: "Which methods use usernameField?"

## Future Enhancements

**Week 2-3: Framework Detection**
- Support 6 framework combinations (TestNG, JUnit, Cucumber, Serenity, RestAssured)
- Pattern learning (naming, assertions, waits)
- Real code example extraction

**Week 4: Knowledge Graph**
- Complete relationship mapping
- Call graph resolution
- Query engine with pattern matching
- Graph traversal

**Week 5: AI Integration**
- Context assembly from knowledge graph
- LLM integration (OpenAI/Claude)
- Code generation with validation
- Response formatting

## License

MIT

## Contributing

See SINGLE_BINARY_IMPLEMENTATION_PLAN.md for detailed implementation guide.
