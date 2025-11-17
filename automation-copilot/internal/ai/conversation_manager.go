package ai

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ConversationManager handles multi-turn conversations
// Remembers last 5 turns for context-aware code generation
type ConversationManager struct {
	db        *sql.DB
	sessionID string
	maxTurns  int
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(db *sql.DB, config Config) *ConversationManager {
	sessionID := uuid.New().String()

	return &ConversationManager{
		db:        db,
		sessionID: sessionID,
		maxTurns:  config.MaxConversationTurns,
	}
}

// LoadSession loads an existing conversation session
func LoadConversationSession(db *sql.DB, sessionID string, config Config) (*ConversationManager, error) {
	// Check if session exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM conversation_sessions WHERE session_id = ?)
	`, sessionID).Scan(&exists)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Update last active time
	_, err = db.Exec(`
		UPDATE conversation_sessions
		SET last_active = ?
		WHERE session_id = ?
	`, time.Now(), sessionID)

	return &ConversationManager{
		db:        db,
		sessionID: sessionID,
		maxTurns:  config.MaxConversationTurns,
	}, nil
}

// StartSession creates a new session in the database
func (cm *ConversationManager) StartSession(projectRoot string) error {
	_, err := cm.db.Exec(`
		INSERT INTO conversation_sessions
		(session_id, project_root, started_at, last_active, message_count, status)
		VALUES (?, ?, ?, ?, 0, 'active')
	`, cm.sessionID, projectRoot, time.Now(), time.Now())

	return err
}

// AddUserMessage adds a user message to the conversation
func (cm *ConversationManager) AddUserMessage(content string) error {
	_, err := cm.db.Exec(`
		INSERT INTO conversation_messages
		(session_id, role, content, timestamp)
		VALUES (?, 'user', ?, ?)
	`, cm.sessionID, content, time.Now())

	if err != nil {
		return err
	}

	// Update session message count
	_, err = cm.db.Exec(`
		UPDATE conversation_sessions
		SET message_count = message_count + 1,
		    last_active = ?
		WHERE session_id = ?
	`, time.Now(), cm.sessionID)

	return err
}

// AddAssistantMessage adds an assistant message to the conversation
func (cm *ConversationManager) AddAssistantMessage(content string, generatedCode *GeneratedCode) error {
	// Serialize generated code to JSON if provided
	var codeJSON sql.NullString
	var metadataJSON sql.NullString

	if generatedCode != nil {
		codeBytes, err := json.Marshal(generatedCode)
		if err == nil {
			codeJSON = sql.NullString{String: string(codeBytes), Valid: true}
		}

		metadata := map[string]interface{}{
			"file_path":  generatedCode.FilePath,
			"class_name": generatedCode.ClassName,
			"language":   generatedCode.Language,
		}

		metadataBytes, err := json.Marshal(metadata)
		if err == nil {
			metadataJSON = sql.NullString{String: string(metadataBytes), Valid: true}
		}
	}

	_, err := cm.db.Exec(`
		INSERT INTO conversation_messages
		(session_id, role, content, code, metadata, timestamp)
		VALUES (?, 'assistant', ?, ?, ?, ?)
	`, cm.sessionID, content, codeJSON, metadataJSON, time.Now())

	if err != nil {
		return err
	}

	// Update session message count
	_, err = cm.db.Exec(`
		UPDATE conversation_sessions
		SET message_count = message_count + 1,
		    last_active = ?
		WHERE session_id = ?
	`, time.Now(), cm.sessionID)

	return err
}

// GetHistory retrieves the conversation history
func (cm *ConversationManager) GetHistory() (*ConversationHistory, error) {
	// Get messages, limiting to last N turns (2 messages per turn: user + assistant)
	rows, err := cm.db.Query(`
		SELECT role, content, code, timestamp
		FROM conversation_messages
		WHERE session_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, cm.sessionID, cm.maxTurns*2)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []ConversationMessage{}

	for rows.Next() {
		var role, content string
		var codeJSON sql.NullString
		var timestamp time.Time

		if err := rows.Scan(&role, &content, &codeJSON, &timestamp); err != nil {
			continue
		}

		msg := ConversationMessage{
			Role:      role,
			Content:   content,
			Timestamp: timestamp,
		}

		// Deserialize code if present
		if codeJSON.Valid && codeJSON.String != "" {
			var code GeneratedCode
			if err := json.Unmarshal([]byte(codeJSON.String), &code); err == nil {
				msg.Code = &code
			}
		}

		messages = append(messages, msg)
	}

	// Reverse to get chronological order (oldest first)
	reverseMessages(messages)

	return &ConversationHistory{
		SessionID: cm.sessionID,
		Messages:  messages,
		MaxTurns:  cm.maxTurns,
	}, nil
}

// BuildContextString builds a string representation of conversation history for prompting
func (cm *ConversationManager) BuildContextString() (string, error) {
	history, err := cm.GetHistory()
	if err != nil {
		return "", err
	}

	if len(history.Messages) == 0 {
		return "", nil
	}

	return history.BuildContext(), nil
}

// GetSessionID returns the current session ID
func (cm *ConversationManager) GetSessionID() string {
	return cm.sessionID
}

// GetMessageCount returns the number of messages in this session
func (cm *ConversationManager) GetMessageCount() (int, error) {
	var count int
	err := cm.db.QueryRow(`
		SELECT message_count
		FROM conversation_sessions
		WHERE session_id = ?
	`, cm.sessionID).Scan(&count)

	return count, err
}

// EndSession marks the session as ended
func (cm *ConversationManager) EndSession() error {
	_, err := cm.db.Exec(`
		UPDATE conversation_sessions
		SET status = 'ended',
		    last_active = ?
		WHERE session_id = ?
	`, time.Now(), cm.sessionID)

	return err
}

// ClearHistory removes all messages from this session
func (cm *ConversationManager) ClearHistory() error {
	_, err := cm.db.Exec(`
		DELETE FROM conversation_messages
		WHERE session_id = ?
	`, cm.sessionID)

	if err != nil {
		return err
	}

	// Reset message count
	_, err = cm.db.Exec(`
		UPDATE conversation_sessions
		SET message_count = 0
		WHERE session_id = ?
	`, cm.sessionID)

	return err
}

// BuildContext builds a context string from conversation history
func (ch *ConversationHistory) BuildContext() string {
	if len(ch.Messages) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("## Previous Conversation\n\n")

	for i, msg := range ch.Messages {
		if msg.Role == "user" {
			sb.WriteString(fmt.Sprintf("**User** (Turn %d): %s\n\n", (i/2)+1, msg.Content))
		} else {
			sb.WriteString(fmt.Sprintf("**Assistant** (Turn %d): %s\n\n", (i/2)+1, msg.Content))

			if msg.Code != nil && msg.Code.Code != "" {
				// Include generated code in context
				sb.WriteString("```" + msg.Code.Language + "\n")
				sb.WriteString(msg.Code.Code)
				sb.WriteString("\n```\n\n")
			}
		}
	}

	return sb.String()
}

// AddMessage adds a message to the history (in-memory)
func (ch *ConversationHistory) AddMessage(msg ConversationMessage) {
	ch.Messages = append(ch.Messages, msg)

	// Trim to max turns
	if len(ch.Messages) > ch.MaxTurns*2 {
		ch.Messages = ch.Messages[len(ch.Messages)-(ch.MaxTurns*2):]
	}
}

// GetLastUserMessage returns the most recent user message
func (ch *ConversationHistory) GetLastUserMessage() *ConversationMessage {
	for i := len(ch.Messages) - 1; i >= 0; i-- {
		if ch.Messages[i].Role == "user" {
			return &ch.Messages[i]
		}
	}
	return nil
}

// GetLastAssistantMessage returns the most recent assistant message
func (ch *ConversationHistory) GetLastAssistantMessage() *ConversationMessage {
	for i := len(ch.Messages) - 1; i >= 0; i-- {
		if ch.Messages[i].Role == "assistant" {
			return &ch.Messages[i]
		}
	}
	return nil
}

// reverseMessages reverses a slice of messages in place
func reverseMessages(messages []ConversationMessage) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}

// ListActiveSessions returns all active conversation sessions
func ListActiveSessions(db *sql.DB) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT session_id, project_root, started_at, last_active, message_count
		FROM conversation_sessions
		WHERE status = 'active'
		ORDER BY last_active DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []map[string]interface{}{}

	for rows.Next() {
		var sessionID, projectRoot string
		var startedAt, lastActive time.Time
		var messageCount int

		if err := rows.Scan(&sessionID, &projectRoot, &startedAt, &lastActive, &messageCount); err != nil {
			continue
		}

		sessions = append(sessions, map[string]interface{}{
			"session_id":    sessionID,
			"project_root":  projectRoot,
			"started_at":    startedAt,
			"last_active":   lastActive,
			"message_count": messageCount,
		})
	}

	return sessions, nil
}
