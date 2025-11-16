-- Test Automation Copilot - SQLite Database Schema
-- Stores complete codebase structure for Cursor-style intelligence

-- ============================================================================
-- FILES: Track all files in workspace
-- ============================================================================
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,  -- 'java', 'gherkin', 'xml', 'properties'
    content_hash TEXT NOT NULL,
    content TEXT,  -- Full file content
    package_name TEXT,  -- Java package or null
    last_modified INTEGER NOT NULL,  -- Unix timestamp
    indexed_at INTEGER NOT NULL,  -- Unix timestamp
    size_bytes INTEGER NOT NULL,
    line_count INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
CREATE INDEX IF NOT EXISTS idx_files_type ON files(type);
CREATE INDEX IF NOT EXISTS idx_files_package ON files(package_name);
CREATE INDEX IF NOT EXISTS idx_files_hash ON files(content_hash);

-- ============================================================================
-- CLASSES: Java classes, interfaces, enums (including Page Objects)
-- ============================================================================
CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    fully_qualified_name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,  -- 'class', 'interface', 'enum', 'annotation'
    package_name TEXT,
    is_abstract BOOLEAN DEFAULT 0,
    is_public BOOLEAN DEFAULT 1,
    extends_class TEXT,  -- Fully qualified name
    implements_interfaces TEXT,  -- JSON array
    annotations TEXT,  -- JSON array
    javadoc TEXT,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_classes_file_id ON classes(file_id);
CREATE INDEX IF NOT EXISTS idx_classes_name ON classes(name);
CREATE INDEX IF NOT EXISTS idx_classes_fqn ON classes(fully_qualified_name);
CREATE INDEX IF NOT EXISTS idx_classes_type ON classes(type);
CREATE INDEX IF NOT EXISTS idx_classes_package ON classes(package_name);

-- ============================================================================
-- METHODS: Methods in Java classes (including test methods)
-- ============================================================================
CREATE TABLE IF NOT EXISTS methods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    signature TEXT NOT NULL,  -- Full method signature
    return_type TEXT,
    parameters TEXT,  -- JSON array of {name, type}
    is_public BOOLEAN DEFAULT 1,
    is_static BOOLEAN DEFAULT 0,
    is_abstract BOOLEAN DEFAULT 0,
    annotations TEXT,  -- JSON array of {name, params}
    javadoc TEXT,
    body TEXT,  -- Method implementation
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_methods_class_id ON methods(class_id);
CREATE INDEX IF NOT EXISTS idx_methods_name ON methods(name);
CREATE INDEX IF NOT EXISTS idx_methods_signature ON methods(signature);

-- ============================================================================
-- FIELDS: Class fields/variables (including WebElements in Page Objects)
-- ============================================================================
CREATE TABLE IF NOT EXISTS fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    is_public BOOLEAN DEFAULT 0,
    is_static BOOLEAN DEFAULT 0,
    is_final BOOLEAN DEFAULT 0,
    annotations TEXT,  -- JSON array (includes @FindBy)
    default_value TEXT,
    locator_type TEXT,  -- 'id', 'name', 'xpath', 'css', etc (for @FindBy)
    locator_value TEXT,  -- Actual locator value
    javadoc TEXT,
    start_line INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_fields_class_id ON fields(class_id);
CREATE INDEX IF NOT EXISTS idx_fields_name ON fields(name);
CREATE INDEX IF NOT EXISTS idx_fields_locator ON fields(locator_type, locator_value);

-- ============================================================================
-- DEPENDENCIES: Relationships between classes
-- ============================================================================
CREATE TABLE IF NOT EXISTS dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_class_id INTEGER NOT NULL,
    to_class_id INTEGER NOT NULL,
    dependency_type TEXT NOT NULL,  -- 'extends', 'implements', 'uses', 'imports', 'test_uses_page'
    context TEXT,  -- Where used (method name, etc)
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (from_class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (to_class_id) REFERENCES classes(id) ON DELETE CASCADE,
    UNIQUE(from_class_id, to_class_id, dependency_type)
);

CREATE INDEX IF NOT EXISTS idx_dependencies_from ON dependencies(from_class_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_to ON dependencies(to_class_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_type ON dependencies(dependency_type);

-- ============================================================================
-- FEATURE_FILES: Gherkin/BDD feature files
-- ============================================================================
CREATE TABLE IF NOT EXISTS feature_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    feature_name TEXT NOT NULL,
    description TEXT,
    tags TEXT,  -- JSON array
    language TEXT DEFAULT 'en',
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_features_file_id ON feature_files(file_id);

-- ============================================================================
-- SCENARIOS: Gherkin scenarios
-- ============================================================================
CREATE TABLE IF NOT EXISTS scenarios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,  -- 'scenario', 'scenario_outline'
    description TEXT,
    tags TEXT,  -- JSON array
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (feature_id) REFERENCES feature_files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scenarios_feature_id ON scenarios(feature_id);

-- ============================================================================
-- STEPS: Gherkin steps (Given/When/Then)
-- ============================================================================
CREATE TABLE IF NOT EXISTS steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scenario_id INTEGER NOT NULL,
    keyword TEXT NOT NULL,  -- 'Given', 'When', 'Then', 'And', 'But'
    text TEXT NOT NULL,
    argument TEXT,  -- DataTable or DocString
    line_number INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (scenario_id) REFERENCES scenarios(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_steps_scenario_id ON steps(scenario_id);

-- ============================================================================
-- STEP_DEFINITIONS: Java step definition methods
-- ============================================================================
CREATE TABLE IF NOT EXISTS step_definitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    method_id INTEGER NOT NULL,
    pattern TEXT NOT NULL,  -- Regex pattern
    keyword TEXT,  -- 'Given', 'When', 'Then' or null for any
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_step_defs_method_id ON step_definitions(method_id);
CREATE INDEX IF NOT EXISTS idx_step_defs_pattern ON step_definitions(pattern);

-- ============================================================================
-- CHAT_SESSIONS: Track chat conversations
-- ============================================================================
CREATE TABLE IF NOT EXISTS chat_sessions (
    id TEXT PRIMARY KEY,  -- UUID
    workspace_path TEXT NOT NULL,
    started_at INTEGER NOT NULL,
    ended_at INTEGER,
    message_count INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_workspace ON chat_sessions(workspace_path);

-- ============================================================================
-- CHAT_HISTORY: Store all chat messages
-- ============================================================================
CREATE TABLE IF NOT EXISTS chat_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,  -- 'user', 'assistant', 'system'
    message TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    context_files TEXT,  -- JSON array of file paths used
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_session_id ON chat_history(session_id);
CREATE INDEX IF NOT EXISTS idx_chat_timestamp ON chat_history(timestamp);

-- ============================================================================
-- FILE_CHANGES: Track all code modifications (for undo/redo)
-- ============================================================================
CREATE TABLE IF NOT EXISTS file_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    change_type TEXT NOT NULL,  -- 'create', 'modify', 'delete'
    before_content TEXT,
    after_content TEXT,
    diff TEXT,  -- Unified diff format
    applied BOOLEAN DEFAULT 0,
    timestamp INTEGER NOT NULL,
    user_approved BOOLEAN DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_changes_session_id ON file_changes(session_id);
CREATE INDEX IF NOT EXISTS idx_changes_file_path ON file_changes(file_path);
CREATE INDEX IF NOT EXISTS idx_changes_timestamp ON file_changes(timestamp);

-- ============================================================================
-- CONFIG: Store workspace configuration
-- ============================================================================
CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_path TEXT NOT NULL UNIQUE,
    framework TEXT,  -- 'selenium-java', 'playwright-java', etc
    test_framework TEXT,  -- 'testng', 'junit'
    bdd_enabled BOOLEAN DEFAULT 0,
    base_package TEXT,
    page_object_package TEXT,
    test_package TEXT,
    config_json TEXT,  -- JSON of additional config
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- ============================================================================
-- EMBEDDINGS: Store vector embeddings for semantic search
-- (Note: Actual vectors stored in chromem-go, this tracks what's indexed)
-- ============================================================================
CREATE TABLE IF NOT EXISTS embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL,  -- 'class', 'method', 'field', 'file'
    entity_id INTEGER NOT NULL,  -- Foreign key to respective table
    embedding_id TEXT NOT NULL,  -- ID in chromem-go
    indexed_at INTEGER NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_embeddings_entity ON embeddings(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_embeddings_id ON embeddings(embedding_id);

-- ============================================================================
-- RECORDING_SESSIONS: Store browser recording metadata
-- ============================================================================
CREATE TABLE IF NOT EXISTS recording_sessions (
    id TEXT PRIMARY KEY,  -- session_xxx
    workspace_path TEXT NOT NULL,
    start_url TEXT,
    started_at INTEGER NOT NULL,
    ended_at INTEGER,
    page_count INTEGER DEFAULT 0,
    interaction_count INTEGER DEFAULT 0,
    recording_data TEXT,  -- JSON of complete RecordingSession
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_recordings_workspace ON recording_sessions(workspace_path);

-- ============================================================================
-- GENERATED_CODE: Track all generated code
-- ============================================================================
CREATE TABLE IF NOT EXISTS generated_code (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recording_session_id TEXT,
    file_path TEXT NOT NULL,
    code_type TEXT NOT NULL,  -- 'page_object', 'test', 'feature', 'steps'
    content TEXT NOT NULL,
    generated_at INTEGER NOT NULL,
    applied BOOLEAN DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (recording_session_id) REFERENCES recording_sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_generated_recording ON generated_code(recording_session_id);
CREATE INDEX IF NOT EXISTS idx_generated_file_path ON generated_code(file_path);
