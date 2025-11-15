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
	parsedClass, err := idx.parser.ParseFile(file.Path)
	if err != nil {
		return fmt.Errorf("failed to parse Java file: %w", err)
	}

	// Extract package name
	packageName := parsedClass.Package
	if packageName != "" {
		file.PackageName = &packageName
	}

	// Use transaction for atomic operations
	return idx.db.WithTransaction(func(tx *sqlx.Tx) error {
		// Delete old data for this file (cascade will handle related records)
		idx.classRepo.DeleteByFileID(file.ID)

		// Index the class
		if err := idx.indexClass(tx, file, parsedClass); err != nil {
			return fmt.Errorf("failed to index class %s: %w", parsedClass.ClassName, err)
		}

		return nil
	})
}

// indexClass indexes a single class/interface/enum
func (idx *JavaIndexer) indexClass(tx *sqlx.Tx, file *database.File, parsedClass *parser.ParsedClass) error {
	// Determine fully qualified name
	fqn := parsedClass.ClassName
	if parsedClass.Package != "" {
		fqn = parsedClass.Package + "." + parsedClass.ClassName
	}

	// Convert annotations to JSON
	annotationsJSON, _ := json.Marshal(parsedClass.Annotations)
	annotationsStr := string(annotationsJSON)

	// Convert implements to JSON (empty for now, as ParsedClass doesn't have this)
	implementsStr := "[]"

	// Determine class type
	classType := determineClassType(parsedClass.Type)

	// Create class record
	class := &database.Class{
		FileID:               file.ID,
		Name:                 parsedClass.ClassName,
		FullyQualifiedName:   fqn,
		Type:                 classType,
		PackageName:          database.StringPtr(parsedClass.Package),
		IsAbstract:           false, // Not provided by parser
		IsPublic:             true,  // Assume public
		ExtendsClass:         database.StringPtr(parsedClass.SuperClass),
		ImplementsInterfaces: &implementsStr,
		Annotations:          &annotationsStr,
		Javadoc:              nil,
		StartLine:            1, // Not provided by parser
		EndLine:              100, // Not provided by parser
	}

	// Insert class
	if err := idx.classRepo.CreateTx(tx, class); err != nil {
		return err
	}

	// Index fields
	for _, field := range parsedClass.Fields {
		if err := idx.indexField(tx, class, field); err != nil {
			return fmt.Errorf("failed to index field %s: %w", field.Name, err)
		}
	}

	// Index methods
	for _, method := range parsedClass.Methods {
		if err := idx.indexMethod(tx, class, method); err != nil {
			return fmt.Errorf("failed to index method %s: %w", method.Name, err)
		}
	}

	return nil
}

// indexField indexes a class field
func (idx *JavaIndexer) indexField(tx *sqlx.Tx, class *database.Class, fld parser.FieldInfo) error {
	// Convert annotations to JSON
	annotationsJSON, _ := json.Marshal(fld.Annotations)
	annotationsStr := string(annotationsJSON)

	// Use locator info from parser
	var locatorType, locatorValue *string
	if fld.LocatorType != "" {
		locatorType = &fld.LocatorType
		locatorValue = &fld.LocatorValue
	}

	field := &database.Field{
		ClassID:      class.ID,
		Name:         fld.Name,
		Type:         fld.Type,
		IsPublic:     false, // Not provided by parser
		IsStatic:     false, // Not provided by parser
		IsFinal:      false, // Not provided by parser
		Annotations:  &annotationsStr,
		DefaultValue: nil,
		LocatorType:  locatorType,
		LocatorValue: locatorValue,
		Javadoc:      nil,
		StartLine:    1, // Not provided by parser
	}

	return idx.fieldRepo.CreateTx(tx, field)
}

// indexMethod indexes a method
func (idx *JavaIndexer) indexMethod(tx *sqlx.Tx, class *database.Class, mth parser.MethodInfo) error {
	// Convert parameters to JSON
	paramsJSON, _ := json.Marshal(mth.Parameters)
	paramsStr := string(paramsJSON)

	// Convert annotations to JSON
	annotationsJSON, _ := json.Marshal(mth.Annotations)
	annotationsStr := string(annotationsJSON)

	// Build signature
	signature := buildSignature(mth)

	method := &database.Method{
		ClassID:     class.ID,
		Name:        mth.Name,
		Signature:   signature,
		ReturnType:  database.StringPtr(mth.ReturnType),
		Parameters:  &paramsStr,
		IsPublic:    true, // Assume public
		IsStatic:    false,
		IsAbstract:  false,
		Annotations: &annotationsStr,
		Javadoc:     nil,
		Body:        database.StringPtr(mth.Body),
		StartLine:   1,  // Not provided by parser
		EndLine:     10, // Not provided by parser
	}

	if err := idx.methodRepo.CreateTx(tx, method); err != nil {
		return err
	}

	// Index step definitions if this is a Cucumber step
	for _, ann := range mth.Annotations {
		if idx.isStepAnnotation(ann) {
			// This is a step definition
			// Extract pattern from annotation
			pattern := idx.extractStepPattern(ann)
			if pattern != "" {
				stepDef := &database.StepDefinition{
					MethodID: method.ID,
					Pattern:  pattern,
					Keyword:  database.StringPtr(idx.getStepKeyword(ann)),
				}
				// Insert step definition
				_, err := tx.Exec(`
					INSERT INTO step_definitions (method_id, pattern, keyword)
					VALUES (?, ?, ?)`,
					stepDef.MethodID,
					stepDef.Pattern,
					stepDef.Keyword,
				)
				if err != nil {
					return err
				}
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

func determineClassType(parserType string) string {
	// Convert parser type to database type
	// Parser types: "pageObject", "test", "step", "utility", "unknown"
	switch parserType {
	case "pageObject":
		return "class"
	case "test":
		return "class"
	case "step":
		return "class"
	default:
		return "class"
	}
}

func buildSignature(mth parser.MethodInfo) string {
	params := make([]string, len(mth.Parameters))
	for i, p := range mth.Parameters {
		params[i] = p.Type + " " + p.Name
	}
	return fmt.Sprintf("%s(%s)", mth.Name, strings.Join(params, ", "))
}

func (idx *JavaIndexer) isStepAnnotation(annotation string) bool {
	stepAnnotations := []string{"@Given", "@When", "@Then", "@And", "@But"}
	for _, step := range stepAnnotations {
		if strings.HasPrefix(annotation, step) {
			return true
		}
	}
	return false
}

func (idx *JavaIndexer) extractStepPattern(annotation string) string {
	// Extract pattern from @Given("I am on login page")
	start := strings.Index(annotation, "(\"")
	if start == -1 {
		return ""
	}
	start += 2
	end := strings.Index(annotation[start:], "\"")
	if end == -1 {
		return ""
	}
	return annotation[start : start+end]
}

func (idx *JavaIndexer) getStepKeyword(annotation string) string {
	if strings.HasPrefix(annotation, "@Given") {
		return "Given"
	} else if strings.HasPrefix(annotation, "@When") {
		return "When"
	} else if strings.HasPrefix(annotation, "@Then") {
		return "Then"
	} else if strings.HasPrefix(annotation, "@And") {
		return "And"
	} else if strings.HasPrefix(annotation, "@But") {
		return "But"
	}
	return ""
}
