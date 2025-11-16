package generator

import (
	"fmt"
	"strings"
)

// ScenarioBasedGenerator generates complete framework from scenarios
type ScenarioBasedGenerator struct {
	baseGenerator *CodeGenerator
}

// NewScenarioBasedGenerator creates a new scenario-based generator
func NewScenarioBasedGenerator(workspacePath string) *ScenarioBasedGenerator {
	return &ScenarioBasedGenerator{
		baseGenerator: NewCodeGeneratorWithWorkspace(workspacePath),
	}
}

// ScenarioGenerationRequest contains all data needed to generate framework from scenarios
type ScenarioGenerationRequest struct {
	Recording          *RecordingData
	Analysis           *AnalysisData
	SelectedScenarios  []ScenarioData
	GeneratePageObject bool
	GenerateTests      bool
	GenerateFeatures   bool
	GenerateSteps      bool
}

// RecordingData represents browser recording data
type RecordingData struct {
	SessionID    string
	URL          string
	Title        string
	Elements     []RecordedElement
	Interactions []RecordedInteraction
}

// RecordedElement represents a recorded DOM element
type RecordedElement struct {
	ID          string
	Type        string
	Label       string
	Placeholder string
	Required    bool
	Pattern     string
	MinLength   int
	MaxLength   int
}

// RecordedInteraction represents a user interaction
type RecordedInteraction struct {
	Action  string
	Element string
	Value   string
	URL     string
}

// AnalysisData represents GPT analysis results
type AnalysisData struct {
	PageType               string
	PagePurpose            string
	ValidationRules        map[string]ValidationRule
	SuccessURL             string
	FailureURL             string
	RecommendedPageObjects []PageObjectSpec
}

// ValidationRule represents an inferred validation rule
type ValidationRule struct {
	Required  bool
	Format    string
	MinLength int
	MaxLength int
	Pattern   string
}

// ScenarioData represents a test scenario
type ScenarioData struct {
	Category       string
	Name           string
	Description    string
	Priority       string
	TestData       map[string]interface{}
	ExpectedResult string
	Reasoning      string
}

// PageObjectSpec represents recommended page object structure
type PageObjectSpec struct {
	Name     string
	Elements []string
	Methods  []string
}

// GenerationResult contains all generated code
type GenerationResult struct {
	PageObjects    []*GeneratedCode
	TestClasses    []*GeneratedCode
	FeatureFiles   []*GeneratedCode
	StepDefinitions []*GeneratedCode
	Summary        string
}

// GenerateFromScenarios generates complete framework from scenarios
func (g *ScenarioBasedGenerator) GenerateFromScenarios(req ScenarioGenerationRequest) (*GenerationResult, error) {
	result := &GenerationResult{
		PageObjects:     []*GeneratedCode{},
		TestClasses:     []*GeneratedCode{},
		FeatureFiles:    []*GeneratedCode{},
		StepDefinitions: []*GeneratedCode{},
	}

	// 1. Generate Page Objects
	if req.GeneratePageObject {
		for _, poSpec := range req.Analysis.RecommendedPageObjects {
			pageObject, err := g.generatePageObjectFromRecording(req.Recording, poSpec)
			if err != nil {
				return nil, fmt.Errorf("failed to generate page object %s: %w", poSpec.Name, err)
			}
			result.PageObjects = append(result.PageObjects, pageObject)
		}
	}

	// 2. Generate Test Classes (Data-Driven with TestNG)
	if req.GenerateTests {
		testClass, err := g.generateDataDrivenTest(req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate test class: %w", err)
		}
		result.TestClasses = append(result.TestClasses, testClass)
	}

	// 3. Generate Feature Files (BDD)
	if req.GenerateFeatures {
		featureFile, err := g.generateFeatureFromScenarios(req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate feature file: %w", err)
		}
		result.FeatureFiles = append(result.FeatureFiles, featureFile)
	}

	// 4. Generate Step Definitions
	if req.GenerateSteps {
		stepDefs, err := g.generateStepDefinitionsFromScenarios(req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate step definitions: %w", err)
		}
		result.StepDefinitions = append(result.StepDefinitions, stepDefs)
	}

	// Generate summary
	result.Summary = g.generateSummary(result)

	return result, nil
}

// generatePageObjectFromRecording creates a Page Object from recording
func (g *ScenarioBasedGenerator) generatePageObjectFromRecording(recording *RecordingData, spec PageObjectSpec) (*GeneratedCode, error) {
	var code strings.Builder

	// Package and imports
	code.WriteString("package pages;\n\n")
	code.WriteString("import org.openqa.selenium.WebDriver;\n")
	code.WriteString("import org.openqa.selenium.WebElement;\n")
	code.WriteString("import org.openqa.selenium.support.FindBy;\n")
	code.WriteString("import org.openqa.selenium.support.PageFactory;\n\n")

	// Class declaration
	code.WriteString(fmt.Sprintf("public class %s {\n", spec.Name))
	code.WriteString("    private WebDriver driver;\n\n")

	// Elements (with @FindBy annotations)
	for _, elem := range recording.Elements {
		if elem.ID != "" {
			code.WriteString(fmt.Sprintf("    @FindBy(id = \"%s\")\n", elem.ID))
			elementName := g.toFieldName(elem.ID)
			code.WriteString(fmt.Sprintf("    private WebElement %s;\n\n", elementName))
		}
	}

	// Constructor
	code.WriteString(fmt.Sprintf("    public %s(WebDriver driver) {\n", spec.Name))
	code.WriteString("        this.driver = driver;\n")
	code.WriteString("        PageFactory.initElements(driver, this);\n")
	code.WriteString("    }\n\n")

	// Methods based on recording interactions
	methods := g.inferMethodsFromInteractions(recording)
	for _, method := range methods {
		code.WriteString(method)
		code.WriteString("\n")
	}

	// Utility methods
	code.WriteString("    public boolean isPageLoaded() {\n")
	if len(recording.Elements) > 0 {
		firstElement := g.toFieldName(recording.Elements[0].ID)
		code.WriteString(fmt.Sprintf("        return %s.isDisplayed();\n", firstElement))
	} else {
		code.WriteString("        return true;\n")
	}
	code.WriteString("    }\n")

	code.WriteString("}\n")

	// Extract page name from spec
	pageName := strings.Replace(spec.Name, "Page", "", 1)
	if pageName == spec.Name {
		pageName = spec.Name
	}

	return &GeneratedCode{
		FileName: spec.Name + ".java",
		FilePath: "src/main/java/pages/" + spec.Name + ".java",
		Content:  code.String(),
		Language: "java",
		Type:     "pageObject",
	}, nil
}

// generateDataDrivenTest creates TestNG test with DataProvider
func (g *ScenarioBasedGenerator) generateDataDrivenTest(req ScenarioGenerationRequest) (*GeneratedCode, error) {
	var code strings.Builder

	// Extract test name from page type
	testName := g.capitalizeFirst(req.Analysis.PageType) + "Test"

	// Package and imports
	code.WriteString("package tests;\n\n")
	code.WriteString("import org.testng.annotations.*;\n")
	code.WriteString("import org.testng.Assert;\n")
	code.WriteString("import org.openqa.selenium.WebDriver;\n")
	code.WriteString("import pages.*;\n\n")

	// Class declaration
	code.WriteString(fmt.Sprintf("public class %s extends BaseTest {\n\n", testName))

	// DataProvider
	code.WriteString(fmt.Sprintf("    @DataProvider(name = \"%sScenarios\")\n", req.Analysis.PageType))
	code.WriteString("    public Object[][] getTestData() {\n")
	code.WriteString("        return new Object[][] {\n")

	// Generate data rows from scenarios
	for _, scenario := range req.SelectedScenarios {
		code.WriteString("            {")

		// Add test data values
		values := []string{}
		for field, value := range scenario.TestData {
			values = append(values, fmt.Sprintf("\"%s\": \"%v\"", field, value))
		}

		// Add expected result
		shouldSucceed := scenario.ExpectedResult == "success"
		code.WriteString(fmt.Sprintf("%t, \"%s\"", shouldSucceed, scenario.Name))

		// Add test data
		for _, elem := range req.Recording.Elements {
			if val, ok := scenario.TestData[elem.ID]; ok {
				code.WriteString(fmt.Sprintf(", \"%v\"", val))
			}
		}

		code.WriteString("},\n")
	}

	code.WriteString("        };\n")
	code.WriteString("    }\n\n")

	// Test method
	code.WriteString(fmt.Sprintf("    @Test(dataProvider = \"%sScenarios\")\n", req.Analysis.PageType))
	code.WriteString("    public void test" + g.capitalizeFirst(req.Analysis.PageType) + "(")

	// Method parameters
	params := []string{"boolean shouldSucceed", "String scenario"}
	for _, elem := range req.Recording.Elements {
		if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
			params = append(params, fmt.Sprintf("String %s", elem.ID))
		}
	}
	code.WriteString(strings.Join(params, ", "))
	code.WriteString(") {\n")

	// Test implementation
	if len(req.Analysis.RecommendedPageObjects) > 0 {
		mainPageName := req.Analysis.RecommendedPageObjects[0].Name
		code.WriteString(fmt.Sprintf("        %s page = new %s(driver);\n", mainPageName, mainPageName))
		code.WriteString(fmt.Sprintf("        driver.get(\"%s\");\n\n", req.Recording.URL))

		// Fill form fields
		code.WriteString("        // Fill form\n")
		for _, elem := range req.Recording.Elements {
			if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
				fieldName := g.toFieldName(elem.ID)
				code.WriteString(fmt.Sprintf("        page.%s.clear();\n", fieldName))
				code.WriteString(fmt.Sprintf("        page.%s.sendKeys(%s);\n", fieldName, elem.ID))
			}
		}

		// Submit
		code.WriteString("\n        // Submit\n")
		for _, elem := range req.Recording.Elements {
			if elem.Type == "submit" || elem.Type == "button" {
				fieldName := g.toFieldName(elem.ID)
				code.WriteString(fmt.Sprintf("        page.%s.click();\n", fieldName))
				break
			}
		}

		// Validation
		code.WriteString("\n        // Validate\n")
		code.WriteString("        if (shouldSucceed) {\n")
		code.WriteString(fmt.Sprintf("            Assert.assertEquals(driver.getCurrentUrl(), \"%s\", scenario + \": Should navigate to success page\");\n", req.Analysis.SuccessURL))
		code.WriteString("        } else {\n")
		if req.Analysis.FailureURL != "" {
			code.WriteString(fmt.Sprintf("            Assert.assertEquals(driver.getCurrentUrl(), \"%s\", scenario + \": Should stay on same page\");\n", req.Analysis.FailureURL))
		} else {
			code.WriteString("            // Check for error message\n")
			code.WriteString("            Assert.assertTrue(page.hasError(), scenario + \": Should show error\");\n")
		}
		code.WriteString("        }\n")
	}

	code.WriteString("    }\n")
	code.WriteString("}\n")

	return &GeneratedCode{
		FileName: testName + ".java",
		FilePath: "src/test/java/tests/" + testName + ".java",
		Content:  code.String(),
		Language: "java",
		Type:     "test",
	}, nil
}

// generateFeatureFromScenarios creates Gherkin feature file
func (g *ScenarioBasedGenerator) generateFeatureFromScenarios(req ScenarioGenerationRequest) (*GeneratedCode, error) {
	var code strings.Builder

	// Feature header
	featureName := g.capitalizeFirst(req.Analysis.PageType) + " Functionality"
	code.WriteString(fmt.Sprintf("Feature: %s\n", featureName))
	code.WriteString(fmt.Sprintf("  %s\n\n", req.Analysis.PagePurpose))

	// Scenario Outline
	code.WriteString("  Scenario Outline: Test " + req.Analysis.PageType + " with various inputs\n")
	code.WriteString(fmt.Sprintf("    Given I am on the %s page\n", req.Analysis.PageType))

	// When steps (from recording interactions)
	for _, elem := range req.Recording.Elements {
		if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
			code.WriteString(fmt.Sprintf("    When I enter \"<%s>\" in %s field\n", elem.ID, elem.Label))
		}
	}

	// Find submit button
	for _, elem := range req.Recording.Elements {
		if elem.Type == "submit" || elem.Type == "button" {
			code.WriteString("    And I click the submit button\n")
			break
		}
	}

	// Then step
	code.WriteString("    Then I should <result>\n\n")

	// Examples table
	code.WriteString("    Examples:\n")

	// Header row
	headers := []string{}
	for _, elem := range req.Recording.Elements {
		if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
			headers = append(headers, elem.ID)
		}
	}
	headers = append(headers, "result")

	code.WriteString("      | " + strings.Join(headers, " | ") + " |\n")

	// Data rows from scenarios
	for _, scenario := range req.SelectedScenarios {
		values := []string{}
		for _, header := range headers[:len(headers)-1] {
			if val, ok := scenario.TestData[header]; ok {
				values = append(values, fmt.Sprintf("%v", val))
			} else {
				values = append(values, "")
			}
		}

		// Result
		if scenario.ExpectedResult == "success" {
			values = append(values, "see success page")
		} else {
			values = append(values, "see error message")
		}

		code.WriteString("      | " + strings.Join(values, " | ") + " |\n")
	}

	fileName := strings.ToLower(req.Analysis.PageType) + ".feature"

	return &GeneratedCode{
		FileName: fileName,
		FilePath: "src/test/resources/features/" + fileName,
		Content:  code.String(),
		Language: "gherkin",
		Type:     "feature",
	}, nil
}

// generateStepDefinitionsFromScenarios creates Cucumber step definitions
func (g *ScenarioBasedGenerator) generateStepDefinitionsFromScenarios(req ScenarioGenerationRequest) (*GeneratedCode, error) {
	var code strings.Builder

	stepClassName := g.capitalizeFirst(req.Analysis.PageType) + "Steps"

	// Package and imports
	code.WriteString("package steps;\n\n")
	code.WriteString("import io.cucumber.java.en.*;\n")
	code.WriteString("import org.testng.Assert;\n")
	code.WriteString("import pages.*;\n\n")

	// Class declaration
	code.WriteString(fmt.Sprintf("public class %s extends BaseSteps {\n\n", stepClassName))

	if len(req.Analysis.RecommendedPageObjects) > 0 {
		mainPageName := req.Analysis.RecommendedPageObjects[0].Name
		code.WriteString(fmt.Sprintf("    private %s page;\n\n", mainPageName))

		// Given step
		code.WriteString(fmt.Sprintf("    @Given(\"I am on the %s page\")\n", req.Analysis.PageType))
		code.WriteString("    public void navigateToPage() {\n")
		code.WriteString(fmt.Sprintf("        driver.get(\"%s\");\n", req.Recording.URL))
		code.WriteString(fmt.Sprintf("        page = new %s(driver);\n", mainPageName))
		code.WriteString("    }\n\n")

		// When steps for each field
		for _, elem := range req.Recording.Elements {
			if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
				label := elem.Label
				if label == "" {
					label = elem.ID
				}
				code.WriteString(fmt.Sprintf("    @When(\"I enter {string} in %s field\")\n", label))
				code.WriteString(fmt.Sprintf("    public void enter%s(String value) {\n", g.capitalizeFirst(elem.ID)))
				fieldName := g.toFieldName(elem.ID)
				code.WriteString(fmt.Sprintf("        page.%s.clear();\n", fieldName))
				code.WriteString(fmt.Sprintf("        page.%s.sendKeys(value);\n", fieldName))
				code.WriteString("    }\n\n")
			}
		}

		// Submit step
		code.WriteString("    @When(\"I click the submit button\")\n")
		code.WriteString("    public void clickSubmit() {\n")
		for _, elem := range req.Recording.Elements {
			if elem.Type == "submit" || elem.Type == "button" {
				fieldName := g.toFieldName(elem.ID)
				code.WriteString(fmt.Sprintf("        page.%s.click();\n", fieldName))
				break
			}
		}
		code.WriteString("    }\n\n")

		// Then steps
		code.WriteString("    @Then(\"I should see success page\")\n")
		code.WriteString("    public void verifySuccess() {\n")
		code.WriteString(fmt.Sprintf("        Assert.assertEquals(driver.getCurrentUrl(), \"%s\");\n", req.Analysis.SuccessURL))
		code.WriteString("    }\n\n")

		code.WriteString("    @Then(\"I should see error message\")\n")
		code.WriteString("    public void verifyError() {\n")
		code.WriteString("        // Verify error message is displayed\n")
		code.WriteString("        Assert.assertTrue(page.hasError());\n")
		code.WriteString("    }\n")
	}

	code.WriteString("}\n")

	return &GeneratedCode{
		FileName: stepClassName + ".java",
		FilePath: "src/test/java/steps/" + stepClassName + ".java",
		Content:  code.String(),
		Language: "java",
		Type:     "stepDefinition",
	}, nil
}

// Helper methods

func (g *ScenarioBasedGenerator) toFieldName(id string) string {
	// Convert id to camelCase field name
	parts := strings.Split(id, "-")
	if len(parts) == 1 {
		parts = strings.Split(id, "_")
	}

	if len(parts) == 1 {
		return id + "Field"
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += g.capitalizeFirst(parts[i])
	}
	return result + "Field"
}

func (g *ScenarioBasedGenerator) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (g *ScenarioBasedGenerator) inferMethodsFromInteractions(recording *RecordingData) []string {
	methods := []string{}

	// Generate a main action method based on interactions
	var methodCode strings.Builder
	methodCode.WriteString("    public void performAction(")

	// Parameters from input elements
	params := []string{}
	for _, elem := range recording.Elements {
		if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
			params = append(params, fmt.Sprintf("String %s", elem.ID))
		}
	}

	methodCode.WriteString(strings.Join(params, ", "))
	methodCode.WriteString(") {\n")

	// Method body
	for _, elem := range recording.Elements {
		if elem.Type == "email" || elem.Type == "text" || elem.Type == "password" {
			fieldName := g.toFieldName(elem.ID)
			methodCode.WriteString(fmt.Sprintf("        %s.clear();\n", fieldName))
			methodCode.WriteString(fmt.Sprintf("        %s.sendKeys(%s);\n", fieldName, elem.ID))
		}
	}

	// Click submit
	for _, elem := range recording.Elements {
		if elem.Type == "submit" || elem.Type == "button" {
			fieldName := g.toFieldName(elem.ID)
			methodCode.WriteString(fmt.Sprintf("        %s.click();\n", fieldName))
			break
		}
	}

	methodCode.WriteString("    }\n")
	methods = append(methods, methodCode.String())

	// Add hasError method
	methods = append(methods, `    public boolean hasError() {
        // Check for error message - implement based on your application
        // Example: return driver.findElements(By.className("error-message")).size() > 0;
        return false;
    }`)

	return methods
}

func (g *ScenarioBasedGenerator) generateSummary(result *GenerationResult) string {
	var summary strings.Builder

	summary.WriteString("Code Generation Complete!\n\n")
	summary.WriteString(fmt.Sprintf("Generated %d Page Objects:\n", len(result.PageObjects)))
	for _, po := range result.PageObjects {
		summary.WriteString(fmt.Sprintf("  - %s\n", po.FileName))
	}

	summary.WriteString(fmt.Sprintf("\nGenerated %d Test Classes:\n", len(result.TestClasses)))
	for _, test := range result.TestClasses {
		summary.WriteString(fmt.Sprintf("  - %s\n", test.FileName))
	}

	summary.WriteString(fmt.Sprintf("\nGenerated %d Feature Files:\n", len(result.FeatureFiles)))
	for _, feature := range result.FeatureFiles {
		summary.WriteString(fmt.Sprintf("  - %s\n", feature.FileName))
	}

	summary.WriteString(fmt.Sprintf("\nGenerated %d Step Definition Classes:\n", len(result.StepDefinitions)))
	for _, step := range result.StepDefinitions {
		summary.WriteString(fmt.Sprintf("  - %s\n", step.FileName))
	}

	return summary.String()
}
