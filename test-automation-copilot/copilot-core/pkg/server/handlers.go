package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/indexer"
)

// IndexWorkspaceRequest represents a workspace indexing request
type IndexWorkspaceRequest struct {
	WorkspacePath string `json:"workspacePath" binding:"required"`
}

// SearchRequest represents a code search request
type SearchRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

// GeneratePageObjectRequest represents a page object generation request
type GeneratePageObjectRequest struct {
	Spec     string              `json:"spec" binding:"required"`
	Elements []map[string]string `json:"elements"`
}

// GenerateTestRequest represents a test generation request
type GenerateTestRequest struct {
	Spec string `json:"spec" binding:"required"`
}

// FixCodeRequest represents a code fix request
type FixCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	Error string `json:"error" binding:"required"`
}

// ChatMessageRequest represents a chat message request
type ChatMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

// indexWorkspace handles workspace indexing
func (s *Server) indexWorkspace(c *gin.Context) {
	var req IndexWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start indexing in background
	go func() {
		// Create a new indexer for this workspace
		idx := indexer.NewWorkspaceIndexer(s.db, indexer.IndexerConfig{
			WorkspacePath: req.WorkspacePath,
		})

		// Index the workspace
		stats, err := idx.IndexWorkspace()
		if err != nil {
			log.Printf("Error indexing workspace: %v", err)
			return
		}

		log.Printf("Indexing complete: %d files indexed", stats.IndexedFiles)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Indexing started",
		"path":    req.WorkspacePath,
	})
}

// getWorkspaceStats returns workspace statistics
func (s *Server) getWorkspaceStats(c *gin.Context) {
	// Get database stats
	classRepo := database.NewClassRepository(s.db)
	methodRepo := database.NewMethodRepository(s.db)
	fileRepo := database.NewFileRepository(s.db)

	allClasses, _ := classRepo.GetAll()
	testMethods, _ := methodRepo.GetTestMethods()
	allFiles, _ := fileRepo.GetAll()

	// Count page objects (heuristic: classes ending with "Page")
	pageObjectCount := 0
	for _, class := range allClasses {
		if len(class.Name) > 4 && class.Name[len(class.Name)-4:] == "Page" {
			pageObjectCount++
		}
	}

	// Get vector store stats
	vectorStats := s.vectorStore.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"files":        len(allFiles),
		"classes":      len(allClasses),
		"pageObjects":  pageObjectCount,
		"testMethods":  len(testMethods),
		"vectorDocs":   vectorStats.DocumentCount,
		"collectionName": vectorStats.CollectionName,
	})
}

// searchCode handles semantic code search
func (s *Server) searchCode(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	ctx := c.Request.Context()
	results, err := s.semanticSearch.SearchCode(ctx, req.Query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"query":   req.Query,
		"count":   len(results),
	})
}

// searchPageObjects handles page object search
func (s *Server) searchPageObjects(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 5
	}

	ctx := c.Request.Context()
	results, err := s.semanticSearch.SearchPageObjects(ctx, req.Query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"query":   req.Query,
		"count":   len(results),
	})
}

// searchTests handles test method search
func (s *Server) searchTests(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 5
	}

	ctx := c.Request.Context()
	results, err := s.semanticSearch.SearchTests(ctx, req.Query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"query":   req.Query,
		"count":   len(results),
	})
}

// generatePageObject handles page object generation
func (s *Server) generatePageObject(c *gin.Context) {
	var req GeneratePageObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Build prompt with context
	completionReq, err := s.contextBuilder.BuildPageObjectPrompt(ctx, req.Spec, req.Elements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate code
	response, err := s.llmClient.Complete(ctx, completionReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":        response.Content,
		"tokensUsed":  response.TokensUsed,
		"model":       response.Model,
		"finishReason": response.FinishReason,
	})
}

// generateTest handles test case generation
func (s *Server) generateTest(c *gin.Context) {
	var req GenerateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Build prompt with context
	completionReq, err := s.contextBuilder.BuildTestCasePrompt(ctx, req.Spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate code
	response, err := s.llmClient.Complete(ctx, completionReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":        response.Content,
		"tokensUsed":  response.TokensUsed,
		"model":       response.Model,
		"finishReason": response.FinishReason,
	})
}

// fixCode handles code fixing
func (s *Server) fixCode(c *gin.Context) {
	var req FixCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Build prompt with context
	completionReq, err := s.contextBuilder.BuildFixPrompt(ctx, req.Code, req.Error)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate fix
	response, err := s.llmClient.Complete(ctx, completionReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fix":         response.Content,
		"tokensUsed":  response.TokensUsed,
		"model":       response.Model,
		"finishReason": response.FinishReason,
	})
}

// chatMessage handles chat messages
func (s *Server) chatMessage(c *gin.Context) {
	var req ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Build prompt with context
	completionReq, err := s.contextBuilder.BuildChatPrompt(ctx, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response
	response, err := s.llmClient.Complete(ctx, completionReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response":    response.Content,
		"tokensUsed":  response.TokensUsed,
		"model":       response.Model,
		"finishReason": response.FinishReason,
	})
}

// getClasses returns all classes
func (s *Server) getClasses(c *gin.Context) {
	classRepo := database.NewClassRepository(s.db)

	classes, err := classRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"classes": classes,
		"count":   len(classes),
	})
}

// getClass returns a specific class
func (s *Server) getClass(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class ID"})
		return
	}

	classRepo := database.NewClassRepository(s.db)

	class, err := classRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"class": class,
	})
}

// getMethodsByClass returns methods for a class
func (s *Server) getMethodsByClass(c *gin.Context) {
	idStr := c.Param("classId")
	classID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class ID"})
		return
	}

	methodRepo := database.NewMethodRepository(s.db)

	methods, err := methodRepo.GetByClassID(classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"methods": methods,
		"count":   len(methods),
	})
}
