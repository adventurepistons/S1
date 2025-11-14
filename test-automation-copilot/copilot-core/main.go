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
			fmt.Println("Usage: copilot-core analyze <workspace-path>")
			os.Exit(1)
		}
		workspacePath := os.Args[2]
		a := analyzer.NewWorkspaceAnalyzer()
		result, err := a.Analyze(workspacePath)
		if err != nil {
			log.Fatal(err)
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Copilot Core - Test Automation AI Engine")
	fmt.Println("\nUsage:")
	fmt.Println("  copilot-core server              Start HTTP server")
	fmt.Println("  copilot-core parse <file>        Parse a Java file")
	fmt.Println("  copilot-core analyze <workspace> Analyze workspace")
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
