package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestNewDB(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(DBConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(DBConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify tables exist
	tables := []string{
		"files", "classes", "methods", "fields", "dependencies",
		"feature_files", "scenarios", "steps", "step_definitions",
		"chat_sessions", "chat_history", "file_changes", "config",
		"embeddings", "recording_sessions", "generated_code",
	}

	for _, table := range tables {
		var count int
		err := db.conn.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s does not exist", table)
		}
	}
}

func TestFileRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(db)

	// Create file
	file := &File{
		Path:         "/test/LoginPage.java",
		Type:         "java",
		ContentHash:  ComputeHash("test content"),
		Content:      "test content",
		PackageName:  StringPtr("com.example.pages"),
		LastModified: Now(),
		IndexedAt:    Now(),
		SizeBytes:    100,
		LineCount:    10,
	}

	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	if file.ID == 0 {
		t.Error("File ID was not set")
	}

	// Get file by path
	retrieved, err := repo.GetByPath(file.Path)
	if err != nil {
		t.Fatalf("Failed to get file by path: %v", err)
	}

	if retrieved.Path != file.Path {
		t.Errorf("Expected path %s, got %s", file.Path, retrieved.Path)
	}

	// Update file
	file.LineCount = 20
	err = repo.Update(file)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	updated, err := repo.GetByID(file.ID)
	if err != nil {
		t.Fatalf("Failed to get updated file: %v", err)
	}

	if updated.LineCount != 20 {
		t.Errorf("Expected line count 20, got %d", updated.LineCount)
	}

	// Count files
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Failed to count files: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 file, got %d", count)
	}

	// Delete file
	err = repo.Delete(file.ID)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	_, err = repo.GetByID(file.ID)
	if err == nil {
		t.Error("Expected error when getting deleted file")
	}
}

func TestClassRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fileRepo := NewFileRepository(db)
	classRepo := NewClassRepository(db)

	// Create file first
	file := &File{
		Path:         "/test/LoginPage.java",
		Type:         "java",
		ContentHash:  ComputeHash("test"),
		Content:      "test",
		PackageName:  StringPtr("com.example.pages"),
		LastModified: Now(),
		IndexedAt:    Now(),
		SizeBytes:    100,
		LineCount:    10,
	}
	fileRepo.Create(file)

	// Create class
	class := &Class{
		FileID:             file.ID,
		Name:               "LoginPage",
		FullyQualifiedName: "com.example.pages.LoginPage",
		Type:               "class",
		PackageName:        StringPtr("com.example.pages"),
		IsPublic:           true,
		StartLine:          1,
		EndLine:            50,
	}

	err := classRepo.Create(class)
	if err != nil {
		t.Fatalf("Failed to create class: %v", err)
	}

	if class.ID == 0 {
		t.Error("Class ID was not set")
	}

	// Get class by FQN
	retrieved, err := classRepo.GetByFQN(class.FullyQualifiedName)
	if err != nil {
		t.Fatalf("Failed to get class by FQN: %v", err)
	}

	if retrieved.Name != class.Name {
		t.Errorf("Expected name %s, got %s", class.Name, retrieved.Name)
	}

	// Get classes by file
	classes, err := classRepo.GetByFileID(file.ID)
	if err != nil {
		t.Fatalf("Failed to get classes by file: %v", err)
	}

	if len(classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(classes))
	}
}

func TestChatRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewChatRepository(db)

	// Create session
	session, err := repo.CreateSession("/test/workspace")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID was not set")
	}

	// Add messages
	msg1, err := repo.AddMessage(session.ID, "user", "Hello", nil)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	msg2, err := repo.AddMessage(session.ID, "assistant", "Hi there!", []string{"/test/file.java"})
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Get messages
	messages, err := repo.GetMessages(session.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].ID != msg1.ID || messages[1].ID != msg2.ID {
		t.Error("Messages not in correct order")
	}

	// Get message count
	count, err := repo.GetMessageCount(session.ID)
	if err != nil {
		t.Fatalf("Failed to get message count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 messages, got %d", count)
	}

	// End session
	err = repo.EndSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to end session: %v", err)
	}

	retrieved, err := repo.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.EndedAt == nil {
		t.Error("Session was not ended")
	}
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.FileCount != 0 {
		t.Errorf("Expected 0 files, got %d", stats.FileCount)
	}

	// Add some data
	fileRepo := NewFileRepository(db)
	file := &File{
		Path:         "/test/Test.java",
		Type:         "java",
		ContentHash:  ComputeHash("test"),
		Content:      "test",
		LastModified: Now(),
		IndexedAt:    Now(),
		SizeBytes:    100,
		LineCount:    10,
	}
	fileRepo.Create(file)

	stats, err = db.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.FileCount != 1 {
		t.Errorf("Expected 1 file, got %d", stats.FileCount)
	}
}

func TestTransactions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fileRepo := NewFileRepository(db)

	// Test successful transaction
	err := db.WithTransaction(func(tx *sqlx.Tx) error {
		file := &File{
			Path:         "/test/File1.java",
			Type:         "java",
			ContentHash:  ComputeHash("test"),
			Content:      "test",
			LastModified: Now(),
			IndexedAt:    Now(),
			SizeBytes:    100,
			LineCount:    10,
		}
		return fileRepo.UpsertTx(tx, file)
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Verify file was created
	count, _ := fileRepo.Count()
	if count != 1 {
		t.Errorf("Expected 1 file after successful transaction, got %d", count)
	}

	// Test failed transaction (should rollback)
	err = db.WithTransaction(func(tx *sqlx.Tx) error {
		file := &File{
			Path:         "/test/File2.java",
			Type:         "java",
			ContentHash:  ComputeHash("test"),
			Content:      "test",
			LastModified: Now(),
			IndexedAt:    Now(),
			SizeBytes:    100,
			LineCount:    10,
		}
		fileRepo.UpsertTx(tx, file)
		return &testError{msg: "intentional error"}
	})

	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// Verify rollback
	count, _ = fileRepo.Count()
	if count != 1 {
		t.Errorf("Expected 1 file after rollback, got %d", count)
	}
}

// Helper functions

func setupTestDB(t *testing.T) *DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(DBConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return db
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
