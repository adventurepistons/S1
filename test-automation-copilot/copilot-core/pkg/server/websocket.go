package server

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourusername/copilot-core/pkg/cloud"
	"github.com/yourusername/copilot-core/pkg/indexer"
)

// WebSocket message types
const (
	MessageTypeRequest  = "request"
	MessageTypeResponse = "response"
	MessageTypeChunk    = "chunk"
	MessageTypeComplete = "complete"
	MessageTypeError    = "error"
	MessageTypeProgress = "progress" // For indexing progress
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSGenerateRequest represents a code generation request over WebSocket
type WSGenerateRequest struct {
	Action        string              `json:"action"` // "pageobject", "test", "chat", "fix", "index"
	Spec          string              `json:"spec"`
	Elements      []map[string]string `json:"elements,omitempty"`
	Code          string              `json:"code,omitempty"`
	Error         string              `json:"error,omitempty"`
	Message       string              `json:"message,omitempty"`
	WorkspacePath string              `json:"workspacePath,omitempty"` // For indexing
}

// WSResponse represents a WebSocket response
type WSResponse struct {
	Action       string                 `json:"action"`
	Content      string                 `json:"content,omitempty"`
	Chunk        string                 `json:"chunk,omitempty"`
	TokensUsed   int                    `json:"tokensUsed,omitempty"`
	Model        string                 `json:"model,omitempty"`
	FinishReason string                 `json:"finishReason,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Progress     *IndexingProgressInfo  `json:"progress,omitempty"` // For indexing progress
}

// IndexingProgressInfo represents indexing progress
type IndexingProgressInfo struct {
	Message         string  `json:"message"`
	FilesTotal      int     `json:"filesTotal"`
	FilesIndexed    int     `json:"filesIndexed"`
	FilesSkipped    int     `json:"filesSkipped"`
	FilesFailed     int     `json:"filesFailed"`
	CurrentFile     string  `json:"currentFile,omitempty"`
	PercentComplete float64 `json:"percentComplete"`
	JavaFiles       int     `json:"javaFiles"`
	GherkinFiles    int     `json:"gherkinFiles"`
	XMLFiles        int     `json:"xmlFiles"`
}

// handleWebSocket handles WebSocket connections for streaming
func (s *Server) handleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket client connected from %s", conn.RemoteAddr())

	// Handle messages
	for {
		// Read message
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			s.sendWSError(conn, "Invalid message format")
			continue
		}

		// Handle request
		if wsMsg.Type == MessageTypeRequest {
			s.handleWSRequest(conn, wsMsg.Payload)
		}
	}

	log.Printf("WebSocket client disconnected")
}

// handleWSRequest handles a WebSocket request
func (s *Server) handleWSRequest(conn *websocket.Conn, payload interface{}) {
	// Convert payload to request
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.sendWSError(conn, "Failed to parse request")
		return
	}

	var req WSGenerateRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		s.sendWSError(conn, "Invalid request format")
		return
	}

	ctx := context.Background()

	// Extract context payload based on action
	var contextPayload cloud.ContextPayload
	var action string

	switch req.Action {
	case "pageobject":
		action = "pageobject"
		// Convert elements to cloud.Element format
		elements := make([]cloud.Element, len(req.Elements))
		for i, elem := range req.Elements {
			elements[i] = cloud.Element{
				Name:         elem["name"],
				LocatorType:  elem["locatorType"],
				LocatorValue: elem["locatorValue"],
			}
		}
		contextPayload, err = s.contextExtractor.ExtractPageObjectContext(ctx, req.Spec, elements)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "test":
		action = "test"
		contextPayload, err = s.contextExtractor.ExtractTestContext(ctx, req.Spec)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "chat":
		action = "chat"
		contextPayload, err = s.contextExtractor.ExtractChatContext(ctx, req.Message)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "fix":
		action = "fix"
		contextPayload, err = s.contextExtractor.ExtractFixContext(ctx, req.Code, req.Error, "")
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "index":
		// Handle workspace indexing with progress updates
		s.handleWSIndexing(conn, req.WorkspacePath)
		return // Indexing handles its own completion

	default:
		s.sendWSError(conn, "Unknown action: "+req.Action)
		return
	}

	// Stream from cloud with callback
	callback := func(chunk string) error {
		return s.sendWSChunk(conn, action, chunk)
	}

	response, err := s.cloudClient.GenerateStream(ctx, contextPayload, callback)
	if err != nil {
		s.sendWSError(conn, err.Error())
		return
	}

	// Send completion message
	s.sendWSComplete(conn, action, response)
}

// sendWSChunk sends a chunk of streaming response
func (s *Server) sendWSChunk(conn *websocket.Conn, action, chunk string) error {
	msg := WSMessage{
		Type: MessageTypeChunk,
		Payload: WSResponse{
			Action: action,
			Chunk:  chunk,
		},
	}

	return conn.WriteJSON(msg)
}

// sendWSComplete sends a completion message
func (s *Server) sendWSComplete(conn *websocket.Conn, action string, response *cloud.GenerationResult) error {
	msg := WSMessage{
		Type: MessageTypeComplete,
		Payload: WSResponse{
			Action:       action,
			Content:      response.Code,
			TokensUsed:   response.TokensUsed,
			Model:        response.Model,
			FinishReason: response.FinishReason,
		},
	}

	return conn.WriteJSON(msg)
}

// sendWSError sends an error message
func (s *Server) sendWSError(conn *websocket.Conn, errMsg string) error {
	msg := WSMessage{
		Type: MessageTypeError,
		Payload: WSResponse{
			Error: errMsg,
		},
	}

	return conn.WriteJSON(msg)
}

// sendWSProgress sends indexing progress update
func (s *Server) sendWSProgress(conn *websocket.Conn, progress *IndexingProgressInfo) error {
	msg := WSMessage{
		Type: MessageTypeProgress,
		Payload: WSResponse{
			Action:   "index",
			Progress: progress,
		},
	}

	return conn.WriteJSON(msg)
}

// handleWSIndexing handles workspace indexing with real-time progress updates
func (s *Server) handleWSIndexing(conn *websocket.Conn, workspacePath string) {
	log.Printf("Starting WebSocket indexing for: %s", workspacePath)

	// Send immediate acknowledgment
	s.sendWSProgress(conn, &IndexingProgressInfo{
		Message:         "Starting indexing...",
		PercentComplete: 0,
	})

	// Create progress callback
	progressCallback := func(stats *indexer.IndexingStats) {
		var percentComplete float64
		if stats.TotalFiles > 0 {
			percentComplete = (float64(stats.IndexedFiles+stats.SkippedFiles+stats.FailedFiles) / float64(stats.TotalFiles)) * 100
		}

		// Send progress update
		progress := &IndexingProgressInfo{
			Message:         "Indexing files...",
			FilesTotal:      stats.TotalFiles,
			FilesIndexed:    stats.IndexedFiles,
			FilesSkipped:    stats.SkippedFiles,
			FilesFailed:     stats.FailedFiles,
			PercentComplete: percentComplete,
			JavaFiles:       stats.JavaFiles,
			GherkinFiles:    stats.GherkinFiles,
			XMLFiles:        stats.XMLFiles,
		}

		s.sendWSProgress(conn, progress)
	}

	// Create indexer with progress callback
	idx := indexer.NewWorkspaceIndexer(s.db, indexer.IndexerConfig{
		WorkspacePath:    workspacePath,
		ProgressCallback: progressCallback,
	})

	// Run indexing
	stats, err := idx.IndexWorkspace()
	if err != nil {
		s.sendWSError(conn, "Indexing failed: "+err.Error())
		return
	}

	// Send final completion message
	s.sendWSProgress(conn, &IndexingProgressInfo{
		Message:         "Indexing complete!",
		FilesTotal:      stats.TotalFiles,
		FilesIndexed:    stats.IndexedFiles,
		FilesSkipped:    stats.SkippedFiles,
		FilesFailed:     stats.FailedFiles,
		PercentComplete: 100,
		JavaFiles:       stats.JavaFiles,
		GherkinFiles:    stats.GherkinFiles,
		XMLFiles:        stats.XMLFiles,
	})

	log.Printf("Indexing completed: %d files indexed in %v", stats.IndexedFiles, stats.Duration)
}
