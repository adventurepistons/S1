package ai

import (
	"database/sql"
	"fmt"
	"strings"
)

// ContextEnricher adds explanatory context to code chunks (Anthropic method)
// Research shows this improves retrieval accuracy by 49%!
type ContextEnricher struct {
	db *sql.DB
}

// NewContextEnricher creates a new context enricher
func NewContextEnricher(db *sql.DB) *ContextEnricher {
	return &ContextEnricher{
		db: db,
	}
}

// EnrichChunk adds contextual information to a chunk
func (ce *ContextEnricher) EnrichChunk(chunkID int, chunkType string, entityID int) (string, error) {
	switch chunkType {
	case "class":
		return ce.enrichClassChunk(entityID)
	case "method":
		return ce.enrichMethodChunk(entityID)
	case "field":
		return ce.enrichFieldChunk(entityID)
	default:
		return "", fmt.Errorf("unknown chunk type: %s", chunkType)
	}
}

// EnrichAllChunks enriches all chunks in the database
func (ce *ContextEnricher) EnrichAllChunks() error {
	fmt.Println("📝 Enriching chunks with context...")

	// Get all chunks
	rows, err := ce.db.Query(`
		SELECT id, chunk_type, entity_id, content
		FROM chunks
		WHERE enriched_content IS NULL OR enriched_content = ''
		ORDER BY id
	`)

	if err != nil {
		return fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		var chunkID, entityID int
		var chunkType, content string

		if err := rows.Scan(&chunkID, &chunkType, &entityID, &content); err != nil {
			continue
		}

		// Enrich the chunk
		enrichedContent, err := ce.EnrichChunk(chunkID, chunkType, entityID)
		if err != nil {
			fmt.Printf("  ⚠️  Warning: Failed to enrich chunk %d: %v\n", chunkID, err)
			continue
		}

		// Save enriched content
		_, err = ce.db.Exec(`
			UPDATE chunks
			SET enriched_content = ?
			WHERE id = ?
		`, enrichedContent, chunkID)

		if err != nil {
			fmt.Printf("  ⚠️  Warning: Failed to save enriched content for chunk %d: %v\n", chunkID, err)
			continue
		}

		count++
	}

	fmt.Printf("  ✅ Enriched %d chunks\n", count)
	return nil
}

// enrichClassChunk adds context to a class chunk
func (ce *ContextEnricher) enrichClassChunk(classID int) (string, error) {
	var sb strings.Builder

	// Get class data
	var className, packageName, filePath string
	var isPageObject, isTestClass bool

	err := ce.db.QueryRow(`
		SELECT c.class_name, f.package, f.file_path, c.is_page_object, c.is_test_class
		FROM classes c
		JOIN files f ON c.file_id = f.id
		WHERE c.id = ?
	`, classID).Scan(&className, &packageName, &filePath, &isPageObject, &isTestClass)

	if err != nil {
		return "", err
	}

	// Build context header
	sb.WriteString("CONTEXT:\n")
	sb.WriteString(fmt.Sprintf("This is %s from package %s (file: %s).\n", className, packageName, filePath))

	// Determine class type
	if isPageObject {
		sb.WriteString("It is a Page Object class representing a web page.\n")
	} else if isTestClass {
		sb.WriteString("It is a Test class containing test methods.\n")
	} else {
		sb.WriteString("It is a utility/helper class.\n")
	}

	// Get parent class
	var extends sql.NullString
	ce.db.QueryRow(`SELECT extends FROM classes WHERE id = ?`, classID).Scan(&extends)
	if extends.Valid && extends.String != "" {
		sb.WriteString(fmt.Sprintf("It extends %s.\n", extends.String))
	}

	// Get interfaces
	interfaceRows, _ := ce.db.Query(`
		SELECT interface_name FROM class_implements WHERE class_id = ?
	`, classID)
	if interfaceRows != nil {
		defer interfaceRows.Close()
		interfaces := []string{}
		for interfaceRows.Next() {
			var interfaceName string
			interfaceRows.Scan(&interfaceName)
			interfaces = append(interfaces, interfaceName)
		}
		if len(interfaces) > 0 {
			sb.WriteString(fmt.Sprintf("It implements: %s.\n", strings.Join(interfaces, ", ")))
		}
	}

	// Get WebElements (for Page Objects)
	if isPageObject {
		elementRows, _ := ce.db.Query(`
			SELECT field_name, locator_strategy, locator_value
			FROM fields
			WHERE class_id = ? AND is_web_element = 1
			ORDER BY line_number
		`, classID)

		if elementRows != nil {
			defer elementRows.Close()
			elements := []string{}
			for elementRows.Next() {
				var fieldName string
				var locatorStrategy, locatorValue sql.NullString
				elementRows.Scan(&fieldName, &locatorStrategy, &locatorValue)

				if locatorStrategy.Valid {
					elements = append(elements, fmt.Sprintf("%s (@FindBy %s=\"%s\")",
						fieldName, locatorStrategy.String, locatorValue.String))
				} else {
					elements = append(elements, fieldName)
				}
			}

			if len(elements) > 0 {
				sb.WriteString(fmt.Sprintf("It defines %d WebElements: %s.\n",
					len(elements), strings.Join(elements, ", ")))
			}
		}
	}

	// Get test methods (for Test classes)
	if isTestClass {
		testRows, _ := ce.db.Query(`
			SELECT method_name, test_type
			FROM methods
			WHERE class_id = ? AND is_test = 1
			ORDER BY line_start
		`, classID)

		if testRows != nil {
			defer testRows.Close()
			tests := []string{}
			for testRows.Next() {
				var methodName, testType string
				testRows.Scan(&methodName, &testType)
				tests = append(tests, methodName)
			}

			if len(tests) > 0 {
				sb.WriteString(fmt.Sprintf("It contains %d test methods: %s.\n",
					len(tests), strings.Join(tests, ", ")))
			}
		}
	}

	// Get classes that use this class
	usedByRows, _ := ce.db.Query(`
		SELECT DISTINCT c2.class_name
		FROM method_calls mc
		JOIN methods m1 ON mc.callee_method_id = m1.id
		JOIN methods m2 ON mc.caller_method_id = m2.id
		JOIN classes c1 ON m1.class_id = c1.id
		JOIN classes c2 ON m2.class_id = c2.id
		WHERE c1.id = ? AND c2.id != ?
		LIMIT 5
	`, classID, classID)

	if usedByRows != nil {
		defer usedByRows.Close()
		usedBy := []string{}
		for usedByRows.Next() {
			var className string
			usedByRows.Scan(&className)
			usedBy = append(usedBy, className)
		}

		if len(usedBy) > 0 {
			sb.WriteString(fmt.Sprintf("It is used by: %s.\n", strings.Join(usedBy, ", ")))
		}
	}

	// Get actual code
	sb.WriteString("\nCODE:\n")

	var content string
	ce.db.QueryRow(`SELECT content FROM chunks WHERE chunk_type = 'class' AND entity_id = ?`, classID).Scan(&content)
	sb.WriteString(content)

	return sb.String(), nil
}

// enrichMethodChunk adds context to a method chunk
func (ce *ContextEnricher) enrichMethodChunk(methodID int) (string, error) {
	var sb strings.Builder

	// Get method data
	var methodName, returnType, className string
	var isTest bool
	var testType sql.NullString

	err := ce.db.QueryRow(`
		SELECT m.method_name, m.return_type, c.class_name, m.is_test, m.test_type
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		WHERE m.id = ?
	`, methodID).Scan(&methodName, &returnType, &className, &isTest, &testType)

	if err != nil {
		return "", err
	}

	// Build context
	sb.WriteString("CONTEXT:\n")
	sb.WriteString(fmt.Sprintf("This is the %s() method from %s.\n", methodName, className))

	if isTest {
		sb.WriteString(fmt.Sprintf("It is a %s test method.\n", testType.String))
	}

	sb.WriteString(fmt.Sprintf("Return type: %s.\n", returnType))

	// Get parameters
	paramRows, _ := ce.db.Query(`
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
			params = append(params, fmt.Sprintf("%s %s", paramType, paramName))
		}

		if len(params) > 0 {
			sb.WriteString(fmt.Sprintf("Parameters: %s.\n", strings.Join(params, ", ")))
		}
	}

	// Get method calls
	callRows, _ := ce.db.Query(`
		SELECT method_name, object_name
		FROM method_calls
		WHERE caller_method_id = ?
		ORDER BY line_number
		LIMIT 10
	`, methodID)

	if callRows != nil {
		defer callRows.Close()
		calls := []string{}
		for callRows.Next() {
			var methodName string
			var objectName sql.NullString
			callRows.Scan(&methodName, &objectName)

			if objectName.Valid && objectName.String != "" {
				calls = append(calls, fmt.Sprintf("%s.%s()", objectName.String, methodName))
			} else {
				calls = append(calls, fmt.Sprintf("%s()", methodName))
			}
		}

		if len(calls) > 0 {
			sb.WriteString(fmt.Sprintf("It calls: %s.\n", strings.Join(calls, ", ")))
		}
	}

	// Get fields accessed
	fieldRows, _ := ce.db.Query(`
		SELECT DISTINCT f.field_name, f.field_type, f.is_web_element
		FROM field_access fa
		JOIN fields f ON fa.field_id = f.id
		WHERE fa.method_id = ?
		LIMIT 10
	`, methodID)

	if fieldRows != nil {
		defer fieldRows.Close()
		fields := []string{}
		for fieldRows.Next() {
			var fieldName, fieldType string
			var isWebElement bool
			fieldRows.Scan(&fieldName, &fieldType, &isWebElement)

			if isWebElement {
				fields = append(fields, fmt.Sprintf("%s (WebElement)", fieldName))
			} else {
				fields = append(fields, fieldName)
			}
		}

		if len(fields) > 0 {
			sb.WriteString(fmt.Sprintf("It uses fields: %s.\n", strings.Join(fields, ", ")))
		}
	}

	// Get assertions (for test methods)
	if isTest {
		var assertionCount int
		ce.db.QueryRow(`
			SELECT COUNT(*) FROM assertions WHERE method_id = ?
		`, methodID).Scan(&assertionCount)

		if assertionCount > 0 {
			sb.WriteString(fmt.Sprintf("It contains %d assertions.\n", assertionCount))
		}
	}

	// Get waits
	var usesExplicitWait bool
	var waitTimeout sql.NullInt64
	ce.db.QueryRow(`
		SELECT uses_explicit_wait, wait_timeout
		FROM methods
		WHERE id = ?
	`, methodID).Scan(&usesExplicitWait, &waitTimeout)

	if usesExplicitWait {
		if waitTimeout.Valid {
			sb.WriteString(fmt.Sprintf("It uses explicit waits (timeout: %d seconds).\n", waitTimeout.Int64))
		} else {
			sb.WriteString("It uses explicit waits.\n")
		}
	}

	// Get actual code
	sb.WriteString("\nCODE:\n")

	var content string
	ce.db.QueryRow(`SELECT content FROM chunks WHERE chunk_type = 'method' AND entity_id = ?`, methodID).Scan(&content)
	sb.WriteString(content)

	return sb.String(), nil
}

// enrichFieldChunk adds context to a field chunk (WebElements)
func (ce *ContextEnricher) enrichFieldChunk(fieldID int) (string, error) {
	var sb strings.Builder

	// Get field data
	var fieldName, fieldType, className string
	var isWebElement bool
	var locatorStrategy, locatorValue sql.NullString

	err := ce.db.QueryRow(`
		SELECT f.field_name, f.field_type, c.class_name, f.is_web_element, f.locator_strategy, f.locator_value
		FROM fields f
		JOIN classes c ON f.class_id = c.id
		WHERE f.id = ?
	`, fieldID).Scan(&fieldName, &fieldType, &className, &isWebElement, &locatorStrategy, &locatorValue)

	if err != nil {
		return "", err
	}

	// Build context
	sb.WriteString("CONTEXT:\n")
	sb.WriteString(fmt.Sprintf("This is the %s field from %s.\n", fieldName, className))
	sb.WriteString(fmt.Sprintf("Type: %s.\n", fieldType))

	if isWebElement && locatorStrategy.Valid {
		sb.WriteString(fmt.Sprintf("It is a WebElement using @FindBy(%s = \"%s\").\n",
			locatorStrategy.String, locatorValue.String))
	}

	// Get methods that use this field
	usedInRows, _ := ce.db.Query(`
		SELECT DISTINCT m.method_name
		FROM field_access fa
		JOIN methods m ON fa.method_id = m.id
		WHERE fa.field_id = ?
		LIMIT 10
	`, fieldID)

	if usedInRows != nil {
		defer usedInRows.Close()
		methods := []string{}
		for usedInRows.Next() {
			var methodName string
			usedInRows.Scan(&methodName)
			methods = append(methods, methodName)
		}

		if len(methods) > 0 {
			sb.WriteString(fmt.Sprintf("It is used in: %s.\n", strings.Join(methods, ", ")))
		}
	}

	// For WebElements, describe what it likely represents
	if isWebElement {
		fieldLower := strings.ToLower(fieldName)

		if strings.Contains(fieldLower, "button") {
			sb.WriteString("This is likely a clickable button element.\n")
		} else if strings.Contains(fieldLower, "input") || strings.Contains(fieldLower, "field") {
			sb.WriteString("This is likely an input field for user data entry.\n")
		} else if strings.Contains(fieldLower, "link") {
			sb.WriteString("This is likely a clickable link element.\n")
		} else if strings.Contains(fieldLower, "text") || strings.Contains(fieldLower, "label") {
			sb.WriteString("This is likely a text display element.\n")
		}
	}

	// Get actual code (field declaration)
	sb.WriteString("\nCODE:\n")

	var modifiers sql.NullString
	ce.db.QueryRow(`SELECT modifiers FROM fields WHERE id = ?`, fieldID).Scan(&modifiers)

	if isWebElement && locatorStrategy.Valid {
		sb.WriteString(fmt.Sprintf("@FindBy(%s = \"%s\")\n", locatorStrategy.String, locatorValue.String))
	}

	if modifiers.Valid && modifiers.String != "" {
		sb.WriteString(fmt.Sprintf("%s %s %s;\n", modifiers.String, fieldType, fieldName))
	} else {
		sb.WriteString(fmt.Sprintf("private %s %s;\n", fieldType, fieldName))
	}

	return sb.String(), nil
}

// ClearEnrichedContent removes all enriched content (useful for re-indexing)
func (ce *ContextEnricher) ClearEnrichedContent() error {
	_, err := ce.db.Exec(`UPDATE chunks SET enriched_content = NULL`)
	return err
}
