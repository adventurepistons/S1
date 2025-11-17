package ai

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// Chunker creates cAST-aware code chunks for retrieval
// Respects AST boundaries to never break methods mid-execution
type Chunker struct {
	db        *sql.DB
	tokenizer *Tokenizer
	config    Config
}

// NewChunker creates a new chunker
func NewChunker(db *sql.DB, config Config) *Chunker {
	return &Chunker{
		db:        db,
		tokenizer: NewTokenizer(),
		config:    config,
	}
}

// Chunk represents a code chunk ready for indexing
type Chunk struct {
	ChunkType string   // "class", "method", "field"
	EntityID  int      // ID from classes/methods/fields table
	Content   string   // The actual code
	Tokens    []string // Tokenized for BM25
	TokenCount int     // Estimated tokens
}

// ChunkProject creates chunks for the entire project
func (c *Chunker) ChunkProject() error {
	fmt.Println("📦 Creating code chunks...")

	// Get all classes
	rows, err := c.db.Query(`
		SELECT id, class_name FROM classes ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("failed to query classes: %w", err)
	}
	defer rows.Close()

	classCount := 0
	chunkCount := 0

	for rows.Next() {
		var classID int
		var className string

		if err := rows.Scan(&classID, &className); err != nil {
			continue
		}

		// Create chunks for this class
		chunks, err := c.chunkClass(classID)
		if err != nil {
			fmt.Printf("  ⚠️  Warning: Failed to chunk class %s: %v\n", className, err)
			continue
		}

		// Save chunks to database
		for _, chunk := range chunks {
			if err := c.saveChunk(chunk); err != nil {
				fmt.Printf("  ⚠️  Warning: Failed to save chunk: %v\n", err)
				continue
			}
			chunkCount++
		}

		classCount++
	}

	fmt.Printf("  ✅ Created %d chunks from %d classes\n", chunkCount, classCount)
	return nil
}

// chunkClass creates chunks for a single class using cAST strategy
func (c *Chunker) chunkClass(classID int) ([]*Chunk, error) {
	// Get class data
	var className, packageName string
	var lineStart, lineEnd int

	err := c.db.QueryRow(`
		SELECT class_name, file_id, line_start, line_end
		FROM classes WHERE id = ?
	`, classID).Scan(&className, &packageName, &lineStart, &lineEnd)

	if err != nil {
		return nil, err
	}

	// Get class code
	classCode, err := c.getClassCode(classID)
	if err != nil {
		return nil, err
	}

	// Estimate tokens in full class
	estimatedTokens := c.tokenizer.EstimateTokens(classCode)

	// Strategy: If class < MaxChunkTokens, create single chunk
	// Otherwise, chunk by methods
	if estimatedTokens <= c.config.MaxChunkTokens {
		// Small class - single chunk
		return []*Chunk{
			{
				ChunkType:  "class",
				EntityID:   classID,
				Content:    classCode,
				Tokens:     c.tokenizer.Tokenize(classCode),
				TokenCount: estimatedTokens,
			},
		}, nil
	}

	// Large class - chunk by methods
	return c.chunkClassByMethods(classID)
}

// chunkClassByMethods chunks a large class into method-level chunks
func (c *Chunker) chunkClassByMethods(classID int) ([]*Chunk, error) {
	chunks := []*Chunk{}

	// Get class header (package, imports, class declaration, fields)
	classHeader, err := c.getClassHeader(classID)
	if err != nil {
		return nil, err
	}

	// Create a chunk for class header (includes all fields)
	headerTokens := c.tokenizer.EstimateTokens(classHeader)
	chunks = append(chunks, &Chunk{
		ChunkType:  "class",
		EntityID:   classID,
		Content:    classHeader,
		Tokens:     c.tokenizer.Tokenize(classHeader),
		TokenCount: headerTokens,
	})

	// Get all methods
	methodRows, err := c.db.Query(`
		SELECT id, method_name, body_source
		FROM methods
		WHERE class_id = ?
		ORDER BY line_start
	`, classID)

	if err != nil {
		return nil, err
	}
	defer methodRows.Close()

	for methodRows.Next() {
		var methodID int
		var methodName, bodySource string

		if err := methodRows.Scan(&methodID, &methodName, &bodySource); err != nil {
			continue
		}

		// Get full method code (with annotations, signature, body)
		methodCode, err := c.getMethodCode(methodID)
		if err != nil {
			continue
		}

		methodTokens := c.tokenizer.EstimateTokens(methodCode)

		chunks = append(chunks, &Chunk{
			ChunkType:  "method",
			EntityID:   methodID,
			Content:    methodCode,
			Tokens:     c.tokenizer.Tokenize(methodCode),
			TokenCount: methodTokens,
		})
	}

	return chunks, nil
}

// getClassCode retrieves the complete code for a class
func (c *Chunker) getClassCode(classID int) (string, error) {
	var sb strings.Builder

	// Get package and imports
	var packageName string
	var fileID int

	err := c.db.QueryRow(`
		SELECT c.class_name, f.package, c.file_id
		FROM classes c
		JOIN files f ON c.file_id = f.id
		WHERE c.id = ?
	`, classID).Scan(&packageName, &packageName, &fileID)

	if err != nil {
		return "", err
	}

	// Add package
	if packageName != "" {
		sb.WriteString(fmt.Sprintf("package %s;\n\n", packageName))
	}

	// Add imports
	importRows, err := c.db.Query(`
		SELECT import_path, is_static
		FROM imports
		WHERE file_id = ?
		ORDER BY is_static DESC, import_path
	`, fileID)

	if err == nil {
		defer importRows.Close()
		for importRows.Next() {
			var importPath string
			var isStatic bool
			importRows.Scan(&importPath, &isStatic)

			if isStatic {
				sb.WriteString(fmt.Sprintf("import static %s;\n", importPath))
			} else {
				sb.WriteString(fmt.Sprintf("import %s;\n", importPath))
			}
		}
		sb.WriteString("\n")
	}

	// Add class declaration
	var className, extends, modifiers string
	c.db.QueryRow(`
		SELECT class_name, extends, modifiers
		FROM classes WHERE id = ?
	`, classID).Scan(&className, &extends, &modifiers)

	if modifiers != "" {
		sb.WriteString(modifiers + " ")
	}
	sb.WriteString("class " + className)
	if extends != "" {
		sb.WriteString(" extends " + extends)
	}
	sb.WriteString(" {\n\n")

	// Add fields
	fieldRows, err := c.db.Query(`
		SELECT field_name, field_type, modifiers, initializer
		FROM fields
		WHERE class_id = ?
		ORDER BY line_number
	`, classID)

	if err == nil {
		defer fieldRows.Close()
		for fieldRows.Next() {
			var fieldName, fieldType string
			var modifiers, initializer sql.NullString

			fieldRows.Scan(&fieldName, &fieldType, &modifiers, &initializer)

			sb.WriteString("    ")
			if modifiers.Valid && modifiers.String != "" {
				sb.WriteString(modifiers.String + " ")
			}
			sb.WriteString(fieldType + " " + fieldName)
			if initializer.Valid && initializer.String != "" {
				sb.WriteString(" = " + initializer.String)
			}
			sb.WriteString(";\n")
		}
		sb.WriteString("\n")
	}

	// Add methods (signatures only, not full bodies for class-level chunk)
	methodRows, err := c.db.Query(`
		SELECT method_name, return_type, modifiers
		FROM methods
		WHERE class_id = ?
		ORDER BY line_start
	`, classID)

	if err == nil {
		defer methodRows.Close()
		for methodRows.Next() {
			var methodName, returnType string
			var modifiers sql.NullString

			methodRows.Scan(&methodName, &returnType, &modifiers)

			sb.WriteString("    ")
			if modifiers.Valid && modifiers.String != "" {
				sb.WriteString(modifiers.String + " ")
			}
			sb.WriteString(returnType + " " + methodName + "() { ... }\n")
		}
	}

	sb.WriteString("}\n")

	return sb.String(), nil
}

// getClassHeader gets class declaration with fields but not method bodies
func (c *Chunker) getClassHeader(classID int) (string, error) {
	var sb strings.Builder

	// Get package
	var packageName, className string
	var fileID int

	c.db.QueryRow(`
		SELECT f.package, c.class_name, c.file_id
		FROM classes c
		JOIN files f ON c.file_id = f.id
		WHERE c.id = ?
	`, classID).Scan(&packageName, &className, &fileID)

	if packageName != "" {
		sb.WriteString(fmt.Sprintf("package %s;\n\n", packageName))
	}

	// Add key imports
	importRows, _ := c.db.Query(`
		SELECT import_path FROM imports WHERE file_id = ? LIMIT 10
	`, fileID)
	if importRows != nil {
		defer importRows.Close()
		for importRows.Next() {
			var importPath string
			importRows.Scan(&importPath)
			sb.WriteString(fmt.Sprintf("import %s;\n", importPath))
		}
		sb.WriteString("\n")
	}

	// Add class declaration
	sb.WriteString(fmt.Sprintf("public class %s {\n\n", className))

	// Add all fields (including @FindBy annotations)
	fieldRows, _ := c.db.Query(`
		SELECT field_name, field_type, is_web_element, locator_strategy, locator_value
		FROM fields
		WHERE class_id = ?
		ORDER BY line_number
	`, classID)

	if fieldRows != nil {
		defer fieldRows.Close()
		for fieldRows.Next() {
			var fieldName, fieldType string
			var isWebElement bool
			var locatorStrategy, locatorValue sql.NullString

			fieldRows.Scan(&fieldName, &fieldType, &isWebElement, &locatorStrategy, &locatorValue)

			if isWebElement && locatorStrategy.Valid {
				sb.WriteString(fmt.Sprintf("    @FindBy(%s = \"%s\")\n",
					locatorStrategy.String, locatorValue.String))
			}
			sb.WriteString(fmt.Sprintf("    private %s %s;\n\n", fieldType, fieldName))
		}
	}

	sb.WriteString("}\n")

	return sb.String(), nil
}

// getMethodCode retrieves the complete code for a method
func (c *Chunker) getMethodCode(methodID int) (string, error) {
	var sb strings.Builder

	// Get method data
	var methodName, returnType, modifiers string
	var bodySource sql.NullString

	err := c.db.QueryRow(`
		SELECT method_name, return_type, modifiers, body_source
		FROM methods WHERE id = ?
	`, methodID).Scan(&methodName, &returnType, &modifiers, &bodySource)

	if err != nil {
		return "", err
	}

	// Add method annotations
	annotationRows, _ := c.db.Query(`
		SELECT annotation_type, description, priority
		FROM method_annotations
		WHERE method_id = ?
		ORDER BY line_number
	`, methodID)

	if annotationRows != nil {
		defer annotationRows.Close()
		for annotationRows.Next() {
			var annotationType string
			var description sql.NullString
			var priority sql.NullInt64

			annotationRows.Scan(&annotationType, &description, &priority)

			sb.WriteString("    @" + annotationType)
			if description.Valid || priority.Valid {
				sb.WriteString("(")
				parts := []string{}
				if description.Valid {
					parts = append(parts, fmt.Sprintf("description = \"%s\"", description.String))
				}
				if priority.Valid {
					parts = append(parts, fmt.Sprintf("priority = %d", priority.Int64))
				}
				sb.WriteString(strings.Join(parts, ", "))
				sb.WriteString(")")
			}
			sb.WriteString("\n")
		}
	}

	// Add method signature
	sb.WriteString("    ")
	if modifiers != "" {
		sb.WriteString(modifiers + " ")
	}
	sb.WriteString(returnType + " " + methodName + "(")

	// Add parameters
	paramRows, _ := c.db.Query(`
		SELECT param_name, param_type
		FROM method_parameters
		WHERE method_id = ?
		ORDER BY param_order
	`, methodID)

	if paramRows != nil {
		defer paramRows.Close()
		params := []string{}
		for paramRows.Next() {
			var paramName, paramType string
			paramRows.Scan(&paramName, &paramType)
			params = append(params, paramType+" "+paramName)
		}
		sb.WriteString(strings.Join(params, ", "))
	}

	sb.WriteString(") {\n")

	// Add body
	if bodySource.Valid && bodySource.String != "" {
		// Indent body
		lines := strings.Split(bodySource.String, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				sb.WriteString("        " + line + "\n")
			}
		}
	}

	sb.WriteString("    }\n")

	return sb.String(), nil
}

// saveChunk saves a chunk to the database
func (c *Chunker) saveChunk(chunk *Chunk) error {
	// Serialize tokens to JSON
	tokensJSON, err := json.Marshal(chunk.Tokens)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(`
		INSERT OR REPLACE INTO chunks
		(chunk_type, entity_id, content, bm25_tokens, token_count)
		VALUES (?, ?, ?, ?, ?)
	`, chunk.ChunkType, chunk.EntityID, chunk.Content, string(tokensJSON), chunk.TokenCount)

	return err
}

// GetChunkCount returns the number of chunks in the database
func (c *Chunker) GetChunkCount() (int, error) {
	var count int
	err := c.db.QueryRow(`SELECT COUNT(*) FROM chunks`).Scan(&count)
	return count, err
}

// ClearChunks removes all chunks (useful for re-indexing)
func (c *Chunker) ClearChunks() error {
	_, err := c.db.Exec(`DELETE FROM chunks`)
	return err
}
