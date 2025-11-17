package graph

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/adventurepistons/automation-copilot/pkg/models"
)

// KnowledgeGraph builds and queries the knowledge graph
type KnowledgeGraph struct {
	db *sql.DB
}

// NewKnowledgeGraph creates a new knowledge graph
func NewKnowledgeGraph(db *sql.DB) *KnowledgeGraph {
	return &KnowledgeGraph{db: db}
}

// BuildGraph creates ALL relationships between entities
func (kg *KnowledgeGraph) BuildGraph() error {
	fmt.Println("Building knowledge graph...")

	// 1. Build class inheritance relationships
	fmt.Println("  - Building class relationships (EXTENDS, IMPLEMENTS)...")
	if err := kg.buildClassRelationships(); err != nil {
		return fmt.Errorf("failed to build class relationships: %w", err)
	}

	// 2. Build method call graph
	fmt.Println("  - Building method call graph (CALLS)...")
	if err := kg.buildMethodCallGraph(); err != nil {
		return fmt.Errorf("failed to build method call graph: %w", err)
	}

	// 3. Build field usage graph
	fmt.Println("  - Building field usage graph (USES)...")
	if err := kg.buildFieldUsageGraph(); err != nil {
		return fmt.Errorf("failed to build field usage graph: %w", err)
	}

	// 4. Build test-to-page-object relationships
	fmt.Println("  - Building test-page object links (TEST_USES_PAGE)...")
	if err := kg.buildTestPageObjectLinks(); err != nil {
		return fmt.Errorf("failed to build test-page links: %w", err)
	}

	fmt.Println("✓ Knowledge graph built successfully!")
	return nil
}

// buildClassRelationships creates EXTENDS and IMPLEMENTS relationships
func (kg *KnowledgeGraph) buildClassRelationships() error {
	// Get all classes
	rows, err := kg.db.Query(`
		SELECT id, class_name, extends FROM classes
		WHERE extends IS NOT NULL AND extends != ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var classID int
		var className, extends string
		if err := rows.Scan(&classID, &className, &extends); err != nil {
			continue
		}

		// Find parent class
		var parentID int
		err := kg.db.QueryRow("SELECT id FROM classes WHERE class_name = ?", extends).Scan(&parentID)
		if err == nil {
			// Create EXTENDS relationship
			kg.createRelationship("EXTENDS", "class", classID, "class", parentID, nil)
			count++
		}
	}

	// Build IMPLEMENTS relationships
	implRows, err := kg.db.Query(`
		SELECT class_id, interface_name FROM class_implements
	`)
	if err != nil {
		return err
	}
	defer implRows.Close()

	for implRows.Next() {
		var classID int
		var interfaceName string
		if err := implRows.Scan(&classID, &interfaceName); err != nil {
			continue
		}

		// Find interface class (if it exists in our codebase)
		var interfaceID int
		err := kg.db.QueryRow("SELECT id FROM classes WHERE class_name = ?", interfaceName).Scan(&interfaceID)
		if err == nil {
			kg.createRelationship("IMPLEMENTS", "class", classID, "class", interfaceID, nil)
			count++
		}
	}

	fmt.Printf("    Created %d class relationships\n", count)
	return nil
}

// buildMethodCallGraph creates CALLS relationships
func (kg *KnowledgeGraph) buildMethodCallGraph() error {
	// Get all method calls
	rows, err := kg.db.Query(`
		SELECT id, caller_method_id, method_name, object_name, line_number
		FROM method_calls
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type MethodCallRow struct {
		ID             int
		CallerMethodID int
		MethodName     string
		ObjectName     sql.NullString
		LineNumber     int
	}

	var calls []MethodCallRow
	for rows.Next() {
		var call MethodCallRow
		if err := rows.Scan(&call.ID, &call.CallerMethodID, &call.MethodName, &call.ObjectName, &call.LineNumber); err != nil {
			continue
		}
		calls = append(calls, call)
	}

	count := 0
	for _, call := range calls {
		// Resolve which method is being called
		calleeMethodID := kg.resolveMethodCall(call.CallerMethodID, call.MethodName, call.ObjectName.String)

		if calleeMethodID > 0 {
			// Create CALLS relationship
			metadata := map[string]interface{}{
				"line_number": call.LineNumber,
				"method_name": call.MethodName,
			}
			kg.createRelationship("CALLS", "method", call.CallerMethodID, "method", calleeMethodID, metadata)

			// Update method_calls table with resolved callee
			kg.db.Exec("UPDATE method_calls SET callee_method_id = ? WHERE id = ?", calleeMethodID, call.ID)
			count++
		}
	}

	fmt.Printf("    Resolved %d method calls\n", count)
	return nil
}

// resolveMethodCall determines which method is being called
func (kg *KnowledgeGraph) resolveMethodCall(callerMethodID int, methodName, objectName string) int {
	// Get caller method's class
	var callerClassID int
	err := kg.db.QueryRow("SELECT class_id FROM methods WHERE id = ?", callerMethodID).Scan(&callerClassID)
	if err != nil {
		return 0
	}

	// Strategy 1: If object name is a field, find field type and search in that class
	if objectName != "" {
		var fieldType string
		err := kg.db.QueryRow(`
			SELECT field_type FROM fields
			WHERE class_id = ? AND field_name = ?
		`, callerClassID, objectName).Scan(&fieldType)

		if err == nil {
			// Find the class for this field type
			var targetClassID int
			err := kg.db.QueryRow("SELECT id FROM classes WHERE class_name = ?", fieldType).Scan(&targetClassID)
			if err == nil {
				// Find method in target class
				var methodID int
				err := kg.db.QueryRow(`
					SELECT id FROM methods
					WHERE class_id = ? AND method_name = ?
					LIMIT 1
				`, targetClassID, methodName).Scan(&methodID)
				if err == nil {
					return methodID
				}
			}
		}
	}

	// Strategy 2: Search in same class (this.methodName or just methodName)
	var methodID int
	err = kg.db.QueryRow(`
		SELECT id FROM methods
		WHERE class_id = ? AND method_name = ?
		LIMIT 1
	`, callerClassID, methodName).Scan(&methodID)
	if err == nil {
		return methodID
	}

	// Strategy 3: Search in parent class (if extends)
	var parentClassName string
	err = kg.db.QueryRow("SELECT extends FROM classes WHERE id = ?", callerClassID).Scan(&parentClassName)
	if err == nil && parentClassName != "" {
		var parentClassID int
		err := kg.db.QueryRow("SELECT id FROM classes WHERE class_name = ?", parentClassName).Scan(&parentClassID)
		if err == nil {
			err := kg.db.QueryRow(`
				SELECT id FROM methods
				WHERE class_id = ? AND method_name = ?
				LIMIT 1
			`, parentClassID, methodName).Scan(&methodID)
			if err == nil {
				return methodID
			}
		}
	}

	return 0
}

// buildFieldUsageGraph creates USES relationships
func (kg *KnowledgeGraph) buildFieldUsageGraph() error {
	// Get all field accesses
	rows, err := kg.db.Query(`
		SELECT id, method_id, field_name, line_number
		FROM field_access
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var accessID, methodID, lineNumber int
		var fieldName string
		if err := rows.Scan(&accessID, &methodID, &fieldName, &lineNumber); err != nil {
			continue
		}

		// Get method's class
		var classID int
		err := kg.db.QueryRow("SELECT class_id FROM methods WHERE id = ?", methodID).Scan(&classID)
		if err != nil {
			continue
		}

		// Find field in class
		var fieldID int
		err = kg.db.QueryRow(`
			SELECT id FROM fields
			WHERE class_id = ? AND field_name = ?
		`, classID, fieldName).Scan(&fieldID)

		if err == nil {
			// Create USES relationship
			metadata := map[string]interface{}{
				"line_number": lineNumber,
				"access_type": "read",
			}
			kg.createRelationship("USES", "method", methodID, "field", fieldID, metadata)

			// Update field_access table
			kg.db.Exec("UPDATE field_access SET field_id = ? WHERE id = ?", fieldID, accessID)
			count++
		}
	}

	fmt.Printf("    Created %d field usage relationships\n", count)
	return nil
}

// buildTestPageObjectLinks creates TEST_USES_PAGE relationships
func (kg *KnowledgeGraph) buildTestPageObjectLinks() error {
	// Get all test classes
	rows, err := kg.db.Query("SELECT id FROM classes WHERE is_test_class = 1")
	if err != nil {
		return err
	}
	defer rows.Close()

	var testClassIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		testClassIDs = append(testClassIDs, id)
	}

	count := 0
	for _, testClassID := range testClassIDs {
		// Find all page object fields in test class
		fieldRows, err := kg.db.Query(`
			SELECT field_type FROM fields
			WHERE class_id = ?
		`, testClassID)
		if err != nil {
			continue
		}

		for fieldRows.Next() {
			var fieldType string
			if err := fieldRows.Scan(&fieldType); err != nil {
				continue
			}

			// Check if field type is a page object
			var pageClassID int
			var isPageObject bool
			err := kg.db.QueryRow(`
				SELECT id, is_page_object FROM classes
				WHERE class_name = ?
			`, fieldType).Scan(&pageClassID, &isPageObject)

			if err == nil && isPageObject {
				kg.createRelationship("TEST_USES_PAGE", "class", testClassID, "class", pageClassID, nil)
				count++
			}
		}
		fieldRows.Close()
	}

	fmt.Printf("    Created %d test-page relationships\n", count)
	return nil
}

// createRelationship creates a relationship in the database
func (kg *KnowledgeGraph) createRelationship(relType, fromEntityType string, fromEntityID int, toEntityType string, toEntityID int, metadata map[string]interface{}) error {
	var metadataJSON []byte
	if metadata != nil {
		metadataJSON, _ = json.Marshal(metadata)
	}

	_, err := kg.db.Exec(`
		INSERT INTO relationships (
			relationship_type, from_entity_type, from_entity_id,
			to_entity_type, to_entity_id, metadata
		) VALUES (?, ?, ?, ?, ?, ?)
	`, relType, fromEntityType, fromEntityID, toEntityType, toEntityID, string(metadataJSON))

	return err
}

// QUERY METHODS

// FindTestsUsingPageObject finds all test classes using a specific page object
func (kg *KnowledgeGraph) FindTestsUsingPageObject(pageObjectName string) ([]models.ClassData, error) {
	// Find page object class ID
	var pageClassID int
	err := kg.db.QueryRow("SELECT id FROM classes WHERE class_name = ?", pageObjectName).Scan(&pageClassID)
	if err != nil {
		return nil, fmt.Errorf("page object not found: %s", pageObjectName)
	}

	// Find all test classes that use this page object
	rows, err := kg.db.Query(`
		SELECT DISTINCT c.class_name, c.fully_qualified_name, f.file_path
		FROM relationships r
		JOIN classes c ON r.from_entity_id = c.id
		JOIN files f ON c.file_id = f.id
		WHERE r.relationship_type = 'TEST_USES_PAGE'
		  AND r.to_entity_id = ?
		  AND r.from_entity_type = 'class'
	`, pageClassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []models.ClassData
	for rows.Next() {
		var class models.ClassData
		if err := rows.Scan(&class.ClassName, &class.Package, &class.FilePath); err != nil {
			continue
		}
		classes = append(classes, class)
	}

	return classes, nil
}

// FindFieldDefinition finds where a field is defined
func (kg *KnowledgeGraph) FindFieldDefinition(fieldName string) (*FieldLocation, error) {
	var location FieldLocation

	err := kg.db.QueryRow(`
		SELECT f.field_name, f.locator_strategy, f.locator_value, f.line_number,
		       c.class_name, fi.file_path
		FROM fields f
		JOIN classes c ON f.class_id = c.id
		JOIN files fi ON c.file_id = fi.id
		WHERE f.field_name = ?
		LIMIT 1
	`, fieldName).Scan(
		&location.FieldName, &location.LocatorStrategy, &location.LocatorValue,
		&location.LineNumber, &location.ClassName, &location.FilePath,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("field not found: %s", fieldName)
	}

	if err != nil {
		return nil, err
	}

	return &location, nil
}

// GetCompleteCallChain gets full call chain for a test method
func (kg *KnowledgeGraph) GetCompleteCallChain(methodName string) ([]models.Method, error) {
	// Find method ID
	var methodID int
	err := kg.db.QueryRow(`
		SELECT id FROM methods
		WHERE method_name = ? AND is_test = 1
		LIMIT 1
	`, methodName).Scan(&methodID)
	if err != nil {
		return nil, fmt.Errorf("test method not found: %s", methodName)
	}

	chain := []models.Method{}
	visited := make(map[int]bool)
	kg.traverseCallGraph(methodID, &chain, visited)

	return chain, nil
}

// traverseCallGraph recursively traverses the call graph
func (kg *KnowledgeGraph) traverseCallGraph(methodID int, chain *[]models.Method, visited map[int]bool) {
	if visited[methodID] {
		return // Prevent infinite loops
	}
	visited[methodID] = true

	// Get method details
	var method models.Method
	var className string
	err := kg.db.QueryRow(`
		SELECT m.method_name, m.return_type, c.class_name
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		WHERE m.id = ?
	`, methodID).Scan(&method.Name, &method.ReturnType, &className)

	if err != nil {
		return
	}

	*chain = append(*chain, method)

	// Get all methods called by this method
	rows, err := kg.db.Query(`
		SELECT to_entity_id FROM relationships
		WHERE relationship_type = 'CALLS'
		  AND from_entity_type = 'method'
		  AND from_entity_id = ?
	`, methodID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var calleeID int
		if err := rows.Scan(&calleeID); err != nil {
			continue
		}
		kg.traverseCallGraph(calleeID, chain, visited)
	}
}

// GetMethodsUsingField finds all methods that use a specific field
func (kg *KnowledgeGraph) GetMethodsUsingField(fieldName string) ([]MethodInfo, error) {
	// Find field ID
	var fieldID int
	err := kg.db.QueryRow("SELECT id FROM fields WHERE field_name = ?", fieldName).Scan(&fieldID)
	if err != nil {
		return nil, fmt.Errorf("field not found: %s", fieldName)
	}

	// Find all methods that use this field
	rows, err := kg.db.Query(`
		SELECT DISTINCT m.method_name, m.line_start, c.class_name, f.file_path
		FROM relationships r
		JOIN methods m ON r.from_entity_id = m.id
		JOIN classes c ON m.class_id = c.id
		JOIN files f ON c.file_id = f.id
		WHERE r.relationship_type = 'USES'
		  AND r.to_entity_id = ?
		  AND r.to_entity_type = 'field'
	`, fieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []MethodInfo
	for rows.Next() {
		var method MethodInfo
		if err := rows.Scan(&method.MethodName, &method.LineNumber, &method.ClassName, &method.FilePath); err != nil {
			continue
		}
		methods = append(methods, method)
	}

	return methods, nil
}

// GetAllPageObjects returns all page object classes
func (kg *KnowledgeGraph) GetAllPageObjects() ([]PageObjectInfo, error) {
	rows, err := kg.db.Query(`
		SELECT c.class_name, f.file_path, COUNT(DISTINCT fld.id) as element_count
		FROM classes c
		JOIN files f ON c.file_id = f.id
		LEFT JOIN fields fld ON c.id = fld.class_id AND fld.is_web_element = 1
		WHERE c.is_page_object = 1
		GROUP BY c.id, c.class_name, f.file_path
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pageObjects []PageObjectInfo
	for rows.Next() {
		var po PageObjectInfo
		if err := rows.Scan(&po.ClassName, &po.FilePath, &po.ElementCount); err != nil {
			continue
		}
		pageObjects = append(pageObjects, po)
	}

	return pageObjects, nil
}

// Helper types for query results

type FieldLocation struct {
	FieldName       string
	ClassName       string
	FilePath        string
	LineNumber      int
	LocatorStrategy string
	LocatorValue    string
}

type MethodInfo struct {
	MethodName string
	ClassName  string
	FilePath   string
	LineNumber int
}

type PageObjectInfo struct {
	ClassName    string
	FilePath     string
	ElementCount int
}
