package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adventurepistons/automation-copilot/internal/detector"
	"github.com/adventurepistons/automation-copilot/internal/parser"
	"github.com/adventurepistons/automation-copilot/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: copilot <command> [args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  parse <file.java>     - Parse a Java file and show extracted data")
		fmt.Println("  index <file.java>     - Parse and save to database")
		fmt.Println("  query <class-name>    - Query database for class info")
		fmt.Println("  detect <project-dir>  - Detect framework and coding patterns")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		if len(os.Args) < 3 {
			log.Fatal("Usage: copilot parse <file.java>")
		}
		parseFile(os.Args[2])

	case "index":
		if len(os.Args) < 3 {
			log.Fatal("Usage: copilot index <file.java>")
		}
		indexFile(os.Args[2])

	case "query":
		if len(os.Args) < 3 {
			log.Fatal("Usage: copilot query <class-name>")
		}
		queryClass(os.Args[2])

	case "detect":
		if len(os.Args) < 3 {
			log.Fatal("Usage: copilot detect <project-dir>")
		}
		detectProject(os.Args[2])

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

// parseFile parses a Java file and prints the extracted data
func parseFile(filePath string) {
	fmt.Printf("Parsing: %s\n", filePath)
	fmt.Println(strings.Repeat("=", 80))

	javaParser := parser.NewJavaParser()
	classData, err := javaParser.ParseFile(filePath)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	// Print results as formatted JSON
	jsonData, err := json.MarshalIndent(classData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))

	// Print summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Summary:\n")
	fmt.Printf("  Package: %s\n", classData.Package)
	fmt.Printf("  Class: %s\n", classData.ClassName)
	fmt.Printf("  Extends: %s\n", classData.Extends)
	fmt.Printf("  Imports: %d\n", len(classData.Imports))
	fmt.Printf("  Fields: %d\n", len(classData.Fields))
	fmt.Printf("  Methods: %d\n", len(classData.Methods))
	fmt.Printf("  Is Page Object: %v\n", classData.PageObjectModel)
	fmt.Printf("  Is Test Class: %v\n", classData.TestClass)

	// Print WebElement details
	webElementCount := 0
	for _, field := range classData.Fields {
		if field.IsWebElement {
			webElementCount++
		}
	}
	if webElementCount > 0 {
		fmt.Printf("\n  WebElements (@FindBy): %d\n", webElementCount)
		for _, field := range classData.Fields {
			if field.IsWebElement {
				fmt.Printf("    - %s: %s = \"%s\"\n", field.Name, field.LocatorStrategy, field.LocatorValue)
			}
		}
	}

	// Print test methods
	testCount := 0
	for _, method := range classData.Methods {
		if method.IsTest {
			testCount++
		}
	}
	if testCount > 0 {
		fmt.Printf("\n  Test Methods: %d\n", testCount)
		for _, method := range classData.Methods {
			if method.IsTest {
				fmt.Printf("    - %s() [%s test]\n", method.Name, method.TestType)
				if len(method.MethodCalls) > 0 {
					fmt.Printf("      Method calls: %d\n", len(method.MethodCalls))
				}
				if len(method.Assertions) > 0 {
					fmt.Printf("      Assertions: %d\n", len(method.Assertions))
				}
			}
		}
	}
}

// indexFile parses a Java file and saves it to the database
func indexFile(filePath string) {
	fmt.Printf("Indexing: %s\n", filePath)

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Parse file
	javaParser := parser.NewJavaParser()
	classData, err := javaParser.ParseFile(filePath)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	// Save to database
	if err := db.SaveClassData(classData); err != nil {
		log.Fatalf("Failed to save to database: %v", err)
	}

	fmt.Printf("✓ Successfully indexed: %s.%s\n", classData.Package, classData.ClassName)
	fmt.Printf("  - %d fields\n", len(classData.Fields))
	fmt.Printf("  - %d methods\n", len(classData.Methods))

	webElementCount := 0
	for _, field := range classData.Fields {
		if field.IsWebElement {
			webElementCount++
		}
	}
	if webElementCount > 0 {
		fmt.Printf("  - %d WebElements\n", webElementCount)
	}
}

// queryClass queries the database for a class
func queryClass(className string) {
	fmt.Printf("Querying class: %s\n", className)
	fmt.Println(strings.Repeat("=", 80))

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Query class
	classData, err := db.GetClassByName(className)
	if err != nil {
		log.Fatalf("Failed to query class: %v", err)
	}

	fmt.Printf("Class: %s\n", classData.ClassName)
	fmt.Printf("Package: %s\n", classData.Package)
	fmt.Printf("File: %s\n", classData.FilePath)
	fmt.Printf("Lines: %d-%d\n", classData.LineStart, classData.LineEnd)

	// Query WebElements
	fields, err := db.GetWebElementFields(className)
	if err != nil {
		log.Fatalf("Failed to query fields: %v", err)
	}

	if len(fields) > 0 {
		fmt.Printf("\nWebElements:\n")
		for _, field := range fields {
			fmt.Printf("  %s:%d - %s: %s = \"%s\"\n",
				classData.FilePath, field.LineNumber,
				field.Name, field.LocatorStrategy, field.LocatorValue)
		}
	}
}

// detectProject detects framework and coding patterns from project
func detectProject(projectDir string) {
	fmt.Printf("Detecting framework and patterns for: %s\n", projectDir)
	fmt.Println(strings.Repeat("=", 80))

	// Get absolute path
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run framework detection
	fmt.Println("\n🔍 Detecting Framework...")
	fmt.Println(strings.Repeat("-", 80))

	frameworkDetector := detector.NewFrameworkDetector(absPath, db.GetDB())
	frameworkConfig, err := frameworkDetector.Detect()
	if err != nil {
		log.Fatalf("Failed to detect framework: %v", err)
	}

	// Print framework results
	fmt.Printf("\n✓ Framework Detection Complete\n\n")
	fmt.Printf("Framework Combination:\n")
	fmt.Printf("  %s\n\n", frameworkDetector.GetFrameworkCombination(frameworkConfig))

	fmt.Printf("Details:\n")
	if frameworkConfig.TestFramework != "" {
		fmt.Printf("  Test Framework:    %s\n", frameworkConfig.TestFramework)
	}
	if frameworkConfig.BDDFramework != "" {
		fmt.Printf("  BDD Framework:     %s\n", frameworkConfig.BDDFramework)
	}
	if frameworkConfig.APIFramework != "" {
		fmt.Printf("  API Framework:     %s\n", frameworkConfig.APIFramework)
	}
	if frameworkConfig.SeleniumVersion != "" {
		fmt.Printf("  Selenium Version:  %s\n", frameworkConfig.SeleniumVersion)
	}

	fmt.Printf("\nArchitecture Patterns:\n")
	fmt.Printf("  Page Object Model: %v\n", frameworkConfig.UsesPageObjectModel)
	fmt.Printf("  Page Factory:      %v\n", frameworkConfig.UsesPageFactory)
	fmt.Printf("  Screenplay:        %v\n", frameworkConfig.UsesScreenplay)

	fmt.Printf("\nConfiguration Files:\n")
	if frameworkConfig.POMFilePath != "" {
		fmt.Printf("  pom.xml:           %s\n", frameworkConfig.POMFilePath)
	}
	if frameworkConfig.TestNGXMLPath != "" {
		fmt.Printf("  testng.xml:        %s\n", frameworkConfig.TestNGXMLPath)
	}
	if frameworkConfig.CucumberFeatures != "" {
		fmt.Printf("  Cucumber Features: %s\n", frameworkConfig.CucumberFeatures)
	}

	// Run pattern detection
	fmt.Println("\n\n🔍 Detecting Coding Patterns...")
	fmt.Println(strings.Repeat("-", 80))

	patternDetector := detector.NewPatternDetector(absPath, db.GetDB())
	codingPatterns, err := patternDetector.DetectPatterns()
	if err != nil {
		log.Fatalf("Failed to detect patterns: %v", err)
	}

	// Print pattern results
	fmt.Printf("\n✓ Pattern Detection Complete\n\n")

	fmt.Printf("Naming Conventions:\n")
	fmt.Printf("  Test Methods:      %s\n", codingPatterns.TestMethodNaming)
	fmt.Printf("  Page Objects:      %s\n", codingPatterns.PageObjectNaming)
	fmt.Printf("  WebElements:       %s\n", codingPatterns.WebElementNaming)
	fmt.Printf("  Methods:           %s\n", codingPatterns.MethodNaming)

	fmt.Printf("\nTest Structure:\n")
	fmt.Printf("  AAA Pattern:       %v\n", codingPatterns.UsesAAAPattern)
	fmt.Printf("  Given-When-Then:   %v\n", codingPatterns.UsesGivenWhenThen)

	fmt.Printf("\nWait Strategies:\n")
	fmt.Printf("  Preferred Type:    %s\n", codingPatterns.PreferredWaitType)
	fmt.Printf("  Default Timeout:   %d seconds\n", codingPatterns.DefaultWaitTimeout)

	fmt.Printf("\nAssertions:\n")
	fmt.Printf("  Library:           %s\n", codingPatterns.AssertionLibrary)
	fmt.Printf("  Uses Messages:     %v\n", codingPatterns.UsesAssertMessages)

	fmt.Printf("\nData Patterns:\n")
	fmt.Printf("  Data Providers:    %v\n", codingPatterns.UsesDataProviders)
	fmt.Printf("  CSV Files:         %v\n", codingPatterns.UsesCSVFiles)
	fmt.Printf("  Excel Files:       %v\n", codingPatterns.UsesExcelFiles)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✓ Detection complete! Results saved to database.")
}
