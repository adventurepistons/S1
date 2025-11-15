package prompts

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// 🔒 THIS IS YOUR IP - Prompt engineering is the competitive advantage

type PromptBuilder struct {
	framework    string
	testRunner   string
	codingStyle  map[string]interface{}
	existingCode []map[string]interface{}
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		codingStyle:  make(map[string]interface{}),
		existingCode: []map[string]interface{}{},
	}
}

func (pb *PromptBuilder) SetFramework(framework, testRunner string) *PromptBuilder {
	pb.framework = framework
	pb.testRunner = testRunner
	return pb
}

func (pb *PromptBuilder) SetCodingStyle(style map[string]interface{}) *PromptBuilder {
	pb.codingStyle = style
	return pb
}

func (pb *PromptBuilder) AddContext(code map[string]interface{}) *PromptBuilder {
	pb.existingCode = append(pb.existingCode, code)
	return pb
}

// ===== Page Object Generation =====

func (pb *PromptBuilder) BuildPageObjectPrompt(spec string, elements []map[string]string) string {
	var prompt strings.Builder

	// System context
	prompt.WriteString("You are an expert Senior QA Automation Engineer with 10+ years of experience.\n")
	prompt.WriteString("You specialize in writing clean, maintainable, production-ready Page Object Model code.\n\n")

	// Framework context
	prompt.WriteString(fmt.Sprintf("Framework: %s with %s\n", pb.framework, pb.testRunner))
	prompt.WriteString(fmt.Sprintf("Task: Create a Page Object class for: %s\n\n", spec))

	// Elements to include
	if len(elements) > 0 {
		prompt.WriteString("ELEMENTS TO INCLUDE:\n")
		prompt.WriteString("```\n")
		for _, elem := range elements {
			locatorType := elem["locatorType"]
			locatorValue := elem["locatorValue"]
			name := elem["name"]

			// Format as Java @FindBy annotation
			prompt.WriteString(fmt.Sprintf("@FindBy(%s = \"%s\")\n", pb.convertToFindByType(locatorType), locatorValue))
			prompt.WriteString(fmt.Sprintf("private WebElement %s;\n\n", name))
		}
		prompt.WriteString("```\n\n")
	}

	// Example code for style matching
	if len(pb.existingCode) > 0 {
		prompt.WriteString("EXISTING CODE STYLE (match this exactly):\n")
		prompt.WriteString("```java\n")
		if example := pb.findExamplePageObject(); example != "" {
			prompt.WriteString(example)
		} else {
			// Provide default example
			prompt.WriteString(pb.getDefaultPageObjectExample())
		}
		prompt.WriteString("\n```\n\n")
	} else {
		// No existing code, provide best practice example
		prompt.WriteString("FOLLOW THIS PATTERN:\n")
		prompt.WriteString("```java\n")
		prompt.WriteString(pb.getDefaultPageObjectExample())
		prompt.WriteString("\n```\n\n")
	}

	// Detailed requirements
	prompt.WriteString("REQUIREMENTS (MUST FOLLOW ALL):\n")
	prompt.WriteString("1. Package Declaration:\n")
	prompt.WriteString("   - Use 'package pages;'\n\n")

	prompt.WriteString("2. Imports:\n")
	prompt.WriteString("   - Import only what you need\n")
	prompt.WriteString("   - Use: org.openqa.selenium.WebDriver, WebElement, support.FindBy, support.PageFactory\n\n")

	prompt.WriteString("3. Class Structure:\n")
	prompt.WriteString("   - Class name MUST end with 'Page' (e.g., LoginPage, CheckoutPage)\n")
	prompt.WriteString("   - Use PascalCase for class name\n")
	prompt.WriteString("   - Include private WebDriver driver field\n\n")

	prompt.WriteString("4. Element Locators:\n")
	prompt.WriteString("   - Use @FindBy annotations for ALL elements\n")
	prompt.WriteString("   - Make all WebElement fields private\n")
	prompt.WriteString("   - Use camelCase for field names (e.g., usernameField, loginButton)\n")
	prompt.WriteString("   - Group related elements together (forms, buttons, links)\n")
	prompt.WriteString("   - Prefer stable locators: id > css > xpath\n\n")

	prompt.WriteString("5. Constructor:\n")
	prompt.WriteString("   - Accept WebDriver as parameter\n")
	prompt.WriteString("   - Initialize with PageFactory.initElements(driver, this);\n\n")

	prompt.WriteString("6. Action Methods:\n")
	prompt.WriteString("   - Create methods for user actions (not getters for every element)\n")
	prompt.WriteString("   - Use clear, action-oriented names (e.g., login(), enterUsername(), clickSubmit())\n")
	prompt.WriteString("   - Return void for actions, or return next Page Object for navigation\n")
	prompt.WriteString("   - Add proper waits (WebDriverWait) where needed\n")
	prompt.WriteString("   - Clear fields before entering text\n\n")

	prompt.WriteString("7. Helper Methods:\n")
	prompt.WriteString("   - Add validation methods (e.g., isDisplayed(), isEnabled())\n")
	prompt.WriteString("   - Add getter methods only when needed for assertions\n\n")

	prompt.WriteString("8. Code Quality:\n")
	prompt.WriteString("   - NO Thread.sleep() - use explicit waits instead\n")
	prompt.WriteString("   - NO hardcoded waits\n")
	prompt.WriteString("   - NO commented code\n")
	prompt.WriteString("   - Add meaningful method names\n")
	prompt.WriteString("   - Follow DRY principle\n\n")

	prompt.WriteString("9. Best Practices:\n")
	prompt.WriteString("   - One action per method\n")
	prompt.WriteString("   - Methods should be small and focused\n")
	prompt.WriteString("   - Use method chaining where appropriate\n")
	prompt.WriteString("   - Handle common scenarios (login, logout, navigation)\n\n")

	// Output format
	prompt.WriteString("OUTPUT:\n")
	prompt.WriteString("- Generate ONLY the complete Java class code\n")
	prompt.WriteString("- Do NOT include any explanations or comments outside the code\n")
	prompt.WriteString("- The code should be production-ready and follow all requirements above\n")
	prompt.WriteString("- Include proper package declaration and imports\n")
	prompt.WriteString("- Format the code properly with correct indentation\n\n")

	prompt.WriteString("BEGIN GENERATION:\n")

	return prompt.String()
}

// Convert locator type to @FindBy format
func (pb *PromptBuilder) convertToFindByType(locatorType string) string {
	switch {
	case strings.Contains(locatorType, "id"):
		return "id"
	case strings.Contains(locatorType, "css"):
		return "css"
	case strings.Contains(locatorType, "xpath"):
		return "xpath"
	case strings.Contains(locatorType, "name"):
		return "name"
	case strings.Contains(locatorType, "className"):
		return "className"
	case strings.Contains(locatorType, "tagName"):
		return "tagName"
	case strings.Contains(locatorType, "linkText"):
		return "linkText"
	case strings.Contains(locatorType, "partialLinkText"):
		return "partialLinkText"
	default:
		return "css"
	}
}

// Get default Page Object example
func (pb *PromptBuilder) getDefaultPageObjectExample() string {
	return `package pages;

import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.support.FindBy;
import org.openqa.selenium.support.PageFactory;
import org.openqa.selenium.support.ui.WebDriverWait;
import org.openqa.selenium.support.ui.ExpectedConditions;
import java.time.Duration;

public class LoginPage {
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

    public void clickLoginButton() {
        wait.until(ExpectedConditions.elementToBeClickable(loginButton));
        loginButton.click();
    }

    public void login(String username, String password) {
        enterUsername(username);
        enterPassword(password);
        clickLoginButton();
    }

    // Validations
    public boolean isLoginButtonDisplayed() {
        return loginButton.isDisplayed();
    }

    public String getErrorMessage() {
        wait.until(ExpectedConditions.visibilityOf(errorMessage));
        return errorMessage.getText();
    }

    public boolean isErrorDisplayed() {
        try {
            return errorMessage.isDisplayed();
        } catch (Exception e) {
            return false;
        }
    }
}`
}

// ===== Test Case Generation =====

func (pb *PromptBuilder) BuildTestCasePrompt(spec string, pageObjects []map[string]interface{}) string {
	var prompt strings.Builder

	// System context - treat AI like a time-constrained senior engineer
	prompt.WriteString("You are a Senior QA Automation Engineer with 10+ years of experience.\n")
	prompt.WriteString("You specialize in writing clean, maintainable, production-ready test automation code.\n")
	prompt.WriteString("Think step-by-step through the test scenario before generating code.\n\n")

	// Framework context
	prompt.WriteString(fmt.Sprintf("Framework: %s with %s\n", pb.framework, pb.testRunner))
	prompt.WriteString(fmt.Sprintf("Task: Create a test case for: %s\n\n", spec))

	// Chain-of-Thought: Guide AI through reasoning steps
	prompt.WriteString("REASONING PROCESS (think through these steps):\n")
	prompt.WriteString("1. What is the test scenario trying to validate?\n")
	prompt.WriteString("2. What are the preconditions (setup)?\n")
	prompt.WriteString("3. What actions need to be performed (using page objects)?\n")
	prompt.WriteString("4. What should be asserted (expected outcomes)?\n")
	prompt.WriteString("5. What cleanup is needed (teardown)?\n\n")

	// Available page objects - explicit context
	if len(pageObjects) > 0 {
		prompt.WriteString("AVAILABLE PAGE OBJECTS (use these, DO NOT create new ones):\n")
		prompt.WriteString("```\n")
		for _, po := range pageObjects {
			prompt.WriteString(fmt.Sprintf("Class: %s\n", po["className"]))
			if methods, ok := po["methods"].([]interface{}); ok {
				prompt.WriteString("Methods:\n")
				for _, m := range methods {
					if method, ok := m.(map[string]interface{}); ok {
						prompt.WriteString(fmt.Sprintf("  - %s()\n", method["name"]))
					}
				}
			}
			prompt.WriteString("\n")
		}
		prompt.WriteString("```\n\n")
	}

	// Example test from existing code
	if len(pb.existingCode) > 0 {
		prompt.WriteString("EXISTING CODE STYLE (match this exactly):\n")
		prompt.WriteString("```java\n")
		if example := pb.findExampleTest(); example != "" {
			prompt.WriteString(example)
		} else {
			prompt.WriteString(pb.getDefaultTestExample())
		}
		prompt.WriteString("\n```\n\n")
	} else {
		prompt.WriteString("FOLLOW THIS PATTERN:\n")
		prompt.WriteString("```java\n")
		prompt.WriteString(pb.getDefaultTestExample())
		prompt.WriteString("\n```\n\n")
	}

	// Test Runner specific instructions
	prompt.WriteString("TEST RUNNER CONFIGURATION:\n")
	switch pb.testRunner {
	case "testng":
		prompt.WriteString("- Use TestNG annotations: @Test, @BeforeMethod, @AfterMethod, @AfterClass\n")
		prompt.WriteString("- Use Assert.assertEquals(), Assert.assertTrue(), etc.\n")
		prompt.WriteString("- WebDriver should be initialized in @BeforeMethod\n")
	case "junit":
		prompt.WriteString("- Use JUnit annotations: @Test, @Before, @After, @BeforeClass, @AfterClass\n")
		prompt.WriteString("- Use org.junit.Assert methods\n")
		prompt.WriteString("- WebDriver should be initialized in @Before\n")
	case "cucumber":
		prompt.WriteString("- Create step definitions with @Given, @When, @Then annotations\n")
		prompt.WriteString("- Use regex patterns in annotations\n")
		prompt.WriteString("- Store state in instance variables\n")
	}
	prompt.WriteString("\n")

	// Detailed requirements
	prompt.WriteString("REQUIREMENTS (MUST FOLLOW ALL):\n\n")

	prompt.WriteString("1. Package & Imports:\n")
	prompt.WriteString("   - Package: 'package tests;'\n")
	prompt.WriteString("   - Import necessary page objects from 'pages' package\n")
	prompt.WriteString("   - Import test framework annotations\n")
	prompt.WriteString("   - Import assertion library\n\n")

	prompt.WriteString("2. Class Structure:\n")
	prompt.WriteString("   - Class name MUST end with 'Test' (e.g., LoginTest, CheckoutTest)\n")
	prompt.WriteString("   - Use PascalCase for class name\n")
	prompt.WriteString("   - Declare WebDriver as private instance variable\n")
	prompt.WriteString("   - Declare page objects as instance variables\n\n")

	prompt.WriteString("3. Setup & Teardown:\n")
	prompt.WriteString("   - Create setup method to initialize WebDriver\n")
	prompt.WriteString("   - Initialize page objects in setup\n")
	prompt.WriteString("   - Create teardown method to quit WebDriver\n")
	prompt.WriteString("   - Use proper annotations based on test runner\n\n")

	prompt.WriteString("4. Test Methods:\n")
	prompt.WriteString("   - Follow AAA pattern: Arrange, Act, Assert\n")
	prompt.WriteString("   - Use descriptive test method names (testLoginWithValidCredentials)\n")
	prompt.WriteString("   - One logical test scenario per method\n")
	prompt.WriteString("   - Use page object methods, NOT direct WebDriver calls\n\n")

	prompt.WriteString("5. Assertions:\n")
	prompt.WriteString("   - Add meaningful assertions to verify expected behavior\n")
	prompt.WriteString("   - Assert against page object validation methods\n")
	prompt.WriteString("   - Use specific assertion methods (assertEquals, assertTrue, assertFalse)\n")
	prompt.WriteString("   - Add assertion messages for better debugging\n\n")

	prompt.WriteString("6. Best Practices:\n")
	prompt.WriteString("   - NO direct WebDriver calls in test methods\n")
	prompt.WriteString("   - NO Thread.sleep() - rely on page object waits\n")
	prompt.WriteString("   - NO hardcoded URLs - use configuration or constants\n")
	prompt.WriteString("   - Each test should be independent and repeatable\n")
	prompt.WriteString("   - Tests should clean up after themselves\n\n")

	prompt.WriteString("DO's:\n")
	prompt.WriteString("✓ Use page object methods exclusively\n")
	prompt.WriteString("✓ Add clear, descriptive test names\n")
	prompt.WriteString("✓ Include multiple assertions when validating complex scenarios\n")
	prompt.WriteString("✓ Handle both positive and negative test cases\n")
	prompt.WriteString("✓ Initialize page objects properly\n\n")

	prompt.WriteString("DON'Ts:\n")
	prompt.WriteString("✗ DO NOT use driver.findElement() in test code\n")
	prompt.WriteString("✗ DO NOT use Thread.sleep()\n")
	prompt.WriteString("✗ DO NOT hardcode test data in test methods\n")
	prompt.WriteString("✗ DO NOT create page object instances without proper setup\n")
	prompt.WriteString("✗ DO NOT skip teardown (always quit driver)\n\n")

	// Output format
	prompt.WriteString("OUTPUT:\n")
	prompt.WriteString("- Generate ONLY the complete Java test class code\n")
	prompt.WriteString("- Do NOT include explanations or markdown formatting\n")
	prompt.WriteString("- The code should be production-ready\n")
	prompt.WriteString("- Include proper package, imports, and all methods\n")
	prompt.WriteString("- Format with correct indentation\n\n")

	prompt.WriteString("BEGIN GENERATION:\n")

	return prompt.String()
}

// Get default test example
func (pb *PromptBuilder) getDefaultTestExample() string {
	if pb.testRunner == "testng" {
		return `package tests;

import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import org.testng.Assert;
import org.testng.annotations.AfterMethod;
import org.testng.annotations.BeforeMethod;
import org.testng.annotations.Test;
import pages.LoginPage;
import pages.DashboardPage;

public class LoginTest {
    private WebDriver driver;
    private LoginPage loginPage;
    private DashboardPage dashboardPage;

    @BeforeMethod
    public void setup() {
        driver = new ChromeDriver();
        driver.manage().window().maximize();
        driver.get("https://example.com/login");

        // Initialize page objects
        loginPage = new LoginPage(driver);
        dashboardPage = new DashboardPage(driver);
    }

    @Test
    public void testLoginWithValidCredentials() {
        // Arrange
        String username = "testuser@example.com";
        String password = "SecurePass123";

        // Act
        loginPage.login(username, password);

        // Assert
        Assert.assertTrue(dashboardPage.isWelcomeMessageDisplayed(),
            "Welcome message should be displayed after successful login");
        Assert.assertEquals(dashboardPage.getUsername(), username,
            "Displayed username should match logged in user");
    }

    @Test
    public void testLoginWithInvalidCredentials() {
        // Arrange
        String username = "invalid@example.com";
        String password = "wrongpass";

        // Act
        loginPage.login(username, password);

        // Assert
        Assert.assertTrue(loginPage.isErrorDisplayed(),
            "Error message should be displayed for invalid credentials");
        Assert.assertEquals(loginPage.getErrorMessage(), "Invalid username or password",
            "Error message should indicate invalid credentials");
    }

    @AfterMethod
    public void teardown() {
        if (driver != null) {
            driver.quit();
        }
    }
}`
	}
	// JUnit example
	return `package tests;

import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import pages.LoginPage;
import static org.junit.Assert.*;

public class LoginTest {
    private WebDriver driver;
    private LoginPage loginPage;

    @Before
    public void setup() {
        driver = new ChromeDriver();
        driver.get("https://example.com/login");
        loginPage = new LoginPage(driver);
    }

    @Test
    public void testLoginWithValidCredentials() {
        loginPage.login("user@test.com", "password123");
        assertTrue("Should navigate to dashboard",
            driver.getCurrentUrl().contains("/dashboard"));
    }

    @After
    public void teardown() {
        if (driver != null) {
            driver.quit();
        }
    }
}`
}

// ===== Chat/Conversation Prompts =====

func (pb *PromptBuilder) BuildChatPrompt(userMessage string, context map[string]interface{}) string {
	var prompt strings.Builder

	// System context - consultant persona
	prompt.WriteString("You are a Senior QA Automation Consultant and Architect with 15+ years of experience.\n")
	prompt.WriteString("You specialize in helping teams build and maintain enterprise-grade test automation frameworks.\n")
	prompt.WriteString("You communicate clearly and concisely, like helping a time-constrained colleague.\n\n")

	// Framework context
	prompt.WriteString("PROJECT CONTEXT:\n")
	if pb.framework != "" {
		prompt.WriteString(fmt.Sprintf("Framework: %s\n", pb.framework))
		prompt.WriteString(fmt.Sprintf("Test Runner: %s\n", pb.testRunner))
	} else {
		prompt.WriteString("Framework: Not yet detected\n")
	}

	// Add framework summary if available
	if analysis, ok := context["analysis"].(map[string]interface{}); ok {
		if pageObjects, ok := analysis["pageObjects"].([]interface{}); ok {
			prompt.WriteString(fmt.Sprintf("Page Objects: %d\n", len(pageObjects)))
		}
		if tests, ok := analysis["testCases"].([]interface{}); ok {
			prompt.WriteString(fmt.Sprintf("Test Cases: %d\n", len(tests)))
		}
		if stepDefs, ok := analysis["stepDefinitions"].([]interface{}); ok && len(stepDefs) > 0 {
			prompt.WriteString(fmt.Sprintf("Step Definitions: %d (Cucumber detected)\n", len(stepDefs)))
		}
	}
	prompt.WriteString("\n")

	// Add relevant code context - explicit scoping like Cursor's @code
	if relevantCode, ok := context["relevantCode"].([]interface{}); ok && len(relevantCode) > 0 {
		prompt.WriteString("RELEVANT CODE CONTEXT (scoped to your question):\n")
		prompt.WriteString("These files were selected as most relevant using semantic search:\n\n")

		for i, code := range relevantCode {
			if codeMap, ok := code.(map[string]interface{}); ok {
				className := codeMap["className"]
				filePath := codeMap["filePath"]

				prompt.WriteString(fmt.Sprintf("## File %d: %s\n", i+1, className))
				if filePath != nil {
					prompt.WriteString(fmt.Sprintf("Path: %s\n", filePath))
				}

				if codeStr, ok := codeMap["code"].(string); ok {
					// Limit code context to avoid token bloat
					if len(codeStr) > 2000 {
						prompt.WriteString("```java\n" + codeStr[:2000] + "\n... (truncated)\n```\n\n")
					} else {
						prompt.WriteString("```java\n" + codeStr + "\n```\n\n")
					}
				}
			}
		}
	}

	// User's question
	prompt.WriteString("USER QUESTION:\n")
	prompt.WriteString(fmt.Sprintf("%s\n\n", userMessage))

	// Response instructions
	prompt.WriteString("INSTRUCTIONS:\n")
	prompt.WriteString("1. Analyze the question and available context carefully\n")
	prompt.WriteString("2. Think step-by-step if the question requires reasoning\n")
	prompt.WriteString("3. Provide a concise, accurate answer\n\n")

	prompt.WriteString("If suggesting code changes:\n")
	prompt.WriteString("✓ Match the existing framework's style exactly\n")
	prompt.WriteString("✓ Use the same naming conventions, package structure, and patterns\n")
	prompt.WriteString("✓ Follow Page Object Model best practices\n")
	prompt.WriteString("✓ Provide complete, production-ready code\n")
	prompt.WriteString("✓ Explain your reasoning briefly\n\n")

	prompt.WriteString("✗ DO NOT suggest anti-patterns (Thread.sleep, direct WebDriver in tests, etc.)\n")
	prompt.WriteString("✗ DO NOT provide incomplete code snippets\n")
	prompt.WriteString("✗ DO NOT introduce new dependencies without mentioning them\n\n")

	prompt.WriteString("Response style:\n")
	prompt.WriteString("- Be direct and practical\n")
	prompt.WriteString("- Use code blocks with language tags\n")
	prompt.WriteString("- Highlight important points\n")
	prompt.WriteString("- Suggest next steps if relevant\n\n")

	prompt.WriteString("BEGIN RESPONSE:\n")

	return prompt.String()
}

// ===== Code Fix/Refactor Prompts =====

func (pb *PromptBuilder) BuildFixPrompt(brokenCode string, error string, context map[string]interface{}) string {
	var prompt strings.Builder

	// System context - debugging expert
	prompt.WriteString("You are a Senior QA Automation Engineer and Expert Debugger.\n")
	prompt.WriteString("You have 15+ years of experience debugging Selenium, TestNG, JUnit, and Playwright code.\n")
	prompt.WriteString("You excel at root cause analysis and providing surgical fixes.\n\n")

	// The broken code
	prompt.WriteString("BROKEN CODE:\n")
	prompt.WriteString("```java\n" + brokenCode + "\n```\n\n")

	// The error
	prompt.WriteString("ERROR MESSAGE:\n")
	prompt.WriteString("```\n" + error + "\n```\n\n")

	// Framework context
	if pb.framework != "" {
		prompt.WriteString(fmt.Sprintf("Framework: %s with %s\n\n", pb.framework, pb.testRunner))
	}

	// Working example for reference
	if workingExample := pb.findSimilarWorkingCode(context); workingExample != "" {
		prompt.WriteString("SIMILAR WORKING CODE (for reference):\n")
		prompt.WriteString("```java\n" + workingExample + "\n```\n\n")
	}

	// Chain-of-Thought debugging process
	prompt.WriteString("DEBUGGING PROCESS (think through these steps):\n")
	prompt.WriteString("1. What type of error is this? (compilation, runtime, assertion, timeout, etc.)\n")
	prompt.WriteString("2. What is the root cause?\n")
	prompt.WriteString("3. What line(s) need to be changed?\n")
	prompt.WriteString("4. What is the correct implementation?\n")
	prompt.WriteString("5. Are there any side effects to consider?\n\n")

	// Common error patterns
	prompt.WriteString("COMMON ERROR PATTERNS TO CHECK:\n")
	prompt.WriteString("- NullPointerException → Element not initialized or not found\n")
	prompt.WriteString("- NoSuchElementException → Wrong locator or element not visible\n")
	prompt.WriteString("- StaleElementReferenceException → DOM changed, need to re-find element\n")
	prompt.WriteString("- TimeoutException → Element not ready, increase wait or fix locator\n")
	prompt.WriteString("- ElementNotInteractableException → Element covered or not visible\n")
	prompt.WriteString("- WebDriverException → Browser crash or session issue\n\n")

	// Fix requirements
	prompt.WriteString("FIX REQUIREMENTS:\n")
	prompt.WriteString("1. Identify and fix the root cause (not just symptoms)\n")
	prompt.WriteString("2. Maintain the same functionality\n")
	prompt.WriteString("3. Follow framework best practices\n")
	prompt.WriteString("4. Add proper waits if timing-related\n")
	prompt.WriteString("5. Improve locator stability if element-related\n")
	prompt.WriteString("6. Add null checks where appropriate\n\n")

	prompt.WriteString("DO's:\n")
	prompt.WriteString("✓ Fix the actual root cause\n")
	prompt.WriteString("✓ Use explicit waits instead of Thread.sleep\n")
	prompt.WriteString("✓ Improve locator strategies if needed\n")
	prompt.WriteString("✓ Add defensive null checks\n")
	prompt.WriteString("✓ Explain what was wrong and why\n\n")

	prompt.WriteString("DON'Ts:\n")
	prompt.WriteString("✗ DO NOT just add try-catch to hide errors\n")
	prompt.WriteString("✗ DO NOT add Thread.sleep as a fix\n")
	prompt.WriteString("✗ DO NOT change functionality to avoid the error\n")
	prompt.WriteString("✗ DO NOT introduce new anti-patterns\n\n")

	// Output format
	prompt.WriteString("OUTPUT FORMAT:\n")
	prompt.WriteString("Provide your response in this exact format:\n\n")
	prompt.WriteString("## Root Cause\n")
	prompt.WriteString("[Explain what caused the error]\n\n")
	prompt.WriteString("## Fixed Code\n")
	prompt.WriteString("```java\n")
	prompt.WriteString("[Complete fixed code]\n")
	prompt.WriteString("```\n\n")
	prompt.WriteString("## What Changed\n")
	prompt.WriteString("[Explain the specific changes made and why]\n\n")
	prompt.WriteString("## Prevention\n")
	prompt.WriteString("[Optional: How to prevent this error in the future]\n\n")

	prompt.WriteString("BEGIN ANALYSIS AND FIX:\n")

	return prompt.String()
}

// ===== Data-Driven Test Prompts =====

func (pb *PromptBuilder) BuildDataDrivenPrompt(testSpec string, dataSource string) string {
	var prompt strings.Builder

	// System context
	prompt.WriteString("You are a Senior QA Automation Engineer specializing in data-driven testing.\n")
	prompt.WriteString("You have extensive experience with TestNG DataProviders, JUnit Parameterized tests, and external data sources.\n\n")

	// Task definition
	prompt.WriteString("TASK:\n")
	prompt.WriteString(fmt.Sprintf("Create a data-driven test for: %s\n", testSpec))
	prompt.WriteString(fmt.Sprintf("Data source type: %s\n\n", dataSource))

	// Framework-specific instructions
	prompt.WriteString("FRAMEWORK CONFIGURATION:\n")
	if pb.testRunner == "testng" {
		prompt.WriteString("Using: TestNG @DataProvider pattern\n\n")

		prompt.WriteString("PATTERN TO FOLLOW:\n")
		prompt.WriteString("```java\n")
		prompt.WriteString("// Data Provider Method\n")
		prompt.WriteString("@DataProvider(name = \"loginData\")\n")
		prompt.WriteString("public Object[][] provideLoginData() {\n")
		prompt.WriteString("    return new Object[][] {\n")
		prompt.WriteString("        {\"validUser\", \"validPass\", true, \"Should login successfully\"},\n")
		prompt.WriteString("        {\"invalidUser\", \"wrongPass\", false, \"Should show error message\"},\n")
		prompt.WriteString("        {\"\", \"pass123\", false, \"Should handle empty username\"}\n")
		prompt.WriteString("    };\n")
		prompt.WriteString("}\n\n")
		prompt.WriteString("// Test Method\n")
		prompt.WriteString("@Test(dataProvider = \"loginData\")\n")
		prompt.WriteString("public void testLogin(String username, String password, boolean shouldSucceed, String description) {\n")
		prompt.WriteString("    // Arrange\n")
		prompt.WriteString("    LoginPage loginPage = new LoginPage(driver);\n\n")
		prompt.WriteString("    // Act\n")
		prompt.WriteString("    loginPage.login(username, password);\n\n")
		prompt.WriteString("    // Assert\n")
		prompt.WriteString("    if (shouldSucceed) {\n")
		prompt.WriteString("        Assert.assertTrue(dashboardPage.isDisplayed(), description);\n")
		prompt.WriteString("    } else {\n")
		prompt.WriteString("        Assert.assertTrue(loginPage.isErrorDisplayed(), description);\n")
		prompt.WriteString("    }\n")
		prompt.WriteString("}\n")
		prompt.WriteString("```\n\n")
	} else if pb.testRunner == "junit" {
		prompt.WriteString("Using: JUnit @ParameterizedTest pattern\n\n")

		prompt.WriteString("PATTERN TO FOLLOW:\n")
		prompt.WriteString("```java\n")
		prompt.WriteString("@ParameterizedTest\n")
		prompt.WriteString("@CsvSource({\n")
		prompt.WriteString("    \"validUser, validPass, true\",\n")
		prompt.WriteString("    \"invalidUser, wrongPass, false\"\n")
		prompt.WriteString("})\n")
		prompt.WriteString("public void testLogin(String username, String password, boolean shouldSucceed) {\n")
		prompt.WriteString("    // Test implementation\n")
		prompt.WriteString("}\n")
		prompt.WriteString("```\n\n")
	}

	// Data source handling
	if dataSource == "excel" || dataSource == "xlsx" {
		prompt.WriteString("DATA SOURCE: Excel File\n")
		prompt.WriteString("Requirements:\n")
		prompt.WriteString("1. Use Apache POI library to read Excel\n")
		prompt.WriteString("2. Create utility method to read Excel and convert to Object[][]\n")
		prompt.WriteString("3. Handle both .xls and .xlsx formats\n")
		prompt.WriteString("4. Include proper exception handling\n\n")

		prompt.WriteString("Example Excel reader:\n")
		prompt.WriteString("```java\n")
		prompt.WriteString("private Object[][] readExcelData(String filePath) throws IOException {\n")
		prompt.WriteString("    FileInputStream file = new FileInputStream(filePath);\n")
		prompt.WriteString("    Workbook workbook = new XSSFWorkbook(file);\n")
		prompt.WriteString("    Sheet sheet = workbook.getSheetAt(0);\n")
		prompt.WriteString("    // Read data and convert to Object[][]\n")
		prompt.WriteString("}\n")
		prompt.WriteString("```\n\n")
	} else if dataSource == "json" {
		prompt.WriteString("DATA SOURCE: JSON File\n")
		prompt.WriteString("Requirements:\n")
		prompt.WriteString("1. Use Jackson or Gson to parse JSON\n")
		prompt.WriteString("2. Create data model class if needed\n")
		prompt.WriteString("3. Convert JSON array to Object[][]\n\n")
	} else if dataSource == "csv" {
		prompt.WriteString("DATA SOURCE: CSV File\n")
		prompt.WriteString("Requirements:\n")
		prompt.WriteString("1. Use OpenCSV or built-in CSV reader\n")
		prompt.WriteString("2. Handle headers properly\n")
		prompt.WriteString("3. Convert to Object[][]\n\n")
	}

	// Requirements
	prompt.WriteString("REQUIREMENTS:\n")
	prompt.WriteString("1. Create data provider method with descriptive name\n")
	prompt.WriteString("2. Include diverse test scenarios (positive, negative, edge cases)\n")
	prompt.WriteString("3. Use clear parameter names in test method\n")
	prompt.WriteString("4. Add description/message parameter for better test reports\n")
	prompt.WriteString("5. Handle data source errors gracefully\n")
	prompt.WriteString("6. Follow AAA pattern in test method\n\n")

	prompt.WriteString("DO's:\n")
	prompt.WriteString("✓ Include multiple test scenarios\n")
	prompt.WriteString("✓ Use meaningful data provider names\n")
	prompt.WriteString("✓ Add clear assertions with messages\n")
	prompt.WriteString("✓ Handle edge cases (null, empty, special chars)\n")
	prompt.WriteString("✓ Use external data files for large datasets\n\n")

	prompt.WriteString("DON'Ts:\n")
	prompt.WriteString("✗ DO NOT hardcode test data in test method\n")
	prompt.WriteString("✗ DO NOT duplicate test logic for each scenario\n")
	prompt.WriteString("✗ DO NOT skip negative test cases\n")
	prompt.WriteString("✗ DO NOT ignore data file errors\n\n")

	prompt.WriteString("OUTPUT:\n")
	prompt.WriteString("- Generate complete test class with data provider\n")
	prompt.WriteString("- Include necessary imports\n")
	prompt.WriteString("- Add utility methods if reading external files\n")
	prompt.WriteString("- Include sample data file structure if applicable\n\n")

	prompt.WriteString("BEGIN GENERATION:\n")

	return prompt.String()
}

// ===== Helper Methods (Find examples from existing code) =====

func (pb *PromptBuilder) findExamplePageObject() string {
	for _, code := range pb.existingCode {
		if code["type"] == "pageObject" {
			if codeStr, ok := code["code"].(string); ok {
				return codeStr
			}
		}
	}
	return ""
}

func (pb *PromptBuilder) findExampleTest() string {
	for _, code := range pb.existingCode {
		if code["type"] == "test" {
			if codeStr, ok := code["code"].(string); ok {
				return codeStr
			}
		}
	}
	return ""
}

func (pb *PromptBuilder) findSimilarWorkingCode(context map[string]interface{}) string {
	// Logic to find similar working code from context
	// This is a simplified version
	if similarCode, ok := context["similarCode"].(string); ok {
		return similarCode
	}
	return ""
}

// ===== System Prompts =====

func GetSystemPrompt(role string) string {
	prompts := map[string]string{
		"code_generator": `You are an expert QA automation engineer. Generate clean, maintainable test automation code following industry best practices. Always use Page Object Model pattern, explicit waits, and meaningful names.`,

		"code_reviewer": `You are a senior QA automation architect reviewing code. Identify issues with stability, maintainability, and best practices. Suggest improvements.`,

		"consultant": `You are an expert QA automation consultant. Provide strategic guidance on test automation frameworks, best practices, and implementation approaches. Be concise and practical.`,

		"debugger": `You are an expert at debugging test automation code. Analyze errors, identify root causes, and provide precise fixes. Explain what went wrong and why.`,
	}

	if prompt, ok := prompts[role]; ok {
		return prompt
	}
	return prompts["consultant"]
}

// ===== Configuration Rules (.testcopilot file support) =====
// Inspired by Cursor's .cursorrules - project-specific coding standards

type ProjectRules struct {
	NamingConventions  map[string]string `json:"naming_conventions"`
	ImportPreferences  []string          `json:"import_preferences"`
	WaitStrategy       string            `json:"wait_strategy"`
	LocatorPriority    []string          `json:"locator_priority"`
	PackageStructure   map[string]string `json:"package_structure"`
	CustomInstructions string            `json:"custom_instructions"`
}

// Load project rules from .testcopilot file
func LoadProjectRules(workspacePath string) (*ProjectRules, error) {
	// Try multiple possible locations
	possiblePaths := []string{
		workspacePath + "/.testcopilot",
		workspacePath + "/.testcopilot.json",
		workspacePath + "/testcopilot.config.json",
	}

	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			var rules ProjectRules
			if err := json.Unmarshal(data, &rules); err == nil {
				return &rules, nil
			}
		}
	}

	// Return default rules if no config found
	return GetDefaultRules(), nil
}

// Get default project rules
func GetDefaultRules() *ProjectRules {
	return &ProjectRules{
		NamingConventions: map[string]string{
			"pageObject": "PascalCase ending with 'Page'",
			"testClass":  "PascalCase ending with 'Test'",
			"testMethod": "camelCase starting with 'test'",
			"element":    "camelCase descriptive (e.g., usernameField, submitButton)",
		},
		ImportPreferences: []string{
			"org.openqa.selenium.*",
			"org.openqa.selenium.support.ui.*",
			"org.testng.Assert.*",
		},
		WaitStrategy: "Explicit waits with WebDriverWait, timeout: 10 seconds",
		LocatorPriority: []string{
			"id (most stable)",
			"data-testid",
			"name",
			"css selector",
			"xpath (least stable)",
		},
		PackageStructure: map[string]string{
			"pages": "src/test/java/pages",
			"tests": "src/test/java/tests",
			"utils": "src/test/java/utils",
		},
		CustomInstructions: "",
	}
}

// Apply project rules to a prompt
func (pb *PromptBuilder) ApplyProjectRules(basePrompt string, rules *ProjectRules) string {
	var enhanced strings.Builder

	enhanced.WriteString(basePrompt)

	// Add project-specific rules
	enhanced.WriteString("\n\n=== PROJECT-SPECIFIC RULES ===\n")
	enhanced.WriteString("(These rules MUST be followed for this specific project)\n\n")

	// Naming conventions
	if len(rules.NamingConventions) > 0 {
		enhanced.WriteString("NAMING CONVENTIONS:\n")
		for key, value := range rules.NamingConventions {
			enhanced.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		enhanced.WriteString("\n")
	}

	// Import preferences
	if len(rules.ImportPreferences) > 0 {
		enhanced.WriteString("PREFERRED IMPORTS:\n")
		for _, imp := range rules.ImportPreferences {
			enhanced.WriteString(fmt.Sprintf("- %s\n", imp))
		}
		enhanced.WriteString("\n")
	}

	// Wait strategy
	if rules.WaitStrategy != "" {
		enhanced.WriteString(fmt.Sprintf("WAIT STRATEGY: %s\n\n", rules.WaitStrategy))
	}

	// Locator priority
	if len(rules.LocatorPriority) > 0 {
		enhanced.WriteString("LOCATOR PRIORITY (use in this order):\n")
		for i, locator := range rules.LocatorPriority {
			enhanced.WriteString(fmt.Sprintf("%d. %s\n", i+1, locator))
		}
		enhanced.WriteString("\n")
	}

	// Package structure
	if len(rules.PackageStructure) > 0 {
		enhanced.WriteString("PACKAGE STRUCTURE:\n")
		for key, value := range rules.PackageStructure {
			enhanced.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		enhanced.WriteString("\n")
	}

	// Custom instructions
	if rules.CustomInstructions != "" {
		enhanced.WriteString("CUSTOM PROJECT INSTRUCTIONS:\n")
		enhanced.WriteString(rules.CustomInstructions)
		enhanced.WriteString("\n\n")
	}

	enhanced.WriteString("=== END PROJECT RULES ===\n\n")

	return enhanced.String()
}

// Generate a sample .testcopilot file for users
func GenerateSampleConfig() string {
	rules := map[string]interface{}{
		"naming_conventions": map[string]string{
			"pageObject": "PascalCase ending with 'Page' (e.g., LoginPage, CheckoutPage)",
			"testClass":  "PascalCase ending with 'Test' (e.g., LoginTest, CheckoutTest)",
			"testMethod": "camelCase starting with 'test' (e.g., testLoginWithValidCredentials)",
			"element":    "camelCase descriptive (e.g., usernameField, submitButton, errorMessage)",
		},
		"import_preferences": []string{
			"org.openqa.selenium.WebDriver",
			"org.openqa.selenium.WebElement",
			"org.openqa.selenium.support.FindBy",
			"org.openqa.selenium.support.PageFactory",
			"org.openqa.selenium.support.ui.WebDriverWait",
			"org.testng.Assert",
		},
		"wait_strategy":    "Always use explicit waits with WebDriverWait (timeout: 10 seconds). Never use Thread.sleep().",
		"locator_priority": []string{"id", "data-testid", "name", "css", "xpath"},
		"package_structure": map[string]string{
			"pages": "src/test/java/pages",
			"tests": "src/test/java/tests",
			"utils": "src/test/java/utils",
			"steps": "src/test/java/steps",
		},
		"custom_instructions": "Always include JavaDoc comments for page objects. Use fluent interface pattern where appropriate. All page object methods should return 'this' or the next page object for method chaining.",
	}

	formatted, _ := json.MarshalIndent(rules, "", "  ")
	return string(formatted)
}

// ===== Utility Functions =====

func FormatContext(context map[string]interface{}) string {
	formatted, _ := json.MarshalIndent(context, "", "  ")
	return string(formatted)
}

func ExtractCodeBlocks(response string) []string {
	var blocks []string
	lines := strings.Split(response, "\n")

	var inBlock bool
	var currentBlock strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inBlock {
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			} else {
				inBlock = true
			}
		} else if inBlock {
			currentBlock.WriteString(line + "\n")
		}
	}

	return blocks
}
