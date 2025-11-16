package database

import (
	"crypto/sha256"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// FileRepository handles file-related database operations
type FileRepository struct {
	db *DB
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create inserts a new file
func (r *FileRepository) Create(file *File) error {
	result, err := r.db.conn.Exec(`
		INSERT INTO files (
			path, type, content_hash, content, package_name,
			last_modified, indexed_at, size_bytes, line_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.Path,
		file.Type,
		file.ContentHash,
		file.Content,
		file.PackageName,
		file.LastModified,
		file.IndexedAt,
		file.SizeBytes,
		file.LineCount,
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	file.ID = id
	return nil
}

// Update updates an existing file
func (r *FileRepository) Update(file *File) error {
	_, err := r.db.conn.Exec(`
		UPDATE files
		SET content_hash = ?,
		    content = ?,
		    package_name = ?,
		    last_modified = ?,
		    indexed_at = ?,
		    size_bytes = ?,
		    line_count = ?
		WHERE id = ?`,
		file.ContentHash,
		file.Content,
		file.PackageName,
		file.LastModified,
		file.IndexedAt,
		file.SizeBytes,
		file.LineCount,
		file.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// Delete removes a file
func (r *FileRepository) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM files WHERE id = ?", id)
	return err
}

// DeleteByPath removes a file by path
func (r *FileRepository) DeleteByPath(path string) error {
	_, err := r.db.conn.Exec("DELETE FROM files WHERE path = ?", path)
	return err
}

// GetByID retrieves a file by ID
func (r *FileRepository) GetByID(id int64) (*File, error) {
	var file File
	err := r.db.conn.Get(&file, "SELECT * FROM files WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByPath retrieves a file by path
func (r *FileRepository) GetByPath(path string) (*File, error) {
	var file File
	err := r.db.conn.Get(&file, "SELECT * FROM files WHERE path = ?", path)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetAll retrieves all files
func (r *FileRepository) GetAll() ([]*File, error) {
	var files []*File
	err := r.db.conn.Select(&files, "SELECT * FROM files ORDER BY path")
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetByType retrieves files by type
func (r *FileRepository) GetByType(fileType string) ([]*File, error) {
	var files []*File
	err := r.db.conn.Select(&files, "SELECT * FROM files WHERE type = ? ORDER BY path", fileType)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetByPackage retrieves files by package name
func (r *FileRepository) GetByPackage(packageName string) ([]*File, error) {
	var files []*File
	err := r.db.conn.Select(&files, "SELECT * FROM files WHERE package_name = ? ORDER BY path", packageName)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Exists checks if a file exists by path
func (r *FileRepository) Exists(path string) (bool, error) {
	var count int
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM files WHERE path = ?", path)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasChanged checks if file content has changed
func (r *FileRepository) HasChanged(path string, newHash string) (bool, error) {
	var currentHash string
	err := r.db.conn.Get(&currentHash, "SELECT content_hash FROM files WHERE path = ?", path)
	if err != nil {
		return true, nil // File doesn't exist, so it's "changed"
	}
	return currentHash != newHash, nil
}

// Upsert inserts or updates a file
func (r *FileRepository) Upsert(file *File) error {
	exists, err := r.Exists(file.Path)
	if err != nil {
		return err
	}

	if exists {
		// Get existing file ID
		existing, err := r.GetByPath(file.Path)
		if err != nil {
			return err
		}
		file.ID = existing.ID
		return r.Update(file)
	}

	return r.Create(file)
}

// UpsertTx inserts or updates a file within a transaction
func (r *FileRepository) UpsertTx(tx *sqlx.Tx, file *File) error {
	// Check if exists
	var existingID int64
	err := tx.Get(&existingID, "SELECT id FROM files WHERE path = ?", file.Path)

	if err == nil {
		// Update existing
		file.ID = existingID
		_, err = tx.Exec(`
			UPDATE files
			SET content_hash = ?,
			    content = ?,
			    package_name = ?,
			    last_modified = ?,
			    indexed_at = ?,
			    size_bytes = ?,
			    line_count = ?
			WHERE id = ?`,
			file.ContentHash,
			file.Content,
			file.PackageName,
			file.LastModified,
			file.IndexedAt,
			file.SizeBytes,
			file.LineCount,
			file.ID,
		)
		return err
	}

	// Insert new
	result, err := tx.Exec(`
		INSERT INTO files (
			path, type, content_hash, content, package_name,
			last_modified, indexed_at, size_bytes, line_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.Path,
		file.Type,
		file.ContentHash,
		file.Content,
		file.PackageName,
		file.LastModified,
		file.IndexedAt,
		file.SizeBytes,
		file.LineCount,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	file.ID = id
	return nil
}

// ComputeHash computes SHA256 hash of content
func ComputeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// GetModifiedSince retrieves files modified since timestamp
func (r *FileRepository) GetModifiedSince(timestamp int64) ([]*File, error) {
	var files []*File
	err := r.db.conn.Select(&files,
		"SELECT * FROM files WHERE last_modified > ? ORDER BY last_modified DESC",
		timestamp)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Count returns total number of files
func (r *FileRepository) Count() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM files")
	return count, err
}

// CountByType returns number of files by type
func (r *FileRepository) CountByType(fileType string) (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM files WHERE type = ?", fileType)
	return count, err
}
