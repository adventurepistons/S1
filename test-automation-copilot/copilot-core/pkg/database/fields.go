package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// FieldRepository handles field-related database operations
type FieldRepository struct {
	db *DB
}

// NewFieldRepository creates a new field repository
func NewFieldRepository(db *DB) *FieldRepository {
	return &FieldRepository{db: db}
}

// Create inserts a new field
func (r *FieldRepository) Create(field *Field) error {
	result, err := r.db.conn.Exec(`
		INSERT INTO fields (
			class_id, name, type, is_public, is_static, is_final,
			annotations, default_value, locator_type, locator_value,
			javadoc, start_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		field.ClassID,
		field.Name,
		field.Type,
		field.IsPublic,
		field.IsStatic,
		field.IsFinal,
		field.Annotations,
		field.DefaultValue,
		field.LocatorType,
		field.LocatorValue,
		field.Javadoc,
		field.StartLine,
	)

	if err != nil {
		return fmt.Errorf("failed to create field: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	field.ID = id
	return nil
}

// CreateTx inserts a new field within a transaction
func (r *FieldRepository) CreateTx(tx *sqlx.Tx, field *Field) error {
	result, err := tx.Exec(`
		INSERT INTO fields (
			class_id, name, type, is_public, is_static, is_final,
			annotations, default_value, locator_type, locator_value,
			javadoc, start_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		field.ClassID,
		field.Name,
		field.Type,
		field.IsPublic,
		field.IsStatic,
		field.IsFinal,
		field.Annotations,
		field.DefaultValue,
		field.LocatorType,
		field.LocatorValue,
		field.Javadoc,
		field.StartLine,
	)

	if err != nil {
		return fmt.Errorf("failed to create field: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	field.ID = id
	return nil
}

// Update updates an existing field
func (r *FieldRepository) Update(field *Field) error {
	_, err := r.db.conn.Exec(`
		UPDATE fields
		SET name = ?,
		    type = ?,
		    is_public = ?,
		    is_static = ?,
		    is_final = ?,
		    annotations = ?,
		    default_value = ?,
		    locator_type = ?,
		    locator_value = ?,
		    javadoc = ?,
		    start_line = ?
		WHERE id = ?`,
		field.Name,
		field.Type,
		field.IsPublic,
		field.IsStatic,
		field.IsFinal,
		field.Annotations,
		field.DefaultValue,
		field.LocatorType,
		field.LocatorValue,
		field.Javadoc,
		field.StartLine,
		field.ID,
	)

	return err
}

// Delete removes a field
func (r *FieldRepository) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM fields WHERE id = ?", id)
	return err
}

// DeleteByClassID removes all fields in a class
func (r *FieldRepository) DeleteByClassID(classID int64) error {
	_, err := r.db.conn.Exec("DELETE FROM fields WHERE class_id = ?", classID)
	return err
}

// GetByID retrieves a field by ID
func (r *FieldRepository) GetByID(id int64) (*Field, error) {
	var field Field
	err := r.db.conn.Get(&field, "SELECT * FROM fields WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &field, nil
}

// GetByClassID retrieves all fields in a class
func (r *FieldRepository) GetByClassID(classID int64) ([]*Field, error) {
	var fields []*Field
	err := r.db.conn.Select(&fields, "SELECT * FROM fields WHERE class_id = ? ORDER BY start_line", classID)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// GetByName searches for fields by name pattern
func (r *FieldRepository) GetByName(pattern string) ([]*Field, error) {
	var fields []*Field
	err := r.db.conn.Select(&fields,
		"SELECT * FROM fields WHERE name LIKE ? ORDER BY name",
		"%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// GetWebElements retrieves all fields with @FindBy annotations (WebElements)
func (r *FieldRepository) GetWebElements() ([]*Field, error) {
	var fields []*Field
	err := r.db.conn.Select(&fields, `
		SELECT * FROM fields
		WHERE annotations LIKE '%@FindBy%'
		  AND locator_type IS NOT NULL
		ORDER BY class_id, start_line
	`)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// GetWebElementsByClass retrieves WebElements for a specific class
func (r *FieldRepository) GetWebElementsByClass(classID int64) ([]*Field, error) {
	var fields []*Field
	err := r.db.conn.Select(&fields, `
		SELECT * FROM fields
		WHERE class_id = ?
		  AND annotations LIKE '%@FindBy%'
		  AND locator_type IS NOT NULL
		ORDER BY start_line
	`, classID)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// Count returns total number of fields
func (r *FieldRepository) Count() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM fields")
	return count, err
}

// CountByType returns number of fields by type
func (r *FieldRepository) CountByType(fieldType string) (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM fields WHERE type = ?", fieldType)
	return count, err
}

// FieldWithClass combines field and class information
type FieldWithClass struct {
	Field
	ClassName          string `db:"class_name"`
	FullyQualifiedName string `db:"fully_qualified_name"`
	FilePath           string `db:"file_path"`
}

// GetFieldsWithContext retrieves fields with class and file context
func (r *FieldRepository) GetFieldsWithContext() ([]*FieldWithClass, error) {
	var results []*FieldWithClass
	err := r.db.conn.Select(&results, `
		SELECT
			f.*,
			c.name as class_name,
			c.fully_qualified_name,
			fi.path as file_path
		FROM fields f
		INNER JOIN classes c ON c.id = f.class_id
		INNER JOIN files fi ON fi.id = c.file_id
		ORDER BY f.name
	`)
	if err != nil {
		return nil, err
	}
	return results, nil
}
