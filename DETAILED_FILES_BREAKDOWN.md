# Detailed File-by-File Breakdown

## Go Backend Files (automation-copilot)

### Entry Points (cmd/)

#### 1. cmd/copilot/main.go (813 lines)
**Purpose**: Main CLI entry point for the Copilot system
**Functions**:
- main() - Parses CLI arguments and routes to commands
- parseFile() - Parses Java file and prints AST
- indexFile() - Parses and saves to database
- queryClass() - Queries database for class info
- detectProject() - Runs framework and pattern detection
- buildGraph() - Builds knowledge graph from indexed data
- askQuestion() - Natural language queries
- buildAIIndex() - Builds all AI indexes (chunks, BM25, embeddings)
- generateCode() - Full code generation pipeline
- generateCodeWithLLM() - Calls Claude API for generation
- generatePromptOnly() - Builds prompt without API call
- interactiveChat() - Multi-turn chat mode
- showStats() - Display statistics

**CLI Commands Exposed**:
- parse, index, query, detect, build-graph, ask
- build-index, generate, chat, stats

**Key Dependencies**:
- parser.JavaParser
- ai.ContextBuilder
- detector (FrameworkDetector, PatternDetector)
- graph (KnowledgeGraph, QueryEngine)
- storage.Database

---

#### 2. cmd/lsp-server/main.go
**Purpose**: LSP server for VSCode extension integration
**Implements**:
- Language Server Protocol
- Connects to copilot backend
- Handles LSP requests from VSCode
- Bridges TypeScript frontend to Go backend

---

### AI/LLM Integration (internal/ai/)

#### 1. context_builder.go (200+ lines)
**Type**: ContextBuilder
**Purpose**: Main orchestrator for AI code generation
**Key Methods**:
- NewContextBuilder() - Initialize with config
- BuildContext() - 6-step context building pipeline
- GenerateCode() - Full code generation with LLM
- GenerateCodePromptOnly() - Just build prompt
- GetStats() - Return statistics
- GetConversationManager() - Access conversation history

**Pipeline Steps**:
1. Analyze request (RequestAnalyzer)
2. Retrieve context (HybridRetriever)
3. Enrich context (ContextEnricher)
4. Select examples (FewShotSelector)
5. Manage conversation (ConversationManager)
6. Assemble prompt (PromptAssembler)

**Components Used**:
- RequestAnalyzer, HybridRetriever, ContextEnricher
- FewShotSelector, ConversationManager, PromptAssembler
- EmbeddingClient, LLMClient, ResponseParser

---

#### 2. request_analyzer.go
**Type**: RequestAnalyzer
**Purpose**: Parses user requests to extract intent and entities
**Methods**:
- Analyze() - Main analysis method
- extractIntent() - Identifies task type (create_test, etc.)
- extractEntities() - Finds pages, elements, methods
- generateEmbedding() - Creates vector representation

**Returns**: UserRequest with intent, entities, embedding

---

#### 3. hybrid_retriever.go (90+ lines)
**Type**: HybridRetriever
**Purpose**: Combines BM25 and semantic search
**Methods**:
- Retrieve() - Main retrieval method
- fuseScores() - Combines BM25 and semantic scores

**Algorithm**: 
```
final_score = 0.65 * semantic + 0.35 * bm25
```

**Fallback Strategy**:
- If embeddings fail, uses BM25 only
- Graceful degradation

---

#### 4. bm25.go
**Type**: BM25Index
**Purpose**: Keyword-based retrieval using BM25 ranking
**Methods**:
- Search() - BM25 search implementation
- ComputeIDF() - Calculate IDF scores
- TokenizeQuery() - Parse query to tokens

**Algorithm**:
```
score = Σ IDF(term) × (tf × (k1+1)) / (tf + k1 × (1-b + b×(len/avgLen)))
k1=1.2, b=0.75
```

---

#### 5. semantic_index.go
**Type**: SemanticIndex
**Purpose**: Embedding-based similarity search
**Methods**:
- Search() - Find similar chunks using embeddings
- cosineSimilarity() - Calculate similarity scores
- buildIndex() - Create semantic index

**Features**:
- Stores embeddings as BLOB in database
- Cosine similarity calculation
- Uses local embedding service

---

#### 6. chunker.go (100+ lines)
**Type**: Chunker
**Purpose**: cAST-aware code chunking for retrieval
**Methods**:
- ChunkProject() - Create chunks for entire project
- chunkClass() - Chunk a single class
- chunkMethod() - Create method chunk

**Features**:
- Respects AST boundaries
- Never breaks methods mid-execution
- Optimal size: ~500 tokens
- 50-token overlap between chunks

---

#### 7. tokenizer.go
**Type**: Tokenizer
**Purpose**: Token counting for cost estimation
**Methods**:
- EstimateTokens() - Estimate token count
- Count() - Count tokens in text

**Uses**: tiktoken library or similar

---

#### 8. context_enricher.go
**Type**: ContextEnricher
**Purpose**: Adds explanatory context to retrieved chunks
**Methods**:
- Enrich() - Add context to chunks

**Benefit**: +49% retrieval accuracy improvement

---

#### 9. fewshot_selector.go
**Type**: FewShotSelector
**Purpose**: Dynamic few-shot example selection
**Methods**:
- SelectExamples() - Choose best examples
- selectPositive() - Select good patterns
- selectNegative() - Select anti-patterns

**Features**:
- 3-5 positive examples
- 1-2 negative examples
- Semantic similarity-based selection
- +7.3% F1-score improvement

---

#### 10. conversation_manager.go
**Type**: ConversationManager
**Purpose**: Manage multi-turn conversation history
**Methods**:
- AddMessage() - Add to conversation
- GetHistory() - Retrieve conversation
- GetLastNTurns() - Get recent turns (default: 5)
- GetSessionID() - Session identifier

**Storage**: SQLite conversation_messages table

---

#### 11. llm_client.go
**Type**: LLMClient (interface)
**Purpose**: Abstract LLM client interface
**Methods**:
- Generate() - Generate code from prompt
- (implementations: Claude, OpenAI, etc.)

---

#### 12. claude_client.go
**Type**: ClaudeClient
**Purpose**: Anthropic Claude API implementation
**Features**:
- Prompt caching (90% cost reduction)
- Token counting
- Cost calculation
- Cache read tokens tracking

**Methods**:
- Generate() - Call Claude API
- calculateCost() - Compute API cost

**Caching**:
- System prompt cached (512+ tokens)
- Dynamic content not cached
- Cache hit: 90% cheaper

---

#### 13. prompt_assembler.go
**Type**: PromptAssembler
**Purpose**: Construct complete prompt for LLM
**Methods**:
- Assemble() - Build prompt from components
- buildSystemPrompt() - Create cached system prompt
- buildUserPrompt() - Create dynamic user prompt

**Components**:
- System prompt (static, cacheable)
- User prompt (dynamic)
- Retrieved context
- Few-shot examples
- Conversation history
- Self-RAG instructions

---

#### 14. response_parser.go
**Type**: ResponseParser
**Purpose**: Parse LLM response to extract code
**Methods**:
- Parse() - Extract code from response
- validateCode() - Check syntax
- extractMetadata() - Get class, package, type

**Output**: ParsedResponse with GeneratedCode

---

#### 15. code_writer.go
**Type**: CodeWriter
**Purpose**: Write generated code to files
**Methods**:
- WriteCodeInteractive() - Save with user confirmation
- suggestPath() - Recommend file path
- createDirectories() - Create package directories

---

#### 16. embedding_client.go
**Type**: EmbeddingClient
**Purpose**: REST client for embedding service
**Methods**:
- Embed() - Get embedding for text
- HealthCheck() - Verify service is available

**Endpoint**: http://localhost:5000

---

#### 17. models.go (160 lines)
**Type Definitions**:
- UserRequest - Parsed user input
- SearchResult - Retrieved code chunk
- Example - Few-shot learning example
- AssembledPrompt - Complete prompt for LLM
- GeneratedCode - Code output from LLM
- ConversationMessage - Chat message
- ConversationHistory - Full chat history
- ContextBuilderMetrics - Performance metrics
- CodeGenerationResult - Complete generation result
- Config - Configuration struct
- DefaultConfig() - Default settings

---

---

### Parser (internal/parser/)

#### 1. java_parser.go (800+ lines)
**Type**: JavaParser
**Purpose**: Tree-sitter based Java AST parsing
**Technology**: Tree-sitter with Java grammar

**Key Methods**:
- ParseFile() - Main entry point
- extractPackage() - Get package declaration
- extractImports() - Get all imports
- extractClassDeclaration() - Get class metadata
- extractFields() - Get all fields including @FindBy
- extractMethods() - Get methods with bodies
- extractAnnotations() - Parse annotations
- detectFrameworkHints() - Identify test/page patterns

**Returns**: ClassData with complete code information

**Extracted Details**:
- Package, imports, class name
- Modifiers (public, abstract, final)
- Extends, implements, annotations
- Fields with types, initializers, @FindBy details
- Methods with parameters, return types, bodies
- Annotations (@Test, @BeforeMethod, @FindBy, etc.)

---

### Detector (internal/detector/)

#### 1. framework_detector.go (200+ lines)
**Type**: FrameworkDetector
**Purpose**: Detect framework combination from project

**Detection Strategy**:
1. Parse pom.xml for Maven dependencies
2. Check Java imports for frameworks
3. Find TestNG XML configuration
4. Discover Cucumber feature files
5. Detect architecture patterns

**6 Framework Combinations**:
1. Selenium + TestNG
2. Selenium + TestNG + Cucumber
3. Selenium + JUnit 5
4. Selenium + JUnit 5 + Cucumber
5. Selenium + Serenity BDD
6. Selenium + RestAssured

**Methods**:
- Detect() - Main detection
- detectFromPOM() - Parse pom.xml
- detectFromImports() - Analyze imports
- detectArchitecturePatterns() - Find POM, PageFactory, Screenplay
- findTestNGXML() - Locate testng.xml
- findCucumberFeatures() - Find .feature files

**Returns**: FrameworkConfig

---

#### 2. pattern_detector.go (200+ lines)
**Type**: PatternDetector
**Purpose**: Learn coding patterns from codebase

**Analyzes**:
- Test method naming conventions
- Page object naming patterns
- WebElement field naming
- General method naming
- Test structure (AAA vs BDD)
- Wait strategies
- Assertion libraries
- Data provider usage

**Methods**:
- DetectPatterns() - Main detection
- analyzeNamingConventions() - Naming analysis
- analyzeTestStructure() - Test pattern analysis
- analyzeWaitStrategies() - Wait pattern analysis
- analyzeAssertions() - Assertion library detection

**Returns**: CodingPatterns used in code generation

---

### Graph (internal/graph/)

#### 1. knowledge_graph.go (200+ lines)
**Type**: KnowledgeGraph
**Purpose**: Build and manage knowledge graph

**Relationships Built**:
- EXTENDS - class inheritance
- IMPLEMENTS - interface implementation
- CALLS - method invocations
- USES - field usage
- TEST_USES_PAGE - test to page object

**Methods**:
- BuildGraph() - Build all relationships
- buildClassRelationships() - EXTENDS/IMPLEMENTS
- buildMethodCallGraph() - CALLS relationships
- buildFieldUsageGraph() - USES relationships
- buildTestPageObjectLinks() - TEST_USES_PAGE
- createRelationship() - Add edge to graph

**Storage**: relationships table

---

#### 2. query_engine.go (150+ lines)
**Type**: QueryEngine
**Purpose**: Natural language queries on knowledge graph

**Supported Queries**:
- "Where is <element>?" - Find element location
- "What tests use <page>?" - Tests using page
- "Elements in <page>" - Page contents
- "Who uses <field>?" - Field usage
- "Methods in <class>" - Class methods
- "List all page objects" - All page objects
- "Show all tests" - All tests

**Methods**:
- Execute() - Process query
- parseQuery() - Extract intent
- queryGraph() - Execute query
- formatResult() - Format response

---

### Storage (internal/storage/)

#### 1. database.go (300+ lines)
**Type**: Database
**Purpose**: SQLite database management

**Methods**:
- NewDatabase() - Initialize DB
- SaveClassData() - Save complete class
- GetClassByName() - Query class
- GetWebElementFields() - Find @FindBy
- Query operations - Read/write to all tables

**Features**:
- Embedded schema
- Transaction support
- Foreign key enforcement
- Prepared statements

---

#### 2. schema.sql (380+ lines)
**Tables** (25+ total):

**Code Structure**:
- files - File metadata
- imports - Import statements
- classes - Class definitions
- class_implements - Interface implementation
- fields - Field declarations
- field_annotations - @FindBy, etc.
- methods - Method definitions
- method_parameters - Method parameters
- method_annotations - @Test, etc.
- method_calls - Method invocations
- field_access - Field usage
- local_variables - Local vars
- assertions - Assert statements

**Framework Detection**:
- framework_config - Detected frameworks
- coding_patterns - Naming/style conventions
- pattern_examples - Few-shot examples

**Knowledge Graph**:
- relationships - Entity relationships

**AI/RAG**:
- chunks - Code chunks with embeddings
- bm25_stats - IDF scores
- retrieval_cache - Query cache
- conversation_messages - Chat history
- conversation_sessions - Sessions

**Metadata**:
- indexing_progress - Indexing status
- parsing_errors - Error tracking

---

### Models (pkg/models/)

#### 1. class_data.go
**Types**:
- ClassData - Complete parsed class
- Import - Import statement
- Annotation - @Annotation definition
- Field - Field declaration
- Method - Method declaration
- Parameter - Method parameter
- Assertion - Assert statement
- MethodCall - Method invocation

---

#### 2. framework_config.go
**Type**: FrameworkConfig
**Fields**:
- TestFramework (TestNG, JUnit)
- BDDFramework (Cucumber, Serenity)
- APIFramework (RestAssured)
- SeleniumVersion
- JavaVersion
- Architecture flags (POM, PageFactory, Screenplay)
- Config file paths

---

### Scripts (scripts/)

#### 1. Python Embedding Service
**Purpose**: Provide free local embeddings
**Endpoint**: http://localhost:5000
**API**:
- /health - Health check
- /embed - Get embeddings

**Model**: all-MiniLM-L6-v2 (384-dim)
**Setup**: make setup-embeddings, make start-embeddings

---

## VSCode Extension Files (vscode-extension)

### Source Code (src/)

#### 1. extension.ts (200+ lines)
**Purpose**: Main extension entry point
**Features**:
- Extension activation on .java files
- LSP server launch
- Command registration
- Status bar item creation
- Backend initialization

**Commands Registered**:
- testCopilot.generate
- testCopilot.chat
- testCopilot.buildIndex
- testCopilot.showStats
- testCopilot.signIn
- testCopilot.signOut
- testCopilot.generateForClass

**Components**:
- LanguageClient setup
- CommandManager
- StatusBarItem

---

#### 2. chatPanel.ts (200+ lines)
**Type**: ChatPanel
**Purpose**: Interactive chat panel UI
**Features**:
- Webview-based UI
- Message history
- Typing indicators
- Cost tracking

**User Actions**:
- send - Submit request
- save - Save code
- insert - Insert at cursor

**Methods**:
- show() - Display panel
- handleUserMessage() - Process input
- handleSaveCode() - Save generation
- handleInsertCode() - Insert code
- updateWebview() - Update UI

---

#### 3. authService.ts (150+ lines)
**Type**: AuthService
**Purpose**: Firebase authentication
**Features**:
- Google OAuth login
- User data management
- Token storage
- Tier management
- Usage tracking

**Methods**:
- signIn() - Google login
- signOutUser() - Logout
- getCurrentUser() - Get user
- getUserData() - Get user info
- isSignedIn() - Check auth
- getToken() - Get access token

---

#### 4. firebase.ts
**Purpose**: Firebase configuration
**Services**:
- Authentication
- Firestore
- Real-time sync

---

#### 5. frameworkDetector.ts (150+ lines)
**Type**: FrameworkDetector
**Purpose**: Detect test framework

**Supported Frameworks**:
- Selenium Java
- Playwright (TS/JS) - Pro
- Cypress (JS/TS) - Pro
- Pytest Python - Pro
- WebdriverIO JS - Pro

**Methods**:
- detectFramework() - Main detection
- detectFromPackageJson() - Node.js
- detectFromPom() - Java
- detectFromGradle() - Gradle

---

#### 6. extension-with-auth.ts
**Purpose**: Auth-enabled variant of extension
**Enhancement**: Adds Firebase authentication

---

### Configuration Files

#### 1. package.json
**Metadata**:
- name: test-automation-copilot
- version: 1.0.0
- publisher: adventurepistons
- engines: VSCode 1.80.0+

**Commands** (9 total):
- signIn, signOut, showAccount
- switchMode
- generate, chat, buildIndex
- showStats
- generateForClass

**Configuration Properties**:
- apiKey - Anthropic API key
- model - Claude model selection
- maxTokens - Token limit
- showCost - Cost display

**Dependencies**:
- vscode-languageclient
- firebase

---

#### 2. tsconfig.json
**TypeScript Configuration**
- Target: ES2020
- Module: commonjs
- Strict mode enabled

---

---

## Summary Statistics

### Go Backend (automation-copilot)
- **Total Files**: 32
- **Total Lines**: ~8,000+
- **Packages**: 6 (ai, parser, detector, graph, storage, models)
- **Main Components**: 23 types/interfaces

### VSCode Extension (vscode-extension)
- **Total Files**: 8 (6 source + 2 config)
- **Total Lines**: ~1,500+
- **Types/Interfaces**: 10+

### Database
- **Tables**: 25+
- **Relationships**: Foreign keys for data integrity
- **Indexes**: 25+ for performance

### Total Project
- **Files**: 40+
- **Lines of Code**: ~9,500+
- **Complexity**: Medium-High
- **Maturity**: Production-ready

---

## Key Architectural Patterns

1. **Modular Design**: Separate concerns (parsing, detection, AI, storage)
2. **Pipeline Architecture**: 6-step context building
3. **Graceful Degradation**: Fallback strategies (embeddings → BM25)
4. **Layered Architecture**: Controllers → Services → Data access
5. **Strategy Pattern**: LLMClient interface for multiple providers
6. **Observer Pattern**: LSP server for UI updates
7. **Builder Pattern**: Prompt and context assembly

