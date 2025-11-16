-- ============================================================================
-- COMPLETE DATABASE SCHEMA FOR DEEP CODEBASE UNDERSTANDING
-- ============================================================================

-- File metadata
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT UNIQUE NOT NULL,
    package TEXT,
    last_modified TIMESTAMP,
    last_indexed TIMESTAMP,
    file_hash TEXT
);

CREATE INDEX IF NOT EXISTS idx_files_package ON files(package);

-- Class metadata
CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    class_name TEXT NOT NULL,
    fully_qualified_name TEXT UNIQUE NOT NULL,
    extends TEXT,
    modifiers TEXT,
    is_page_object BOOLEAN DEFAULT 0,
    is_test_class BOOLEAN DEFAULT 0,
    is_step_definition BOOLEAN DEFAULT 0,
    line_start INTEGER,
    line_end INTEGER,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_classes_name ON classes(class_name);
CREATE INDEX IF NOT EXISTS idx_classes_page_object ON classes(is_page_object);
CREATE INDEX IF NOT EXISTS idx_classes_test_class ON classes(is_test_class);

-- Interfaces implemented by classes
CREATE TABLE IF NOT EXISTS class_implements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    interface_name TEXT NOT NULL,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_class_implements_class_id ON class_implements(class_id);

-- Import statements
CREATE TABLE IF NOT EXISTS imports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    import_path TEXT NOT NULL,
    is_static BOOLEAN DEFAULT 0,
    is_wildcard BOOLEAN DEFAULT 0,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_imports_file_id ON imports(file_id);
CREATE INDEX IF NOT EXISTS idx_imports_path ON imports(import_path);

-- Fields (with SPECIAL focus on @FindBy WebElements)
CREATE TABLE IF NOT EXISTS fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    field_name TEXT NOT NULL,
    field_type TEXT NOT NULL,
    modifiers TEXT,
    initializer TEXT,
    is_web_element BOOLEAN DEFAULT 0,
    locator_strategy TEXT,
    locator_value TEXT,
    line_number INTEGER,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_fields_class_id ON fields(class_id);
CREATE INDEX IF NOT EXISTS idx_fields_name ON fields(field_name);
CREATE INDEX IF NOT EXISTS idx_fields_web_element ON fields(is_web_element);
CREATE INDEX IF NOT EXISTS idx_fields_locator_strategy ON fields(locator_strategy);

-- Annotations on fields (especially @FindBy)
CREATE TABLE IF NOT EXISTS field_annotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    field_id INTEGER NOT NULL,
    annotation_type TEXT NOT NULL,
    parameters TEXT,
    how TEXT,
    using TEXT,
    line_number INTEGER,
    FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_field_annotations_field_id ON field_annotations(field_id);
CREATE INDEX IF NOT EXISTS idx_field_annotations_type ON field_annotations(annotation_type);

-- Methods
CREATE TABLE IF NOT EXISTS methods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    method_name TEXT NOT NULL,
    return_type TEXT NOT NULL,
    modifiers TEXT,
    is_test BOOLEAN DEFAULT 0,
    test_type TEXT,
    uses_explicit_wait BOOLEAN DEFAULT 0,
    uses_implicit_wait BOOLEAN DEFAULT 0,
    wait_timeout INTEGER,
    if_statements INTEGER DEFAULT 0,
    for_loops INTEGER DEFAULT 0,
    try_catch_blocks INTEGER DEFAULT 0,
    line_start INTEGER,
    line_end INTEGER,
    body_source TEXT,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_methods_class_id ON methods(class_id);
CREATE INDEX IF NOT EXISTS idx_methods_name ON methods(method_name);
CREATE INDEX IF NOT EXISTS idx_methods_is_test ON methods(is_test);

-- Method parameters
CREATE TABLE IF NOT EXISTS method_parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    param_name TEXT NOT NULL,
    param_type TEXT NOT NULL,
    param_order INTEGER NOT NULL,
    annotations TEXT,
    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_method_parameters_method_id ON method_parameters(method_id);

-- Annotations on methods (especially @Test, @BeforeMethod, etc.)
CREATE TABLE IF NOT EXISTS method_annotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    annotation_type TEXT NOT NULL,
    parameters TEXT,
    description TEXT,
    priority INTEGER,
    enabled BOOLEAN DEFAULT 1,
    line_number INTEGER,
    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_method_annotations_method_id ON method_annotations(method_id);
CREATE INDEX IF NOT EXISTS idx_method_annotations_type ON method_annotations(annotation_type);

-- Method calls (WHO calls WHAT - critical for understanding flow)
CREATE TABLE IF NOT EXISTS method_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    caller_method_id INTEGER NOT NULL,
    method_name TEXT NOT NULL,
    object_name TEXT,
    arguments TEXT,
    chained_from TEXT,
    line_number INTEGER,
    callee_method_id INTEGER,
    FOREIGN KEY (caller_method_id) REFERENCES methods(id) ON DELETE CASCADE,
    FOREIGN KEY (callee_method_id) REFERENCES methods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_method_calls_caller ON method_calls(caller_method_id);
CREATE INDEX IF NOT EXISTS idx_method_calls_method_name ON method_calls(method_name);
CREATE INDEX IF NOT EXISTS idx_method_calls_callee ON method_calls(callee_method_id);

-- Field access (WHO uses WHAT field - critical for page object usage)
CREATE TABLE IF NOT EXISTS field_access (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    field_name TEXT NOT NULL,
    field_id INTEGER,
    access_type TEXT,
    line_number INTEGER,
    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE,
    FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_field_access_method_id ON field_access(method_id);
CREATE INDEX IF NOT EXISTS idx_field_access_field_id ON field_access(field_id);
CREATE INDEX IF NOT EXISTS idx_field_access_field_name ON field_access(field_name);

-- Local variables in methods
CREATE TABLE IF NOT EXISTS local_variables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    var_name TEXT NOT NULL,
    var_type TEXT NOT NULL,
    init_value TEXT,
    line_number INTEGER,
    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_local_variables_method_id ON local_variables(method_id);

-- Assertions
CREATE TABLE IF NOT EXISTS assertions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    assertion_type TEXT NOT NULL,
    expected TEXT,
    actual TEXT,
    message TEXT,
    line_number INTEGER,
    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assertions_method_id ON assertions(method_id);
CREATE INDEX IF NOT EXISTS idx_assertions_type ON assertions(assertion_type);

-- ============================================================================
-- FRAMEWORK DETECTION TABLES
-- ============================================================================

-- Detected framework configuration
CREATE TABLE IF NOT EXISTS framework_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_root TEXT UNIQUE NOT NULL,
    test_framework TEXT,
    bdd_framework TEXT,
    api_framework TEXT,
    selenium_version TEXT,
    java_version TEXT,
    uses_page_object_model BOOLEAN DEFAULT 0,
    uses_page_factory BOOLEAN DEFAULT 0,
    uses_screenplay BOOLEAN DEFAULT 0,
    testng_xml_path TEXT,
    pom_file_path TEXT,
    cucumber_features_path TEXT,
    last_detected TIMESTAMP
);

-- Coding patterns detected in the project
CREATE TABLE IF NOT EXISTS coding_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_root TEXT UNIQUE NOT NULL,
    test_method_naming TEXT,
    page_object_naming TEXT,
    web_element_naming TEXT,
    method_naming TEXT,
    uses_aaa_pattern BOOLEAN DEFAULT 0,
    uses_given_when_then BOOLEAN DEFAULT 0,
    preferred_wait_type TEXT,
    default_wait_timeout INTEGER,
    assertion_library TEXT,
    uses_assert_messages BOOLEAN DEFAULT 0,
    uses_data_providers BOOLEAN DEFAULT 0,
    uses_csv_files BOOLEAN DEFAULT 0,
    uses_excel_files BOOLEAN DEFAULT 0,
    last_analyzed TIMESTAMP
);

-- Pattern examples (for learning by example when generating code)
CREATE TABLE IF NOT EXISTS pattern_examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern_type TEXT NOT NULL,
    example_code TEXT NOT NULL,
    file_path TEXT,
    line_start INTEGER,
    line_end INTEGER,
    frequency INTEGER DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_pattern_examples_type ON pattern_examples(pattern_type);

-- ============================================================================
-- KNOWLEDGE GRAPH TABLES
-- ============================================================================

-- Relationships between entities (for building knowledge graph)
CREATE TABLE IF NOT EXISTS relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    relationship_type TEXT NOT NULL,
    from_entity_type TEXT NOT NULL,
    from_entity_id INTEGER NOT NULL,
    to_entity_type TEXT NOT NULL,
    to_entity_id INTEGER NOT NULL,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_relationships_from ON relationships(from_entity_type, from_entity_id);
CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_entity_type, to_entity_id);

-- ============================================================================
-- METADATA AND TRACKING
-- ============================================================================

-- Indexing progress tracking
CREATE TABLE IF NOT EXISTS indexing_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_root TEXT UNIQUE NOT NULL,
    total_files INTEGER,
    indexed_files INTEGER,
    failed_files INTEGER,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    status TEXT
);

-- Errors encountered during parsing
CREATE TABLE IF NOT EXISTS parsing_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    error_message TEXT NOT NULL,
    error_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_parsing_errors_file_path ON parsing_errors(file_path);
