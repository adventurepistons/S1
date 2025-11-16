# 🚀 Implementation Roadmap - AI-Driven Framework Generator

## 📋 Current Status

### ✅ **Already Implemented (Foundation)**
- [x] Browser Recording Infrastructure
- [x] Element Analyzer
- [x] Java Parser (Tree-sitter)
- [x] Gherkin Parser
- [x] POM Parser (Maven)
- [x] TestNG Parser
- [x] Workspace Analyzer
- [x] Code Generator (basic)
- [x] LLM Client (OpenAI)
- [x] Prompt Builder

### ⏳ **New Components Needed**

---

## 🎯 Phase 1: Scenario Generation Engine (Week 1-2)

### **Component: `ScenarioGenerator`**

**Location:** `copilot-core/pkg/scenario/scenario_generator.go`

**Purpose:** Use GPT to generate all possible test scenarios from recording

**Interface:**
```go
type ScenarioGenerator struct {
    llmClient *llm.Client
}

type GenerateScenarioRequest struct {
    Recording    *RecordingSession  // From browser
    UserIntent   string              // "I want to test login"
    AppContext   *ApplicationContext // Optional: app type, domain
}

type TestScenarios struct {
    Positive     []Scenario `json:"positive"`
    Negative     []Scenario `json:"negative"`
    EdgeCases    []Scenario `json:"edgeCases"`
    Security     []Scenario `json:"security"`
}

type Scenario struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    TestData    map[string]interface{} `json:"testData"`
    Expected    ExpectedResult         `json:"expected"`
    Priority    string                 `json:"priority"` // high, medium, low
}

func (s *ScenarioGenerator) GenerateScenarios(req GenerateScenarioRequest) (*TestScenarios, error)
```

**GPT Prompt Template:**
```
You are a senior QA engineer. Based on this recording of a web form:

Elements:
- Username field (id: username, type: email, required: true)
- Password field (id: password, type: password, required: true)
- Remember Me checkbox (id: remember-me, optional)
- Login button (id: login-btn)

User flow:
1. Enter username: test@example.com
2. Enter password: Pass123!
3. Click login
4. Navigate to dashboard

Validation elements on success:
- Welcome message: "Welcome, Test User"
- Logout button visible

Generate ALL possible test scenarios including:
1. Positive scenarios (happy path)
2. Negative scenarios (invalid inputs)
3. Edge cases (boundary conditions)
4. Security tests (injection, XSS)

For each scenario, provide:
- Name
- Description
- Test data (input values)
- Expected result (pass/fail)
- Priority (high/medium/low)

Output as JSON matching this schema:
{
  "positive": [...],
  "negative": [...],
  "edgeCases": [...],
  "security": [...]
}
```

**Implementation Tasks:**
1. Create scenario_generator.go
2. Design GPT prompt for scenario generation
3. Parse GPT response into structured scenarios
4. Add scenario filtering (user can select which to generate)
5. Add scenario validation logic

**Testing:**
- Unit test with sample recording
- Test with various form types (login, registration, checkout)
- Verify scenario quality and coverage

---

## 🎯 Phase 2: Framework State Manager (Week 2-3)

### **Component: `FrameworkStateManager`**

**Location:** `copilot-core/pkg/framework/state_manager.go`

**Purpose:** Track existing framework state to enable smart reuse

**Interface:**
```go
type FrameworkStateManager struct {
    workspacePath string
    analyzer      *analyzer.WorkspaceAnalyzer
    state         *FrameworkState
}

type FrameworkState struct {
    PageObjects     map[string]*PageObjectState   // key: class name
    TestCases       map[string]*TestCaseState
    StepDefinitions map[string]*StepDefState
    FeatureFiles    map[string]*FeatureState
    LastUpdated     time.Time
}

type PageObjectState struct {
    ClassName    string
    FilePath     string
    Elements     []ElementState
    Methods      []MethodState
    LastModified time.Time
}

type ElementState struct {
    Name         string
    LocatorType  string
    LocatorValue string
    SourceType   string // "recorded" or "manual"
}

// Core methods
func (m *FrameworkStateManager) LoadState() (*FrameworkState, error)
func (m *FrameworkStateManager) SaveState() error
func (m *FrameworkStateManager) FindPageObject(name string) (*PageObjectState, bool)
func (m *FrameworkStateManager) HasElement(pageObject, elementName string) bool
func (m *FrameworkStateManager) GetReusableComponents(recording *RecordingSession) *ReusableComponents
```

**State File:** `.copilot-state.json` (gitignored)

**Example State:**
```json
{
  "pageObjects": {
    "LoginPage": {
      "className": "LoginPage",
      "filePath": "src/main/java/pages/LoginPage.java",
      "elements": [
        {
          "name": "usernameField",
          "locatorType": "id",
          "locatorValue": "username",
          "sourceType": "recorded"
        }
      ],
      "methods": [
        {
          "name": "login",
          "parameters": ["String username", "String password", "boolean rememberMe"]
        }
      ],
      "lastModified": "2025-11-15T10:30:00Z"
    }
  },
  "lastUpdated": "2025-11-15T10:30:00Z"
}
```

**Implementation Tasks:**
1. Create state_manager.go
2. Implement state loading/saving
3. Create state diff algorithm
4. Build reusable component detector
5. Add state versioning

---

## 🎯 Phase 3: Smart Code Merger (Week 3-4)

### **Component: `CodeMerger`**

**Location:** `copilot-core/pkg/merger/code_merger.go`

**Purpose:** Intelligently merge new code with existing framework

**Interface:**
```go
type CodeMerger struct {
    stateManager *framework.FrameworkStateManager
    javaParser   *parser.JavaParser
}

type MergeRequest struct {
    NewCode          *GeneratedCode
    ExistingFilePath string
    Strategy         MergeStrategy // append, update, skip
}

type MergeStrategy string

const (
    MergeStrategyAppend MergeStrategy = "append"  // Add new elements/methods
    MergeStrategyUpdate MergeStrategy = "update"  // Replace existing
    MergeStrategySkip   MergeStrategy = "skip"    // Don't change
)

type MergeResult struct {
    Action       string // "created", "updated", "skipped"
    FilePath     string
    Changes      []Change
    MergedCode   string
}

type Change struct {
    Type        string // "element_added", "method_added", "element_updated"
    Location    string // Line number or element name
    Description string
}

func (m *CodeMerger) MergePageObject(req MergeRequest) (*MergeResult, error)
func (m *CodeMerger) MergeTestCase(req MergeRequest) (*MergeResult, error)
func (m *CodeMerger) MergeFeatureFile(req MergeRequest) (*MergeResult, error)
```

**Merge Logic for Page Objects:**

```go
func (m *CodeMerger) MergePageObject(req MergeRequest) (*MergeResult, error) {
    // 1. Parse existing file
    existing, err := m.javaParser.ParseFile(req.ExistingFilePath)
    if err != nil {
        return nil, err
    }

    result := &MergeResult{
        FilePath: req.ExistingFilePath,
        Changes:  []Change{},
    }

    // 2. Extract new elements from generated code
    newElements := extractElements(req.NewCode)
    newMethods := extractMethods(req.NewCode)

    // 3. Merge elements
    for _, newElem := range newElements {
        if !hasElement(existing, newElem) {
            // ADD new element
            existing.Fields = append(existing.Fields, newElem)
            result.Changes = append(result.Changes, Change{
                Type:        "element_added",
                Location:    newElem.Name,
                Description: fmt.Sprintf("Added %s element", newElem.Name),
            })
        }
    }

    // 4. Merge methods
    for _, newMethod := range newMethods {
        if !hasMethod(existing, newMethod) {
            // ADD new method
            existing.Methods = append(existing.Methods, newMethod)
            result.Changes = append(result.Changes, Change{
                Type:        "method_added",
                Location:    newMethod.Name,
                Description: fmt.Sprintf("Added %s() method", newMethod.Name),
            })
        }
    }

    // 5. Generate updated code
    result.MergedCode = generateJavaCode(existing)
    result.Action = "updated"

    return result, nil
}
```

**Implementation Tasks:**
1. Create code_merger.go
2. Implement element/method extraction
3. Build smart diff algorithm
4. Add conflict detection
5. Create merge preview (show user what will change)

---

## 🎯 Phase 4: Test Data Generator (Week 4-5)

### **Component: `TestDataGenerator`**

**Location:** `copilot-core/pkg/testdata/data_generator.go`

**Purpose:** Generate test data for DataProvider from scenarios

**Interface:**
```go
type TestDataGenerator struct {
    scenarios *scenario.TestScenarios
}

type DataProviderSpec struct {
    MethodName  string
    Parameters  []Parameter
    DataRows    []DataRow
}

type Parameter struct {
    Name string
    Type string // "String", "int", "boolean"
}

type DataRow struct {
    Values      []interface{}
    Description string
}

// Generate TestNG DataProvider code
func (g *TestDataGenerator) GenerateDataProvider(scenarios *scenario.TestScenarios) (*DataProviderSpec, error)

// Generate CSV file
func (g *TestDataGenerator) GenerateCSV(scenarios *scenario.TestScenarios, filePath string) error

// Generate Excel file
func (g *TestDataGenerator) GenerateExcel(scenarios *scenario.TestScenarios, filePath string) error
```

**Example Output:**

```java
@DataProvider(name = "loginScenarios")
public Object[][] getLoginData() {
    return new Object[][] {
        // {username, password, rememberMe, shouldSucceed, scenario}
        {"test@example.com", "Pass123!", true, true, "Valid login"},
        {"test@example.com", "wrong", false, false, "Invalid password"},
        {"", "Pass123!", false, false, "Empty username"},
        {"admin' OR '1'='1", "test", false, false, "SQL injection"},
    };
}
```

**Implementation Tasks:**
1. Create data_generator.go
2. Build DataProvider code generator
3. Add CSV export
4. Add Excel export (using excelize library)
5. Support SQL integration (future)

---

## 🎯 Phase 5: Enhanced Code Generator (Week 5-6)

### **Component: Enhanced `CodeGenerator`**

**Location:** `copilot-core/pkg/generator/generator.go` (enhance existing)

**New Features:**

1. **Data-Driven Test Generation:**
```go
func (g *CodeGenerator) GenerateDataDrivenTest(
    scenarios *scenario.TestScenarios,
    pageObjects []string,
) (*GeneratedCode, error)
```

2. **Complete Test Class:**
```go
func (g *CodeGenerator) GenerateCompleteTestClass(req CompleteTestRequest) (*GeneratedCode, error)

type CompleteTestRequest struct {
    FeatureName     string
    Scenarios       *scenario.TestScenarios
    PageObjects     []string
    ValidationPage  string
    Framework       string // "testng" or "junit"
}
```

**Generated Test Structure:**
```java
package tests;

import org.testng.annotations.*;
import org.testng.Assert;
import pages.*;

public class LoginTest extends BaseTest {

    private LoginPage loginPage;
    private DashboardPage dashboardPage;

    @BeforeMethod
    public void setup() {
        driver.get("https://app.example.com/login");
        loginPage = new LoginPage(driver);
    }

    @DataProvider(name = "loginScenarios")
    public Object[][] getLoginData() {
        // Auto-generated from scenarios
    }

    @Test(dataProvider = "loginScenarios")
    public void testLogin(String username, String password,
                          boolean rememberMe, boolean shouldSucceed,
                          String scenario) {
        // Auto-generated test logic
        loginPage.login(username, password, rememberMe);
        dashboardPage = new DashboardPage(driver);

        if (shouldSucceed) {
            Assert.assertTrue(dashboardPage.isLoggedIn(), scenario);
        } else {
            Assert.assertTrue(loginPage.hasError(), scenario);
        }
    }
}
```

**Implementation Tasks:**
1. Enhance GenerateRequest struct
2. Add data-driven test template
3. Integrate with TestDataGenerator
4. Add validation logic generation
5. Support both TestNG and JUnit (TestNG first)

---

## 🎯 Phase 6: Conversational Interface (Week 6-7)

### **Component: `ConversationalEngine`**

**Location:** `copilot-core/pkg/conversation/engine.go`

**Purpose:** Handle natural language requests and map to actions

**Interface:**
```go
type ConversationalEngine struct {
    llmClient       *llm.Client
    stateManager    *framework.FrameworkStateManager
    recordingStore  *RecordingStore
    scenarioGen     *scenario.ScenarioGenerator
    codeGen         *generator.CodeGenerator
    merger          *merger.CodeMerger
}

type UserRequest struct {
    Message       string
    WorkspacePath string
    Context       map[string]interface{}
}

type ActionPlan struct {
    Intent          string // "create_test", "modify_test", "add_scenario"
    TargetFeature   string // "login", "checkout"
    RecordingsNeeded []string
    Actions         []Action
}

type Action struct {
    Type        string // "generate_page_object", "generate_test", "merge_code"
    Description string
    Parameters  map[string]interface{}
}

func (e *ConversationalEngine) ProcessRequest(req UserRequest) (*ActionPlan, error)
func (e *ConversationalEngine) ExecutePlan(plan *ActionPlan) (*ExecutionResult, error)
```

**Example Flow:**

```go
// User: "I want to test login"
request := UserRequest{
    Message: "I want to test login",
    WorkspacePath: "/path/to/project",
}

// 1. Parse intent
plan, err := engine.ProcessRequest(request)
// plan.Intent = "create_test"
// plan.TargetFeature = "login"
// plan.Actions = [
//   {Type: "find_recording", Parameters: {keyword: "login"}},
//   {Type: "generate_scenarios"},
//   {Type: "generate_page_objects"},
//   {Type: "generate_tests"},
// ]

// 2. Execute plan
result, err := engine.ExecutePlan(plan)
// result.FilesCreated = ["LoginPage.java", "LoginTest.java", "login.feature"]
// result.FilesUpdated = []
// result.Summary = "Created login test with 8 scenarios"
```

**Intent Detection Prompt:**
```
You are an AI assistant for test automation. Parse this user request and determine the intent.

User: "I want to test login functionality"

Possible intents:
- create_test: User wants to create a new test
- modify_test: User wants to change existing test
- add_scenario: User wants to add more test scenarios
- fix_test: User wants to fix failing test
- explain_code: User wants explanation

Output JSON:
{
  "intent": "create_test",
  "feature": "login",
  "confidence": 0.95
}
```

**Implementation Tasks:**
1. Create conversation engine
2. Build intent parser (GPT-powered)
3. Create action planner
4. Implement execution engine
5. Add user confirmation step

---

## 🎯 Phase 7: VS Code Integration (Week 7-8)

### **Component: Chat Interface UI**

**Location:** `src/chat/ChatPanel.ts`

**UI Features:**

1. **Chat Interface:**
```typescript
class ChatPanel {
    private messages: Message[] = [];

    async sendMessage(userMessage: string) {
        // Show user message
        this.addMessage({
            role: 'user',
            content: userMessage,
        });

        // Call Go binary
        const response = await vscode.commands.executeCommand(
            'copilot.processChat',
            userMessage
        );

        // Show AI response
        this.addMessage({
            role: 'assistant',
            content: response.message,
            actions: response.actions, // Show generated files
        });
    }
}
```

2. **Action Preview:**
```typescript
// Show user what will be generated/modified
interface ActionPreview {
    action: string; // "create", "update"
    files: FileChange[];
}

interface FileChange {
    path: string;
    type: 'create' | 'update';
    diff?: string; // For updates, show diff
    preview: string; // Code preview
}
```

3. **User Confirmation:**
```typescript
async confirmGeneration(preview: ActionPreview) {
    const result = await vscode.window.showQuickPick(
        ['Generate all files', 'Select files to generate', 'Cancel'],
        { placeHolder: 'AI will create/update the following files:' }
    );

    if (result === 'Generate all files') {
        await this.executeGeneration(preview);
    }
}
```

**Implementation Tasks:**
1. Create ChatPanel webview
2. Add message rendering
3. Build action preview UI
4. Add file diff viewer
5. Implement user confirmation flow

---

## 📊 Complete Implementation Timeline

### **Week 1-2: Scenario Generation**
- [ ] ScenarioGenerator component
- [ ] GPT prompt engineering
- [ ] Scenario validation
- [ ] Unit tests

### **Week 3-4: Framework State & Merging**
- [ ] FrameworkStateManager
- [ ] CodeMerger component
- [ ] Merge preview
- [ ] Conflict detection

### **Week 4-5: Test Data**
- [ ] TestDataGenerator
- [ ] DataProvider generation
- [ ] CSV/Excel export
- [ ] Data validation

### **Week 5-6: Enhanced Generator**
- [ ] Data-driven test templates
- [ ] Complete test class generation
- [ ] Validation logic
- [ ] Integration with scenarios

### **Week 6-7: Conversational Engine**
- [ ] Intent parser
- [ ] Action planner
- [ ] Execution engine
- [ ] User confirmation

### **Week 7-8: VS Code UI**
- [ ] Chat interface
- [ ] Action preview
- [ ] File diff viewer
- [ ] User confirmation flow

---

## 🎯 Success Criteria

**Phase 1 Complete When:**
- ✅ User records login → AI generates 10+ scenarios
- ✅ Scenarios cover positive, negative, edge cases, security
- ✅ User can select which scenarios to generate

**MVP Complete When:**
- ✅ User: "I want to test login"
- ✅ System: Generates complete framework (Page Objects + Tests + Features + Data)
- ✅ All code compiles and runs
- ✅ Smart reuse: Second feature reuses existing code
- ✅ Modification: User can change tests conversationally

**Production Ready When:**
- ✅ Handles 20+ recordings
- ✅ Generates 100+ test scenarios
- ✅ 80%+ code reuse rate
- ✅ < 5 min from "I want to test X" to runnable tests

---

## 🚀 Next Steps

**Immediate:** Start with **Phase 1: Scenario Generation**

This is the foundation. Without good scenarios, the rest doesn't matter.

**First Task:** Create `scenario_generator.go` and GPT prompt template.

Ready to start? 🎯
