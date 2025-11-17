package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourusername/copilot-core/pkg/cloud"
	pkgcontext "github.com/yourusername/copilot-core/pkg/context"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/embeddings"
)

// Server represents the HTTP/WebSocket server
type Server struct {
	router           *gin.Engine
	db               *database.DB
	vectorStore      *embeddings.VectorStore
	semanticSearch   *embeddings.SemanticSearch
	cloudClient      *cloud.CloudClient
	contextExtractor *pkgcontext.ContextExtractor
	upgrader         websocket.Upgrader
	port             int
}

// ServerConfig configures the server
type ServerConfig struct {
	Port             int
	DB               *database.DB
	VectorStore      *embeddings.VectorStore
	SemanticSearch   *embeddings.SemanticSearch
	CloudClient      *cloud.CloudClient
	ContextExtractor *pkgcontext.ContextExtractor
}

// NewServer creates a new server
func NewServer(config ServerConfig) (*Server, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database is required")
	}
	if config.SemanticSearch == nil {
		return nil, fmt.Errorf("semantic search is required")
	}
	if config.CloudClient == nil {
		return nil, fmt.Errorf("cloud client is required")
	}
	if config.ContextExtractor == nil {
		return nil, fmt.Errorf("context extractor is required")
	}

	port := config.Port
	if port == 0 {
		port = 8080
	}

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		router:           gin.Default(),
		db:               config.DB,
		vectorStore:      config.VectorStore,
		semanticSearch:   config.SemanticSearch,
		cloudClient:      config.CloudClient,
		contextExtractor: config.ContextExtractor,
		port:             port,
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

		// Recording sessions
		recordings := v1.Group("/recordings")
		{
			recordings.POST("/save", s.saveRecordingSession)
			recordings.GET("/:id", s.getRecordingSession)
			recordings.POST("/list", s.getRecordingSessions)
			recordings.DELETE("/:id", s.deleteRecordingSession)
			recordings.GET("/stats", s.getRecordingStats)
		}

		// API Testing
		apiTests := v1.Group("/api-tests")
		{
			apiTests.POST("/save", s.saveApiTest)
			apiTests.GET("/:id", s.getApiTest)
			apiTests.GET("/project/:projectId", s.getApiTestsByProject)
			apiTests.DELETE("/:id", s.deleteApiTest)
			apiTests.POST("/results/save", s.saveApiTestResult)
			apiTests.POST("/results/list", s.getApiTestResults)
			apiTests.GET("/stats", s.getApiTestStats)
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
	// Check cloud API health
	cloudHealthy := false
	ctx := c.Request.Context()
	if err := s.cloudClient.CheckHealth(ctx); err == nil {
		cloudHealthy = true
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"services": gin.H{
			"database":    s.db != nil,
			"vectorStore": s.vectorStore != nil,
			"cloudAPI":    cloudHealthy,
		},
	})
}
