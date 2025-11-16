# 🏗️ Architecture V2 - Framework Understanding First, AI Second

**Strategic Principle**: Master context building and framework understanding at 100%, THEN layer AI on top

**Monetization**: Free credits for all users (no separate free tier), AI from day one

---

## 🎯 Architecture Boundaries

### **Extension (TypeScript - VS Code UI)**

**Responsibility**: User interface and VS Code integration

**What Lives Here**:
```
✅ Webview UI (chat interface)
✅ User input capture
✅ File tree display
✅ Diff preview UI
✅ File read/write operations (via VS Code API)
✅ WebSocket client (for streaming)
✅ Configuration UI
✅ Application Map UI
✅ Command palette integration
```

**What Does NOT Live Here**:
```
❌ Parsing logic
❌ Framework detection
❌ Context building
❌ Code generation
❌ LLM calls
❌ Business logic
```

**Size**: ~5,000 lines TypeScript

---

### **Go Binary (Local - Heavy Lifting)**

**Responsibility**: Workspace understanding, framework mastery, context building

**What Lives Here**:
```
✅ Workspace indexing (SQLite storage)
✅ Tree-sitter parsing (Java, Gherkin, XML, properties)
✅ Framework detection (TestNG, JUnit, Cucumber, Serenity, Rest-Assured)
✅ AST analysis (extract classes, methods, annotations)
✅ Dependency analysis (POM.xml, build.gradle)
✅ Page Object detection and extraction
✅ Test case mapping
✅ Step definition extraction (Cucumber)
✅ Application Map storage
✅ Context builder (assemble relevant code for LLM)
✅ Pattern extraction (naming conventions, coding style)
✅ File structure analysis
✅ WebSocket server (for streaming to extension)
✅ Local vector embeddings (chromem-go)
✅ Semantic search
```

**What Does NOT Live Here**:
```
❌ LLM API calls
❌ Prompt engineering (templates only)
❌ Usage tracking
❌ Payment/billing
❌ User authentication
```

**Size**: ~15,000 lines Go (already ~80% built)

---

### **Cloud Backend (Node.js/Python - AI Layer)**

**Responsibility**: LLM integration, prompt orchestration, usage tracking

**What Lives Here**:
```
✅ OpenAI/Claude API integration
✅ Prompt assembly (using context from Go binary)
✅ Response streaming
✅ Token counting and usage tracking
✅ Free credit management
✅ User authentication
✅ Rate limiting
✅ Prompt caching coordination
✅ LLM provider abstraction (OpenAI, Claude, DeepSeek)
✅ Advanced prompt techniques (HyDE, Self-RAG) - Phase 2
✅ Analytics and logging
```

**What Does NOT Live Here**:
```
❌ Code parsing
❌ Framework detection
❌ File operations
❌ Context building (receives pre-built context from Go)
```

**Size**: ~3,000 lines Node.js/TypeScript

---

## 📊 Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         VS Code Extension                        │
│  - User types: "create login test"                              │
│  - Shows chat UI, displays response                             │
└────────────────────────┬────────────────────────────────────────┘
                         │ (WebSocket)
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                         Go Binary (Local)                        │
│                                                                   │
│  1. Parse user message                                           │
│  2. Detect intent: "create_test"                                 │
│  3. Check existing code:                                         │
│     - Does LoginPage.java exist? YES                             │
│     - What methods? login(), enterUsername(), etc.               │
│     - Does LoginTest.java exist? NO (need to create)             │
│  4. Check Application Map:                                       │
│     - Login URL: https://app.com/login                           │
│     - Elements: username, password, submit button                │
│  5. Detect framework: TestNG + Page Object Model                 │
│  6. Build context package:                                       │
│     {                                                             │
│       "intent": "create_test",                                   │
│       "framework": "testng",                                     │
│       "pattern": "page_object_model",                            │
│       "existing_code": {                                         │
│         "page_object": "LoginPage.java (full code)",            │
│         "similar_tests": ["DashboardTest.java"],                │
│         "test_base": "BaseTest.java"                             │
│       },                                                          │
│       "application_map": {                                       │
│         "page": "login",                                         │
│         "url": "https://app.com/login",                          │
│         "elements": [...]                                        │
│       },                                                          │
│       "coding_style": {                                          │
│         "naming": "camelCase",                                   │
│         "assertions": "Assert.assertTrue with messages",         │
│         "waits": "WebDriverWait, 10 seconds"                     │
│       }                                                           │
│     }                                                             │
│  7. Send to cloud backend                                        │
└────────────────────────┬────────────────────────────────────────┘
                         │ (HTTPS API)
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Cloud Backend (API)                         │
│                                                                   │
│  1. Receive context package                                      │
│  2. Check user credits (deduct tokens)                           │
│  3. Build prompt:                                                │
│     - System: "You are Senior QA Engineer..."                   │
│     - Context: [Insert all code from Go binary]                 │
│     - Task: "Create LoginTest.java following style"             │
│  4. Call OpenAI API (GPT-4o)                                     │
│  5. Stream response back                                         │
│  6. Track usage                                                  │
└────────────────────────┬────────────────────────────────────────┘
                         │ (Stream back)
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                         Go Binary (Local)                        │
│                                                                   │
│  1. Receive generated code                                       │
│  2. Validate:                                                    │
│     - Syntax check (Java parser)                                 │
│     - Follows framework conventions                              │
│     - No anti-patterns (Thread.sleep, etc.)                      │
│  3. Determine file operation:                                    │
│     - Create new file: src/test/java/tests/LoginTest.java       │
│     OR                                                            │
│     - Edit existing file: Add test method to existing test      │
│  4. Generate diff preview                                        │
│  5. Send to extension                                            │
└────────────────────────┬────────────────────────────────────────┘
                         │ (WebSocket)
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                         VS Code Extension                        │
│                                                                   │
│  1. Show diff preview UI                                         │
│  2. User clicks "Accept"                                         │
│  3. Write file to disk (VS Code API)                             │
│  4. Show success message                                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🧩 Application Map Concept

**Problem**: AI doesn't know about your application's structure, pages, elements

**Solution**: User-defined Application Map (one-time setup)

### What is an Application Map?

```json
{
  "application": {
    "name": "E-Commerce App",
    "base_url": "https://myapp.com",
    "environments": {
      "dev": "https://dev.myapp.com",
      "staging": "https://staging.myapp.com",
      "prod": "https://myapp.com"
    }
  },
  "pages": [
    {
      "name": "Login",
      "url": "/login",
      "description": "User authentication page",
      "elements": [
        {
          "name": "usernameField",
          "locator": "id=username",
          "type": "input",
          "description": "Email or username input"
        },
        {
          "name": "passwordField",
          "locator": "id=password",
          "type": "input",
          "description": "Password input"
        },
        {
          "name": "loginButton",
          "locator": "css=button[type='submit']",
          "type": "button",
          "description": "Submit login form"
        },
        {
          "name": "errorMessage",
          "locator": "className=error-message",
          "type": "text",
          "description": "Error message display"
        }
      ],
      "actions": [
        "login(username, password) -> Dashboard",
        "displayError() -> stays on Login"
      ]
    },
    {
      "name": "Dashboard",
      "url": "/dashboard",
      "description": "Main user dashboard after login",
      "elements": [
        {
          "name": "welcomeMessage",
          "locator": "css=.welcome-text",
          "type": "text",
          "description": "Welcome message with username"
        },
        {
          "name": "logoutButton",
          "locator": "id=logout",
          "type": "button",
          "description": "Logout button"
        }
      ]
    },
    {
      "name": "ProductSearch",
      "url": "/products",
      "description": "Product search and browse page",
      "elements": [
        {
          "name": "searchBox",
          "locator": "name=search",
          "type": "input"
        },
        {
          "name": "searchButton",
          "locator": "css=button.search-btn",
          "type": "button"
        },
        {
          "name": "productList",
          "locator": "css=.product-item",
          "type": "list"
        }
      ]
    }
  ],
  "user_flows": [
    {
      "name": "Complete Purchase",
      "steps": [
        "Login -> Dashboard",
        "Dashboard -> ProductSearch",
        "ProductSearch -> ProductDetails",
        "ProductDetails -> Cart",
        "Cart -> Checkout",
        "Checkout -> OrderConfirmation"
      ]
    }
  ],
  "test_data": {
    "valid_users": [
      {"username": "testuser@example.com", "password": "Test@123"}
    ],
    "invalid_users": [
      {"username": "invalid@example.com", "password": "wrong"}
    ]
  }
}
```

### How to Create Application Map

**Option 1**: Manual creation (JSON file)
```
.testcopilot/
  application-map.json
```

**Option 2**: UI-based builder (Extension command)
```
Command: "Test Copilot: Create Application Map"
→ Opens form:
   - Page name: [Login]
   - URL: [/login]
   - Add element: [username] [id=username] [input]
   - Add element: [password] [id=password] [input]
   - Save
```

**Option 3**: Browser extension (future - record while browsing)
```
User browses app → Extension captures:
- Page URLs
- Element locators (auto-generated)
- User actions
→ Generates application-map.json
```

### How Application Map is Used

**Scenario 1**: User says "create login test"

```
Go Binary:
1. Check: Does LoginPage.java exist?
   - NO → Need to create page object first

2. Look up Application Map → "Login" page
   - Found elements: username, password, loginButton
   - Found actions: login(username, password) -> Dashboard

3. Build context for LLM:
   {
     "task": "create_page_object",
     "page": "Login",
     "elements": [...from Application Map],
     "framework": "Selenium + TestNG",
     "pattern": "Page Object Model"
   }

4. LLM generates LoginPage.java

5. Then create LoginTest.java using the new page object
```

**Scenario 2**: User says "create test for product search"

```
Go Binary:
1. Check: Does ProductSearchPage.java exist? YES
2. Check: Does ProductSearchTest.java exist? NO

3. Look up Application Map → "ProductSearch" page
   - Elements: searchBox, searchButton, productList

4. Build context:
   {
     "task": "create_test",
     "existing_page_object": "ProductSearchPage.java (full code)",
     "similar_tests": ["LoginTest.java"],
     "application_map": {...ProductSearch page},
     "framework": "TestNG"
   }

5. LLM generates ProductSearchTest.java
```

---

## 🔧 Framework Understanding - All Combinations

### Priority 1: Selenium + Java (6 combinations)

| Framework | Test Runner | BDD | Structure |
|-----------|-------------|-----|-----------|
| **1. Selenium + TestNG** | TestNG | No | Page Object Model |
| **2. Selenium + TestNG + Cucumber** | Cucumber | Yes | POM + Step Definitions |
| **3. Selenium + JUnit 5** | JUnit 5 | No | Page Object Model |
| **4. Selenium + JUnit 5 + Cucumber** | Cucumber | Yes | POM + Step Definitions |
| **5. Selenium + Serenity BDD** | Serenity | Yes | Screenplay pattern |
| **6. Selenium + Rest-Assured (API + UI)** | TestNG/JUnit | No | Hybrid testing |

### What We Need to Understand Per Framework

#### **1. Selenium + TestNG (Most Common)**

**Detection Signals**:
```xml
<!-- pom.xml -->
<dependency>
    <groupId>org.testng</groupId>
    <artifactId>testng</artifactId>
</dependency>
<dependency>
    <groupId>org.seleniumhq.selenium</groupId>
    <artifactId>selenium-java</artifactId>
</dependency>
```

**Patterns to Recognize**:
```java
// Test structure
@Test
public void testLogin() { }

// Setup/Teardown
@BeforeMethod
public void setup() { }

@AfterMethod
public void teardown() { }

// Assertions
Assert.assertEquals(actual, expected, "message");
Assert.assertTrue(condition, "message");

// Data-driven
@DataProvider(name = "loginData")
public Object[][] provideData() { }

@Test(dataProvider = "loginData")
public void testWithData(String user, String pass) { }

// Groups
@Test(groups = {"smoke", "regression"})

// Dependencies
@Test(dependsOnMethods = "testLogin")
```

**File Structure**:
```
src/test/java/
├── pages/
│   ├── BasePage.java
│   ├── LoginPage.java
│   └── DashboardPage.java
├── tests/
│   ├── BaseTest.java
│   ├── LoginTest.java
│   └── DashboardTest.java
└── utils/
    ├── DriverFactory.java
    └── TestDataProvider.java

src/test/resources/
├── testng.xml
└── test-data.xlsx
```

**Code Generation Rules**:
- Use TestNG annotations
- Extend BaseTest for all tests
- Page objects use PageFactory
- Data providers for data-driven tests
- testng.xml for suite configuration

---

#### **2. Selenium + TestNG + Cucumber**

**Detection Signals**:
```xml
<!-- pom.xml -->
<dependency>
    <groupId>io.cucumber</groupId>
    <artifactId>cucumber-testng</artifactId>
</dependency>
<dependency>
    <groupId>io.cucumber</groupId>
    <artifactId>cucumber-java</artifactId>
</dependency>
```

**Additional Files to Parse**:
```
src/test/resources/features/
├── login.feature
└── product-search.feature

src/test/java/
├── pages/           (same as #1)
├── steps/
│   ├── LoginSteps.java
│   └── ProductSteps.java
└── runners/
    └── TestRunner.java
```

**Gherkin Understanding**:
```gherkin
# login.feature
Feature: User Authentication

  Scenario: Successful login with valid credentials
    Given user is on login page
    When user enters username "test@example.com"
    And user enters password "Test@123"
    And user clicks login button
    Then user should see dashboard
    And welcome message should display "Welcome, Test User"

  Scenario Outline: Login with multiple users
    Given user is on login page
    When user logs in with "<username>" and "<password>"
    Then login should "<result>"

    Examples:
      | username          | password | result  |
      | valid@example.com | Pass123  | succeed |
      | invalid@test.com  | wrong    | fail    |
```

**Step Definition Patterns**:
```java
public class LoginSteps {
    private LoginPage loginPage;
    private DashboardPage dashboardPage;

    @Given("user is on login page")
    public void userIsOnLoginPage() {
        loginPage = new LoginPage(driver);
    }

    @When("user enters username {string}")
    public void userEntersUsername(String username) {
        loginPage.enterUsername(username);
    }

    @Then("user should see dashboard")
    public void userShouldSeeDashboard() {
        Assert.assertTrue(dashboardPage.isDisplayed());
    }
}
```

**What Go Binary Must Extract**:
1. Parse .feature files (Gherkin AST)
2. Extract scenarios, steps, examples
3. Map steps to step definitions
4. Detect missing step definitions
5. Understand data tables and scenario outlines

**Code Generation Rules**:
- Generate .feature files first
- Create step definitions matching Gherkin
- Use regex in @Given/@When/@Then
- Step definitions use page objects
- Runner class extends AbstractTestNGCucumberTests

---

#### **3. Selenium + JUnit 5**

**Detection Signals**:
```xml
<dependency>
    <groupId>org.junit.jupiter</groupId>
    <artifactId>junit-jupiter</artifactId>
</dependency>
```

**Patterns to Recognize**:
```java
// Test structure
@Test
void testLogin() { }

// Display names
@DisplayName("Verify successful login with valid credentials")
@Test
void testSuccessfulLogin() { }

// Setup/Teardown
@BeforeEach
void setup() { }

@AfterEach
void teardown() { }

@BeforeAll
static void setupClass() { }

// Assertions
assertEquals(expected, actual, "message");
assertTrue(condition, "message");
assertAll(
    () -> assertTrue(condition1),
    () -> assertEquals(expected, actual)
);

// Parameterized tests
@ParameterizedTest
@CsvSource({
    "user1, pass1, true",
    "user2, pass2, false"
})
void testLogin(String username, String password, boolean shouldSucceed) { }

// Tags
@Tag("smoke")
@Tag("regression")
@Test
void testLogin() { }
```

**Code Generation Rules**:
- Use JUnit 5 annotations (not JUnit 4!)
- Static imports: `import static org.junit.jupiter.api.Assertions.*;`
- Use @DisplayName for readability
- @ParameterizedTest for data-driven

---

#### **4. Selenium + JUnit 5 + Cucumber**

Similar to #2 but:
- Runner extends `io.cucumber.junit.platform.engine.Cucumber`
- Uses JUnit 5 @Suite annotation
- junit-platform.properties for configuration

---

#### **5. Selenium + Serenity BDD**

**Detection Signals**:
```xml
<dependency>
    <groupId>net.serenity-bdd</groupId>
    <artifactId>serenity-cucumber</artifactId>
</dependency>
```

**Screenplay Pattern** (not Page Object):
```java
// Actor-based
Actor user = Actor.named("Test User");

// Tasks
user.attemptsTo(
    NavigateTo.theLoginPage(),
    Enter.theValue("test@example.com").into(LoginForm.USERNAME),
    Enter.theValue("Pass123").into(LoginForm.PASSWORD),
    Click.on(LoginForm.SUBMIT_BUTTON)
);

// Questions (assertions)
user.should(
    seeThat("Dashboard is displayed",
        DashboardPage.WELCOME_MESSAGE,
        isDisplayed())
);
```

**Different architecture**:
```
src/test/java/
├── screenplay/
│   ├── tasks/
│   │   ├── NavigateTo.java
│   │   └── Login.java
│   ├── questions/
│   │   └── TheWelcomeMessage.java
│   └── ui/
│       └── LoginForm.java
├── features/
│   └── login.feature
└── stepdefinitions/
    └── LoginSteps.java
```

**Code Generation Rules**:
- Use Screenplay pattern (Tasks, Questions, Abilities)
- Actors perform tasks
- Questions for assertions
- Target elements instead of page objects

---

#### **6. Selenium + Rest-Assured (Hybrid API + UI)**

**Detection Signals**:
```xml
<dependency>
    <groupId>io.rest-assured</groupId>
    <artifactId>rest-assured</artifactId>
</dependency>
```

**Hybrid test structure**:
```java
@Test
public void testUserRegistrationE2E() {
    // Step 1: Create user via API
    String userId = given()
        .contentType("application/json")
        .body("{\"username\":\"test\",\"password\":\"pass\"}")
    .when()
        .post("/api/users")
    .then()
        .statusCode(201)
        .extract().path("id");

    // Step 2: Verify UI shows new user
    LoginPage loginPage = new LoginPage(driver);
    loginPage.login("test", "pass");

    DashboardPage dashboard = new DashboardPage(driver);
    Assert.assertEquals(dashboard.getUserId(), userId);

    // Step 3: Cleanup via API
    given()
        .when().delete("/api/users/" + userId)
        .then().statusCode(204);
}
```

**Code Generation Rules**:
- Combine API calls (Rest-Assured) with UI (Selenium)
- Use API for setup/cleanup
- Use UI for critical user journeys
- Separate concerns: api/ and ui/ packages

---

### What Go Binary Must Do for Each Framework

**For EVERY framework combination**:

1. **Detection**:
   ```go
   type FrameworkDetector struct {
       pomParser  *POMParser
       fileScanner *FileScanner
   }

   func (fd *FrameworkDetector) Detect() *Framework {
       // 1. Parse pom.xml
       dependencies := fd.pomParser.GetDependencies()

       // 2. Check for framework signatures
       hasSelenium := dependencies.Contains("selenium-java")
       hasTestNG := dependencies.Contains("testng")
       hasJUnit := dependencies.Contains("junit-jupiter")
       hasCucumber := dependencies.Contains("cucumber")
       hasSerenity := dependencies.Contains("serenity")
       hasRestAssured := dependencies.Contains("rest-assured")

       // 3. Check file structure
       hasFeatures := fd.fileScanner.DirectoryExists("src/test/resources/features")
       hasSteps := fd.fileScanner.DirectoryExists("src/test/java/**/steps")
       hasScreenplay := fd.fileScanner.DirectoryExists("src/test/java/**/screenplay")

       // 4. Determine combination
       if hasSelenium && hasTestNG && hasCucumber && hasFeatures {
           return &Framework{
               Type: "selenium-testng-cucumber",
               Runner: "cucumber",
               BDD: true,
               Pattern: "page_object_model"
           }
       }
       // ... etc for all combinations
   }
   ```

2. **Pattern Extraction**:
   ```go
   type PatternExtractor struct {
       framework *Framework
   }

   func (pe *PatternExtractor) ExtractCodingStyle() *CodingStyle {
       // Analyze existing test files
       tests := pe.findAllTests()

       return &CodingStyle{
           NamingConvention: pe.detectNaming(tests),
           AssertionStyle: pe.detectAssertions(tests),
           WaitStrategy: pe.detectWaits(tests),
           PackageStructure: pe.detectPackages(),
           ImportStyle: pe.detectImports(tests)
       }
   }
   ```

3. **Context Building**:
   ```go
   func (cb *ContextBuilder) BuildContext(userMessage string) *Context {
       intent := cb.parseIntent(userMessage)

       return &Context{
           Intent: intent,                           // "create_test"
           Framework: cb.framework,                  // Detected framework
           ExistingCode: cb.findRelevantCode(intent), // Related files
           ApplicationMap: cb.loadAppMap(),          // User's app structure
           CodingStyle: cb.extractedStyle,           // Project patterns
           Templates: cb.loadTemplates(),            // Framework templates
       }
   }
   ```

---

## 🔄 End-to-End User Flow

### Flow 1: First-Time User (No Code Exists)

```
1. User opens workspace with just:
   - pom.xml (Selenium + TestNG dependencies)
   - No tests yet

2. Extension detects: New project, no Application Map

3. Copilot suggests:
   "I see you have Selenium + TestNG set up. Would you like to:
    [1] Create Application Map (recommended)
    [2] Start with a basic test
    [3] Tour of features"

4. User clicks [1] Create Application Map

5. UI opens:
   "Let's map your application. What's the first page to test?"

   Page name: [Login____________]
   URL: [/login__________]

   Elements:
   [+ Add element]

6. User adds:
   - username → id=username → input
   - password → id=password → input
   - loginButton → css=button[type='submit'] → button

7. Saves → .testcopilot/application-map.json created

8. User types in chat: "create login test"

9. Go Binary:
   - Detects: Selenium + TestNG
   - Finds: No LoginPage.java exists
   - Checks: Application Map has "Login" page
   - Decision: Need to create LoginPage first, then LoginTest

10. Creates context package:
    {
      "tasks": [
        {
          "type": "create_page_object",
          "page": "Login",
          "file": "src/test/java/pages/LoginPage.java",
          "elements": [...from app map]
        },
        {
          "type": "create_test",
          "test_name": "LoginTest",
          "file": "src/test/java/tests/LoginTest.java",
          "page_object": "LoginPage"
        },
        {
          "type": "create_base_test",
          "file": "src/test/java/tests/BaseTest.java"
        }
      ],
      "framework": "selenium-testng"
    }

11. Sends to Cloud Backend

12. Cloud Backend:
    - Builds prompts for each task
    - Calls GPT-4o (3 separate calls, or 1 with multi-file)
    - Streams back code

13. Go Binary receives:
    - BaseTest.java (setup/teardown)
    - LoginPage.java (page object)
    - LoginTest.java (test cases)

14. Extension shows diff preview:
    "I'll create 3 new files:"

    ✓ src/test/java/tests/BaseTest.java
      [Show code preview]

    ✓ src/test/java/pages/LoginPage.java
      [Show code preview]

    ✓ src/test/java/tests/LoginTest.java
      [Show code preview]

    [Accept All] [Reject] [Edit]

15. User clicks [Accept All]

16. Extension writes files

17. Success! User now has working test structure
```

---

### Flow 2: Existing Project (Add to existing tests)

```
1. User has project with:
   - LoginPage.java, LoginTest.java
   - DashboardPage.java, DashboardTest.java
   - Application Map exists

2. User types: "add test for product search"

3. Go Binary:
   - Detects: Selenium + TestNG
   - Checks: ProductSearchPage.java → NOT FOUND
   - Checks: Application Map → "ProductSearch" page FOUND
   - Analyzes existing patterns:
     * All page objects extend BasePage
     * Use @FindBy with PageFactory
     * Tests extend BaseTest
     * Use TestNG assertions with messages
     * Follow naming: testMethodName()

4. Decision: Create ProductSearchPage.java and ProductSearchTest.java

5. Context package:
   {
     "tasks": [
       {
         "type": "create_page_object",
         "page": "ProductSearch",
         "file": "src/test/java/pages/ProductSearchPage.java",
         "elements": [...from app map],
         "extend": "BasePage",
         "follow_pattern": {
           "example": "LoginPage.java (full code)"
         }
       },
       {
         "type": "create_test",
         "file": "src/test/java/tests/ProductSearchTest.java",
         "page_object": "ProductSearchPage",
         "extend": "BaseTest",
         "follow_pattern": {
           "example": "LoginTest.java (full code)"
         }
       }
     ]
   }

6. Cloud generates code matching EXACT style of existing code

7. User accepts

8. New files created, perfectly matching existing patterns
```

---

### Flow 3: BDD/Cucumber Project

```
1. User has: Selenium + TestNG + Cucumber

2. File structure:
   src/test/resources/features/
   src/test/java/steps/
   src/test/java/pages/
   src/test/java/runners/

3. User types: "create checkout feature"

4. Go Binary:
   - Detects: Cucumber BDD
   - Checks: checkout.feature → NOT FOUND
   - Checks: Application Map → "Checkout" page FOUND
   - Analyzes existing features:
     * Given/When/Then structure
     * Scenario Outline for data-driven
     * Examples tables

5. Context package:
   {
     "tasks": [
       {
         "type": "create_feature",
         "file": "src/test/resources/features/checkout.feature",
         "page": "Checkout",
         "scenarios": ["successful checkout", "invalid payment"]
       },
       {
         "type": "create_step_definitions",
         "file": "src/test/java/steps/CheckoutSteps.java",
         "feature": "checkout.feature"
       },
       {
         "type": "create_page_object",
         "file": "src/test/java/pages/CheckoutPage.java",
         "elements": [...from app map]
       }
     ]
   }

6. Cloud generates:
   - checkout.feature (Gherkin)
   - CheckoutSteps.java (step definitions)
   - CheckoutPage.java (page object)

7. All perfectly integrated with existing Cucumber setup
```

---

## 💳 Free Credits Model

**No separate free tier** - AI from day one with credits

**Pricing Structure**:

```
New User:
✓ 10,000 free tokens (~20-30 code generations)
✓ Experience full product
✓ No feature limitations

After free credits:
→ Subscribe: $20/month unlimited
→ Pay-as-you-go: $0.10 per 1000 tokens

Credits displayed in extension:
"Credits remaining: 7,342 tokens (~15 generations left)"
```

**Why this works**:
- Users try REAL product (not crippled version)
- If they generate 20 tests and love it → they'll pay
- If they don't use it → no loss to us
- Simple to understand

---

## 🎯 Revised Implementation Priority

### Phase 1: Foundation (Week 1-2) ✅ MOSTLY DONE

**Go Binary**:
- ✅ Workspace indexing
- ✅ Tree-sitter parsing
- ✅ Framework detection (basic)
- ✅ Semantic search
- ⬜ **ENHANCE**: Deep framework detection (all 6 combinations)
- ⬜ **NEW**: Application Map storage and retrieval

**Extension**:
- ✅ Chat UI
- ✅ WebSocket streaming
- ⬜ **NEW**: Application Map UI
- ⬜ **NEW**: Diff preview UI

**Cloud Backend**:
- ⬜ **NEW**: OpenAI integration (2 hours)
- ⬜ **NEW**: Free credits tracking
- ⬜ **NEW**: Usage monitoring

---

### Phase 2: Framework Mastery (Week 3-4) 🎯 FOCUS HERE

**Priority**: Deep understanding of ALL 6 framework combinations

**Tasks**:
1. **Enhance framework detector** (3 days)
   ```go
   // Detect all combinations accurately
   // Extract framework-specific patterns
   // Understand file structures
   ```

2. **Pattern extraction** (3 days)
   ```go
   // Learn project coding style
   // Extract naming conventions
   // Detect assertion patterns
   // Understand wait strategies
   ```

3. **Application Map** (2 days)
   ```go
   // Storage (SQLite)
   // UI for creation
   // Integration with context builder
   ```

4. **Context builder enhancement** (3 days)
   ```go
   // Assemble PERFECT context for LLM
   // Include all relevant code
   // Include Application Map
   // Include coding style
   ```

5. **Testing** (2 days)
   - Test with real Selenium projects
   - Verify framework detection
   - Validate context quality

---

### Phase 3: AI Integration (Week 5) 🤖

**ONLY after Phase 2 is perfect**

1. Cloud backend setup (2 days)
2. Prompt templates per framework (2 days)
3. Response streaming (1 day)
4. Free credits system (1 day)
5. End-to-end testing (1 day)

---

### Phase 4: Code Operations (Week 6)

1. File creation logic
2. File editing logic (merge with existing)
3. Diff preview UI
4. Validation (syntax check, anti-patterns)

---

## 📊 Success Metrics

**Framework Understanding**:
- ✅ Detects all 6 combinations with 100% accuracy
- ✅ Extracts coding style with 95% accuracy
- ✅ Builds context that includes ALL relevant code

**AI Quality** (depends on Phase 2 being perfect):
- ✅ Generated code matches existing style 95% of time
- ✅ Follows framework conventions 100%
- ✅ No anti-patterns in output

**User Experience**:
- ✅ "create login test" → generates 3 perfect files
- ✅ Code accepted without edits 70% of time
- ✅ Free credits get users hooked

---

## 🎯 Decision: What to Build First

**Your vision is correct**:

> "Concentrate on how good we understand frameworks... then we will work on AI part once our context building and code workspace understanding at its best"

**Action Plan**:

1. **This Week**:
   - Enhance framework detector (all 6 combinations)
   - Build Application Map foundation
   - Test with real projects

2. **Next Week**:
   - Perfect pattern extraction
   - Perfect context building
   - Validate with manual review (is context complete?)

3. **Week 3**:
   - Add AI integration (should be easy if context is perfect)
   - Test end-to-end
   - Launch MVP

**The key insight**: If context is perfect, AI will be great. If context is mediocre, no amount of fancy prompting will help.

---

**Next step**: Should I start enhancing the framework detector to recognize all 6 combinations?
