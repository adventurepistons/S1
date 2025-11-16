package storage

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/adventurepistons/automation-copilot/pkg/models"
)

//go:embed schema.sql
var schemaFS embed.FS

// Database handles all SQLite operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	database := &Database{db: db}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// initSchema creates all tables from schema.sql
func (d *Database) initSchema() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	if _, err := d.db.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// SaveClassData saves complete class data to database
func (d *Database) SaveClassData(classData *models.ClassData) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Save file
	fileID, err := d.saveFile(tx, classData)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// 2. Save imports
	if err := d.saveImports(tx, fileID, classData.Imports); err != nil {
		return fmt.Errorf("failed to save imports: %w", err)
	}

	// 3. Save class
	classID, err := d.saveClass(tx, fileID, classData)
	if err != nil {
		return fmt.Errorf("failed to save class: %w", err)
	}

	// 4. Save implements
	if err := d.saveImplements(tx, classID, classData.Implements); err != nil {
		return fmt.Errorf("failed to save implements: %w", err)
	}

	// 5. Save fields
	if err := d.saveFields(tx, classID, classData.Fields); err != nil {
		return fmt.Errorf("failed to save fields: %w", err)
	}

	// 6. Save methods
	if err := d.saveMethods(tx, classID, classData.Methods); err != nil {
		return fmt.Errorf("failed to save methods: %w", err)
	}

	return tx.Commit()
}

// saveFile saves file metadata
func (d *Database) saveFile(tx *sql.Tx, classData *models.ClassData) (int64, error) {
	now := time.Now()

	result, err := tx.Exec(`
		INSERT INTO files (file_path, package, last_modified, last_indexed)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			package = excluded.package,
			last_modified = excluded.last_modified,
			last_indexed = excluded.last_indexed
	`, classData.FilePath, classData.Package, now, now)

	if err != nil {
		return 0, err
	}

	// Get the file ID
	var fileID int64
	err = tx.QueryRow("SELECT id FROM files WHERE file_path = ?", classData.FilePath).Scan(&fileID)
	if err != nil {
		return 0, err
	}

	return fileID, nil
}

// saveImports saves import statements
func (d *Database) saveImports(tx *sql.Tx, fileID int64, imports []models.Import) error {
	// Delete old imports
	if _, err := tx.Exec("DELETE FROM imports WHERE file_id = ?", fileID); err != nil {
		return err
	}

	// Insert new imports
	stmt, err := tx.Prepare(`
		INSERT INTO imports (file_id, import_path, is_static, is_wildcard)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, imp := range imports {
		if _, err := stmt.Exec(fileID, imp.Path, imp.IsStatic, imp.IsWildcard); err != nil {
			return err
		}
	}

	return nil
}

// saveClass saves class metadata
func (d *Database) saveClass(tx *sql.Tx, fileID int64, classData *models.ClassData) (int64, error) {
	modifiersJSON, _ := json.Marshal(classData.Modifiers)
	fullyQualifiedName := classData.Package + "." + classData.ClassName

	result, err := tx.Exec(`
		INSERT INTO classes (
			file_id, class_name, fully_qualified_name, extends, modifiers,
			is_page_object, is_test_class, is_step_definition,
			line_start, line_end
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(fully_qualified_name) DO UPDATE SET
			file_id = excluded.file_id,
			class_name = excluded.class_name,
			extends = excluded.extends,
			modifiers = excluded.modifiers,
			is_page_object = excluded.is_page_object,
			is_test_class = excluded.is_test_class,
			is_step_definition = excluded.is_step_definition,
			line_start = excluded.line_start,
			line_end = excluded.line_end
	`,
		fileID, classData.ClassName, fullyQualifiedName, classData.Extends,
		string(modifiersJSON), classData.PageObjectModel, classData.TestClass,
		classData.StepDefinition, classData.LineStart, classData.LineEnd,
	)

	if err != nil {
		return 0, err
	}

	// Get the class ID
	var classID int64
	err = tx.QueryRow("SELECT id FROM classes WHERE fully_qualified_name = ?", fullyQualifiedName).Scan(&classID)
	if err != nil {
		return 0, err
	}

	return classID, nil
}

// saveImplements saves implemented interfaces
func (d *Database) saveImplements(tx *sql.Tx, classID int64, implements []string) error {
	// Delete old implements
	if _, err := tx.Exec("DELETE FROM class_implements WHERE class_id = ?", classID); err != nil {
		return err
	}

	// Insert new implements
	stmt, err := tx.Prepare("INSERT INTO class_implements (class_id, interface_name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, iface := range implements {
		if _, err := stmt.Exec(classID, iface); err != nil {
			return err
		}
	}

	return nil
}

// saveFields saves all fields including @FindBy WebElements
func (d *Database) saveFields(tx *sql.Tx, classID int64, fields []models.Field) error {
	// Delete old fields
	if _, err := tx.Exec("DELETE FROM fields WHERE class_id = ?", classID); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO fields (
			class_id, field_name, field_type, modifiers, initializer,
			is_web_element, locator_strategy, locator_value, line_number
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, field := range fields {
		modifiersJSON, _ := json.Marshal(field.Modifiers)

		result, err := stmt.Exec(
			classID, field.Name, field.Type, string(modifiersJSON), field.Initializer,
			field.IsWebElement, field.LocatorStrategy, field.LocatorValue, field.LineNumber,
		)
		if err != nil {
			return err
		}

		fieldID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Save field annotations
		if err := d.saveFieldAnnotations(tx, fieldID, field.Annotations); err != nil {
			return err
		}
	}

	return nil
}

// saveFieldAnnotations saves annotations on fields
func (d *Database) saveFieldAnnotations(tx *sql.Tx, fieldID int64, annotations []models.Annotation) error {
	stmt, err := tx.Prepare(`
		INSERT INTO field_annotations (field_id, annotation_type, parameters, how, using, line_number)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ann := range annotations {
		paramsJSON, _ := json.Marshal(ann.Parameters)

		if _, err := stmt.Exec(fieldID, ann.Type, string(paramsJSON), ann.How, ann.Using, ann.LineNumber); err != nil {
			return err
		}
	}

	return nil
}

// saveMethods saves all methods with complete body analysis
func (d *Database) saveMethods(tx *sql.Tx, classID int64, methods []models.Method) error {
	// Delete old methods
	if _, err := tx.Exec("DELETE FROM methods WHERE class_id = ?", classID); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO methods (
			class_id, method_name, return_type, modifiers, is_test, test_type,
			uses_explicit_wait, uses_implicit_wait, wait_timeout,
			if_statements, for_loops, try_catch_blocks,
			line_start, line_end, body_source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, method := range methods {
		modifiersJSON, _ := json.Marshal(method.Modifiers)

		result, err := stmt.Exec(
			classID, method.Name, method.ReturnType, string(modifiersJSON),
			method.IsTest, method.TestType,
			method.UsesExplicitWait, method.UsesImplicitWait, method.WaitTimeout,
			method.IfStatements, method.ForLoops, method.TryCatchBlocks,
			method.LineStart, method.LineEnd, method.BodySource,
		)
		if err != nil {
			return err
		}

		methodID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Save method parameters
		if err := d.saveMethodParameters(tx, methodID, method.Parameters); err != nil {
			return err
		}

		// Save method annotations
		if err := d.saveMethodAnnotations(tx, methodID, method.Annotations); err != nil {
			return err
		}

		// Save method calls
		if err := d.saveMethodCalls(tx, methodID, method.MethodCalls); err != nil {
			return err
		}

		// Save field accesses
		if err := d.saveFieldAccesses(tx, methodID, method.FieldAccess); err != nil {
			return err
		}

		// Save local variables
		if err := d.saveLocalVariables(tx, methodID, method.LocalVariables); err != nil {
			return err
		}

		// Save assertions
		if err := d.saveAssertions(tx, methodID, method.Assertions); err != nil {
			return err
		}
	}

	return nil
}

// saveMethodParameters saves method parameters
func (d *Database) saveMethodParameters(tx *sql.Tx, methodID int64, params []models.Parameter) error {
	stmt, err := tx.Prepare(`
		INSERT INTO method_parameters (method_id, param_name, param_type, param_order, annotations)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, param := range params {
		annotationsJSON, _ := json.Marshal(param.Annotations)
		if _, err := stmt.Exec(methodID, param.Name, param.Type, i, string(annotationsJSON)); err != nil {
			return err
		}
	}

	return nil
}

// saveMethodAnnotations saves annotations on methods
func (d *Database) saveMethodAnnotations(tx *sql.Tx, methodID int64, annotations []models.Annotation) error {
	stmt, err := tx.Prepare(`
		INSERT INTO method_annotations (
			method_id, annotation_type, parameters, description, priority, enabled, line_number
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ann := range annotations {
		paramsJSON, _ := json.Marshal(ann.Parameters)
		if _, err := stmt.Exec(
			methodID, ann.Type, string(paramsJSON), ann.Description, ann.Priority, ann.Enabled, ann.LineNumber,
		); err != nil {
			return err
		}
	}

	return nil
}

// saveMethodCalls saves method calls
func (d *Database) saveMethodCalls(tx *sql.Tx, methodID int64, calls []models.MethodCall) error {
	stmt, err := tx.Prepare(`
		INSERT INTO method_calls (caller_method_id, method_name, object_name, arguments, chained_from, line_number)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, call := range calls {
		argsJSON, _ := json.Marshal(call.Arguments)
		if _, err := stmt.Exec(
			methodID, call.MethodName, call.ObjectName, string(argsJSON), call.ChainedFrom, call.LineNumber,
		); err != nil {
			return err
		}
	}

	return nil
}

// saveFieldAccesses saves field accesses
func (d *Database) saveFieldAccesses(tx *sql.Tx, methodID int64, fieldAccesses []string) error {
	stmt, err := tx.Prepare(`
		INSERT INTO field_access (method_id, field_name, access_type)
		VALUES (?, ?, 'read')
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, fieldName := range fieldAccesses {
		if _, err := stmt.Exec(methodID, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// saveLocalVariables saves local variables
func (d *Database) saveLocalVariables(tx *sql.Tx, methodID int64, vars []models.LocalVariable) error {
	stmt, err := tx.Prepare(`
		INSERT INTO local_variables (method_id, var_name, var_type, init_value, line_number)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range vars {
		if _, err := stmt.Exec(methodID, v.Name, v.Type, v.InitValue, v.LineNumber); err != nil {
			return err
		}
	}

	return nil
}

// saveAssertions saves assertions
func (d *Database) saveAssertions(tx *sql.Tx, methodID int64, assertions []models.Assertion) error {
	stmt, err := tx.Prepare(`
		INSERT INTO assertions (method_id, assertion_type, expected, actual, message, line_number)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, assertion := range assertions {
		if _, err := stmt.Exec(
			methodID, assertion.Type, assertion.Expected, assertion.Actual, assertion.Message, assertion.LineNumber,
		); err != nil {
			return err
		}
	}

	return nil
}

// Query methods

// GetClassByName retrieves a class by name
func (d *Database) GetClassByName(className string) (*models.ClassData, error) {
	var classData models.ClassData
	var modifiers string

	err := d.db.QueryRow(`
		SELECT c.class_name, c.fully_qualified_name, c.extends, c.modifiers,
		       c.is_page_object, c.is_test_class, c.line_start, c.line_end,
		       f.file_path, f.package
		FROM classes c
		JOIN files f ON c.file_id = f.id
		WHERE c.class_name = ?
	`, className).Scan(
		&classData.ClassName, &modifiers, &classData.Extends,
		&classData.PageObjectModel, &classData.TestClass,
		&classData.LineStart, &classData.LineEnd,
		&classData.FilePath, &classData.Package,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(modifiers), &classData.Modifiers)

	return &classData, nil
}

// GetWebElementFields retrieves all WebElement fields for a class
func (d *Database) GetWebElementFields(className string) ([]models.Field, error) {
	rows, err := d.db.Query(`
		SELECT f.field_name, f.field_type, f.locator_strategy, f.locator_value, f.line_number
		FROM fields f
		JOIN classes c ON f.class_id = c.id
		WHERE c.class_name = ? AND f.is_web_element = 1
	`, className)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []models.Field
	for rows.Next() {
		var field models.Field
		if err := rows.Scan(
			&field.Name, &field.Type, &field.LocatorStrategy, &field.LocatorValue, &field.LineNumber,
		); err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

// GetAllPageObjects retrieves all page object classes
func (d *Database) GetAllPageObjects() ([]string, error) {
	rows, err := d.db.Query("SELECT class_name FROM classes WHERE is_page_object = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pageObjects []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		pageObjects = append(pageObjects, name)
	}

	return pageObjects, nil
}
