package indexer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/parser"
)

// JavaIndexer indexes Java files
type JavaIndexer struct {
	db         *database.DB
	classRepo  *database.ClassRepository
	methodRepo *database.MethodRepository
	fieldRepo  *database.FieldRepository
	depRepo    *database.DependencyRepository
	parser     *parser.JavaParser
}

// NewJavaIndexer creates a new Java indexer
func NewJavaIndexer(db *database.DB) *JavaIndexer {
	return &JavaIndexer{
		db:         db,
		classRepo:  database.NewClassRepository(db),
		methodRepo: database.NewMethodRepository(db),
		fieldRepo:  database.NewFieldRepository(db),
		depRepo:    database.NewDependencyRepository(db),
		parser:     parser.NewJavaParser(),
	}
}

// IndexFile indexes a Java file
func (idx *JavaIndexer) IndexFile(file *database.File, content string) error {
	// Parse Java file
	parseResult, err := idx.parser.Parse(content, file.Path)
	if err != nil {
		return fmt.Errorf("failed to parse Java file: %w", err)
	}

	// Extract package name
	packageName := parseResult.PackageName
	if packageName != "" {
		file.PackageName = &packageName
	}

	// Use transaction for atomic operations
	return idx.db.WithTransaction(func(tx *sqlx.Tx) error {
		// Delete old data for this file (cascade will handle related records)
		idx.classRepo.DeleteByFileID(file.ID)

		// Index each class/interface/enum
		for _, cls := range parseResult.Classes {
			if err := idx.indexClass(tx, file, cls, packageName); err != nil {
				return fmt.Errorf("failed to index class %s: %w", cls.Name, err)
			}
		}

		return nil
	})
}

// indexClass indexes a single class/interface/enum
func (idx *JavaIndexer) indexClass(tx *sqlx.Tx, file *database.File, cls parser.Class, packageName string) error {
	// Determine fully qualified name
	fqn := cls.Name
	if packageName != "" {
		fqn = packageName + "." + cls.Name
	}

	// Convert annotations to JSON
	annotationsJSON, _ := json.Marshal(cls.Annotations)
	annotationsStr := string(annotationsJSON)

	// Convert implements to JSON
	implementsJSON, _ := json.Marshal(cls.Implements)
	implementsStr := string(implementsJSON)

	// Create class record
	class := &database.Class{
		FileID:               file.ID,
		Name:                 cls.Name,
		FullyQualifiedName:   fqn,
		Type:                 cls.Type, // "class", "interface", "enum"
		PackageName:          &packageName,
		IsAbstract:           cls.IsAbstract,
		IsPublic:             cls.IsPublic,
		ExtendsClass:         database.StringPtr(cls.Extends),
		ImplementsInterfaces: &implementsStr,
		Annotations:          &annotationsStr,
		Javadoc:              database.StringPtr(cls.Javadoc),
		StartLine:            cls.StartLine,
		EndLine:              cls.EndLine,
	}

	// Insert class
	if err := idx.classRepo.CreateTx(tx, class); err != nil {
		return err
	}

	// Index fields
	for _, field := range cls.Fields {
		if err := idx.indexField(tx, class, field); err != nil {
			return fmt.Errorf("failed to index field %s: %w", field.Name, err)
		}
	}

	// Index methods
	for _, method := range cls.Methods {
		if err := idx.indexMethod(tx, class, method); err != nil {
			return fmt.Errorf("failed to index method %s: %w", method.Name, err)
		}
	}

	return nil
}

// indexField indexes a class field
func (idx *JavaIndexer) indexField(tx *sqlx.Tx, class *database.Class, fld parser.Field) error {
	// Convert annotations to JSON
	annotationsJSON, _ := json.Marshal(fld.Annotations)
	annotationsStr := string(annotationsJSON)

	// Extract @FindBy locator information
	var locatorType, locatorValue *string
	for _, ann := range fld.Annotations {
		if ann.Name == "FindBy" {
			// Parse @FindBy(id = "username") or @FindBy(xpath = "//input[@id='username']")
			locatorType, locatorValue = idx.extractFindByLocator(ann)
			break
		}
	}

	field := &database.Field{
		ClassID:      class.ID,
		Name:         fld.Name,
		Type:         fld.Type,
		IsPublic:     fld.IsPublic,
		IsStatic:     fld.IsStatic,
		IsFinal:      fld.IsFinal,
		Annotations:  &annotationsStr,
		DefaultValue: database.StringPtr(fld.DefaultValue),
		LocatorType:  locatorType,
		LocatorValue: locatorValue,
		Javadoc:      database.StringPtr(fld.Javadoc),
		StartLine:    fld.StartLine,
	}

	return idx.fieldRepo.CreateTx(tx, field)
}

// indexMethod indexes a method
func (idx *JavaIndexer) indexMethod(tx *sqlx.Tx, class *database.Class, mth parser.Method) error {
	// Convert parameters to JSON
	paramsJSON, _ := json.Marshal(mth.Parameters)
	paramsStr := string(paramsJSON)

	// Convert annotations to JSON
	annotationsJSON, _ := json.Marshal(mth.Annotations)
	annotationsStr := string(annotationsJSON)

	method := &database.Method{
		ClassID:     class.ID,
		Name:        mth.Name,
		Signature:   mth.Signature,
		ReturnType:  database.StringPtr(mth.ReturnType),
		Parameters:  &paramsStr,
		IsPublic:    mth.IsPublic,
		IsStatic:    mth.IsStatic,
		IsAbstract:  mth.IsAbstract,
		Annotations: &annotationsStr,
		Javadoc:     database.StringPtr(mth.Javadoc),
		Body:        database.StringPtr(mth.Body),
		StartLine:   mth.StartLine,
		EndLine:     mth.EndLine,
	}

	if err := idx.methodRepo.CreateTx(tx, method); err != nil {
		return err
	}

	// Index step definitions if this is a Cucumber step
	for _, ann := range mth.Annotations {
		if idx.isStepAnnotation(ann.Name) {
			// This is a step definition
			// Extract pattern from annotation
			pattern := idx.extractStepPattern(ann)
			if pattern != "" {
				stepDef := &database.StepDefinition{
					MethodID: method.ID,
					Pattern:  pattern,
					Keyword:  database.StringPtr(idx.getStepKeyword(ann.Name)),
				}
				// Insert step definition
				// Note: We need a StepDefinitionRepository
				// For now, we'll skip this
				_ = stepDef
			}
		}
	}

	return nil
}

// BuildDependencyGraph builds the dependency graph for all classes
func (idx *JavaIndexer) BuildDependencyGraph() error {
	// Get all classes
	classes, err := idx.classRepo.GetAll()
	if err != nil {
		return err
	}

	classMap := make(map[string]*database.Class)
	for _, cls := range classes {
		classMap[cls.FullyQualifiedName] = cls
	}

	// Build dependencies
	for _, cls := range classes {
		// Extends dependency
		if cls.ExtendsClass != nil && *cls.ExtendsClass != "" {
			if parent, exists := classMap[*cls.ExtendsClass]; exists {
				dep := &database.Dependency{
					FromClassID:    cls.ID,
					ToClassID:      parent.ID,
					DependencyType: "extends",
				}
				idx.depRepo.Create(dep)
			}
		}

		// Implements dependencies
		if cls.ImplementsInterfaces != nil && *cls.ImplementsInterfaces != "" {
			var interfaces []string
			json.Unmarshal([]byte(*cls.ImplementsInterfaces), &interfaces)
			for _, iface := range interfaces {
				if parent, exists := classMap[iface]; exists {
					dep := &database.Dependency{
						FromClassID:    cls.ID,
						ToClassID:      parent.ID,
						DependencyType: "implements",
					}
					idx.depRepo.Create(dep)
				}
			}
		}

		// Field type dependencies
		fields, _ := idx.fieldRepo.GetByClassID(cls.ID)
		for _, field := range fields {
			// Check if field type is another class in our codebase
			fieldType := field.Type
			// Remove generics: List<String> -> List
			if idx := strings.Index(fieldType, "<"); idx >= 0 {
				fieldType = fieldType[:idx]
			}

			// Try to find the class
			for fqn, targetCls := range classMap {
				if strings.HasSuffix(fqn, "."+fieldType) || fqn == fieldType {
					dep := &database.Dependency{
						FromClassID:    cls.ID,
						ToClassID:      targetCls.ID,
						DependencyType: "uses",
						Context:        database.StringPtr("field:" + field.Name),
					}
					idx.depRepo.Create(dep)
					break
				}
			}
		}

		// Test -> Page Object dependencies
		// If this is a test class, find which page objects it uses
		if strings.HasSuffix(cls.Name, "Test") {
			for _, field := range fields {
				fieldType := field.Type
				if strings.HasSuffix(fieldType, "Page") {
					// This is likely a Page Object
					for fqn, targetCls := range classMap {
						if strings.HasSuffix(fqn, "."+fieldType) || fqn == fieldType {
							dep := &database.Dependency{
								FromClassID:    cls.ID,
								ToClassID:      targetCls.ID,
								DependencyType: "test_uses_page",
								Context:        database.StringPtr("field:" + field.Name),
							}
							idx.depRepo.Create(dep)
							break
						}
					}
				}
			}
		}
	}

	return nil
}

// Helper functions

func (idx *JavaIndexer) extractFindByLocator(ann parser.Annotation) (*string, *string) {
	// Parse annotation parameters
	// Example: @FindBy(id = "username") -> ("id", "username")
	// Example: @FindBy(xpath = "//input[@id='username']") -> ("xpath", "//input[@id='username']")

	for _, param := range ann.Parameters {
		// param format: "id = \"username\""
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			locType := strings.TrimSpace(parts[0])
			locValue := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			return &locType, &locValue
		}
	}

	return nil, nil
}

func (idx *JavaIndexer) isStepAnnotation(name string) bool {
	stepAnnotations := []string{"Given", "When", "Then", "And", "But"}
	for _, step := range stepAnnotations {
		if name == step {
			return true
		}
	}
	return false
}

func (idx *JavaIndexer) extractStepPattern(ann parser.Annotation) string {
	// Extract pattern from @Given("I am on login page")
	if len(ann.Parameters) > 0 {
		// Remove quotes
		pattern := strings.Trim(ann.Parameters[0], "\"")
		return pattern
	}
	return ""
}

func (idx *JavaIndexer) getStepKeyword(annName string) string {
	keywords := map[string]string{
		"Given": "Given",
		"When":  "When",
		"Then":  "Then",
		"And":   "And",
		"But":   "But",
	}
	return keywords[annName]
}
