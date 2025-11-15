package server

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourusername/copilot-core/pkg/llm"
)

// WebSocket message types
const (
	MessageTypeRequest  = "request"
	MessageTypeResponse = "response"
	MessageTypeChunk    = "chunk"
	MessageTypeComplete = "complete"
	MessageTypeError    = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSGenerateRequest represents a code generation request over WebSocket
type WSGenerateRequest struct {
	Action   string              `json:"action"` // "pageobject", "test", "chat", "fix"
	Spec     string              `json:"spec"`
	Elements []map[string]string `json:"elements,omitempty"`
	Code     string              `json:"code,omitempty"`
	Error    string              `json:"error,omitempty"`
	Message  string              `json:"message,omitempty"`
}

// WSResponse represents a WebSocket response
type WSResponse struct {
	Action       string `json:"action"`
	Content      string `json:"content,omitempty"`
	Chunk        string `json:"chunk,omitempty"`
	TokensUsed   int    `json:"tokensUsed,omitempty"`
	Model        string `json:"model,omitempty"`
	FinishReason string `json:"finishReason,omitempty"`
	Error        string `json:"error,omitempty"`
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

	// Build completion request based on action
	var completionReq llm.CompletionRequest
	var action string

	switch req.Action {
	case "pageobject":
		action = "pageobject"
		completionReq, err = s.contextBuilder.BuildPageObjectPrompt(ctx, req.Spec, req.Elements)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "test":
		action = "test"
		completionReq, err = s.contextBuilder.BuildTestCasePrompt(ctx, req.Spec)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "chat":
		action = "chat"
		completionReq, err = s.contextBuilder.BuildChatPrompt(ctx, req.Message)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	case "fix":
		action = "fix"
		completionReq, err = s.contextBuilder.BuildFixPrompt(ctx, req.Code, req.Error)
		if err != nil {
			s.sendWSError(conn, err.Error())
			return
		}

	default:
		s.sendWSError(conn, "Unknown action: "+req.Action)
		return
	}

	// Stream completion with callback
	callback := func(chunk string) error {
		return s.sendWSChunk(conn, action, chunk)
	}

	response, err := s.llmClient.CompleteStream(ctx, completionReq, callback)
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
func (s *Server) sendWSComplete(conn *websocket.Conn, action string, response *llm.CompletionResponse) error {
	msg := WSMessage{
		Type: MessageTypeComplete,
		Payload: WSResponse{
			Action:       action,
			Content:      response.Content,
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
