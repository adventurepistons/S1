package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// MethodRepository handles method-related database operations
type MethodRepository struct {
	db *DB
}

// NewMethodRepository creates a new method repository
func NewMethodRepository(db *DB) *MethodRepository {
	return &MethodRepository{db: db}
}

// Create inserts a new method
func (r *MethodRepository) Create(method *Method) error {
	result, err := r.db.conn.Exec(`
		INSERT INTO methods (
			class_id, name, signature, return_type, parameters,
			is_public, is_static, is_abstract, annotations,
			javadoc, body, start_line, end_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		method.ClassID,
		method.Name,
		method.Signature,
		method.ReturnType,
		method.Parameters,
		method.IsPublic,
		method.IsStatic,
		method.IsAbstract,
		method.Annotations,
		method.Javadoc,
		method.Body,
		method.StartLine,
		method.EndLine,
	)

	if err != nil {
		return fmt.Errorf("failed to create method: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	method.ID = id
	return nil
}

// CreateTx inserts a new method within a transaction
func (r *MethodRepository) CreateTx(tx *sqlx.Tx, method *Method) error {
	result, err := tx.Exec(`
		INSERT INTO methods (
			class_id, name, signature, return_type, parameters,
			is_public, is_static, is_abstract, annotations,
			javadoc, body, start_line, end_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		method.ClassID,
		method.Name,
		method.Signature,
		method.ReturnType,
		method.Parameters,
		method.IsPublic,
		method.IsStatic,
		method.IsAbstract,
		method.Annotations,
		method.Javadoc,
		method.Body,
		method.StartLine,
		method.EndLine,
	)

	if err != nil {
		return fmt.Errorf("failed to create method: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	method.ID = id
	return nil
}

// Update updates an existing method
func (r *MethodRepository) Update(method *Method) error {
	_, err := r.db.conn.Exec(`
		UPDATE methods
		SET name = ?,
		    signature = ?,
		    return_type = ?,
		    parameters = ?,
		    is_public = ?,
		    is_static = ?,
		    is_abstract = ?,
		    annotations = ?,
		    javadoc = ?,
		    body = ?,
		    start_line = ?,
		    end_line = ?
		WHERE id = ?`,
		method.Name,
		method.Signature,
		method.ReturnType,
		method.Parameters,
		method.IsPublic,
		method.IsStatic,
		method.IsAbstract,
		method.Annotations,
		method.Javadoc,
		method.Body,
		method.StartLine,
		method.EndLine,
		method.ID,
	)

	return err
}

// Delete removes a method
func (r *MethodRepository) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM methods WHERE id = ?", id)
	return err
}

// DeleteByClassID removes all methods in a class
func (r *MethodRepository) DeleteByClassID(classID int64) error {
	_, err := r.db.conn.Exec("DELETE FROM methods WHERE class_id = ?", classID)
	return err
}

// GetByID retrieves a method by ID
func (r *MethodRepository) GetByID(id int64) (*Method, error) {
	var method Method
	err := r.db.conn.Get(&method, "SELECT * FROM methods WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &method, nil
}

// GetByClassID retrieves all methods in a class
func (r *MethodRepository) GetByClassID(classID int64) ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods, "SELECT * FROM methods WHERE class_id = ? ORDER BY start_line", classID)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetByName searches for methods by name pattern
func (r *MethodRepository) GetByName(pattern string) ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods,
		"SELECT * FROM methods WHERE name LIKE ? ORDER BY name",
		"%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetTestMethods retrieves all methods with @Test annotation
func (r *MethodRepository) GetTestMethods() ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods, `
		SELECT * FROM methods
		WHERE annotations LIKE '%@Test%'
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetTestMethodsByClass retrieves test methods for a specific class
func (r *MethodRepository) GetTestMethodsByClass(classID int64) ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods, `
		SELECT * FROM methods
		WHERE class_id = ?
		  AND annotations LIKE '%@Test%'
		ORDER BY start_line
	`, classID)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetDataProviderMethods retrieves methods with @DataProvider annotation
func (r *MethodRepository) GetDataProviderMethods() ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods, `
		SELECT * FROM methods
		WHERE annotations LIKE '%@DataProvider%'
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetStepDefinitionMethods retrieves methods that are step definitions
func (r *MethodRepository) GetStepDefinitionMethods() ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods, `
		SELECT DISTINCT m.* FROM methods m
		INNER JOIN step_definitions sd ON sd.method_id = m.id
		ORDER BY m.name
	`)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// Count returns total number of methods
func (r *MethodRepository) Count() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM methods")
	return count, err
}

// MethodWithClass combines method and class information
type MethodWithClass struct {
	Method
	ClassName            string  `db:"class_name"`
	FullyQualifiedName   string  `db:"fully_qualified_name"`
	FilePath             string  `db:"file_path"`
}

// GetMethodsWithContext retrieves methods with class and file context
func (r *MethodRepository) GetMethodsWithContext() ([]*MethodWithClass, error) {
	var results []*MethodWithClass
	err := r.db.conn.Select(&results, `
		SELECT
			m.*,
			c.name as class_name,
			c.fully_qualified_name,
			f.path as file_path
		FROM methods m
		INNER JOIN classes c ON c.id = m.class_id
		INNER JOIN files f ON f.id = c.file_id
		ORDER BY m.name
	`)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// SearchBySignature searches for methods by signature pattern
func (r *MethodRepository) SearchBySignature(pattern string) ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods,
		"SELECT * FROM methods WHERE signature LIKE ? ORDER BY name",
		"%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetPublicMethods retrieves all public methods in a class
func (r *MethodRepository) GetPublicMethods(classID int64) ([]*Method, error) {
	var methods []*Method
	err := r.db.conn.Select(&methods, `
		SELECT * FROM methods
		WHERE class_id = ? AND is_public = 1
		ORDER BY start_line
	`, classID)
	if err != nil {
		return nil, err
	}
	return methods, nil
}
