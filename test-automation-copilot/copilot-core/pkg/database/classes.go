package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ClassRepository handles class-related database operations
type ClassRepository struct {
	db *DB
}

// NewClassRepository creates a new class repository
func NewClassRepository(db *DB) *ClassRepository {
	return &ClassRepository{db: db}
}

// Create inserts a new class
func (r *ClassRepository) Create(class *Class) error {
	result, err := r.db.conn.Exec(`
		INSERT INTO classes (
			file_id, name, fully_qualified_name, type, package_name,
			is_abstract, is_public, extends_class, implements_interfaces,
			annotations, javadoc, start_line, end_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		class.FileID,
		class.Name,
		class.FullyQualifiedName,
		class.Type,
		class.PackageName,
		class.IsAbstract,
		class.IsPublic,
		class.ExtendsClass,
		class.ImplementsInterfaces,
		class.Annotations,
		class.Javadoc,
		class.StartLine,
		class.EndLine,
	)

	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	class.ID = id
	return nil
}

// CreateTx inserts a new class within a transaction
func (r *ClassRepository) CreateTx(tx *sqlx.Tx, class *Class) error {
	result, err := tx.Exec(`
		INSERT INTO classes (
			file_id, name, fully_qualified_name, type, package_name,
			is_abstract, is_public, extends_class, implements_interfaces,
			annotations, javadoc, start_line, end_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		class.FileID,
		class.Name,
		class.FullyQualifiedName,
		class.Type,
		class.PackageName,
		class.IsAbstract,
		class.IsPublic,
		class.ExtendsClass,
		class.ImplementsInterfaces,
		class.Annotations,
		class.Javadoc,
		class.StartLine,
		class.EndLine,
	)

	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	class.ID = id
	return nil
}

// Update updates an existing class
func (r *ClassRepository) Update(class *Class) error {
	_, err := r.db.conn.Exec(`
		UPDATE classes
		SET name = ?,
		    fully_qualified_name = ?,
		    type = ?,
		    package_name = ?,
		    is_abstract = ?,
		    is_public = ?,
		    extends_class = ?,
		    implements_interfaces = ?,
		    annotations = ?,
		    javadoc = ?,
		    start_line = ?,
		    end_line = ?
		WHERE id = ?`,
		class.Name,
		class.FullyQualifiedName,
		class.Type,
		class.PackageName,
		class.IsAbstract,
		class.IsPublic,
		class.ExtendsClass,
		class.ImplementsInterfaces,
		class.Annotations,
		class.Javadoc,
		class.StartLine,
		class.EndLine,
		class.ID,
	)

	return err
}

// Delete removes a class
func (r *ClassRepository) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM classes WHERE id = ?", id)
	return err
}

// DeleteByFileID removes all classes in a file
func (r *ClassRepository) DeleteByFileID(fileID int64) error {
	_, err := r.db.conn.Exec("DELETE FROM classes WHERE file_id = ?", fileID)
	return err
}

// GetByID retrieves a class by ID
func (r *ClassRepository) GetByID(id int64) (*Class, error) {
	var class Class
	err := r.db.conn.Get(&class, "SELECT * FROM classes WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &class, nil
}

// GetByFQN retrieves a class by fully qualified name
func (r *ClassRepository) GetByFQN(fqn string) (*Class, error) {
	var class Class
	err := r.db.conn.Get(&class, "SELECT * FROM classes WHERE fully_qualified_name = ?", fqn)
	if err != nil {
		return nil, err
	}
	return &class, nil
}

// GetByFileID retrieves all classes in a file
func (r *ClassRepository) GetByFileID(fileID int64) ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, "SELECT * FROM classes WHERE file_id = ? ORDER BY start_line", fileID)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetByPackage retrieves classes by package name
func (r *ClassRepository) GetByPackage(packageName string) ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, "SELECT * FROM classes WHERE package_name = ? ORDER BY name", packageName)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetByType retrieves classes by type (class, interface, enum)
func (r *ClassRepository) GetByType(classType string) ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, "SELECT * FROM classes WHERE type = ? ORDER BY name", classType)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetAll retrieves all classes
func (r *ClassRepository) GetAll() ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, "SELECT * FROM classes ORDER BY fully_qualified_name")
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// Search searches for classes by name pattern
func (r *ClassRepository) Search(pattern string) ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes,
		"SELECT * FROM classes WHERE name LIKE ? OR fully_qualified_name LIKE ? ORDER BY name",
		"%"+pattern+"%", "%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetPageObjects retrieves all Page Object classes
// Page Objects typically extend BasePage or have "Page" suffix
func (r *ClassRepository) GetPageObjects() ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, `
		SELECT * FROM classes
		WHERE name LIKE '%Page'
		   OR extends_class LIKE '%BasePage'
		   OR extends_class LIKE '%PageObject'
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetTestClasses retrieves all test classes
// Test classes typically have @Test annotations or extend BaseTest
func (r *ClassRepository) GetTestClasses() ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, `
		SELECT DISTINCT c.* FROM classes c
		INNER JOIN methods m ON m.class_id = c.id
		WHERE m.annotations LIKE '%Test%'
		   OR c.extends_class LIKE '%BaseTest%'
		   OR c.name LIKE '%Test'
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetStepDefinitionClasses retrieves all step definition classes
func (r *ClassRepository) GetStepDefinitionClasses() ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, `
		SELECT DISTINCT c.* FROM classes c
		INNER JOIN methods m ON m.class_id = c.id
		INNER JOIN step_definitions sd ON sd.method_id = m.id
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// ExistsByFQN checks if a class exists by fully qualified name
func (r *ClassRepository) ExistsByFQN(fqn string) (bool, error) {
	var count int
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM classes WHERE fully_qualified_name = ?", fqn)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count returns total number of classes
func (r *ClassRepository) Count() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM classes")
	return count, err
}

// CountByType returns number of classes by type
func (r *ClassRepository) CountByType(classType string) (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM classes WHERE type = ?", classType)
	return count, err
}

// ClassWithFile combines class and file information
type ClassWithFile struct {
	Class
	FilePath string `db:"file_path"`
}

// GetClassesWithFiles retrieves classes with their file paths
func (r *ClassRepository) GetClassesWithFiles() ([]*ClassWithFile, error) {
	var results []*ClassWithFile
	err := r.db.conn.Select(&results, `
		SELECT c.*, f.path as file_path
		FROM classes c
		INNER JOIN files f ON f.id = c.file_id
		ORDER BY c.fully_qualified_name
	`)
	if err != nil {
		return nil, err
	}
	return results, nil
}
