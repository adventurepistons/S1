package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adventurepistons/automation-copilot/internal/ai"
	"github.com/adventurepistons/automation-copilot/internal/detector"
	"github.com/adventurepistons/automation-copilot/internal/graph"
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
		fmt.Println("  build-graph           - Build knowledge graph from indexed data")
		fmt.Println("  ask <question>        - Ask a question about the codebase")
		fmt.Println()
		fmt.Println("AI-Powered Code Generation:")
		fmt.Println("  build-index           - Build AI indexes (chunks, BM25, embeddings)")
		fmt.Println("  generate <request>    - Generate code using AI (requires ANTHROPIC_API_KEY)")
		fmt.Println("  chat                  - Interactive chat mode for multi-turn conversations")
		fmt.Println("  stats                 - Show Context Builder statistics")
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

	case "build-graph":
		buildGraph()

	case "ask":
		if len(os.Args) < 3 {
			log.Fatal("Usage: copilot ask <question>")
		}
		// Join all remaining args as the question
		question := strings.Join(os.Args[2:], " ")
		askQuestion(question)

	case "build-index":
		buildAIIndex()

	case "generate":
		if len(os.Args) < 3 {
			log.Fatal("Usage: copilot generate <request>")
		}
		// Join all remaining args as the request
		request := strings.Join(os.Args[2:], " ")
		generateCode(request)

	case "stats":
		showStats()

	case "chat":
		interactiveChat()

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

// buildGraph builds the knowledge graph from indexed data
func buildGraph() {
	fmt.Println("Building Knowledge Graph...")
	fmt.Println(strings.Repeat("=", 80))

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create knowledge graph
	kg := graph.NewKnowledgeGraph(db.GetDB())

	// Build the graph
	if err := kg.BuildGraph(); err != nil {
		log.Fatalf("Failed to build knowledge graph: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✓ Knowledge graph built successfully!")
	fmt.Println("\nYou can now use 'copilot ask <question>' to query the codebase.")
}

// askQuestion asks a natural language question about the codebase
func askQuestion(question string) {
	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create knowledge graph and query engine
	kg := graph.NewKnowledgeGraph(db.GetDB())
	qe := graph.NewQueryEngine(kg)

	// Execute query
	fmt.Printf("Question: %s\n", question)
	fmt.Println(strings.Repeat("=", 80))

	result, err := qe.Execute(question)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	if !result.Success {
		fmt.Printf("❌ %s\n", result.Message)
		fmt.Println("\nSupported queries:")
		fmt.Println("  - Where is <element>?")
		fmt.Println("  - What tests use <page>?")
		fmt.Println("  - Elements in <page>")
		fmt.Println("  - Who uses <field>?")
		fmt.Println("  - Methods in <class>")
		fmt.Println("  - List all page objects")
		fmt.Println("  - Show all tests")
		return
	}

	fmt.Printf("✓ %s\n", result.Message)
}

// buildAIIndex builds all AI indexes (chunks, BM25, embeddings, examples)
func buildAIIndex() {
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("🤖 Building AI Indexes for Code Generation")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create context builder with default config
	config := ai.DefaultConfig()
	contextBuilder, err := ai.NewContextBuilder(db.GetDB(), config)
	if err != nil {
		log.Fatalf("Failed to create context builder: %v", err)
	}

	// Build all indexes
	if err := contextBuilder.BuildIndexes(); err != nil {
		log.Fatalf("Failed to build indexes: %v", err)
	}

	fmt.Println()
	fmt.Println("🎉 All indexes built successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Start embedding service: make start-embeddings")
	fmt.Println("  2. Generate code: copilot generate \"create test for login\"")
	fmt.Println()
}

// generateCode generates code using AI
func generateCode(request string) {
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("🤖 AI-Powered Code Generation")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create context builder with API key from environment
	config := ai.DefaultConfig()
	config.LLMConfig.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	contextBuilder, err := ai.NewContextBuilder(db.GetDB(), config)
	if err != nil {
		log.Fatalf("Failed to create context builder: %v", err)
	}

	// Check if API key is available
	apiKeyAvailable := config.LLMConfig.APIKey != ""

	if apiKeyAvailable {
		// FULL GENERATION: Build context + call LLM + parse response
		fmt.Println("✅ ANTHROPIC_API_KEY found - generating code with Claude API")
		generateCodeWithLLM(contextBuilder, request)
	} else {
		// PROMPT ONLY: Just build context (original behavior)
		fmt.Println("ℹ️  ANTHROPIC_API_KEY not set - building prompt only")
		fmt.Println("   Set ANTHROPIC_API_KEY to enable automatic code generation")
		fmt.Println()
		generatePromptOnly(contextBuilder, request)
	}
}

// generateCodeWithLLM performs full end-to-end code generation
func generateCodeWithLLM(contextBuilder *ai.ContextBuilder, request string) {
	ctx := context.Background()

	// Generate code
	result, err := contextBuilder.GenerateCode(ctx, request)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	// Display generated code
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("📄 Generated Code")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Class: %s\n", result.ParsedResponse.GeneratedCode.ClassName)
	fmt.Printf("Type: %s\n", result.ParsedResponse.GeneratedCode.CodeType)
	fmt.Printf("Package: %s\n", result.ParsedResponse.GeneratedCode.PackageName)
	fmt.Println()

	fmt.Println(result.ParsedResponse.GeneratedCode.Code)
	fmt.Println()

	// Display validation warnings
	if len(result.ValidationErrors) > 0 {
		fmt.Println("════════════════════════════════════════════════════════")
		fmt.Println("⚠️  Validation Warnings")
		fmt.Println("════════════════════════════════════════════════════════")
		fmt.Println()
		for _, verr := range result.ValidationErrors {
			fmt.Printf("  [%s] %s\n", verr.Severity, verr.Message)
		}
		fmt.Println()
	}

	// Display final metrics
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("📊 Generation Metrics")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("  Total Time:            %v\n", result.TotalTime)
	fmt.Printf("  Context Building:      %v\n", result.ContextMetrics.RetrievalLatency+result.ContextMetrics.ExampleSelectionTime)
	fmt.Printf("  LLM Latency:           %v\n", result.LLMResponse.Latency)
	fmt.Printf("  Retrieved Chunks:      %d (avg relevance: %.2f)\n",
		result.ContextMetrics.HybridResultCount,
		result.ContextMetrics.AverageRelevanceScore)
	fmt.Printf("  Input Tokens:          %d (%d cached)\n",
		result.LLMResponse.InputTokens, result.LLMResponse.CacheReadTokens)
	fmt.Printf("  Output Tokens:         %d\n", result.LLMResponse.OutputTokens)
	fmt.Printf("  Actual Cost:           $%.4f\n", result.LLMResponse.Cost)
	fmt.Println()

	// Offer to save
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("💾 Save Code")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Suggested path: %s\n", result.ParsedResponse.GeneratedCode.SuggestedPath)
	fmt.Println()
	fmt.Print("Save to file? (y/n): ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		// Save code
		codeWriter := ai.NewCodeWriter(".")
		writeResult, err := codeWriter.WriteCodeInteractive(result.ParsedResponse.GeneratedCode)
		if err != nil {
			fmt.Printf("❌ Failed to write code: %v\n", err)
		} else {
			fmt.Printf("✅ Code saved to: %s\n", writeResult.FilePath)
		}
	} else {
		fmt.Println("ℹ️  Code not saved. Copy from output above.")
	}

	fmt.Println()
}

// generatePromptOnly builds context without calling LLM
func generatePromptOnly(contextBuilder *ai.ContextBuilder, request string) {
	prompt, metrics, err := contextBuilder.GenerateCodePromptOnly(request)
	if err != nil {
		log.Fatalf("Failed to build context: %v", err)
	}

	// Display prompt and metrics
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("📋 Generated Prompt for LLM")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("SYSTEM PROMPT (CACHED):")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(prompt.SystemPrompt)
	fmt.Println()

	fmt.Println("USER PROMPT (DYNAMIC):")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(prompt.UserPrompt)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("📊 Metrics")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("  Retrieval Time:        %v\n", metrics.RetrievalLatency)
	fmt.Printf("  Example Selection:     %v\n", metrics.ExampleSelectionTime)
	fmt.Printf("  Retrieved Chunks:      %d\n", metrics.HybridResultCount)
	fmt.Printf("  Average Relevance:     %.2f\n", metrics.AverageRelevanceScore)
	fmt.Printf("  Total Tokens:          %d\n", metrics.TotalTokens)
	fmt.Printf("  Cached Tokens:         %d (%.1f%%)\n",
		metrics.CachedTokens,
		float64(metrics.CachedTokens)/float64(metrics.TotalTokens)*100)
	fmt.Printf("  Dynamic Tokens:        %d\n", metrics.DynamicTokens)
	fmt.Printf("  Estimated Cost:        $%.4f\n", metrics.EstimatedCost)
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("💡 Next Steps")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("The prompt above is ready to send to an LLM!")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Copy the prompt and paste into Claude.ai or ChatGPT")
	fmt.Println("  2. Set ANTHROPIC_API_KEY and run again for automatic generation")
	fmt.Println()
	fmt.Println("The LLM will generate code matching your project's exact style.")
	fmt.Println()
}

// showStats shows Context Builder statistics
func showStats() {
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("📊 Context Builder Statistics")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create context builder
	config := ai.DefaultConfig()
	contextBuilder, err := ai.NewContextBuilder(db.GetDB(), config)
	if err != nil {
		log.Fatalf("Failed to create context builder: %v", err)
	}

	// Get statistics
	stats, err := contextBuilder.GetStats()
	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	// Display stats
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
	fmt.Println()

	// Summary
	if hybridStats, ok := stats["hybrid_retrieval"].(map[string]interface{}); ok {
		if bm25Stats, ok := hybridStats["bm25"].(map[string]interface{}); ok {
			fmt.Println("BM25 Index:")
			fmt.Printf("  Documents: %v\n", bm25Stats["total_documents"])
			fmt.Printf("  Terms:     %v\n", bm25Stats["total_terms"])
			fmt.Println()
		}

		if semanticStats, ok := hybridStats["semantic"].(map[string]interface{}); ok {
			fmt.Println("Semantic Index:")
			fmt.Printf("  Chunks:    %v\n", semanticStats["total_chunks"])
			fmt.Printf("  Embedded:  %v\n", semanticStats["with_embeddings"])
			fmt.Printf("  Coverage:  %.1f%%\n", semanticStats["coverage"])
			fmt.Println()
		}
	}

	if exampleStats, ok := stats["examples"].(map[string]interface{}); ok {
		fmt.Println("Pattern Examples:")
		fmt.Printf("  Total:     %v\n", exampleStats["total_examples"])
		fmt.Printf("  Positive:  %v\n", exampleStats["positive_examples"])
		fmt.Printf("  Negative:  %v\n", exampleStats["negative_examples"])
		fmt.Println()
	}

	fmt.Println("════════════════════════════════════════════════════════")
}

// interactiveChat provides an interactive chat interface for multi-turn code generation
func interactiveChat() {
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("💬 Interactive Chat Mode")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("This mode allows multi-turn conversations for iterative code generation.")
	fmt.Println("Type 'exit' or 'quit' to end the session.")
	fmt.Println()

	// Initialize database
	db, err := storage.NewDatabase("copilot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create context builder with API key
	config := ai.DefaultConfig()
	config.LLMConfig.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	if config.LLMConfig.APIKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY not set!")
		fmt.Println("   Please set the environment variable and try again.")
		return
	}

	contextBuilder, err := ai.NewContextBuilder(db.GetDB(), config)
	if err != nil {
		log.Fatalf("Failed to create context builder: %v", err)
	}

	// Get conversation session info
	sessionID := contextBuilder.GetConversationManager().GetSessionID()
	fmt.Printf("Session ID: %s\n", sessionID)
	fmt.Println()

	ctx := context.Background()
	turnNumber := 1

	// Chat loop
	for {
		fmt.Printf("[Turn %d] Your request: ", turnNumber)

		var userInput string
		fmt.Scanln(&userInput)

		// Check for exit
		if strings.ToLower(userInput) == "exit" || strings.ToLower(userInput) == "quit" {
			fmt.Println("\n👋 Ending chat session. Goodbye!")
			break
		}

		if strings.TrimSpace(userInput) == "" {
			continue
		}

		// Generate code
		fmt.Println()
		result, err := contextBuilder.GenerateCode(ctx, userInput)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		// Display generated code (abbreviated)
		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Printf("📄 Generated: %s (%s)\n",
			result.ParsedResponse.GeneratedCode.ClassName,
			result.ParsedResponse.GeneratedCode.CodeType)
		fmt.Println("═══════════════════════════════════════════════════════")

		// Show first 20 lines
		lines := strings.Split(result.ParsedResponse.GeneratedCode.Code, "\n")
		previewLines := 20
		if len(lines) < previewLines {
			previewLines = len(lines)
		}
		for i := 0; i < previewLines; i++ {
			fmt.Println(lines[i])
		}
		if len(lines) > previewLines {
			fmt.Printf("\n... (%d more lines)\n", len(lines)-previewLines)
		}
		fmt.Println()

		// Show metrics
		fmt.Printf("⏱️  Time: %v | 💰 Cost: $%.4f | 📊 Tokens: %d\n",
			result.TotalTime,
			result.LLMResponse.Cost,
			result.LLMResponse.InputTokens+result.LLMResponse.OutputTokens)
		fmt.Println()

		// Offer to save
		fmt.Print("Save to file? (y/n): ")
		var saveResponse string
		fmt.Scanln(&saveResponse)

		if strings.ToLower(saveResponse) == "y" {
			codeWriter := ai.NewCodeWriter(".")
			writeResult, err := codeWriter.WriteCodeInteractive(result.ParsedResponse.GeneratedCode)
			if err != nil {
				fmt.Printf("❌ Failed to save: %v\n", err)
			} else {
				fmt.Printf("✅ Saved to: %s\n", writeResult.FilePath)
			}
		}

		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────")
		fmt.Println()

		turnNumber++
	}

	// Show session summary
	messageCount, _ := contextBuilder.GetConversationManager().GetMessageCount()
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("📊 Session Summary")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("  Total turns:    %d\n", turnNumber-1)
	fmt.Printf("  Messages saved: %d\n", messageCount)
	fmt.Printf("  Session ID:     %s\n", sessionID)
	fmt.Println()
}
