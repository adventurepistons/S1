package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yourusername/copilot-core/pkg/analyzer"
	"github.com/yourusername/copilot-core/pkg/generator"
	"github.com/yourusername/copilot-core/pkg/llm"
	"github.com/yourusername/copilot-core/pkg/parser"
	"github.com/yourusername/copilot-core/pkg/storage"
)

type Server struct {
	parser    *parser.JavaParser
	analyzer  *analyzer.WorkspaceAnalyzer
	generator *generator.CodeGenerator
	llmClient *llm.Client
	storage   *storage.Manager
}

func main() {
	// Check if running as CLI or server
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
	} else {
		runCLI()
	}
}

// Start HTTP server for extension communication
func startServer() {
	port := os.Getenv("COPILOT_PORT")
	if port == "" {
		port = "47823" // Random high port
	}

	server := &Server{
		parser:    parser.NewJavaParser(),
		analyzer:  analyzer.NewWorkspaceAnalyzer(),
		generator: generator.NewCodeGenerator(),
		llmClient: llm.NewClient(),
		storage:   storage.NewManager(),
	}

	// Initialize storage
	if err := server.storage.Initialize(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/analyze", server.handleAnalyze)
	http.HandleFunc("/chat", server.handleChat)
	http.HandleFunc("/generate", server.handleGenerate)
	http.HandleFunc("/search", server.handleSearch)
	http.HandleFunc("/health", server.handleHealth)

	log.Printf("Copilot Core server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// CLI mode for direct execution
func runCLI() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		if len(os.Args) < 3 {
			fmt.Println("Usage: copilot-core parse <file-path>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		p := parser.NewJavaParser()
		result, err := p.ParseFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Usage: copilot-core analyze <workspace-path> [--json]")
			os.Exit(1)
		}
		workspacePath := os.Args[2]

		// Check for JSON flag
		jsonOutput := false
		if len(os.Args) > 3 && os.Args[3] == "--json" {
			jsonOutput = true
		}

		a := analyzer.NewWorkspaceAnalyzer()
		result, err := a.Analyze(workspacePath)
		if err != nil {
			log.Fatal(err)
		}

		if jsonOutput {
			// JSON output for programmatic use
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			// Human-readable health report
			printHealthReport(result)
		}

	case "health":
		// Alias for analyze (non-JSON)
		if len(os.Args) < 3 {
			fmt.Println("Usage: copilot-core health <workspace-path>")
			os.Exit(1)
		}
		workspacePath := os.Args[2]
		a := analyzer.NewWorkspaceAnalyzer()
		result, err := a.Analyze(workspacePath)
		if err != nil {
			log.Fatal(err)
		}
		printHealthReport(result)

	case "generate":
		if len(os.Args) < 5 {
			fmt.Println("Usage: copilot-core generate <type> <workspace> <spec>")
			fmt.Println("Types: pageObject, test, feature, stepDefinition")
			os.Exit(1)
		}
		genType := os.Args[2]
		workspacePath := os.Args[3]
		spec := os.Args[4]

		g := generator.NewCodeGeneratorWithWorkspace(workspacePath)
		req := generator.GenerateRequest{
			Type:          genType,
			Specification: spec,
			Framework:     "selenium-java",
		}

		code, err := g.Generate(req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\n✅ Generated: %s\n", code.FileName)
		fmt.Printf("📁 Path: %s\n\n", code.FilePath)
		fmt.Println(code.Content)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("🤖 Copilot Core - Test Automation AI Engine")
	fmt.Println("\nUsage:")
	fmt.Println("  copilot-core server                         Start HTTP server")
	fmt.Println("  copilot-core parse <file>                   Parse a Java file")
	fmt.Println("  copilot-core analyze <workspace> [--json]   Analyze workspace (full report)")
	fmt.Println("  copilot-core health <workspace>             Project health report")
	fmt.Println("  copilot-core generate <type> <workspace> <spec>")
	fmt.Println("      Types: pageObject, test, feature, stepDefinition")
	fmt.Println("\nExamples:")
	fmt.Println("  copilot-core analyze /path/to/project")
	fmt.Println("  copilot-core health /path/to/project")
	fmt.Println("  copilot-core generate pageObject . \"Login Page\"")
}

func printHealthReport(analysis *analyzer.FrameworkAnalysis) {
	fmt.Println("\n═══════════════════════════════════════════════════════")
	fmt.Println("       📊 PROJECT HEALTH REPORT")
	fmt.Println("═══════════════════════════════════════════════════════\n")

	// 1. Framework Information
	fmt.Println("🔧 FRAMEWORK")
	fmt.Printf("   Framework: %s\n", analysis.Framework)
	fmt.Printf("   Test Runner: %s\n", analysis.TestRunner)
	if analysis.ParallelMode != "" {
		fmt.Printf("   Parallel Execution: %s (%d threads)\n", analysis.ParallelMode, analysis.ThreadCount)
	}

	// 2. Dependencies
	if len(analysis.Dependencies) > 0 {
		fmt.Printf("   Dependencies: %d\n", len(analysis.Dependencies))
		if analysis.PomInfo != nil {
			// Show key dependencies
			if seleniumDep := analysis.PomInfo.GetDependencyByArtifact("selenium-java"); seleniumDep != nil {
				fmt.Printf("     • Selenium: %s\n", seleniumDep.Version)
			}
			if testngDep := analysis.PomInfo.GetDependencyByArtifact("testng"); testngDep != nil {
				fmt.Printf("     • TestNG: %s\n", testngDep.Version)
			}
		}
	}

	// 3. Code Statistics
	fmt.Println("\n📂 CODE STATISTICS")
	fmt.Printf("   Page Objects: %d\n", len(analysis.PageObjects))
	fmt.Printf("   Test Cases: %d\n", len(analysis.TestCases))
	if len(analysis.StepDefinitions) > 0 {
		fmt.Printf("   Step Definitions: %d\n", len(analysis.StepDefinitions))
	}
	if len(analysis.FeatureFiles) > 0 {
		fmt.Printf("   Feature Files: %d\n", len(analysis.FeatureFiles))
	}
	fmt.Printf("   Utility Classes: %d\n", len(analysis.Utilities))

	// 4. Page Objects Detail
	if len(analysis.PageObjects) > 0 {
		fmt.Println("\n📄 PAGE OBJECTS")
		totalElements := 0
		for _, po := range analysis.PageObjects {
			totalElements += len(po.Elements)
			fmt.Printf("   • %s (%d elements, %d methods)\n",
				po.ClassName, len(po.Elements), len(po.Methods))
		}
		fmt.Printf("   Total Elements: %d\n", totalElements)
	}

	// 5. Test Cases Detail
	if len(analysis.TestCases) > 0 {
		fmt.Println("\n✅ TEST CASES")
		totalTests := 0
		for _, tc := range analysis.TestCases {
			totalTests += len(tc.TestMethods)
			if len(tc.TestMethods) > 0 {
				fmt.Printf("   • %s (%d tests)\n", tc.ClassName, len(tc.TestMethods))
			}
		}
		fmt.Printf("   Total Test Methods: %d\n", totalTests)
	}

	// 6. BDD/Cucumber Info
	if len(analysis.FeatureFiles) > 0 {
		fmt.Println("\n🥒 BDD/CUCUMBER")
		totalScenarios := 0
		for _, feature := range analysis.FeatureFiles {
			totalScenarios += len(feature.Scenarios)
			fmt.Printf("   • %s (%d scenarios)\n",
				feature.Feature.Name, len(feature.Scenarios))
		}
		fmt.Printf("   Total Scenarios: %d\n", totalScenarios)

		// Step matching
		if len(analysis.StepMatches) > 0 {
			fmt.Printf("   Matched Steps: %d\n", len(analysis.StepMatches))
		}
	}

	// 7. Project Structure
	fmt.Println("\n📁 PROJECT STRUCTURE")
	if analysis.Structure.PagesDir != "" {
		fmt.Printf("   Pages: %s\n", analysis.Structure.PagesDir)
	}
	if analysis.Structure.TestsDir != "" {
		fmt.Printf("   Tests: %s\n", analysis.Structure.TestsDir)
	}
	if analysis.Structure.StepsDir != "" {
		fmt.Printf("   Steps: %s\n", analysis.Structure.StepsDir)
	}
	if analysis.Structure.ResourcesDir != "" {
		fmt.Printf("   Resources: %s\n", analysis.Structure.ResourcesDir)
	}

	// 8. Health Score
	score := calculateHealthScore(analysis)
	fmt.Println("\n💯 HEALTH SCORE")
	fmt.Printf("   Overall: %d/100 %s\n", score, getScoreEmoji(score))

	// 9. Recommendations
	recommendations := generateRecommendations(analysis)
	if len(recommendations) > 0 {
		fmt.Println("\n💡 RECOMMENDATIONS")
		for _, rec := range recommendations {
			fmt.Printf("   %s\n", rec)
		}
	}

	fmt.Println("\n═══════════════════════════════════════════════════════\n")
}

func calculateHealthScore(analysis *analyzer.FrameworkAnalysis) int {
	score := 50 // Base score

	// Has page objects (+10)
	if len(analysis.PageObjects) > 0 {
		score += 10
	}

	// Has tests (+10)
	if len(analysis.TestCases) > 0 {
		score += 10
	}

	// Has pom.xml (+10)
	if analysis.PomInfo != nil {
		score += 10
	}

	// Has testng.xml (+5)
	if analysis.TestNGSuite != nil {
		score += 5
	}

	// Has feature files (+5)
	if len(analysis.FeatureFiles) > 0 {
		score += 5
	}

	// Has step definitions (+5)
	if len(analysis.StepDefinitions) > 0 {
		score += 5
	}

	// Parallel execution configured (+5)
	if analysis.ParallelMode != "" {
		score += 5
	}

	return score
}

func getScoreEmoji(score int) string {
	if score >= 90 {
		return "🌟 Excellent"
	} else if score >= 75 {
		return "✅ Good"
	} else if score >= 60 {
		return "⚠️  Fair"
	} else {
		return "❌ Needs Work"
	}
}

func generateRecommendations(analysis *analyzer.FrameworkAnalysis) []string {
	var recommendations []string

	if len(analysis.PageObjects) == 0 {
		recommendations = append(recommendations, "• Add Page Object Model pattern for better maintainability")
	}

	if len(analysis.TestCases) == 0 {
		recommendations = append(recommendations, "• Create test cases for your page objects")
	}

	if analysis.PomInfo == nil {
		recommendations = append(recommendations, "• Add pom.xml for dependency management")
	}

	if analysis.ParallelMode == "" && len(analysis.TestCases) > 5 {
		recommendations = append(recommendations, "• Enable parallel execution in testng.xml for faster test runs")
	}

	if len(analysis.FeatureFiles) > 0 && len(analysis.StepDefinitions) == 0 {
		recommendations = append(recommendations, "• Add step definitions for your feature files")
	}

	if len(analysis.StepMatches) > 0 {
		// Check for unmatched steps
		recommendations = append(recommendations, "• Review step-to-definition matching")
	}

	return recommendations
}

// ===== HTTP Handlers =====

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WorkspacePath string `json:"workspacePath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	analysis, err := s.analyzer.Analyze(req.WorkspacePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store in database
	if err := s.storage.StoreAnalysis(analysis); err != nil {
		log.Printf("Failed to store analysis: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message       string                 `json:"message"`
		WorkspacePath string                 `json:"workspacePath"`
		Context       map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get relevant context via semantic search
	relevantCode, err := s.storage.SemanticSearch(req.Message, req.WorkspacePath, 5)
	if err != nil {
		log.Printf("Semantic search failed: %v", err)
	}

	// Build context for LLM
	context := s.buildContext(relevantCode, req.Context)

	// Call LLM
	response, err := s.llmClient.Chat(req.Message, context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save chat history
	s.storage.SaveChatMessage(req.Message, "user")
	s.storage.SaveChatMessage(response, "assistant")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response": response,
	})
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type          string                 `json:"type"` // "pageObject", "test", "utility"
		Specification string                 `json:"specification"`
		WorkspacePath string                 `json:"workspacePath"`
		Context       map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get relevant context
	relevantCode, _ := s.storage.SemanticSearch(req.Specification, req.WorkspacePath, 3)
	context := s.buildContext(relevantCode, req.Context)

	// Generate code
	code, err := s.generator.Generate(req.Type, req.Specification, context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code":     code.Content,
		"fileName": code.FileName,
		"path":     code.FilePath,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query         string `json:"query"`
		WorkspacePath string `json:"workspacePath"`
		TopK          int    `json:"topK"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	results, err := s.storage.SemanticSearch(req.Query, req.WorkspacePath, req.TopK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *Server) buildContext(relevantCode []interface{}, additionalContext map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	context["relevantCode"] = relevantCode

	for k, v := range additionalContext {
		context[k] = v
	}

	return context
}
