package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/adventurepistons/automation-copilot/internal/ai"
	"github.com/adventurepistons/automation-copilot/internal/storage"
)

// LSP server for Test Automation Copilot
// Communicates with VSCode extension via JSON-RPC over stdio

func main() {
	// Setup logging to file (can't use stdout - used for JSON-RPC)
	logFile, err := os.OpenFile("copilot-lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Starting Test Automation Copilot LSP server...")

	// Initialize server
	server := NewServer()
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Server handles LSP communication
type Server struct {
	contextBuilder *ai.ContextBuilder
	db             *storage.Database
	requestID      int
}

// NewServer creates a new LSP server
func NewServer() *Server {
	return &Server{
		requestID: 0,
	}
}

// Start begins listening for JSON-RPC messages on stdin
func (s *Server) Start() error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var msg JSONRPCMessage
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("Decode error: %v", err)
			continue
		}

		log.Printf("Received: %s", msg.Method)

		// Handle message
		response := s.handleMessage(msg)

		// Send response
		if response != nil {
			if err := encoder.Encode(response); err != nil {
				log.Printf("Encode error: %v", err)
			}
		}
	}
}

// handleMessage processes incoming JSON-RPC messages
func (s *Server) handleMessage(msg JSONRPCMessage) *JSONRPCMessage {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)

	case "initialized":
		// Client finished initialization
		return nil

	case "testCopilot/initialize":
		return s.handleCopilotInitialize(msg)

	case "testCopilot/generate":
		return s.handleGenerate(msg)

	case "testCopilot/chat":
		return s.handleChat(msg)

	case "testCopilot/buildIndex":
		return s.handleBuildIndex(msg)

	case "testCopilot/getStats":
		return s.handleGetStats(msg)

	case "shutdown":
		return s.handleShutdown(msg)

	default:
		log.Printf("Unknown method: %s", msg.Method)
		return s.errorResponse(msg.ID, -32601, "Method not found")
	}
}

// handleInitialize handles LSP initialize request
func (s *Server) handleInitialize(msg JSONRPCMessage) *JSONRPCMessage {
	capabilities := map[string]interface{}{
		"textDocumentSync": 1,
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"capabilities": capabilities,
		},
	}
}

// handleCopilotInitialize initializes the copilot backend
func (s *Server) handleCopilotInitialize(msg JSONRPCMessage) *JSONRPCMessage {
	var params struct {
		WorkspaceRoot string `json:"workspaceRoot"`
		APIKey        string `json:"apiKey"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, -32602, "Invalid params")
	}

	log.Printf("Initializing copilot for workspace: %s", params.WorkspaceRoot)

	// Initialize database
	dbPath := params.WorkspaceRoot + "/.copilot/copilot.db"
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		return s.errorResponse(msg.ID, -32603, fmt.Sprintf("Failed to initialize database: %v", err))
	}
	s.db = db

	// Create context builder
	config := ai.DefaultConfig()
	config.LLMConfig.APIKey = params.APIKey

	contextBuilder, err := ai.NewContextBuilder(db.GetDB(), config)
	if err != nil {
		return s.errorResponse(msg.ID, -32603, fmt.Sprintf("Failed to create context builder: %v", err))
	}
	s.contextBuilder = contextBuilder

	log.Println("Copilot initialized successfully")

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"success": true,
			"message": "Copilot initialized",
		},
	}
}

// handleGenerate handles code generation requests
func (s *Server) handleGenerate(msg JSONRPCMessage) *JSONRPCMessage {
	var params struct {
		Request string `json:"request"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, -32602, "Invalid params")
	}

	if s.contextBuilder == nil {
		return s.errorResponse(msg.ID, -32603, "Copilot not initialized")
	}

	log.Printf("Generating code for: %s", params.Request)

	// Generate code
	ctx := context.Background()
	result, err := s.contextBuilder.GenerateCode(ctx, params.Request)
	if err != nil {
		return s.errorResponse(msg.ID, -32603, fmt.Sprintf("Generation failed: %v", err))
	}

	// Build response
	response := map[string]interface{}{
		"code":          result.ParsedResponse.GeneratedCode.Code,
		"className":     result.ParsedResponse.GeneratedCode.ClassName,
		"packageName":   result.ParsedResponse.GeneratedCode.PackageName,
		"codeType":      result.ParsedResponse.GeneratedCode.CodeType,
		"suggestedPath": result.ParsedResponse.GeneratedCode.SuggestedPath,
		"cost":          result.LLMResponse.Cost,
		"tokens": map[string]interface{}{
			"input":  result.LLMResponse.InputTokens,
			"output": result.LLMResponse.OutputTokens,
			"cached": result.LLMResponse.CacheReadTokens,
		},
		"latency":    result.TotalTime.Milliseconds(),
		"validation": result.ValidationErrors,
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  response,
	}
}

// handleChat handles chat requests
func (s *Server) handleChat(msg JSONRPCMessage) *JSONRPCMessage {
	// Same as handleGenerate for now
	return s.handleGenerate(msg)
}

// handleBuildIndex handles index building
func (s *Server) handleBuildIndex(msg JSONRPCMessage) *JSONRPCMessage {
	if s.contextBuilder == nil {
		return s.errorResponse(msg.ID, -32603, "Copilot not initialized")
	}

	log.Println("Building indexes...")

	if err := s.contextBuilder.BuildIndexes(); err != nil {
		return s.errorResponse(msg.ID, -32603, fmt.Sprintf("Build index failed: %v", err))
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"success": true,
			"message": "Indexes built successfully",
		},
	}
}

// handleGetStats handles stats requests
func (s *Server) handleGetStats(msg JSONRPCMessage) *JSONRPCMessage {
	if s.contextBuilder == nil {
		return s.errorResponse(msg.ID, -32603, "Copilot not initialized")
	}

	stats, err := s.contextBuilder.GetStats()
	if err != nil {
		return s.errorResponse(msg.ID, -32603, fmt.Sprintf("Get stats failed: %v", err))
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  stats,
	}
}

// handleShutdown handles shutdown request
func (s *Server) handleShutdown(msg JSONRPCMessage) *JSONRPCMessage {
	log.Println("Shutting down...")

	if s.db != nil {
		s.db.Close()
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  nil,
	}
}

// errorResponse creates an error response
func (s *Server) errorResponse(id interface{}, code int, message string) *JSONRPCMessage {
	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}

// JSONRPCMessage represents a JSON-RPC 2.0 message
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
