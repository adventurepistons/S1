# Comprehensive Codebase Catalog

## Project Overview

**Test Automation Copilot** is an AI-powered code generation system for test automation. It uses advanced Retrieval-Augmented Generation (RAG) with Claude API to generate production-ready Selenium test code matching a project's coding patterns and style.

### Key Stats
- **Language**: Go (backend) + TypeScript (VSCode extension)
- **Architecture**: Modular, multi-component system
- **Database**: SQLite with 25+ tables
- **AI Integration**: Claude API (Anthropic)
- **Embeddings**: Local (sentence-transformers) - FREE
- **Cost**: ~$0.012 per test generated

---

## Part 1: Go Backend (automation-copilot)

### Directory Structure
```
automation-copilot/
├── cmd/                          # Entry points
│   ├── copilot/                 # Main CLI application
│   └── lsp-server/              # VSCode LSP server integration
├── internal/                     # Core implementation (non-exposed)
│   ├── ai/                      # AI/LLM integration (18 files)
│   ├── parser/                  # Java code parsing
│   ├── detector/                # Framework detection & pattern learning
│   ├── graph/                   # Knowledge graph building
│   ├── storage/                 # Database layer
│   └── extractor/               # Additional extraction utilities
├── pkg/models/                   # Data models & types
├── scripts/                      # Python embedding service
└── test_samples/                 # Sample test files

### 1.1 Main Entry Point: cmd/copilot/main.go

**Purpose**: CLI interface for all copilot operations

**CLI Commands**:
```
PARSING & INDEXING:
  parse <file.java>               # Parse Java file and show AST
  index <file.java>               # Parse and save to database
  query <class-name>              # Query indexed class info
  detect <project-dir>            # Detect frameworks and patterns
  build-graph                     # Build knowledge graph
  ask <question>                  # Ask natural language questions

AI CODE GENERATION:
  build-index                     # Build all AI indexes
  generate <request>              # Generate code with AI
  chat                           # Interactive multi-turn chat
  stats                          # Show Context Builder statistics
```

**Key Functions**:
- parseFile() - Extracts structured data from Java files
- indexFile() - Saves parsed data to SQLite database
- detectProject() - Runs framework and pattern detection
- generateCode() - Full code generation pipeline
- interactiveChat() - Multi-turn conversation mode
- showStats() - Display indexing and retrieval statistics

---

### 1.2 Core Packages in internal/

#### A. **internal/parser/** - Java Code Extraction

**java_parser.go** (1 file)
- **Type**: `JavaParser`
- **Purpose**: AST-based Java code parsing using Tree-sitter
- **Key Methods**:
  - ParseFile() - Main entry point
  - extractPackage() - Extract package declaration
  - extractImports() - Extract all import statements
  - extractClassDeclaration() - Extract class metadata (name, extends, implements, annotations)
  - extractFields() - Extract all fields including @FindBy WebElements
  - extractMethods() - Extract methods with complete bodies
  - detectFrameworkHints() - Identify test/page object patterns

**Extracted Data Includes**:
- Package name, file path
- Class metadata (modifiers, interfaces, annotations)
- Fields with type, annotations, locator strategies (@FindBy)
- Methods with full source code, parameters, return types
- Test annotations (@Test, @BeforeMethod, etc.)
- Method calls and field access patterns

---

#### B. **internal/ai/** - AI/LLM Integration (18 files)

**Core Orchestrator**:
- **context_builder.go** - Main orchestrator combining all AI components
  - Builds complete context through a 6-step pipeline
  - Calls LLM or generates prompt-only
  - Manages full code generation workflow

**Request Analysis**:
- **request_analyzer.go** - Parses user intent
  - Extracts intent (create_test, create_page_object, etc.)
  - Identifies entities (pages, elements, methods)
  - Generates embeddings for semantic search

**Retrieval (RAG Core)**:
- **hybrid_retriever.go** - Combines keyword + semantic search
  - Uses α=0.65 semantic + 0.35 BM25 fusion
  - Achieves 81% Recall@10 accuracy
  - Falls back gracefully if embeddings fail
  
- **bm25.go** - Keyword-based retrieval
  - BM25 ranking algorithm
  - IDF (Inverse Document Frequency) scoring
  - Stores/computes statistics
  
- **semantic_index.go** - Embedding-based retrieval
  - Stores embeddings as BLOB in database
  - Cosine similarity search
  - Uses free local embeddings

**Context Enhancement**:
- **context_enricher.go** - Adds explanatory context
  - +49% retrieval accuracy improvement
  - Implements Anthropic's contextual retrieval method
  
- **chunker.go** - cAST-aware code chunking
  - Respects AST boundaries
  - Never breaks methods mid-execution
  - Creates chunks optimized for retrieval

- **tokenizer.go** - Token counting and estimation
  - Estimates tokens for cost calculation
  - Handles caching token estimates

**Few-Shot Learning**:
- **fewshot_selector.go** - Dynamic example selection
  - Selects 3-5 positive examples
  - Selects 1-2 anti-patterns (negative examples)
  - Uses semantic similarity for selection
  - +7.3% F1-score improvement

**Conversation Management**:
- **conversation_manager.go** - Multi-turn conversation
  - Keeps last 5 turns of history
  - Stores to database
  - Enables iterative refinement

**LLM Integration**:
- **llm_client.go** - LLM client interface
  - Abstract interface for any LLM provider
  
- **claude_client.go** - Claude API implementation
  - Uses Anthropic SDK
  - Implements prompt caching (90% cost reduction)
  - Tracks input/output tokens and costs
  - Supports cache read tokens

**Prompt Generation**:
- **prompt_assembler.go** - Constructs complete prompt
  - System prompt (cached)
  - Dynamic user prompt
  - Self-RAG instructions
  - Manages cache breakpoints

**Output Processing**:
- **response_parser.go** - Parses LLM responses
  - Extracts generated code
  - Parses class name, package, type
  - Identifies imports
  
- **code_writer.go** - Writes code to files
  - Creates package directory structure
  - Formats and saves Java code
  - Suggests file paths

**Models**:
- **models.go** - Type definitions
  - UserRequest, SearchResult, Example
  - AssembledPrompt, GeneratedCode
  - ContextBuilderMetrics, Config
  - ConversationMessage, ConversationHistory

---

#### C. **internal/detector/** - Framework & Pattern Detection (2 files)

**framework_detector.go**
- **Type**: `FrameworkDetector`
- **Purpose**: Detects framework combination from project
- **Detects** (6 combinations):
  1. Selenium + TestNG
  2. Selenium + TestNG + Cucumber
  3. Selenium + JUnit 5
  4. Selenium + JUnit 5 + Cucumber
  5. Selenium + Serenity BDD
  6. Selenium + Rest-Assured
  
- **Detection Methods**:
  - Reads pom.xml for Maven dependencies
  - Checks imports in Java files
  - Finds TestNG XML config
  - Discovers Cucumber feature files
  - Detects architecture patterns (Page Object, PageFactory, Screenplay)
  
- **Returns**: FrameworkConfig with all detected info

**pattern_detector.go**
- **Type**: `PatternDetector`
- **Purpose**: Learns coding conventions from codebase
- **Analyzes**:
  - Naming conventions (test methods, page objects, web elements)
  - Test structure (AAA pattern vs Given-When-Then)
  - Wait strategies (explicit vs implicit)
  - Assertion libraries and patterns
  - Data provider usage
  
- **Returns**: CodingPatterns used for code generation

---

#### D. **internal/graph/** - Knowledge Graph (2 files)

**knowledge_graph.go**
- **Type**: `KnowledgeGraph`
- **Purpose**: Builds relationships between code entities
- **Relationships Built**:
  - EXTENDS - class inheritance
  - IMPLEMENTS - interface implementation
  - CALLS - method invocations
  - USES - field usage
  - TEST_USES_PAGE - test to page object linkage

**query_engine.go**
- **Type**: `QueryEngine`
- **Purpose**: Natural language queries on knowledge graph
- **Supported Queries**:
  - "Where is <element>?"
  - "What tests use <page>?"
  - "Elements in <page>"
  - "Who uses <field>?"
  - "Methods in <class>"
  - "List all page objects"
  - "Show all tests"

---

#### E. **internal/storage/** - Database Layer (1 file + schema)

**database.go**
- **Type**: `Database`
- **Purpose**: SQLite database management
- **Key Methods**:
  - SaveClassData() - Saves complete parsed class
  - GetClassByName() - Queries class by name
  - GetWebElementFields() - Finds @FindBy annotations
  
**schema.sql** - Complete database schema with:
- **Core Tables** (files, classes, fields, methods):
  - files - File metadata
  - classes - Class definitions with test/page object flags
  - fields - All field declarations with WebElement info
  - methods - Methods with test annotations and body source
  - method_calls - Method invocations (caller → callee)
  - field_access - Field usage tracking
  - assertions - Assertion statements
  - local_variables - Local variable declarations

- **Framework Detection Tables**:
  - framework_config - Detected framework combination
  - coding_patterns - Learned naming/style conventions
  - pattern_examples - Few-shot learning examples

- **Knowledge Graph Tables**:
  - relationships - Edges between entities
  - class_implements - Interface implementations

- **AI/RAG Tables**:
  - chunks - cAST-aware code chunks with embeddings
  - bm25_stats - IDF scores for keyword retrieval
  - retrieval_cache - Query result caching
  - conversation_messages - Chat history
  - conversation_sessions - Session tracking

---

### 1.3 Data Models: pkg/models/

**class_data.go**
- ClassData - Complete parsed class information
  - FilePath, Package, Imports
  - ClassName, Modifiers, Extends, Implements
  - Fields (with @FindBy details), Methods
  - Flags: PageObjectModel, TestClass, StepDefinition
  - Line numbers

- Supporting types:
  - Import, Annotation, Field, Method
  - Parameter, Assertion, MethodCall

**framework_config.go**
- FrameworkConfig - Detected framework info
  - Test framework (TestNG, JUnit)
  - BDD framework (Cucumber, Serenity)
  - API framework (RestAssured)
  - Architecture patterns
  - Config file paths

---

### 1.4 Python Embedding Service: scripts/

**Purpose**: Provides free local embeddings using sentence-transformers

**Features**:
- REST API (port 5000)
- health check endpoint
- Embedding generation endpoint
- Returns 384-dimensional vectors
- ~90MB model (downloaded once)

**Setup**: `make setup-embeddings` and `make start-embeddings`

---

## Part 2: VSCode Extension (vscode-extension)

### Directory Structure
```
vscode-extension/
├── src/
│   ├── extension.ts           # Main extension entry point
│   ├── extension-with-auth.ts # Auth-enabled variant
│   ├── chatPanel.ts           # Interactive chat UI
│   ├── authService.ts         # Firebase authentication
│   ├── firebase.ts            # Firebase configuration
│   └── frameworkDetector.ts   # Framework detection
├── package.json               # Extension metadata & commands
└── tsconfig.json              # TypeScript configuration

### 2.1 Main Extension: src/extension.ts

**Purpose**: VSCode extension lifecycle and command registration

**Components**:
- Extension activation on `.java` files
- LSP client creation (connects to Go backend)
- Status bar item showing "Test Copilot"
- Command registration
- Backend initialization

**Integration Points**:
- Launches Go LSP server (`bin/copilot-lsp`)
- Communicates via Language Server Protocol
- Passes API key and workspace root

**Configuration in package.json**:
- apiKey: Anthropic API key
- model: Claude model selection (sonnet-4.5, opus-4, etc.)
- maxTokens: Token limit per generation
- showCost: Display cost in status bar

---

### 2.2 Chat Interface: src/chatPanel.ts

**Purpose**: Interactive chat panel for multi-turn code generation

**Features**:
- Webview-based UI
- Message history display
- Typing indicators
- Cost tracking
- Code display with copy/save/insert options

**User Actions**:
- send - Submit generation request
- save - Save generated code to file
- insert - Insert code at cursor position

**Integration**:
- Communicates with Go backend via LSP
- Receives generated code with metadata
- Displays metrics (time, cost, tokens)

---

### 2.3 Authentication: src/authService.ts

**Purpose**: Firebase-based user authentication and tier management

**Features**:
- Google OAuth sign-in
- User data management
- Token storage (secure)
- Tier management (free/pro/team)
- Monthly usage tracking

**User Data Stored**:
- uid, email, displayName
- tier (free/pro/team)
- apiKeyMode (byok/local/managed)
- frameworks (supported frameworks)
- monthlyUsage, usageLimit

**Integration**:
- Uses Firebase Auth & Firestore
- Stores tokens in VSCode secrets
- Tracks last login
- Manages usage quotas

---

### 2.4 Firebase Integration: src/firebase.ts

**Purpose**: Firebase configuration and initialization

**Services**:
- Authentication (Google provider)
- Firestore (user data storage)
- Real-time sync

---

### 2.5 Framework Detection: src/frameworkDetector.ts

**Purpose**: Detects test framework from workspace

**Supported Frameworks**:
- Selenium + Java (free tier)
- Playwright (TypeScript/JavaScript) - Pro
- Cypress (JavaScript/TypeScript) - Pro
- Pytest + Selenium (Python) - Pro
- WebdriverIO (JavaScript) - Pro

**Detection Method**:
- Checks package.json for Node.js projects
- Checks pom.xml/build.gradle for Java
- Analyzes dependencies in config files

---

### 2.6 Extension Commands (from package.json)

**User Commands**:
- `testCopilot.signIn` - Google login
- `testCopilot.signOut` - Logout
- `testCopilot.showAccount` - Account info
- `testCopilot.switchMode` - Switch API mode

**Generation Commands**:
- `testCopilot.generate` - Generate test code
- `testCopilot.chat` - Open chat panel
- `testCopilot.buildIndex` - Index current project
- `testCopilot.showStats` - Display statistics
- `testCopilot.generateForClass` - Generate for current class (context menu)

**Activation Events**:
- `onLanguage:java` - Activate when opening Java files
- `onCommand:*` - Activate on specific command calls

---

## Part 3: Key Features & Capabilities

### 3.1 Code Understanding Features

**Feature**: Complete Java Code Extraction
- Parses ALL code elements using Tree-sitter AST
- Captures @FindBy annotations with locator strategies
- Tracks method bodies, parameters, return types
- Records field types, initializers, annotations
- Identifies test methods and page object patterns

**Feature**: Database Indexing
- SQLite storage with 25+ tables
- Foreign key relationships for querying
- Efficient indexing on critical columns
- Transaction support for consistency

**Feature**: Framework Detection
- Auto-detects 6 framework combinations
- Identifies architecture patterns (POM, PageFactory, Screenplay)
- Parses configuration files (pom.xml, testng.xml, etc.)
- Detects Cucumber features

**Feature**: Pattern Learning
- Analyzes naming conventions
- Learns assertion styles
- Detects wait strategies
- Recognizes test structure patterns
- Identifies data provider usage

---

### 3.2 AI/RAG Features

**Feature**: Hybrid Retrieval
- BM25 keyword search (Elasticsearch-like)
- Semantic search via embeddings
- Weighted fusion (α=0.65 semantic)
- 81% Recall@10 accuracy

**Feature**: Context Enrichment
- Adds explanatory context to retrieved code
- Anthropic's contextual retrieval method
- +49% accuracy improvement

**Feature**: Few-Shot Learning
- Selects 3-5 positive examples (good patterns)
- Selects 1-2 anti-pattern examples (what to avoid)
- Dynamic selection via semantic similarity
- +7.3% F1-score improvement

**Feature**: Conversation Management
- Multi-turn chat with history
- Keeps last 5 turns for context
- Enables iterative refinement
- Stores conversation in database

**Feature**: Prompt Caching
- Caches system prompt with Claude API
- ~90% cost reduction
- 512-token minimum cache size
- Automatic cache management

---

### 3.3 Code Generation Features

**Feature**: Automatic Code Generation
- End-to-end generation pipeline
- Validates generated code
- Parses class/method/imports
- Suggests file paths

**Feature**: Multiple Generation Modes**:
1. **Full Automatic** - Calls Claude API directly
2. **Prompt Only** - Builds context for manual use
3. **Interactive Chat** - Multi-turn refinement

**Feature**: Cost Optimization
- Input: ~6100 tokens (3500 cached)
- Output: ~452 tokens
- Total cost: ~$0.012 per test
- 93% cheaper than manual writing

---

### 3.4 User-Facing Features

**Feature**: CLI Interface
```
copilot parse <file>              # Analyze Java file
copilot index <file>              # Index to database
copilot detect <project>          # Auto-detect framework
copilot build-index               # Build AI indexes
copilot generate "description"    # Generate code
copilot chat                      # Interactive mode
copilot ask "question"            # Query codebase
copilot stats                     # View statistics
```

**Feature**: VSCode Integration
- Side panel chat interface
- Command palette commands
- Context menu (generate for class)
- Status bar cost indicator
- Settings for API key, model, tokens

**Feature**: Interactive Chat
- Multi-turn conversations
- Iterative code refinement
- Save/insert generated code
- Session tracking
- Cost display

**Feature**: Statistics & Monitoring
- Indexing progress
- Retrieval metrics
- Token counting
- Cost calculation
- Index coverage

---

## Part 4: Database Schema Overview

### Core Tables (Code Extraction)
```
files               # File metadata
├─ imports         # Import statements
├─ classes         # Class definitions
│  ├─ class_implements      # Interfaces
│  ├─ fields               # Class fields with @FindBy
│  │  └─ field_annotations # @FindBy details
│  ├─ methods              # Methods with bodies
│  │  ├─ method_parameters # Parameters
│  │  ├─ method_annotations # @Test, @Before, etc
│  │  ├─ method_calls      # Who calls who
│  │  ├─ assertions        # Assert statements
│  │  └─ local_variables   # Local vars
│  └─ field_access         # Field usage
relationships       # Knowledge graph edges
```

### Framework Tables
```
framework_config    # Detected frameworks
coding_patterns     # Naming/style conventions
pattern_examples    # Few-shot learning examples
```

### AI/RAG Tables
```
chunks              # cAST-aware chunks with embeddings
bm25_stats          # IDF scores
retrieval_cache     # Query caching
conversation_messages # Chat history
conversation_sessions # Session tracking
```

---

## Part 5: Architecture Flow

### Complete Code Generation Pipeline

```
User Input: "Create test for login"
    ↓
[Request Analyzer]
    ↓ Extracts: intent, entities, intent embedding
[Hybrid Retriever]
    ├─ BM25 Search → Top-N results
    ├─ Semantic Search → Embedding-based results
    └─ Fusion (α=0.65) → Top-K final chunks
    ↓
[Context Enricher]
    ↓ Adds explanatory context (+49% accuracy)
[Few-Shot Selector]
    ├─ Select 3-5 positive examples
    └─ Select 1-2 anti-patterns
    ↓
[Conversation Manager]
    ↓ Retrieves last 5 turns from database
[Prompt Assembler]
    ├─ System Prompt (cached)
    ├─ Dynamic Content
    └─ Self-RAG instructions
    ↓
[LLM Client - Claude API]
    ├─ Input tokens: ~6100 (3500 cached)
    └─ Output tokens: ~452
    ↓
[Response Parser]
    ├─ Extract code
    ├─ Parse class/method/package
    └─ Identify imports
    ↓
[Code Writer]
    └─ Save to file with suggested path
    ↓
Generated Test Code
```

---

## Part 6: Technology Stack

### Backend (Go)
- **Parsing**: Tree-sitter (java.wasm)
- **Database**: SQLite with go-sqlite3
- **LLM**: Anthropic Claude API
- **Embeddings**: REST client for Python service
- **HTTP/WS**: Standard Go libraries

### Frontend (VSCode Extension)
- **Framework**: VSCode API
- **Language**: TypeScript
- **LSP**: vscode-languageclient
- **Auth**: Firebase
- **Styling**: VSCode WebView CSS

### Infrastructure
- **Embeddings Service**: Python (sentence-transformers)
- **Model**: all-MiniLM-L6-v2 (384-dim, 90MB)

---

## Part 7: Performance Characteristics

| Metric | Value |
|--------|-------|
| **Retrieval Accuracy** | 81% Recall@10 |
| **Context Quality** | +49% improvement |
| **Few-Shot Impact** | +7.3% F1-score |
| **Cost Reduction** | 90% via caching |
| **Time** | 2 min (was 30 min) |
| **Speed** | 93% faster than manual |
| **Cost** | $0.012 per test |
| **Style Consistency** | 100% match |

---

## Part 8: File Summary

### Go Backend Files (32 total)

**cmd/ (2 files)**:
- main.go - CLI entry point with 8+ commands
- cmd/lsp-server/main.go - LSCode integration

**internal/ai/ (18 files)**:
- context_builder.go - Main orchestrator
- llm_client.go, claude_client.go - LLM integration
- hybrid_retriever.go, bm25.go, semantic_index.go - Retrieval
- chunker.go, tokenizer.go - Text processing
- context_enricher.go, fewshot_selector.go - Context enhancement
- conversation_manager.go - Multi-turn context
- prompt_assembler.go - Prompt construction
- request_analyzer.go - Intent extraction
- response_parser.go, code_writer.go - Output handling
- embedding_client.go - Embedding service
- models.go - Type definitions

**internal/parser/ (1 file)**:
- java_parser.go - Tree-sitter AST parsing

**internal/detector/ (2 files)**:
- framework_detector.go - Framework detection
- pattern_detector.go - Pattern learning

**internal/graph/ (2 files)**:
- knowledge_graph.go - Relationship building
- query_engine.go - Graph querying

**internal/storage/ (1 file + schema)**:
- database.go - SQLite operations
- schema.sql - Database schema

**pkg/models/ (2 files)**:
- class_data.go - Parsed code structures
- framework_config.go - Framework info

### VSCode Extension Files (6 files)

**src/**:
- extension.ts - Main extension entry point
- extension-with-auth.ts - Auth variant
- chatPanel.ts - Interactive chat UI
- authService.ts - Firebase authentication
- firebase.ts - Firebase config
- frameworkDetector.ts - Framework detection

**Config**:
- package.json - Extension manifest & commands
- tsconfig.json - TypeScript config

---

## Part 9: Key Algorithms

### BM25 Ranking
```
score = Σ IDF(term) × (tf × (k1+1)) / (tf + k1 × (1-b + b×(len/avgLen)))
Parameters: k1=1.2, b=0.75
```

### Hybrid Score Fusion
```
final_score = 0.65 × semantic_score + 0.35 × bm25_score
Why: Semantic captures intent, BM25 provides precision
```

### Few-Shot Selection
- Select top similar examples via cosine similarity
- 3-5 positive examples for good patterns
- 1-2 negative examples for anti-patterns
- Improves F1-score by 7.3%

### cAST-Aware Chunking
- Respects AST boundaries
- Never breaks methods mid-execution
- Optimal chunk size: ~500 tokens
- 50-token overlap between chunks

---

## Part 10: Configuration

**Default Config** (internal/ai/models.go):
```go
SemanticWeight:       0.65      // Semantic vs BM25 weight
TopK:                 5         // Chunks to retrieve
BM25_K1:              1.2       // Term frequency saturation
BM25_B:               0.75      // Length normalization
PositiveExamples:     3         // Good examples
NegativeExamples:     1         // Anti-patterns
MaxChunkTokens:       500       // Chunk size
ChunkOverlap:         50        // Overlap tokens
EmbeddingServiceURL:  "http://localhost:5000"
EmbeddingDimension:   384       // Vector size
EnableCaching:        true      // Prompt caching
MaxConversationTurns: 5         // History length
```

---

## Conclusion

This codebase is a **production-ready AI code generation system** with:

✅ Complete Java AST parsing  
✅ Intelligent RAG with hybrid retrieval  
✅ Free local embeddings  
✅ Multi-turn conversations  
✅ 90% cost reduction via prompt caching  
✅ Framework auto-detection  
✅ Pattern learning  
✅ VSCode integration  
✅ Beautiful UI  
✅ Firebase authentication  

**Total Implementation**: ~8,000+ lines of Go + TypeScript

