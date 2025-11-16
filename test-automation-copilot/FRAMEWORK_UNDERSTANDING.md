# 🔍 Framework Understanding - Line by Line Deep Dive

**Goal**: From user message "edit login test" → Know EVERYTHING about login test, page objects, elements, methods

**How**: Build a complete knowledge graph of the workspace through deep AST analysis

---

## 📚 Example: Understanding a TestNG + Selenium Project

### Sample Project Structure
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
    └── DriverFactory.java

pom.xml
testng.xml
```

---

## 🔬 Step 1: Framework Detection (High-Level)

### Parse pom.xml

**File**: `pom.xml`
```xml
<dependencies>
    <dependency>
        <groupId>org.seleniumhq.selenium</groupId>
        <artifactId>selenium-java</artifactId>
        <version>4.15.0</version>
    </dependency>
    <dependency>
        <groupId>org.testng</groupId>
        <artifactId>testng</artifactId>
        <version>7.8.0</version>
    </dependency>
</dependencies>
```

**What We Extract**:
```json
{
  "framework_detected": {
    "test_library": "selenium",
    "test_runner": "testng",
    "pattern": "page_object_model",
    "bdd": false,
    "api_testing": false
  },
  "dependencies": [
    {"group": "org.seleniumhq.selenium", "artifact": "selenium-java", "version": "4.15.0"},
    {"group": "org.testng", "artifact": "testng", "version": "7.8.0"}
  ]
}
```

**Go Code**:
```go
// copilot-core/pkg/framework/detector.go

type FrameworkDetector struct {
    pomPath string
}

func (fd *FrameworkDetector) Detect() *Framework {
    pom := fd.parsePOM()

    hasSelenium := pom.HasDependency("selenium-java")
    hasTestNG := pom.HasDependency("testng")
    hasJUnit := pom.HasDependency("junit-jupiter")
    hasCucumber := pom.HasDependency("cucumber")

    if hasSelenium && hasTestNG && !hasCucumber {
        return &Framework{
            Type: "selenium-testng",
            TestRunner: "testng",
            Pattern: "page_object_model",
            BDD: false
        }
    }
    // ... other combinations
}
```

---

## 🗂️ Step 2: File Structure Mapping

### Scan Directory Tree

**What We Do**:
```go
type WorkspaceScanner struct {
    rootPath string
}

func (ws *WorkspaceScanner) Scan() *WorkspaceStructure {
    return &WorkspaceStructure{
        Packages: map[string]*Package{
            "pages": {
                Path: "src/test/java/pages",
                Files: []string{
                    "BasePage.java",
                    "LoginPage.java",
                    "DashboardPage.java"
                }
            },
            "tests": {
                Path: "src/test/java/tests",
                Files: []string{
                    "BaseTest.java",
                    "LoginTest.java",
                    "DashboardTest.java"
                }
            },
            "utils": {
                Path: "src/test/java/utils",
                Files: []string{
                    "DriverFactory.java"
                }
            }
        }
    }
}
```

**What We Store** (SQLite):
```sql
CREATE TABLE files (
    id INTEGER PRIMARY KEY,
    file_path TEXT,
    package_name TEXT,
    file_type TEXT,  -- 'page_object', 'test', 'utility', 'step_definition'
    class_name TEXT,
    last_modified INTEGER
);

INSERT INTO files VALUES
    (1, 'src/test/java/pages/LoginPage.java', 'pages', 'page_object', 'LoginPage', 1699999999),
    (2, 'src/test/java/tests/LoginTest.java', 'tests', 'test', 'LoginTest', 1699999999);
```

---

## 🧬 Step 3: Deep AST Analysis (File-Level Parsing)

### Parse LoginPage.java

**File**: `src/test/java/pages/LoginPage.java`
```java
package pages;

import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.support.FindBy;
import org.openqa.selenium.support.PageFactory;
import org.openqa.selenium.support.ui.WebDriverWait;
import org.openqa.selenium.support.ui.ExpectedConditions;
import java.time.Duration;

public class LoginPage extends BasePage {
    private WebDriver driver;
    private WebDriverWait wait;

    // Elements
    @FindBy(id = "username")
    private WebElement usernameField;

    @FindBy(id = "password")
    private WebElement passwordField;

    @FindBy(css = "button[type='submit']")
    private WebElement loginButton;

    @FindBy(className = "error-message")
    private WebElement errorMessage;

    // Constructor
    public LoginPage(WebDriver driver) {
        super(driver);
        this.driver = driver;
        this.wait = new WebDriverWait(driver, Duration.ofSeconds(10));
        PageFactory.initElements(driver, this);
    }

    // Actions
    public void enterUsername(String username) {
        wait.until(ExpectedConditions.visibilityOf(usernameField));
        usernameField.clear();
        usernameField.sendKeys(username);
    }

    public void enterPassword(String password) {
        passwordField.clear();
        passwordField.sendKeys(password);
    }

    public DashboardPage clickLoginButton() {
        wait.until(ExpectedConditions.elementToBeClickable(loginButton));
        loginButton.click();
        return new DashboardPage(driver);
    }

    public void login(String username, String password) {
        enterUsername(username);
        enterPassword(password);
        clickLoginButton();
    }

    // Validations
    public boolean isErrorDisplayed() {
        try {
            return errorMessage.isDisplayed();
        } catch (Exception e) {
            return false;
        }
    }

    public String getErrorMessage() {
        wait.until(ExpectedConditions.visibilityOf(errorMessage));
        return errorMessage.getText();
    }
}
```

### What We Extract (Line by Line)

**Using Tree-sitter AST Parser**:

```go
// copilot-core/pkg/parser/java_parser.go

type JavaParser struct {
    parser *sitter.Parser
}

func (jp *JavaParser) ParseFile(filePath string) *JavaClass {
    content := readFile(filePath)
    tree := jp.parser.Parse([]byte(content))
    root := tree.RootNode()

    class := &JavaClass{
        FilePath: filePath,
    }

    // Extract package
    class.Package = jp.extractPackage(root)

    // Extract imports
    class.Imports = jp.extractImports(root)

    // Extract class declaration
    classNode := jp.findClassDeclaration(root)
    class.ClassName = jp.extractClassName(classNode)
    class.ExtendsClass = jp.extractSuperclass(classNode)
    class.Implements = jp.extractInterfaces(classNode)
    class.Modifiers = jp.extractModifiers(classNode)

    // Extract fields
    class.Fields = jp.extractFields(classNode)

    // Extract methods
    class.Methods = jp.extractMethods(classNode)

    return class
}
```

**Extracted Data Structure**:
```json
{
  "file_path": "src/test/java/pages/LoginPage.java",
  "package": "pages",
  "imports": [
    "org.openqa.selenium.WebDriver",
    "org.openqa.selenium.WebElement",
    "org.openqa.selenium.support.FindBy",
    "org.openqa.selenium.support.PageFactory",
    "org.openqa.selenium.support.ui.WebDriverWait",
    "org.openqa.selenium.support.ui.ExpectedConditions",
    "java.time.Duration"
  ],
  "class": {
    "name": "LoginPage",
    "modifiers": ["public"],
    "extends": "BasePage",
    "implements": [],
    "fields": [
      {
        "name": "driver",
        "type": "WebDriver",
        "modifiers": ["private"],
        "annotations": []
      },
      {
        "name": "wait",
        "type": "WebDriverWait",
        "modifiers": ["private"],
        "annotations": []
      },
      {
        "name": "usernameField",
        "type": "WebElement",
        "modifiers": ["private"],
        "annotations": [
          {
            "type": "FindBy",
            "params": {
              "id": "username"
            }
          }
        ]
      },
      {
        "name": "passwordField",
        "type": "WebElement",
        "modifiers": ["private"],
        "annotations": [
          {
            "type": "FindBy",
            "params": {
              "id": "password"
            }
          }
        ]
      },
      {
        "name": "loginButton",
        "type": "WebElement",
        "modifiers": ["private"],
        "annotations": [
          {
            "type": "FindBy",
            "params": {
              "css": "button[type='submit']"
            }
          }
        ]
      },
      {
        "name": "errorMessage",
        "type": "WebElement",
        "modifiers": ["private"],
        "annotations": [
          {
            "type": "FindBy",
            "params": {
              "className": "error-message"
            }
          }
        ]
      }
    ],
    "methods": [
      {
        "name": "LoginPage",
        "return_type": null,
        "is_constructor": true,
        "modifiers": ["public"],
        "parameters": [
          {"name": "driver", "type": "WebDriver"}
        ],
        "annotations": [],
        "body_summary": "Calls super(driver), initializes wait, calls PageFactory.initElements",
        "calls": ["super", "PageFactory.initElements"],
        "line_start": 26,
        "line_end": 31
      },
      {
        "name": "enterUsername",
        "return_type": "void",
        "is_constructor": false,
        "modifiers": ["public"],
        "parameters": [
          {"name": "username", "type": "String"}
        ],
        "annotations": [],
        "body_summary": "Waits for usernameField visibility, clears, sends keys",
        "calls": ["wait.until", "usernameField.clear", "usernameField.sendKeys"],
        "interacts_with_fields": ["usernameField", "wait"],
        "line_start": 34,
        "line_end": 38
      },
      {
        "name": "enterPassword",
        "return_type": "void",
        "modifiers": ["public"],
        "parameters": [
          {"name": "password", "type": "String"}
        ],
        "interacts_with_fields": ["passwordField"],
        "line_start": 40,
        "line_end": 43
      },
      {
        "name": "clickLoginButton",
        "return_type": "DashboardPage",
        "modifiers": ["public"],
        "parameters": [],
        "body_summary": "Waits for button, clicks, returns new DashboardPage",
        "returns": "new DashboardPage(driver)",
        "interacts_with_fields": ["loginButton", "wait"],
        "line_start": 45,
        "line_end": 49
      },
      {
        "name": "login",
        "return_type": "void",
        "modifiers": ["public"],
        "parameters": [
          {"name": "username", "type": "String"},
          {"name": "password", "type": "String"}
        ],
        "body_summary": "Calls enterUsername, enterPassword, clickLoginButton",
        "calls": ["enterUsername", "enterPassword", "clickLoginButton"],
        "line_start": 51,
        "line_end": 55
      },
      {
        "name": "isErrorDisplayed",
        "return_type": "boolean",
        "modifiers": ["public"],
        "parameters": [],
        "interacts_with_fields": ["errorMessage"],
        "line_start": 58,
        "line_end": 64
      },
      {
        "name": "getErrorMessage",
        "return_type": "String",
        "modifiers": ["public"],
        "parameters": [],
        "interacts_with_fields": ["errorMessage", "wait"],
        "line_start": 66,
        "line_end": 69
      }
    ]
  }
}
```

### Database Storage

**SQLite Schema**:
```sql
CREATE TABLE classes (
    id INTEGER PRIMARY KEY,
    file_id INTEGER,
    class_name TEXT,
    package_name TEXT,
    extends_class TEXT,
    class_type TEXT,  -- 'page_object', 'test', 'utility', 'step_definition'
    FOREIGN KEY (file_id) REFERENCES files(id)
);

CREATE TABLE fields (
    id INTEGER PRIMARY KEY,
    class_id INTEGER,
    field_name TEXT,
    field_type TEXT,
    modifiers TEXT,  -- JSON array
    annotations TEXT,  -- JSON array
    locator_type TEXT,  -- 'id', 'css', 'xpath', NULL
    locator_value TEXT,
    FOREIGN KEY (class_id) REFERENCES classes(id)
);

CREATE TABLE methods (
    id INTEGER PRIMARY KEY,
    class_id INTEGER,
    method_name TEXT,
    return_type TEXT,
    is_constructor BOOLEAN,
    modifiers TEXT,  -- JSON array
    parameters TEXT,  -- JSON array
    annotations TEXT,  -- JSON array
    body_summary TEXT,
    calls_methods TEXT,  -- JSON array
    interacts_with_fields TEXT,  -- JSON array
    line_start INTEGER,
    line_end INTEGER,
    FOREIGN KEY (class_id) REFERENCES classes(id)
);

CREATE TABLE method_calls (
    id INTEGER PRIMARY KEY,
    from_method_id INTEGER,
    to_method_name TEXT,
    to_class_name TEXT,
    FOREIGN KEY (from_method_id) REFERENCES methods(id)
);
```

**Insert for LoginPage**:
```sql
-- Class
INSERT INTO classes VALUES
    (1, 1, 'LoginPage', 'pages', 'BasePage', 'page_object');

-- Fields (elements)
INSERT INTO fields VALUES
    (1, 1, 'usernameField', 'WebElement', '["private"]', '[{"type":"FindBy","params":{"id":"username"}}]', 'id', 'username'),
    (2, 1, 'passwordField', 'WebElement', '["private"]', '[{"type":"FindBy","params":{"id":"password"}}]', 'id', 'password'),
    (3, 1, 'loginButton', 'WebElement', '["private"]', '[{"type":"FindBy","params":{"css":"button[type=\'submit\']"}}]', 'css', 'button[type=\'submit\']'),
    (4, 1, 'errorMessage', 'WebElement', '["private"]', '[{"type":"FindBy","params":{"className":"error-message"}}]', 'className', 'error-message');

-- Methods
INSERT INTO methods VALUES
    (1, 1, 'LoginPage', NULL, 1, '["public"]', '[{"name":"driver","type":"WebDriver"}]', '[]', 'Constructor', '["super","PageFactory.initElements"]', '[]', 26, 31),
    (2, 1, 'enterUsername', 'void', 0, '["public"]', '[{"name":"username","type":"String"}]', '[]', 'Enters username', '["wait.until","usernameField.clear","usernameField.sendKeys"]', '["usernameField","wait"]', 34, 38),
    (3, 1, 'enterPassword', 'void', 0, '["public"]', '[{"name":"password","type":"String"}]', '[]', 'Enters password', '[]', '["passwordField"]', 40, 43),
    (4, 1, 'clickLoginButton', 'DashboardPage', 0, '["public"]', '[]', '[]', 'Clicks login button', '[]', '["loginButton","wait"]', 45, 49),
    (5, 1, 'login', 'void', 0, '["public"]', '[{"name":"username","type":"String"},{"name":"password","type":"String"}]', '[]', 'Complete login flow', '["enterUsername","enterPassword","clickLoginButton"]', '[]', 51, 55);
```

---

### Parse LoginTest.java

**File**: `src/test/java/tests/LoginTest.java`
```java
package tests;

import org.openqa.selenium.WebDriver;
import org.testng.Assert;
import org.testng.annotations.Test;
import pages.LoginPage;
import pages.DashboardPage;

public class LoginTest extends BaseTest {

    @Test(description = "Verify successful login with valid credentials")
    public void testLoginWithValidCredentials() {
        // Arrange
        LoginPage loginPage = new LoginPage(driver);
        String username = "testuser@example.com";
        String password = "Test@123";

        // Act
        loginPage.login(username, password);
        DashboardPage dashboardPage = new DashboardPage(driver);

        // Assert
        Assert.assertTrue(dashboardPage.isWelcomeMessageDisplayed(),
            "Welcome message should be displayed after successful login");
        Assert.assertEquals(dashboardPage.getUsername(), username,
            "Displayed username should match logged in user");
    }

    @Test(description = "Verify error message with invalid credentials")
    public void testLoginWithInvalidCredentials() {
        // Arrange
        LoginPage loginPage = new LoginPage(driver);
        String username = "invalid@example.com";
        String password = "wrongpass";

        // Act
        loginPage.login(username, password);

        // Assert
        Assert.assertTrue(loginPage.isErrorDisplayed(),
            "Error message should be displayed for invalid credentials");
        Assert.assertEquals(loginPage.getErrorMessage(),
            "Invalid username or password",
            "Error message should indicate invalid credentials");
    }

    @Test(description = "Verify login button is disabled with empty username")
    public void testLoginButtonDisabledWithEmptyUsername() {
        LoginPage loginPage = new LoginPage(driver);

        loginPage.enterPassword("Test@123");

        Assert.assertFalse(loginPage.isLoginButtonEnabled(),
            "Login button should be disabled when username is empty");
    }
}
```

**Extracted Data**:
```json
{
  "file_path": "src/test/java/tests/LoginTest.java",
  "package": "tests",
  "imports": [
    "org.openqa.selenium.WebDriver",
    "org.testng.Assert",
    "org.testng.annotations.Test",
    "pages.LoginPage",
    "pages.DashboardPage"
  ],
  "class": {
    "name": "LoginTest",
    "extends": "BaseTest",
    "class_type": "test",
    "uses_page_objects": ["LoginPage", "DashboardPage"],
    "methods": [
      {
        "name": "testLoginWithValidCredentials",
        "return_type": "void",
        "modifiers": ["public"],
        "annotations": [
          {
            "type": "Test",
            "params": {
              "description": "Verify successful login with valid credentials"
            }
          }
        ],
        "test_type": "positive",
        "creates_page_objects": ["LoginPage", "DashboardPage"],
        "calls_page_methods": [
          {"page": "LoginPage", "method": "login"},
          {"page": "DashboardPage", "method": "isWelcomeMessageDisplayed"},
          {"page": "DashboardPage", "method": "getUsername"}
        ],
        "assertions": [
          {
            "type": "Assert.assertTrue",
            "condition": "dashboardPage.isWelcomeMessageDisplayed()",
            "message": "Welcome message should be displayed after successful login"
          },
          {
            "type": "Assert.assertEquals",
            "actual": "dashboardPage.getUsername()",
            "expected": "username",
            "message": "Displayed username should match logged in user"
          }
        ],
        "test_data": {
          "username": "testuser@example.com",
          "password": "Test@123"
        },
        "line_start": 11,
        "line_end": 27
      },
      {
        "name": "testLoginWithInvalidCredentials",
        "annotations": [
          {
            "type": "Test",
            "params": {
              "description": "Verify error message with invalid credentials"
            }
          }
        ],
        "test_type": "negative",
        "creates_page_objects": ["LoginPage"],
        "calls_page_methods": [
          {"page": "LoginPage", "method": "login"},
          {"page": "LoginPage", "method": "isErrorDisplayed"},
          {"page": "LoginPage", "method": "getErrorMessage"}
        ],
        "assertions": [
          {
            "type": "Assert.assertTrue",
            "condition": "loginPage.isErrorDisplayed()"
          },
          {
            "type": "Assert.assertEquals",
            "actual": "loginPage.getErrorMessage()",
            "expected": "\"Invalid username or password\""
          }
        ],
        "line_start": 29,
        "line_end": 44
      }
    ]
  }
}
```

**Database Storage**:
```sql
-- Class
INSERT INTO classes VALUES
    (2, 2, 'LoginTest', 'tests', 'BaseTest', 'test');

-- Class uses page objects (relationship)
CREATE TABLE class_dependencies (
    id INTEGER PRIMARY KEY,
    class_id INTEGER,
    depends_on_class TEXT,
    dependency_type TEXT,  -- 'uses', 'extends', 'imports'
    FOREIGN KEY (class_id) REFERENCES classes(id)
);

INSERT INTO class_dependencies VALUES
    (1, 2, 'LoginPage', 'uses'),
    (2, 2, 'DashboardPage', 'uses'),
    (3, 2, 'BaseTest', 'extends');

-- Methods
INSERT INTO methods VALUES
    (6, 2, 'testLoginWithValidCredentials', 'void', 0, '["public"]', '[]', '[{"type":"Test","params":{"description":"Verify successful login with valid credentials"}}]', 'Positive test for valid login', '["LoginPage.login","DashboardPage.isWelcomeMessageDisplayed","DashboardPage.getUsername"]', '[]', 11, 27),
    (7, 2, 'testLoginWithInvalidCredentials', 'void', 0, '["public"]', '[]', '[{"type":"Test","params":{"description":"Verify error message with invalid credentials"}}]', 'Negative test for invalid login', '["LoginPage.login","LoginPage.isErrorDisplayed","LoginPage.getErrorMessage"]', '[]', 29, 44);
```

---

## 🕸️ Step 4: Build Knowledge Graph (Relationships)

### Relationship Types We Track

```go
type KnowledgeGraph struct {
    Files       map[string]*FileNode
    Classes     map[string]*ClassNode
    Methods     map[string]*MethodNode
    Fields      map[string]*FieldNode

    Relationships []Relationship
}

type Relationship struct {
    From       string
    To         string
    Type       RelationType
    Metadata   map[string]interface{}
}

type RelationType string

const (
    Uses        RelationType = "uses"
    Extends     RelationType = "extends"
    Implements  RelationType = "implements"
    Calls       RelationType = "calls"
    Creates     RelationType = "creates"
    References  RelationType = "references"
    Tests       RelationType = "tests"
)
```

### Relationships for Our Example

```
LoginTest (class)
  ├─ extends → BaseTest
  ├─ uses → LoginPage
  └─ uses → DashboardPage

testLoginWithValidCredentials (method in LoginTest)
  ├─ creates → LoginPage instance
  ├─ calls → LoginPage.login()
  ├─ creates → DashboardPage instance
  ├─ calls → DashboardPage.isWelcomeMessageDisplayed()
  └─ calls → DashboardPage.getUsername()

LoginPage (class)
  ├─ extends → BasePage
  └─ has_elements → [usernameField, passwordField, loginButton, errorMessage]

LoginPage.login() (method)
  ├─ calls → enterUsername()
  ├─ calls → enterPassword()
  └─ calls → clickLoginButton()

LoginPage.enterUsername() (method)
  └─ uses_field → usernameField

usernameField (field in LoginPage)
  ├─ has_annotation → @FindBy(id="username")
  └─ locator → id=username
```

### Graph Queries We Can Do

```go
// Query: Find all tests that use LoginPage
func (kg *KnowledgeGraph) FindTestsUsingPageObject(pageObjectName string) []string {
    var tests []string
    for _, rel := range kg.Relationships {
        if rel.Type == Uses && rel.To == pageObjectName {
            class := kg.Classes[rel.From]
            if class.Type == "test" {
                tests = append(tests, rel.From)
            }
        }
    }
    return tests
    // Returns: ["LoginTest"]
}

// Query: Find all methods in LoginPage
func (kg *KnowledgeGraph) FindMethodsInClass(className string) []*MethodNode {
    class := kg.Classes[className]
    return class.Methods
    // Returns: [enterUsername, enterPassword, clickLoginButton, login, isErrorDisplayed, getErrorMessage]
}

// Query: Find all elements in LoginPage
func (kg *KnowledgeGraph) FindElementsInPageObject(pageObjectName string) []*FieldNode {
    class := kg.Classes[pageObjectName]
    var elements []*FieldNode
    for _, field := range class.Fields {
        if field.HasAnnotation("FindBy") {
            elements = append(elements, field)
        }
    }
    return elements
    // Returns: [usernameField, passwordField, loginButton, errorMessage]
}

// Query: What does login() method do?
func (kg *KnowledgeGraph) GetMethodImplementation(className, methodName string) *MethodDetails {
    method := kg.Methods[className + "." + methodName]
    return &MethodDetails{
        Name: method.Name,
        Calls: method.Calls,  // ["enterUsername", "enterPassword", "clickLoginButton"]
        UsesFields: method.UsesFields,  // []
        LineStart: method.LineStart,
        LineEnd: method.LineEnd,
        SourceCode: method.SourceCode
    }
}
```

---

## 🔍 Step 5: Intent Parsing (Understanding User Messages)

### User: "edit login test"

**Parse Intent**:
```go
type IntentParser struct {
    kg *KnowledgeGraph
    nlp *NLPProcessor
}

func (ip *IntentParser) Parse(userMessage string) *Intent {
    // 1. Extract keywords
    keywords := ip.nlp.ExtractKeywords(userMessage)
    // ["edit", "login", "test"]

    // 2. Classify intent
    action := ip.classifyAction(keywords)
    // "edit" → MODIFY_CODE

    // 3. Identify target
    target := ip.identifyTarget(keywords, ip.kg)
    // "login test" → could be:
    //   - LoginTest.java (test class)
    //   - LoginTest.testLoginWithValidCredentials() (method)
    //   - LoginPage.login() (page object method)

    // 4. Disambiguate
    candidates := ip.findCandidates(keywords, ip.kg)

    return &Intent{
        Action: action,  // "MODIFY_CODE"
        Targets: candidates,
        Keywords: keywords,
        OriginalMessage: userMessage
    }
}

func (ip *IntentParser) findCandidates(keywords []string, kg *KnowledgeGraph) []Target {
    var candidates []Target

    // Search pattern 1: "login" + "test" → LoginTest class
    if contains(keywords, "login") && contains(keywords, "test") {
        if class := kg.FindClass("LoginTest"); class != nil {
            candidates = append(candidates, Target{
                Type: "test_class",
                Name: "LoginTest",
                FilePath: "src/test/java/tests/LoginTest.java",
                Confidence: 0.9
            })
        }
    }

    // Search pattern 2: "login" → LoginPage.login() method
    if contains(keywords, "login") {
        if method := kg.FindMethod("LoginPage", "login"); method != nil {
            candidates = append(candidates, Target{
                Type: "page_object_method",
                Name: "LoginPage.login()",
                FilePath: "src/test/java/pages/LoginPage.java",
                LineStart: method.LineStart,
                LineEnd: method.LineEnd,
                Confidence: 0.7
            })
        }
    }

    // Search pattern 3: Feature files (if Cucumber)
    if contains(keywords, "login") {
        if feature := kg.FindFeature("login"); feature != nil {
            candidates = append(candidates, Target{
                Type: "feature_file",
                Name: "login.feature",
                FilePath: "src/test/resources/features/login.feature",
                Confidence: 0.6
            })
        }
    }

    // Sort by confidence
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Confidence > candidates[j].Confidence
    })

    return candidates
}
```

**Result**:
```json
{
  "intent": {
    "action": "MODIFY_CODE",
    "targets": [
      {
        "type": "test_class",
        "name": "LoginTest",
        "file_path": "src/test/java/tests/LoginTest.java",
        "confidence": 0.9
      },
      {
        "type": "page_object_method",
        "name": "LoginPage.login()",
        "file_path": "src/test/java/pages/LoginPage.java",
        "line_start": 51,
        "line_end": 55,
        "confidence": 0.7
      }
    ]
  }
}
```

**Clarification** (if needed):
```
Copilot: "I found multiple matches for 'login test'. Which did you mean?"
  [1] LoginTest.java (test class) ← Most likely
  [2] LoginPage.login() method
  [3] All login-related tests

User selects: [1]
```

---

## 🧩 Step 6: Context Assembly (Build Perfect Context)

### User Selected: LoginTest.java

**Context Builder Logic**:
```go
type ContextBuilder struct {
    kg *KnowledgeGraph
    db *Database
}

func (cb *ContextBuilder) BuildContext(target Target, intent Intent) *Context {
    ctx := &Context{
        Target: target,
        Intent: intent,
    }

    // 1. Get the main file
    ctx.MainFile = cb.loadFile(target.FilePath)

    // 2. Get dependencies
    ctx.Dependencies = cb.loadDependencies(target)

    // 3. Get framework context
    ctx.Framework = cb.getFramework()

    // 4. Get coding style
    ctx.CodingStyle = cb.extractCodingStyle(target)

    // 5. Get similar code
    ctx.SimilarCode = cb.findSimilarCode(target)

    return ctx
}

func (cb *ContextBuilder) loadDependencies(target Target) []Dependency {
    var deps []Dependency

    // LoginTest depends on LoginPage and DashboardPage
    classNode := cb.kg.Classes[target.Name]

    for _, rel := range cb.kg.Relationships {
        if rel.From == target.Name && rel.Type == Uses {
            depClass := cb.kg.Classes[rel.To]
            deps = append(deps, Dependency{
                Name: rel.To,
                FilePath: depClass.FilePath,
                FullCode: cb.loadFile(depClass.FilePath),
                Type: "page_object"
            })
        }
    }

    // Also include BaseTest (extends)
    if classNode.Extends != "" {
        baseClass := cb.kg.Classes[classNode.Extends]
        deps = append(deps, Dependency{
            Name: classNode.Extends,
            FilePath: baseClass.FilePath,
            FullCode: cb.loadFile(baseClass.FilePath),
            Type: "base_test"
        })
    }

    return deps
}
```

**Final Context Package**:
```json
{
  "intent": "MODIFY_CODE",
  "target": {
    "type": "test_class",
    "name": "LoginTest",
    "file_path": "src/test/java/tests/LoginTest.java"
  },

  "main_file": {
    "path": "src/test/java/tests/LoginTest.java",
    "content": "... full code ...",
    "structure": {
      "class_name": "LoginTest",
      "extends": "BaseTest",
      "methods": [
        {
          "name": "testLoginWithValidCredentials",
          "line_start": 11,
          "line_end": 27,
          "annotations": ["@Test(description=\"...\")"],
          "calls": ["LoginPage.login", "DashboardPage.isWelcomeMessageDisplayed"]
        },
        {
          "name": "testLoginWithInvalidCredentials",
          "line_start": 29,
          "line_end": 44
        }
      ]
    }
  },

  "dependencies": [
    {
      "name": "LoginPage",
      "type": "page_object",
      "file_path": "src/test/java/pages/LoginPage.java",
      "content": "... full LoginPage.java code ...",
      "elements": [
        {"name": "usernameField", "locator": "id=username"},
        {"name": "passwordField", "locator": "id=password"},
        {"name": "loginButton", "locator": "css=button[type='submit']"},
        {"name": "errorMessage", "locator": "className=error-message"}
      ],
      "methods": [
        {"name": "login", "params": ["String username", "String password"]},
        {"name": "enterUsername", "params": ["String username"]},
        {"name": "enterPassword", "params": ["String password"]},
        {"name": "clickLoginButton", "returns": "DashboardPage"},
        {"name": "isErrorDisplayed", "returns": "boolean"},
        {"name": "getErrorMessage", "returns": "String"}
      ]
    },
    {
      "name": "DashboardPage",
      "type": "page_object",
      "file_path": "src/test/java/pages/DashboardPage.java",
      "content": "... full DashboardPage.java code ...",
      "methods": [
        {"name": "isWelcomeMessageDisplayed", "returns": "boolean"},
        {"name": "getUsername", "returns": "String"}
      ]
    },
    {
      "name": "BaseTest",
      "type": "base_test",
      "file_path": "src/test/java/tests/BaseTest.java",
      "content": "... full BaseTest.java code ...",
      "provides": ["WebDriver driver", "@BeforeMethod setup", "@AfterMethod teardown"]
    }
  ],

  "framework": {
    "type": "selenium-testng",
    "test_runner": "testng",
    "pattern": "page_object_model"
  },

  "coding_style": {
    "naming_conventions": {
      "test_methods": "testActionWithCondition",
      "variables": "camelCase",
      "page_objects": "PascalCase + 'Page'"
    },
    "assertion_style": {
      "library": "org.testng.Assert",
      "pattern": "Assert.assertTrue(condition, message)",
      "always_has_message": true
    },
    "wait_strategy": {
      "type": "WebDriverWait",
      "timeout": 10,
      "unit": "seconds"
    },
    "structure": {
      "pattern": "AAA (Arrange, Act, Assert)",
      "comments": "Yes, before each section"
    }
  },

  "similar_code": [
    {
      "name": "DashboardTest.java",
      "relevance": 0.85,
      "reason": "Similar test structure and BaseTest usage",
      "sample_method": "... testDashboardDisplay() method ..."
    }
  ],

  "application_map": null
}
```

---

## 🎯 Step 7: Handle Different User Queries

### Example 1: "I need to fix the login function"

**Intent Parsing**:
```
Keywords: ["fix", "login", "function"]
Action: MODIFY_CODE
Target: Ambiguous - could be:
  1. LoginPage.login() method (page object) ← Most likely for "function"
  2. LoginTest methods
```

**Context Built**:
```json
{
  "target": {
    "type": "page_object_method",
    "class": "LoginPage",
    "method": "login",
    "file_path": "src/test/java/pages/LoginPage.java",
    "line_start": 51,
    "line_end": 55
  },
  "main_file": {
    "content": "... LoginPage.java full code ...",
    "highlight_method": "login"
  },
  "dependencies": [
    {
      "name": "BasePage",
      "reason": "LoginPage extends BasePage"
    }
  ],
  "related_tests": [
    {
      "name": "LoginTest",
      "methods_that_call_this": ["testLoginWithValidCredentials", "testLoginWithInvalidCredentials"],
      "reason": "These tests will be affected by changes to login() method"
    }
  ]
}
```

### Example 2: "Where is the username element defined?"

**Intent Parsing**:
```
Keywords: ["where", "username", "element", "defined"]
Action: FIND_LOCATION
Target: Field definition
```

**Response**:
```
Copilot: "The username element is defined in LoginPage.java:

File: src/test/java/pages/LoginPage.java
Line: 15

@FindBy(id = "username")
private WebElement usernameField;

This element is used in:
- LoginPage.enterUsername() method (line 34)

It's referenced by:
- LoginTest.testLoginWithValidCredentials()
- LoginTest.testLoginWithInvalidCredentials()
- LoginTest.testLoginButtonDisabledWithEmptyUsername()
"
```

**Context Sent to LLM**:
```json
{
  "query_type": "FIND_LOCATION",
  "answer": {
    "element_name": "usernameField",
    "locator": {
      "type": "id",
      "value": "username"
    },
    "defined_in": {
      "file": "src/test/java/pages/LoginPage.java",
      "class": "LoginPage",
      "line": 15,
      "code": "@FindBy(id = \"username\")\nprivate WebElement usernameField;"
    },
    "used_in_methods": [
      {
        "method": "LoginPage.enterUsername",
        "line": 34,
        "usage": "usernameField.clear(); usernameField.sendKeys(username);"
      }
    ],
    "used_in_tests": [
      "LoginTest.testLoginWithValidCredentials",
      "LoginTest.testLoginWithInvalidCredentials"
    ]
  }
}
```

### Example 3: "Add a test for password reset"

**Intent Parsing**:
```
Keywords: ["add", "test", "password", "reset"]
Action: CREATE_NEW_TEST
Target: New test method or file
```

**Context Built**:
```json
{
  "intent": "CREATE_NEW_TEST",
  "test_scenario": "password reset",

  "check_existing": {
    "password_reset_test_exists": false,
    "password_reset_page_object_exists": false,
    "password_reset_in_application_map": true
  },

  "application_map": {
    "page": "PasswordReset",
    "url": "/reset-password",
    "elements": [
      {"name": "emailField", "locator": "id=email"},
      {"name": "resetButton", "locator": "css=button.reset-btn"},
      {"name": "successMessage", "locator": "className=success-msg"}
    ]
  },

  "similar_code": {
    "test_example": "LoginTest.java",
    "page_example": "LoginPage.java",
    "reason": "Follow same structure for consistency"
  },

  "tasks": [
    {
      "task": "create_page_object",
      "file": "src/test/java/pages/PasswordResetPage.java",
      "based_on": "LoginPage.java pattern"
    },
    {
      "task": "create_test",
      "file": "src/test/java/tests/PasswordResetTest.java",
      "based_on": "LoginTest.java pattern"
    }
  ]
}
```

---

## 🏗️ Implementation in Go Binary

### Main Components

```go
// copilot-core/cmd/server/main.go

type Server struct {
    kg              *KnowledgeGraph
    db              *Database
    frameworkDetector *FrameworkDetector
    intentParser    *IntentParser
    contextBuilder  *ContextBuilder
    cloudClient     *CloudClient
}

func (s *Server) HandleUserMessage(msg string) {
    // 1. Parse intent
    intent := s.intentParser.Parse(msg)

    // 2. Disambiguate if needed
    if len(intent.Targets) > 1 {
        target := s.askUserToChoose(intent.Targets)
        intent.SelectedTarget = target
    } else {
        intent.SelectedTarget = intent.Targets[0]
    }

    // 3. Build context
    ctx := s.contextBuilder.BuildContext(intent.SelectedTarget, intent)

    // 4. Send to cloud backend
    response := s.cloudClient.SendContext(ctx)

    // 5. Stream back to extension
    s.streamToExtension(response)
}
```

### Knowledge Graph Builder

```go
// copilot-core/pkg/knowledge/builder.go

type KnowledgeGraphBuilder struct {
    parser      *JavaParser
    db          *Database
    workspacePath string
}

func (kgb *KnowledgeGraphBuilder) Build() *KnowledgeGraph {
    kg := NewKnowledgeGraph()

    // 1. Find all Java files
    files := kgb.findAllJavaFiles()

    // 2. Parse each file
    for _, file := range files {
        classData := kgb.parser.ParseFile(file)

        // Store in database
        kgb.db.StoreClass(classData)

        // Add to knowledge graph
        kg.AddClass(classData)
    }

    // 3. Build relationships
    kg.BuildRelationships()

    // 4. Index for search
    kg.BuildSearchIndex()

    return kg
}
```

### Search Index for Fast Lookups

```go
// copilot-core/pkg/knowledge/search_index.go

type SearchIndex struct {
    // Map: keyword → locations
    classIndex     map[string][]string  // "Login" → ["LoginTest", "LoginPage", "LoginSteps"]
    methodIndex    map[string][]string  // "login" → ["LoginPage.login", "LoginSteps.userLogsIn"]
    fieldIndex     map[string][]string  // "username" → ["LoginPage.usernameField"]

    // Fuzzy search
    fuzzyMatcher   *FuzzyMatcher

    // Semantic embeddings (for advanced search)
    embeddings     map[string][]float32
}

func (si *SearchIndex) Search(query string) []SearchResult {
    // 1. Exact match
    exact := si.exactMatch(query)

    // 2. Fuzzy match
    fuzzy := si.fuzzyMatcher.Match(query, 0.8)

    // 3. Semantic match (using embeddings)
    semantic := si.semanticMatch(query)

    // 4. Combine and rank
    return si.combineResults(exact, fuzzy, semantic)
}
```

---

## 📊 Summary: From User Message to Perfect Context

### The Complete Flow

```
User: "edit login test"
    ↓
[1] Intent Parser
    - Keywords: ["edit", "login", "test"]
    - Action: MODIFY_CODE
    - Search index for "login" + "test"
    - Find: LoginTest.java (confidence: 0.9)
    ↓
[2] Knowledge Graph Query
    - Get LoginTest class node
    - Find relationships:
      * extends BaseTest
      * uses LoginPage
      * uses DashboardPage
    - Get all methods in LoginTest
    - Get all dependencies
    ↓
[3] Context Builder
    - Load main file: LoginTest.java (full code)
    - Load dependencies:
      * LoginPage.java (full code + elements + methods)
      * DashboardPage.java (full code + methods)
      * BaseTest.java (full code + setup/teardown)
    - Extract coding style from LoginTest
    - Find similar tests: DashboardTest.java
    - Get framework info: TestNG + Selenium
    ↓
[4] Context Package (JSON)
    {
      "target": "LoginTest.java",
      "main_file": {...},
      "dependencies": [...],
      "framework": {...},
      "coding_style": {...},
      "similar_code": [...]
    }
    ↓
[5] Send to Cloud Backend
    Cloud builds prompt with ALL this context
    LLM has EVERYTHING it needs to understand the test
    ↓
[6] AI Response
    Perfect code that matches project style
```

---

## 🎯 Next Steps: Implementation Priority

### Week 1: Deep AST Parsing
1. Enhance Java parser to extract ALL information
2. Store in SQLite with proper relationships
3. Test with real Selenium projects

### Week 2: Knowledge Graph
1. Build relationship mapper
2. Create search index
3. Test queries: "find all tests using LoginPage"

### Week 3: Intent Parser
1. Keyword extraction
2. Target disambiguation
3. Clarification prompts

### Week 4: Context Builder
1. Dependency loader
2. Coding style extractor
3. Context optimization (don't send too much)

### Week 5: Integration
1. Connect to cloud backend
2. End-to-end testing
3. Validate context quality

---

**Key Insight**:

> The better we understand the workspace, the better the AI will perform.
> This deep understanding is our MOAT - no generic copilot does this for test automation.

**Question for you**: Should we start by enhancing the JavaParser to extract all this detailed information (fields with @FindBy, method calls, etc.)?
