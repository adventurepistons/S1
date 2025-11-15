package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/embeddings"
	"github.com/yourusername/copilot-core/pkg/llm"
)

// Server represents the HTTP/WebSocket server
type Server struct {
	router         *gin.Engine
	db             *database.DB
	vectorStore    *embeddings.VectorStore
	semanticSearch *embeddings.SemanticSearch
	llmClient      *llm.Client
	contextBuilder *llm.ContextBuilder
	upgrader       websocket.Upgrader
	port           int
}

// ServerConfig configures the server
type ServerConfig struct {
	Port           int
	DB             *database.DB
	VectorStore    *embeddings.VectorStore
	SemanticSearch *embeddings.SemanticSearch
	LLMClient      *llm.Client
	ContextBuilder *llm.ContextBuilder
}

// NewServer creates a new server
func NewServer(config ServerConfig) (*Server, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database is required")
	}
	if config.SemanticSearch == nil {
		return nil, fmt.Errorf("semantic search is required")
	}
	if config.LLMClient == nil {
		return nil, fmt.Errorf("LLM client is required")
	}
	if config.ContextBuilder == nil {
		return nil, fmt.Errorf("context builder is required")
	}

	port := config.Port
	if port == 0 {
		port = 8080
	}

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		router:         gin.Default(),
		db:             config.DB,
		vectorStore:    config.VectorStore,
		semanticSearch: config.SemanticSearch,
		llmClient:      config.LLMClient,
		contextBuilder: config.ContextBuilder,
		port:           port,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				// TODO: Restrict in production
				return true
			},
		},
	}

	// Setup routes
	s.setupRoutes()

	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// API v1
	v1 := s.router.Group("/api/v1")
	{
		// Workspace operations
		workspace := v1.Group("/workspace")
		{
			workspace.POST("/index", s.indexWorkspace)
			workspace.GET("/stats", s.getWorkspaceStats)
		}

		// Search operations
		search := v1.Group("/search")
		{
			search.POST("/code", s.searchCode)
			search.POST("/pageobjects", s.searchPageObjects)
			search.POST("/tests", s.searchTests)
		}

		// Code generation
		generate := v1.Group("/generate")
		{
			generate.POST("/pageobject", s.generatePageObject)
			generate.POST("/test", s.generateTest)
			generate.POST("/fix", s.fixCode)
		}

		// Chat
		chat := v1.Group("/chat")
		{
			chat.POST("/message", s.chatMessage)
		}

		// Database queries
		db := v1.Group("/db")
		{
			db.GET("/classes", s.getClasses)
			db.GET("/classes/:id", s.getClass)
			db.GET("/methods/:classId", s.getMethodsByClass)
		}
	}

	// WebSocket endpoint for streaming
	s.router.GET("/ws/stream", s.handleWebSocket)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on %s", addr)
	return s.router.Run(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close database
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	// Close vector store
	if err := s.vectorStore.Close(); err != nil {
		return fmt.Errorf("failed to close vector store: %w", err)
	}

	return nil
}

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"services": gin.H{
			"database":    s.db != nil,
			"vectorStore": s.vectorStore != nil,
			"llm":         s.llmClient.IsConfigured(),
		},
	})
}
