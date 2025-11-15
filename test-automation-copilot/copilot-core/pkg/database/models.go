package database

import "time"

// ============================================================================
// FILE MODELS
// ============================================================================

// File represents a source file in the workspace
type File struct {
	ID           int64     `db:"id"`
	Path         string    `db:"path"`
	Type         string    `db:"type"` // java, gherkin, xml, properties
	ContentHash  string    `db:"content_hash"`
	Content      string    `db:"content"`
	PackageName  *string   `db:"package_name"`
	LastModified int64     `db:"last_modified"`
	IndexedAt    int64     `db:"indexed_at"`
	SizeBytes    int64     `db:"size_bytes"`
	LineCount    int       `db:"line_count"`
	CreatedAt    int64     `db:"created_at"`
}

// ============================================================================
// CLASS MODELS
// ============================================================================

// Class represents a Java class, interface, enum, or annotation
type Class struct {
	ID                   int64   `db:"id"`
	FileID               int64   `db:"file_id"`
	Name                 string  `db:"name"`
	FullyQualifiedName   string  `db:"fully_qualified_name"`
	Type                 string  `db:"type"` // class, interface, enum, annotation
	PackageName          *string `db:"package_name"`
	IsAbstract           bool    `db:"is_abstract"`
	IsPublic             bool    `db:"is_public"`
	ExtendsClass         *string `db:"extends_class"`
	ImplementsInterfaces *string `db:"implements_interfaces"` // JSON array
	Annotations          *string `db:"annotations"`           // JSON array
	Javadoc              *string `db:"javadoc"`
	StartLine            int     `db:"start_line"`
	EndLine              int     `db:"end_line"`
	CreatedAt            int64   `db:"created_at"`
}

// ============================================================================
// METHOD MODELS
// ============================================================================

// Method represents a method in a Java class
type Method struct {
	ID          int64   `db:"id"`
	ClassID     int64   `db:"class_id"`
	Name        string  `db:"name"`
	Signature   string  `db:"signature"`
	ReturnType  *string `db:"return_type"`
	Parameters  *string `db:"parameters"`  // JSON array
	IsPublic    bool    `db:"is_public"`
	IsStatic    bool    `db:"is_static"`
	IsAbstract  bool    `db:"is_abstract"`
	Annotations *string `db:"annotations"` // JSON array
	Javadoc     *string `db:"javadoc"`
	Body        *string `db:"body"`
	StartLine   int     `db:"start_line"`
	EndLine     int     `db:"end_line"`
	CreatedAt   int64   `db:"created_at"`
}

// ============================================================================
// FIELD MODELS
// ============================================================================

// Field represents a class field/variable (including WebElements)
type Field struct {
	ID           int64   `db:"id"`
	ClassID      int64   `db:"class_id"`
	Name         string  `db:"name"`
	Type         string  `db:"type"`
	IsPublic     bool    `db:"is_public"`
	IsStatic     bool    `db:"is_static"`
	IsFinal      bool    `db:"is_final"`
	Annotations  *string `db:"annotations"`   // JSON array
	DefaultValue *string `db:"default_value"`
	LocatorType  *string `db:"locator_type"`  // For @FindBy annotations
	LocatorValue *string `db:"locator_value"` // Actual locator
	Javadoc      *string `db:"javadoc"`
	StartLine    int     `db:"start_line"`
	CreatedAt    int64   `db:"created_at"`
}

// ============================================================================
// DEPENDENCY MODELS
// ============================================================================

// Dependency represents relationships between classes
type Dependency struct {
	ID             int64   `db:"id"`
	FromClassID    int64   `db:"from_class_id"`
	ToClassID      int64   `db:"to_class_id"`
	DependencyType string  `db:"dependency_type"` // extends, implements, uses, imports, test_uses_page
	Context        *string `db:"context"`
	CreatedAt      int64   `db:"created_at"`
}

// ============================================================================
// BDD/GHERKIN MODELS
// ============================================================================

// FeatureFile represents a Gherkin feature file
type FeatureFile struct {
	ID          int64   `db:"id"`
	FileID      int64   `db:"file_id"`
	FeatureName string  `db:"feature_name"`
	Description *string `db:"description"`
	Tags        *string `db:"tags"` // JSON array
	Language    string  `db:"language"`
	CreatedAt   int64   `db:"created_at"`
}

// Scenario represents a Gherkin scenario
type Scenario struct {
	ID          int64   `db:"id"`
	FeatureID   int64   `db:"feature_id"`
	Name        string  `db:"name"`
	Type        string  `db:"type"` // scenario, scenario_outline
	Description *string `db:"description"`
	Tags        *string `db:"tags"` // JSON array
	StartLine   int     `db:"start_line"`
	EndLine     int     `db:"end_line"`
	CreatedAt   int64   `db:"created_at"`
}

// Step represents a Gherkin step
type Step struct {
	ID         int64   `db:"id"`
	ScenarioID int64   `db:"scenario_id"`
	Keyword    string  `db:"keyword"` // Given, When, Then, And, But
	Text       string  `db:"text"`
	Argument   *string `db:"argument"` // DataTable or DocString
	LineNumber int     `db:"line_number"`
	CreatedAt  int64   `db:"created_at"`
}

// StepDefinition represents a Java step definition method
type StepDefinition struct {
	ID        int64   `db:"id"`
	MethodID  int64   `db:"method_id"`
	Pattern   string  `db:"pattern"` // Regex pattern
	Keyword   *string `db:"keyword"` // Given, When, Then or null
	CreatedAt int64   `db:"created_at"`
}

// ============================================================================
// CHAT MODELS
// ============================================================================

// ChatSession represents a chat conversation
type ChatSession struct {
	ID           string `db:"id"` // UUID
	WorkspacePath string `db:"workspace_path"`
	StartedAt    int64  `db:"started_at"`
	EndedAt      *int64 `db:"ended_at"`
	MessageCount int    `db:"message_count"`
	CreatedAt    int64  `db:"created_at"`
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	ID           int64   `db:"id"`
	SessionID    string  `db:"session_id"`
	Role         string  `db:"role"` // user, assistant, system
	Message      string  `db:"message"`
	Timestamp    int64   `db:"timestamp"`
	ContextFiles *string `db:"context_files"` // JSON array
	CreatedAt    int64   `db:"created_at"`
}

// ============================================================================
// FILE CHANGE MODELS
// ============================================================================

// FileChange represents a code modification
type FileChange struct {
	ID            int64   `db:"id"`
	SessionID     string  `db:"session_id"`
	FilePath      string  `db:"file_path"`
	ChangeType    string  `db:"change_type"` // create, modify, delete
	BeforeContent *string `db:"before_content"`
	AfterContent  *string `db:"after_content"`
	Diff          *string `db:"diff"` // Unified diff format
	Applied       bool    `db:"applied"`
	Timestamp     int64   `db:"timestamp"`
	UserApproved  bool    `db:"user_approved"`
	CreatedAt     int64   `db:"created_at"`
}

// ============================================================================
// CONFIG MODELS
// ============================================================================

// Config represents workspace configuration
type Config struct {
	ID                int64   `db:"id"`
	WorkspacePath     string  `db:"workspace_path"`
	Framework         *string `db:"framework"`           // selenium-java, playwright-java
	TestFramework     *string `db:"test_framework"`      // testng, junit
	BDDEnabled        bool    `db:"bdd_enabled"`
	BasePackage       *string `db:"base_package"`
	PageObjectPackage *string `db:"page_object_package"`
	TestPackage       *string `db:"test_package"`
	ConfigJSON        *string `db:"config_json"` // Additional config
	CreatedAt         int64   `db:"created_at"`
	UpdatedAt         int64   `db:"updated_at"`
}

// ============================================================================
// EMBEDDING MODELS
// ============================================================================

// Embedding tracks what has been indexed in chromem-go
type Embedding struct {
	ID          int64  `db:"id"`
	EntityType  string `db:"entity_type"` // class, method, field, file
	EntityID    int64  `db:"entity_id"`
	EmbeddingID string `db:"embedding_id"` // ID in chromem-go
	IndexedAt   int64  `db:"indexed_at"`
	CreatedAt   int64  `db:"created_at"`
}

// ============================================================================
// RECORDING MODELS
// ============================================================================

// RecordingSession represents a browser recording session
type RecordingSession struct {
	ID               string  `db:"id"` // session_xxx
	WorkspacePath    string  `db:"workspace_path"`
	StartURL         *string `db:"start_url"`
	StartedAt        int64   `db:"started_at"`
	EndedAt          *int64  `db:"ended_at"`
	PageCount        int     `db:"page_count"`
	InteractionCount int     `db:"interaction_count"`
	RecordingData    *string `db:"recording_data"` // JSON
	CreatedAt        int64   `db:"created_at"`
}

// GeneratedCode represents generated code from recordings
type GeneratedCode struct {
	ID                 int64   `db:"id"`
	RecordingSessionID *string `db:"recording_session_id"`
	FilePath           string  `db:"file_path"`
	CodeType           string  `db:"code_type"` // page_object, test, feature, steps
	Content            string  `db:"content"`
	GeneratedAt        int64   `db:"generated_at"`
	Applied            bool    `db:"applied"`
	CreatedAt          int64   `db:"created_at"`
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// Now returns current Unix timestamp
func Now() int64 {
	return time.Now().Unix()
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Int64Ptr returns a pointer to an int64
func Int64Ptr(i int64) *int64 {
	return &i
}
