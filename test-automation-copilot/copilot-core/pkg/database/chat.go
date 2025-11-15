package database

import (
	"fmt"

	"github.com/google/uuid"
)

// ChatRepository handles chat-related database operations
type ChatRepository struct {
	db *DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// ============================================================================
// CHAT SESSION OPERATIONS
// ============================================================================

// CreateSession creates a new chat session
func (r *ChatRepository) CreateSession(workspacePath string) (*ChatSession, error) {
	session := &ChatSession{
		ID:            uuid.New().String(),
		WorkspacePath: workspacePath,
		StartedAt:     Now(),
		MessageCount:  0,
	}

	_, err := r.db.conn.Exec(`
		INSERT INTO chat_sessions (id, workspace_path, started_at, message_count)
		VALUES (?, ?, ?, ?)`,
		session.ID,
		session.WorkspacePath,
		session.StartedAt,
		session.MessageCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a chat session by ID
func (r *ChatRepository) GetSession(sessionID string) (*ChatSession, error) {
	var session ChatSession
	err := r.db.conn.Get(&session, "SELECT * FROM chat_sessions WHERE id = ?", sessionID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionsByWorkspace retrieves all sessions for a workspace
func (r *ChatRepository) GetSessionsByWorkspace(workspacePath string) ([]*ChatSession, error) {
	var sessions []*ChatSession
	err := r.db.conn.Select(&sessions,
		"SELECT * FROM chat_sessions WHERE workspace_path = ? ORDER BY started_at DESC",
		workspacePath)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetActiveSession retrieves the most recent active session for a workspace
func (r *ChatRepository) GetActiveSession(workspacePath string) (*ChatSession, error) {
	var session ChatSession
	err := r.db.conn.Get(&session, `
		SELECT * FROM chat_sessions
		WHERE workspace_path = ? AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, workspacePath)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// EndSession marks a session as ended
func (r *ChatRepository) EndSession(sessionID string) error {
	_, err := r.db.conn.Exec(`
		UPDATE chat_sessions
		SET ended_at = ?
		WHERE id = ?`,
		Now(),
		sessionID,
	)
	return err
}

// DeleteSession deletes a chat session and all its messages
func (r *ChatRepository) DeleteSession(sessionID string) error {
	_, err := r.db.conn.Exec("DELETE FROM chat_sessions WHERE id = ?", sessionID)
	return err
}

// ============================================================================
// CHAT MESSAGE OPERATIONS
// ============================================================================

// AddMessage adds a message to a chat session
func (r *ChatRepository) AddMessage(sessionID, role, message string, contextFiles []string) (*ChatMessage, error) {
	// Convert context files to JSON
	var contextJSON *string
	if len(contextFiles) > 0 {
		// Simple JSON array encoding
		json := "["
		for i, file := range contextFiles {
			if i > 0 {
				json += ","
			}
			json += fmt.Sprintf(`"%s"`, file)
		}
		json += "]"
		contextJSON = &json
	}

	msg := &ChatMessage{
		SessionID:    sessionID,
		Role:         role,
		Message:      message,
		Timestamp:    Now(),
		ContextFiles: contextJSON,
	}

	result, err := r.db.conn.Exec(`
		INSERT INTO chat_history (session_id, role, message, timestamp, context_files)
		VALUES (?, ?, ?, ?, ?)`,
		msg.SessionID,
		msg.Role,
		msg.Message,
		msg.Timestamp,
		msg.ContextFiles,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	msg.ID = id

	// Update message count
	_, err = r.db.conn.Exec(`
		UPDATE chat_sessions
		SET message_count = message_count + 1
		WHERE id = ?`,
		sessionID,
	)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

// GetMessages retrieves all messages for a session
func (r *ChatRepository) GetMessages(sessionID string) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	err := r.db.conn.Select(&messages,
		"SELECT * FROM chat_history WHERE session_id = ? ORDER BY timestamp ASC",
		sessionID)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetRecentMessages retrieves the N most recent messages for a session
func (r *ChatRepository) GetRecentMessages(sessionID string, limit int) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	err := r.db.conn.Select(&messages, `
		SELECT * FROM chat_history
		WHERE session_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - i - 1
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessageCount returns the number of messages in a session
func (r *ChatRepository) GetMessageCount(sessionID string) (int, error) {
	var count int
	err := r.db.conn.Get(&count,
		"SELECT COUNT(*) FROM chat_history WHERE session_id = ?",
		sessionID)
	return count, err
}

// DeleteMessage deletes a specific message
func (r *ChatRepository) DeleteMessage(messageID int64) error {
	// Get session ID first to update count
	var sessionID string
	err := r.db.conn.Get(&sessionID,
		"SELECT session_id FROM chat_history WHERE id = ?",
		messageID)
	if err != nil {
		return err
	}

	// Delete message
	_, err = r.db.conn.Exec("DELETE FROM chat_history WHERE id = ?", messageID)
	if err != nil {
		return err
	}

	// Update message count
	_, err = r.db.conn.Exec(`
		UPDATE chat_sessions
		SET message_count = message_count - 1
		WHERE id = ?`,
		sessionID,
	)

	return err
}

// ============================================================================
// FILE CHANGE OPERATIONS
// ============================================================================

// AddFileChange records a code modification
func (r *ChatRepository) AddFileChange(change *FileChange) error {
	result, err := r.db.conn.Exec(`
		INSERT INTO file_changes (
			session_id, file_path, change_type, before_content,
			after_content, diff, applied, timestamp, user_approved
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		change.SessionID,
		change.FilePath,
		change.ChangeType,
		change.BeforeContent,
		change.AfterContent,
		change.Diff,
		change.Applied,
		change.Timestamp,
		change.UserApproved,
	)

	if err != nil {
		return fmt.Errorf("failed to add file change: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	change.ID = id
	return nil
}

// GetFileChanges retrieves all file changes for a session
func (r *ChatRepository) GetFileChanges(sessionID string) ([]*FileChange, error) {
	var changes []*FileChange
	err := r.db.conn.Select(&changes,
		"SELECT * FROM file_changes WHERE session_id = ? ORDER BY timestamp ASC",
		sessionID)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// GetFileChangesByFile retrieves all changes for a specific file in a session
func (r *ChatRepository) GetFileChangesByFile(sessionID, filePath string) ([]*FileChange, error) {
	var changes []*FileChange
	err := r.db.conn.Select(&changes, `
		SELECT * FROM file_changes
		WHERE session_id = ? AND file_path = ?
		ORDER BY timestamp ASC
	`, sessionID, filePath)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// GetUnappliedChanges retrieves all unapplied changes for a session
func (r *ChatRepository) GetUnappliedChanges(sessionID string) ([]*FileChange, error) {
	var changes []*FileChange
	err := r.db.conn.Select(&changes, `
		SELECT * FROM file_changes
		WHERE session_id = ? AND applied = 0
		ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// MarkChangeApplied marks a file change as applied
func (r *ChatRepository) MarkChangeApplied(changeID int64, applied bool) error {
	_, err := r.db.conn.Exec(`
		UPDATE file_changes
		SET applied = ?
		WHERE id = ?`,
		applied,
		changeID,
	)
	return err
}

// MarkChangeApproved marks a file change as user-approved
func (r *ChatRepository) MarkChangeApproved(changeID int64, approved bool) error {
	_, err := r.db.conn.Exec(`
		UPDATE file_changes
		SET user_approved = ?
		WHERE id = ?`,
		approved,
		changeID,
	)
	return err
}

// GetChangeHistory retrieves change history for a file
func (r *ChatRepository) GetChangeHistory(filePath string) ([]*FileChange, error) {
	var changes []*FileChange
	err := r.db.conn.Select(&changes, `
		SELECT * FROM file_changes
		WHERE file_path = ?
		ORDER BY timestamp DESC
	`, filePath)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// CountChanges returns total number of file changes
func (r *ChatRepository) CountChanges() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM file_changes")
	return count, err
}

// CountAppliedChanges returns number of applied changes
func (r *ChatRepository) CountAppliedChanges() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM file_changes WHERE applied = 1")
	return count, err
}
