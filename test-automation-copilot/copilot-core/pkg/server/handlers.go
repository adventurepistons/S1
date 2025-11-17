package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/copilot-core/pkg/cloud"
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

	// Convert elements to cloud.Element format
	elements := make([]cloud.Element, len(req.Elements))
	for i, elem := range req.Elements {
		elements[i] = cloud.Element{
			Name:         elem["name"],
			LocatorType:  elem["locatorType"],
			LocatorValue: elem["locatorValue"],
		}
	}

	// Extract context (preprocessing)
	payload, err := s.contextExtractor.ExtractPageObjectContext(ctx, req.Spec, elements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send to cloud for AI processing
	response, err := s.cloudClient.Generate(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         response.Code,
		"tokensUsed":   response.TokensUsed,
		"model":        response.Model,
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

	// Extract context (preprocessing)
	payload, err := s.contextExtractor.ExtractTestContext(ctx, req.Spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send to cloud for AI processing
	response, err := s.cloudClient.Generate(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         response.Code,
		"tokensUsed":   response.TokensUsed,
		"model":        response.Model,
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

	// Extract context (preprocessing)
	payload, err := s.contextExtractor.ExtractFixContext(ctx, req.Code, req.Error, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send to cloud for AI processing
	response, err := s.cloudClient.Generate(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fix":          response.Code,
		"tokensUsed":   response.TokensUsed,
		"model":        response.Model,
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

	// Extract context (preprocessing)
	payload, err := s.contextExtractor.ExtractChatContext(ctx, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send to cloud for AI processing
	response, err := s.cloudClient.Generate(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response":     response.Code,
		"tokensUsed":   response.TokensUsed,
		"model":        response.Model,
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

// ===== Recording Session Handlers =====

// SaveRecordingRequest represents a request to save a recording session
type SaveRecordingRequest struct {
	Session *database.RecordingSession `json:"session" binding:"required"`
}

// GetRecordingRequest represents a request to get recording sessions
type GetRecordingRequest struct {
	WorkspacePath string `json:"workspacePath"`
	Limit         int    `json:"limit"`
}

// saveRecordingSession saves a recording session
func (s *Server) saveRecordingSession(c *gin.Context) {
	var req SaveRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.SaveRecordingSession(req.Session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Recording session saved successfully",
		"sessionId":  req.Session.ID,
		"pageCount":  req.Session.PagesCaptured,
		"elementCount": req.Session.ElementsCaptured,
	})
}

// getRecordingSession retrieves a specific recording session
func (s *Server) getRecordingSession(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := s.db.GetRecordingSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session": session,
	})
}

// getRecordingSessions retrieves recording sessions for a workspace
func (s *Server) getRecordingSessions(c *gin.Context) {
	var req GetRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	sessions, err := s.db.GetRecordingSessionsByWorkspace(req.WorkspacePath, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// deleteRecordingSession deletes a recording session
func (s *Server) deleteRecordingSession(c *gin.Context) {
	sessionID := c.Param("id")

	if err := s.db.DeleteRecordingSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recording session deleted successfully",
	})
}

// getRecordingStats returns statistics about recording sessions
func (s *Server) getRecordingStats(c *gin.Context) {
	workspacePath := c.Query("workspacePath")
	if workspacePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspacePath is required"})
		return
	}

	stats, err := s.db.RecordingSessionStats(workspacePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
