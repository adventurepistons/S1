package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// RecordingSession represents a browser recording session
type RecordingSession struct {
	ID               string    `json:"id"`
	StartTime        int64     `json:"startTime"`
	EndTime          *int64    `json:"endTime,omitempty"`
	PagesCaptured    int       `json:"pagesCaptured"`
	ElementsCaptured int       `json:"elementsCaptured"`
	WorkspacePath    string    `json:"workspacePath"`
	RecordingData    string    `json:"recordingData"` // Full JSON data
	CreatedAt        time.Time `json:"createdAt"`
}

// SaveRecordingSession saves a recording session to the database
func (db *DB) SaveRecordingSession(session *RecordingSession) error {
	query := `
		INSERT INTO recording_sessions (
			id, start_time, end_time, pages_captured, elements_captured,
			workspace_path, recording_data, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			end_time = excluded.end_time,
			pages_captured = excluded.pages_captured,
			elements_captured = excluded.elements_captured,
			recording_data = excluded.recording_data
	`

	createdAt := session.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := db.conn.Exec(query,
		session.ID,
		session.StartTime,
		session.EndTime,
		session.PagesCaptured,
		session.ElementsCaptured,
		session.WorkspacePath,
		session.RecordingData,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save recording session: %w", err)
	}

	return nil
}

// GetRecordingSession retrieves a recording session by ID
func (db *DB) GetRecordingSession(id string) (*RecordingSession, error) {
	query := `
		SELECT id, start_time, end_time, pages_captured, elements_captured,
		       workspace_path, recording_data, created_at
		FROM recording_sessions
		WHERE id = ?
	`

	session := &RecordingSession{}
	err := db.conn.QueryRow(query, id).Scan(
		&session.ID,
		&session.StartTime,
		&session.EndTime,
		&session.PagesCaptured,
		&session.ElementsCaptured,
		&session.WorkspacePath,
		&session.RecordingData,
		&session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recording session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recording session: %w", err)
	}

	return session, nil
}

// GetRecordingSessionsByWorkspace retrieves all recording sessions for a workspace
func (db *DB) GetRecordingSessionsByWorkspace(workspacePath string, limit int) ([]*RecordingSession, error) {
	if limit == 0 {
		limit = 10
	}

	query := `
		SELECT id, start_time, end_time, pages_captured, elements_captured,
		       workspace_path, recording_data, created_at
		FROM recording_sessions
		WHERE workspace_path = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, workspacePath, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recording sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*RecordingSession
	for rows.Next() {
		session := &RecordingSession{}
		err := rows.Scan(
			&session.ID,
			&session.StartTime,
			&session.EndTime,
			&session.PagesCaptured,
			&session.ElementsCaptured,
			&session.WorkspacePath,
			&session.RecordingData,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recording session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteRecordingSession deletes a recording session
func (db *DB) DeleteRecordingSession(id string) error {
	query := `DELETE FROM recording_sessions WHERE id = ?`

	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recording session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("recording session not found: %s", id)
	}

	return nil
}

// SaveGeneratedCodeForRecording links generated code to a recording session
func (db *DB) SaveGeneratedCodeForRecording(recordingID, fileName, filePath, content, codeType string) error {
	query := `
		INSERT INTO generated_code (
			recording_session_id, file_name, file_path, content, type, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query, recordingID, fileName, filePath, content, codeType, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save generated code: %w", err)
	}

	return nil
}

// GetGeneratedCodeForRecording retrieves generated code for a recording session
func (db *DB) GetGeneratedCodeForRecording(recordingID string) ([]map[string]interface{}, error) {
	query := `
		SELECT id, file_name, file_path, content, type, created_at
		FROM generated_code
		WHERE recording_session_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query generated code: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var fileName, filePath, content, codeType string
		var createdAt time.Time

		err := rows.Scan(&id, &fileName, &filePath, &content, &codeType, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generated code: %w", err)
		}

		results = append(results, map[string]interface{}{
			"id":        id,
			"fileName":  fileName,
			"filePath":  filePath,
			"content":   content,
			"type":      codeType,
			"createdAt": createdAt,
		})
	}

	return results, nil
}

// RecordingSessionStats returns statistics about recording sessions
func (db *DB) RecordingSessionStats(workspacePath string) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_sessions,
			SUM(pages_captured) as total_pages,
			SUM(elements_captured) as total_elements,
			MAX(created_at) as last_recording
		FROM recording_sessions
		WHERE workspace_path = ?
	`

	var totalSessions, totalPages, totalElements int
	var lastRecording sql.NullTime

	err := db.conn.QueryRow(query, workspacePath).Scan(
		&totalSessions,
		&totalPages,
		&totalElements,
		&lastRecording,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get recording stats: %w", err)
	}

	stats := map[string]interface{}{
		"totalSessions":  totalSessions,
		"totalPages":     totalPages,
		"totalElements":  totalElements,
		"lastRecording":  nil,
	}

	if lastRecording.Valid {
		stats["lastRecording"] = lastRecording.Time
	}

	return stats, nil
}

// ParseRecordingData parses the JSON recording data
func ParseRecordingData(recordingData string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(recordingData), &data); err != nil {
		return nil, fmt.Errorf("failed to parse recording data: %w", err)
	}
	return data, nil
}
