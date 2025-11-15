package indexer

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/parser"
)

// XMLIndexer indexes XML files (pom.xml, testng.xml)
type XMLIndexer struct {
	db        *database.DB
	pomParser *parser.PomParser
	// testngParser *parser.TestNGParser // Can add later if needed
}

// NewXMLIndexer creates a new XML indexer
func NewXMLIndexer(db *database.DB) *XMLIndexer {
	return &XMLIndexer{
		db:        db,
		pomParser: parser.NewPomParser(),
	}
}

// IndexFile indexes an XML file
func (idx *XMLIndexer) IndexFile(file *database.File, content string) error {
	// Determine XML type from filename
	fileName := strings.ToLower(file.Path)

	if strings.HasSuffix(fileName, "pom.xml") {
		return idx.indexPOM(file, content)
	} else if strings.HasSuffix(fileName, "testng.xml") {
		return idx.indexTestNG(file, content)
	}

	// Unknown XML type, skip
	return nil
}

// indexPOM indexes a pom.xml file
func (idx *XMLIndexer) indexPOM(file *database.File, content string) error {
	// Parse POM
	pom, err := idx.pomParser.ParseFile(file.Path)
	if err != nil {
		return fmt.Errorf("failed to parse POM: %w", err)
	}

	// Store POM information in config table
	// This helps us understand the project structure

	return idx.db.WithTransaction(func(tx *sqlx.Tx) error {
		// Check if config exists for this workspace
		// We'll use the directory containing pom.xml as workspace path
		workspacePath := file.Path[:strings.LastIndex(file.Path, "/")]

		var count int
		err := tx.Get(&count, "SELECT COUNT(*) FROM config WHERE workspace_path = ?", workspacePath)
		if err != nil {
			return err
		}

		if count == 0 {
			// Create config
			_, err = tx.Exec(`
				INSERT INTO config (workspace_path, base_package, created_at, updated_at)
				VALUES (?, ?, ?, ?)`,
				workspacePath,
				database.StringPtr(pom.GroupID),
				database.Now(),
				database.Now(),
			)
			return err
		}

		// Update existing config with POM info
		_, err = tx.Exec(`
			UPDATE config
			SET base_package = ?,
			    updated_at = ?
			WHERE workspace_path = ?`,
			database.StringPtr(pom.GroupID),
			database.Now(),
			workspacePath,
		)
		return err
	})
}

// indexTestNG indexes a testng.xml file
func (idx *XMLIndexer) indexTestNG(file *database.File, content string) error {
	// For now, we just store the file content
	// Later, we can parse and extract test suites, classes, etc.
	return nil
}
