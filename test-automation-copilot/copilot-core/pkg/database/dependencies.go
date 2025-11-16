package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DependencyRepository handles dependency-related database operations
type DependencyRepository struct {
	db *DB
}

// NewDependencyRepository creates a new dependency repository
func NewDependencyRepository(db *DB) *DependencyRepository {
	return &DependencyRepository{db: db}
}

// Create inserts a new dependency
func (r *DependencyRepository) Create(dep *Dependency) error {
	result, err := r.db.conn.Exec(`
		INSERT OR IGNORE INTO dependencies (
			from_class_id, to_class_id, dependency_type, context
		) VALUES (?, ?, ?, ?)`,
		dep.FromClassID,
		dep.ToClassID,
		dep.DependencyType,
		dep.Context,
	)

	if err != nil {
		return fmt.Errorf("failed to create dependency: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	dep.ID = id
	return nil
}

// CreateTx inserts a new dependency within a transaction
func (r *DependencyRepository) CreateTx(tx *sqlx.Tx, dep *Dependency) error {
	result, err := tx.Exec(`
		INSERT OR IGNORE INTO dependencies (
			from_class_id, to_class_id, dependency_type, context
		) VALUES (?, ?, ?, ?)`,
		dep.FromClassID,
		dep.ToClassID,
		dep.DependencyType,
		dep.Context,
	)

	if err != nil {
		return fmt.Errorf("failed to create dependency: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	dep.ID = id
	return nil
}

// Delete removes a dependency
func (r *DependencyRepository) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM dependencies WHERE id = ?", id)
	return err
}

// DeleteByClass removes all dependencies involving a class
func (r *DependencyRepository) DeleteByClass(classID int64) error {
	_, err := r.db.conn.Exec(
		"DELETE FROM dependencies WHERE from_class_id = ? OR to_class_id = ?",
		classID, classID)
	return err
}

// GetByID retrieves a dependency by ID
func (r *DependencyRepository) GetByID(id int64) (*Dependency, error) {
	var dep Dependency
	err := r.db.conn.Get(&dep, "SELECT * FROM dependencies WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// GetDependenciesFrom retrieves all dependencies from a class
func (r *DependencyRepository) GetDependenciesFrom(classID int64) ([]*Dependency, error) {
	var deps []*Dependency
	err := r.db.conn.Select(&deps,
		"SELECT * FROM dependencies WHERE from_class_id = ?",
		classID)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

// GetDependenciesTo retrieves all dependencies to a class
func (r *DependencyRepository) GetDependenciesTo(classID int64) ([]*Dependency, error) {
	var deps []*Dependency
	err := r.db.conn.Select(&deps,
		"SELECT * FROM dependencies WHERE to_class_id = ?",
		classID)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

// GetByType retrieves dependencies by type
func (r *DependencyRepository) GetByType(depType string) ([]*Dependency, error) {
	var deps []*Dependency
	err := r.db.conn.Select(&deps,
		"SELECT * FROM dependencies WHERE dependency_type = ?",
		depType)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

// DependencyWithClasses combines dependency with class names
type DependencyWithClasses struct {
	Dependency
	FromClassName string `db:"from_class_name"`
	ToClassName   string `db:"to_class_name"`
	FromFQN       string `db:"from_fqn"`
	ToFQN         string `db:"to_fqn"`
}

// GetDependenciesWithClasses retrieves dependencies with class information
func (r *DependencyRepository) GetDependenciesWithClasses() ([]*DependencyWithClasses, error) {
	var results []*DependencyWithClasses
	err := r.db.conn.Select(&results, `
		SELECT
			d.*,
			c1.name as from_class_name,
			c2.name as to_class_name,
			c1.fully_qualified_name as from_fqn,
			c2.fully_qualified_name as to_fqn
		FROM dependencies d
		INNER JOIN classes c1 ON c1.id = d.from_class_id
		INNER JOIN classes c2 ON c2.id = d.to_class_id
		ORDER BY d.dependency_type, c1.name
	`)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetTestPageDependencies retrieves test -> page object dependencies
func (r *DependencyRepository) GetTestPageDependencies() ([]*DependencyWithClasses, error) {
	var results []*DependencyWithClasses
	err := r.db.conn.Select(&results, `
		SELECT
			d.*,
			c1.name as from_class_name,
			c2.name as to_class_name,
			c1.fully_qualified_name as from_fqn,
			c2.fully_qualified_name as to_fqn
		FROM dependencies d
		INNER JOIN classes c1 ON c1.id = d.from_class_id
		INNER JOIN classes c2 ON c2.id = d.to_class_id
		WHERE d.dependency_type = 'test_uses_page'
		ORDER BY c1.name
	`)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetRelatedClasses retrieves all classes related to a given class
// (both dependencies and dependents)
func (r *DependencyRepository) GetRelatedClasses(classID int64) ([]*Class, error) {
	var classes []*Class
	err := r.db.conn.Select(&classes, `
		SELECT DISTINCT c.* FROM classes c
		WHERE c.id IN (
			SELECT to_class_id FROM dependencies WHERE from_class_id = ?
			UNION
			SELECT from_class_id FROM dependencies WHERE to_class_id = ?
		)
		ORDER BY c.name
	`, classID, classID)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

// GetDependencyGraph builds a full dependency graph for a set of classes
type DependencyGraph struct {
	Nodes []*Class                 `json:"nodes"`
	Edges []*DependencyWithClasses `json:"edges"`
}

// GetFullGraph retrieves the complete dependency graph
func (r *DependencyRepository) GetFullGraph() (*DependencyGraph, error) {
	// Get all classes
	classRepo := NewClassRepository(r.db)
	classes, err := classRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Get all dependencies with class info
	edges, err := r.GetDependenciesWithClasses()
	if err != nil {
		return nil, err
	}

	return &DependencyGraph{
		Nodes: classes,
		Edges: edges,
	}, nil
}

// Count returns total number of dependencies
func (r *DependencyRepository) Count() (int64, error) {
	var count int64
	err := r.db.conn.Get(&count, "SELECT COUNT(*) FROM dependencies")
	return count, err
}

// CountByType returns number of dependencies by type
func (r *DependencyRepository) CountByType(depType string) (int64, error) {
	var count int64
	err := r.db.conn.Get(&count,
		"SELECT COUNT(*) FROM dependencies WHERE dependency_type = ?",
		depType)
	return count, err
}

// Exists checks if a dependency exists
func (r *DependencyRepository) Exists(fromClassID, toClassID int64, depType string) (bool, error) {
	var count int
	err := r.db.conn.Get(&count, `
		SELECT COUNT(*) FROM dependencies
		WHERE from_class_id = ? AND to_class_id = ? AND dependency_type = ?
	`, fromClassID, toClassID, depType)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
