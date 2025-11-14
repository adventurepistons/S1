package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/copilot-core/pkg/parser"
)

type WorkspaceAnalyzer struct {
	parser *parser.JavaParser
}

type FrameworkAnalysis struct {
	Framework      string                `json:"framework"`
	TestRunner     string                `json:"testRunner"`
	PageObjects    []PageObjectInfo      `json:"pageObjects"`
	TestCases      []TestCaseInfo        `json:"testCases"`
	StepDefinitions []StepDefinitionInfo `json:"stepDefinitions"`
	Utilities      []UtilityInfo         `json:"utilities"`
	Dependencies   []string              `json:"dependencies"`
	Structure      ProjectStructure      `json:"structure"`
}

type PageObjectInfo struct {
	ClassName string         `json:"className"`
	FilePath  string         `json:"filePath"`
	Elements  []ElementInfo  `json:"elements"`
	Methods   []ActionInfo   `json:"methods"`
}

type ElementInfo struct {
	Name         string `json:"name"`
	LocatorType  string `json:"locatorType"`
	LocatorValue string `json:"locatorValue"`
}

type ActionInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters"`
}

type TestCaseInfo struct {
	ClassName       string           `json:"className"`
	FilePath        string           `json:"filePath"`
	TestMethods     []TestMethodInfo `json:"testMethods"`
	UsesPageObjects []string         `json:"usesPageObjects"`
}

type TestMethodInfo struct {
	Name        string   `json:"name"`
	Annotations []string `json:"annotations"`
	Description string   `json:"description"`
}

type StepDefinitionInfo struct {
	ClassName string     `json:"className"`
	FilePath  string     `json:"filePath"`
	Steps     []StepInfo `json:"steps"`
}

type StepInfo struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
	Method  string `json:"method"`
}

type UtilityInfo struct {
	ClassName string `json:"className"`
	FilePath  string `json:"filePath"`
	Purpose   string `json:"purpose"`
}

type ProjectStructure struct {
	RootPath     string `json:"rootPath"`
	PagesDir     string `json:"pagesDir,omitempty"`
	TestsDir     string `json:"testsDir,omitempty"`
	StepsDir     string `json:"stepsDir,omitempty"`
	UtilsDir     string `json:"utilsDir,omitempty"`
	ResourcesDir string `json:"resourcesDir,omitempty"`
}

func NewWorkspaceAnalyzer() *WorkspaceAnalyzer {
	return &WorkspaceAnalyzer{
		parser: parser.NewJavaParser(),
	}
}

func (a *WorkspaceAnalyzer) Analyze(workspacePath string) (*FrameworkAnalysis, error) {
	fmt.Printf("Analyzing workspace: %s\n", workspacePath)

	analysis := &FrameworkAnalysis{
		Framework:       "unknown",
		TestRunner:      "unknown",
		PageObjects:     []PageObjectInfo{},
		TestCases:       []TestCaseInfo{},
		StepDefinitions: []StepDefinitionInfo{},
		Utilities:       []UtilityInfo{},
		Dependencies:    []string{},
		Structure: ProjectStructure{
			RootPath: workspacePath,
		},
	}

	// Detect project structure
	analysis.Structure = a.detectProjectStructure(workspacePath)

	// Parse dependencies
	analysis.Dependencies = a.parseDependencies(workspacePath)

	// Detect framework and test runner
	analysis.Framework = a.detectFramework(analysis.Dependencies, workspacePath)
	analysis.TestRunner = a.detectTestRunner(analysis.Dependencies, workspacePath)

	// Parse all Java files
	srcPath := a.findSrcDirectory(workspacePath)
	if srcPath == "" {
		return analysis, nil
	}

	parsedClasses, err := a.parser.ParseDirectory(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	// Categorize classes
	for _, parsedClass := range parsedClasses {
		switch parsedClass.Type {
		case "pageObject":
			analysis.PageObjects = append(analysis.PageObjects, a.convertToPageObjectInfo(parsedClass))
		case "test":
			analysis.TestCases = append(analysis.TestCases, a.convertToTestCaseInfo(parsedClass))
		case "step":
			analysis.StepDefinitions = append(analysis.StepDefinitions, a.convertToStepDefinitionInfo(parsedClass))
		case "utility":
			analysis.Utilities = append(analysis.Utilities, a.convertToUtilityInfo(parsedClass))
		}
	}

	return analysis, nil
}

func (a *WorkspaceAnalyzer) detectProjectStructure(rootPath string) ProjectStructure {
	structure := ProjectStructure{RootPath: rootPath}

	possibleDirs := map[string][]string{
		"pagesDir":     {"src/test/java/pages", "src/main/java/pages", "pages"},
		"testsDir":     {"src/test/java/tests", "src/test/java", "tests"},
		"stepsDir":     {"src/test/java/steps", "src/test/java/stepDefinitions", "steps"},
		"utilsDir":     {"src/test/java/utils", "src/main/java/utils", "utils"},
		"resourcesDir": {"src/test/resources", "resources"},
	}

	for key, paths := range possibleDirs {
		for _, p := range paths {
			fullPath := filepath.Join(rootPath, p)
			if _, err := os.Stat(fullPath); err == nil {
				switch key {
				case "pagesDir":
					structure.PagesDir = fullPath
				case "testsDir":
					structure.TestsDir = fullPath
				case "stepsDir":
					structure.StepsDir = fullPath
				case "utilsDir":
					structure.UtilsDir = fullPath
				case "resourcesDir":
					structure.ResourcesDir = fullPath
				}
				break
			}
		}
	}

	return structure
}

func (a *WorkspaceAnalyzer) parseDependencies(workspacePath string) []string {
	var dependencies []string

	// Check for pom.xml (Maven)
	pomPath := filepath.Join(workspacePath, "pom.xml")
	if data, err := os.ReadFile(pomPath); err == nil {
		content := string(data)
		// Simple regex-like extraction
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "<artifactId>") {
				start := strings.Index(line, "<artifactId>")
				end := strings.Index(line, "</artifactId>")
				if start != -1 && end != -1 {
					artifactId := line[start+len("<artifactId>") : end]
					dependencies = append(dependencies, artifactId)
				}
			}
		}
	}

	return dependencies
}

func (a *WorkspaceAnalyzer) detectFramework(dependencies []string, workspacePath string) string {
	for _, dep := range dependencies {
		if strings.Contains(dep, "selenium") {
			return "selenium-java"
		}
		if strings.Contains(dep, "playwright") {
			return "playwright-ts"
		}
	}
	return "unknown"
}

func (a *WorkspaceAnalyzer) detectTestRunner(dependencies []string, workspacePath string) string {
	for _, dep := range dependencies {
		if strings.Contains(dep, "testng") {
			return "testng"
		}
		if strings.Contains(dep, "junit") {
			return "junit"
		}
		if strings.Contains(dep, "cucumber") {
			return "cucumber"
		}
	}

	// Check for testng.xml
	if _, err := os.Stat(filepath.Join(workspacePath, "testng.xml")); err == nil {
		return "testng"
	}

	return "unknown"
}

func (a *WorkspaceAnalyzer) findSrcDirectory(workspacePath string) string {
	possiblePaths := []string{
		filepath.Join(workspacePath, "src"),
		filepath.Join(workspacePath, "test"),
		workspacePath,
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

func (a *WorkspaceAnalyzer) convertToPageObjectInfo(parsedClass *parser.ParsedClass) PageObjectInfo {
	var elements []ElementInfo
	for _, field := range parsedClass.Fields {
		if field.LocatorType != "" && field.LocatorValue != "" {
			elements = append(elements, ElementInfo{
				Name:         field.Name,
				LocatorType:  field.LocatorType,
				LocatorValue: field.LocatorValue,
			})
		}
	}

	var methods []ActionInfo
	for _, method := range parsedClass.Methods {
		if method.ReturnType == "void" && !strings.HasPrefix(method.Name, "get") {
			var params []string
			for _, param := range method.Parameters {
				params = append(params, fmt.Sprintf("%s %s", param.Type, param.Name))
			}

			methods = append(methods, ActionInfo{
				Name:        method.Name,
				Description: a.inferMethodDescription(method.Name),
				Parameters:  params,
			})
		}
	}

	return PageObjectInfo{
		ClassName: parsedClass.ClassName,
		FilePath:  parsedClass.FilePath,
		Elements:  elements,
		Methods:   methods,
	}
}

func (a *WorkspaceAnalyzer) convertToTestCaseInfo(parsedClass *parser.ParsedClass) TestCaseInfo {
	var testMethods []TestMethodInfo
	for _, method := range parsedClass.Methods {
		hasTestAnnotation := false
		for _, ann := range method.Annotations {
			if strings.Contains(ann, "@Test") {
				hasTestAnnotation = true
				break
			}
		}

		if hasTestAnnotation {
			testMethods = append(testMethods, TestMethodInfo{
				Name:        method.Name,
				Annotations: method.Annotations,
				Description: a.inferMethodDescription(method.Name),
			})
		}
	}

	var usesPageObjects []string
	for _, field := range parsedClass.Fields {
		if strings.Contains(field.Type, "Page") {
			usesPageObjects = append(usesPageObjects, field.Type)
		}
	}

	return TestCaseInfo{
		ClassName:       parsedClass.ClassName,
		FilePath:        parsedClass.FilePath,
		TestMethods:     testMethods,
		UsesPageObjects: usesPageObjects,
	}
}

func (a *WorkspaceAnalyzer) convertToStepDefinitionInfo(parsedClass *parser.ParsedClass) StepDefinitionInfo {
	var steps []StepInfo

	for _, method := range parsedClass.Methods {
		for _, annotation := range method.Annotations {
			// Extract Given/When/Then patterns
			for _, stepType := range []string{"Given", "When", "Then", "And"} {
				if strings.Contains(annotation, "@"+stepType) {
					// Extract pattern from annotation
					start := strings.Index(annotation, `"`)
					end := strings.LastIndex(annotation, `"`)
					if start != -1 && end != -1 && end > start {
						pattern := annotation[start+1 : end]
						steps = append(steps, StepInfo{
							Type:    stepType,
							Pattern: pattern,
							Method:  method.Name,
						})
					}
				}
			}
		}
	}

	return StepDefinitionInfo{
		ClassName: parsedClass.ClassName,
		FilePath:  parsedClass.FilePath,
		Steps:     steps,
	}
}

func (a *WorkspaceAnalyzer) convertToUtilityInfo(parsedClass *parser.ParsedClass) UtilityInfo {
	purpose := "Utility class"

	if strings.Contains(parsedClass.ClassName, "Driver") {
		purpose = "WebDriver management"
	} else if strings.Contains(parsedClass.ClassName, "Config") {
		purpose = "Configuration management"
	} else if strings.Contains(parsedClass.ClassName, "Data") {
		purpose = "Test data management"
	} else if strings.Contains(parsedClass.ClassName, "Wait") {
		purpose = "Wait utilities"
	}

	return UtilityInfo{
		ClassName: parsedClass.ClassName,
		FilePath:  parsedClass.FilePath,
		Purpose:   purpose,
	}
}

func (a *WorkspaceAnalyzer) inferMethodDescription(methodName string) string {
	// Convert camelCase to readable description
	var result string
	for i, r := range methodName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += " "
		}
		result += string(r)
	}

	if len(result) > 0 {
		result = strings.ToUpper(result[0:1]) + result[1:]
	}

	return result
}
