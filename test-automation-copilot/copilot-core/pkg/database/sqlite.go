package database

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite database connection
type DB struct {
	conn *sqlx.DB
	path string
}

// Config for database initialization
type DBConfig struct {
	Path string // Path to SQLite database file
}

// NewDB creates a new database connection
func NewDB(config DBConfig) (*DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	conn, err := sqlx.Connect("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set pragmas for performance and reliability
	pragmas := []string{
		"PRAGMA journal_mode=WAL",           // Write-Ahead Logging for better concurrency
		"PRAGMA synchronous=NORMAL",         // Balance between safety and speed
		"PRAGMA foreign_keys=ON",            // Enable foreign key constraints
		"PRAGMA cache_size=10000",           // 10MB cache
		"PRAGMA temp_store=MEMORY",          // Store temp tables in memory
		"PRAGMA mmap_size=30000000000",      // Memory-mapped I/O
		"PRAGMA page_size=4096",             // 4KB pages
		"PRAGMA busy_timeout=5000",          // Wait 5s if database is locked
	}

	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	db := &DB{
		conn: conn,
		path: config.Path,
	}

	return db, nil
}

// Initialize creates all tables and indexes
func (db *DB) Initialize() error {
	// Execute schema
	if _, err := db.conn.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConn returns the underlying sqlx connection
func (db *DB) GetConn() *sqlx.DB {
	return db.conn
}

// GetPath returns the database file path
func (db *DB) GetPath() string {
	return db.path
}

// Begin starts a new transaction
func (db *DB) Begin() (*sqlx.Tx, error) {
	return db.conn.Beginx()
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// Stats returns database statistics
type DBStats struct {
	FileCount       int64 `db:"file_count"`
	ClassCount      int64 `db:"class_count"`
	MethodCount     int64 `db:"method_count"`
	FieldCount      int64 `db:"field_count"`
	DependencyCount int64 `db:"dependency_count"`
	FeatureCount    int64 `db:"feature_count"`
	ScenarioCount   int64 `db:"scenario_count"`
	ChatSessionCount int64 `db:"chat_session_count"`
	ChatMessageCount int64 `db:"chat_message_count"`
}

// GetStats returns database statistics
func (db *DB) GetStats() (*DBStats, error) {
	stats := &DBStats{}

	// Get counts from each table
	queries := map[string]*int64{
		"SELECT COUNT(*) FROM files":           &stats.FileCount,
		"SELECT COUNT(*) FROM classes":         &stats.ClassCount,
		"SELECT COUNT(*) FROM methods":         &stats.MethodCount,
		"SELECT COUNT(*) FROM fields":          &stats.FieldCount,
		"SELECT COUNT(*) FROM dependencies":    &stats.DependencyCount,
		"SELECT COUNT(*) FROM feature_files":   &stats.FeatureCount,
		"SELECT COUNT(*) FROM scenarios":       &stats.ScenarioCount,
		"SELECT COUNT(*) FROM chat_sessions":   &stats.ChatSessionCount,
		"SELECT COUNT(*) FROM chat_history":    &stats.ChatMessageCount,
	}

	for query, target := range queries {
		if err := db.conn.Get(target, query); err != nil {
			return nil, fmt.Errorf("failed to get stats: %w", err)
		}
	}

	return stats, nil
}

// Vacuum optimizes the database
func (db *DB) Vacuum() error {
	_, err := db.conn.Exec("VACUUM")
	return err
}

// ClearAll deletes all data from all tables (useful for testing)
func (db *DB) ClearAll() error {
	tables := []string{
		"generated_code",
		"recording_sessions",
		"embeddings",
		"config",
		"file_changes",
		"chat_history",
		"chat_sessions",
		"step_definitions",
		"steps",
		"scenarios",
		"feature_files",
		"dependencies",
		"fields",
		"methods",
		"classes",
		"files",
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, table := range tables {
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}

	return tx.Commit()
}

// ============================================================================
// TRANSACTION HELPERS
// ============================================================================

// TxFunc is a function that operates within a transaction
type TxFunc func(*sqlx.Tx) error

// WithTransaction executes a function within a transaction
func (db *DB) WithTransaction(fn TxFunc) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ============================================================================
// WORKSPACE MANAGEMENT
// ============================================================================

// InitializeWorkspace sets up database for a new workspace
func (db *DB) InitializeWorkspace(workspacePath string, config *Config) error {
	return db.WithTransaction(func(tx *sqlx.Tx) error {
		// Check if workspace already exists
		var count int
		err := tx.Get(&count, "SELECT COUNT(*) FROM config WHERE workspace_path = ?", workspacePath)
		if err != nil {
			return err
		}

		if count > 0 {
			// Update existing config
			_, err = tx.Exec(`
				UPDATE config
				SET framework = ?,
				    test_framework = ?,
				    bdd_enabled = ?,
				    base_package = ?,
				    page_object_package = ?,
				    test_package = ?,
				    config_json = ?,
				    updated_at = ?
				WHERE workspace_path = ?`,
				config.Framework,
				config.TestFramework,
				config.BDDEnabled,
				config.BasePackage,
				config.PageObjectPackage,
				config.TestPackage,
				config.ConfigJSON,
				Now(),
				workspacePath,
			)
			return err
		}

		// Insert new config
		_, err = tx.Exec(`
			INSERT INTO config (
				workspace_path, framework, test_framework, bdd_enabled,
				base_package, page_object_package, test_package, config_json,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			workspacePath,
			config.Framework,
			config.TestFramework,
			config.BDDEnabled,
			config.BasePackage,
			config.PageObjectPackage,
			config.TestPackage,
			config.ConfigJSON,
			Now(),
			Now(),
		)
		return err
	})
}

// GetWorkspaceConfig retrieves workspace configuration
func (db *DB) GetWorkspaceConfig(workspacePath string) (*Config, error) {
	var config Config
	err := db.conn.Get(&config, "SELECT * FROM config WHERE workspace_path = ?", workspacePath)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
