# Test Automation Copilot - Go Backend Server

A powerful AI-driven backend server for intelligent test automation code generation, analysis, and assistance.

## Overview

The Test Automation Copilot backend provides a complete suite of services for building a Cursor IDE-level intelligent assistant for test automation. It combines:

- **SQLite Database** - Fast relational storage for code structure
- **Vector Database (chromem-go)** - Semantic search over codebase
- **OpenAI GPT-4** - Context-aware code generation
- **HTTP/WebSocket API** - Real-time streaming responses

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    VS Code Extension                        │
│  ┌────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │   UI/UX    │  │  CoreClient  │  │  Browser Recorder│   │
│  └────────────┘  └──────┬───────┘  └──────────────────┘   │
└────────────────────────┼─────────────────────────────────────┘
                         │ HTTP/WebSocket
┌────────────────────────┼─────────────────────────────────────┐
│               Go Backend Server (Port 8080)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  HTTP API    │  │  WebSocket   │  │  Health Check    │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────────────┘  │
│         │                 │                                  │
│  ┌──────┴─────────────────┴────────────────────────────┐   │
│  │              Request Handlers                        │   │
│  │  • Workspace Indexing  • Code Search                │   │
│  │  • Code Generation     • Chat                       │   │
│  └──────┬──────────────┬──────────────┬────────────────┘   │
│         │              │              │                     │
│  ┌──────┴────┐  ┌──────┴────┐  ┌─────┴────────┐          │
│  │ SQLite DB │  │  Vector   │  │  LLM Client  │          │
│  │ (Code     │  │ Database  │  │  (OpenAI     │          │
│  │ Structure)│  │(chromem-go)│  │   GPT-4)     │          │
│  └───────────┘  └───────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────┘
```

## Features

### 🗄️ Database Layer
- **SQLite** for structured code storage (classes, methods, fields, files)
- **chromem-go** for vector embeddings and semantic search
- Workspace indexing for Java, Gherkin (.feature), and XML (pom.xml)
- AST-based parsing with tree-sitter

### 🧠 AI Integration
- **OpenAI GPT-4** for code generation
- **OpenAI text-embedding-3-small** for embeddings (1536 dimensions)
- Context-aware prompt building
- Cost estimation and token counting
- Streaming support for real-time responses

### 🌐 API Endpoints

#### Workspace Management
- `POST /api/v1/workspace/index` - Index a workspace
- `GET /api/v1/workspace/stats` - Get workspace statistics

#### Semantic Search
- `POST /api/v1/search/code` - Search all code
- `POST /api/v1/search/pageobjects` - Search page objects
- `POST /api/v1/search/tests` - Search test methods

#### Code Generation
- `POST /api/v1/generate/pageobject` - Generate page object
- `POST /api/v1/generate/test` - Generate test case
- `POST /api/v1/generate/fix` - Fix broken code

#### Chat & Database
- `POST /api/v1/chat/message` - Chat with AI assistant
- `GET /api/v1/db/classes` - Get all classes
- `GET /api/v1/db/classes/:id` - Get specific class
- `GET /api/v1/db/methods/:classId` - Get methods by class

#### Real-time Streaming
- `GET /ws/stream` - WebSocket for streaming responses

#### Health
- `GET /health` - Server health check

## Setup

### Prerequisites

- Go 1.23+
- OpenAI API key (for GPT-4 and embeddings)

### Installation

1. **Clone the repository**
```bash
git clone <repo-url>
cd test-automation-copilot/copilot-core
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set environment variables**
```bash
export OPENAI_API_KEY="your-api-key-here"
```

4. **Build the server**
```bash
go build -o test-copilot-server ./cmd/server
```

### Running the Server

```bash
./test-copilot-server \
  --port 8080 \
  --db ./data/copilot.db \
  --vector ./data/vectordb
```

**Command-line options:**
- `--port` - Server port (default: 8080)
- `--db` - SQLite database path (default: ./data/copilot.db)
- `--vector` - Vector database path (default: ./data/vectordb)

The server will:
1. Initialize SQLite database and schema
2. Initialize embedding service (OpenAI)
3. Initialize vector database (chromem-go)
4. Start HTTP/WebSocket server
5. Display all available endpoints

## Usage Examples

### 1. Index a Workspace

```bash
curl -X POST http://localhost:8080/api/v1/workspace/index \
  -H "Content-Type: application/json" \
  -d '{"workspacePath": "/path/to/your/project"}'
```

**Response:**
```json
{
  "message": "Indexing started",
  "path": "/path/to/your/project"
}
```

### 2. Get Workspace Statistics

```bash
curl http://localhost:8080/api/v1/workspace/stats
```

**Response:**
```json
{
  "files": 150,
  "classes": 75,
  "pageObjects": 25,
  "testMethods": 120,
  "vectorDocs": 200,
  "collectionName": "test-copilot"
}
```

### 3. Semantic Code Search

```bash
curl -X POST http://localhost:8080/api/v1/search/code \
  -H "Content-Type: application/json" \
  -d '{
    "query": "login page object with username and password",
    "limit": 5
  }'
```

**Response:**
```json
{
  "results": [
    {
      "id": "1",
      "name": "LoginPage",
      "content": "public class LoginPage { ... }",
      "filePath": "src/test/java/pages/LoginPage.java",
      "type": "class",
      "similarity": 0.95
    }
  ],
  "count": 5
}
```

### 4. Generate Page Object

```bash
curl -X POST http://localhost:8080/api/v1/generate/pageobject \
  -H "Content-Type: application/json" \
  -d '{
    "spec": "Login page with username, password fields and submit button",
    "elements": [
      {"name": "usernameField", "locatorType": "id", "locatorValue": "username"},
      {"name": "passwordField", "locatorType": "id", "locatorValue": "password"},
      {"name": "loginButton", "locatorType": "css", "locatorValue": "button[type=submit]"}
    ]
  }'
```

**Response:**
```json
{
  "code": "package pages;\n\nimport org.openqa.selenium.WebDriver;\n...",
  "tokensUsed": 450,
  "model": "gpt-4-turbo-preview",
  "finishReason": "stop"
}
```

### 5. Chat with AI

```bash
curl -X POST http://localhost:8080/api/v1/chat/message \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How do I create a page object for the checkout page?"
  }'
```

**Response:**
```json
{
  "response": "To create a page object for the checkout page...",
  "tokensUsed": 320,
  "model": "gpt-4-turbo-preview",
  "finishReason": "stop"
}
```

### 6. WebSocket Streaming

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/stream');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'request',
    payload: {
      action: 'pageobject',
      spec: 'Login page',
      elements: [...]
    }
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  switch (message.type) {
    case 'chunk':
      console.log('Chunk:', message.payload.chunk);
      break;
    case 'complete':
      console.log('Complete:', message.payload.content);
      break;
    case 'error':
      console.error('Error:', message.payload.error);
      break;
  }
};
```

## Development

### Project Structure

```
copilot-core/
├── cmd/
│   └── server/          # Server entry point
│       └── main.go
├── pkg/
│   ├── database/        # SQLite database layer
│   ├── embeddings/      # Vector database & embeddings
│   ├── indexer/         # Workspace indexing
│   ├── llm/             # LLM client & context builder
│   ├── parser/          # Code parsers (Java, Gherkin, XML)
│   ├── prompts/         # Prompt templates
│   └── server/          # HTTP/WebSocket server
├── data/                # Runtime data (created automatically)
│   ├── copilot.db       # SQLite database
│   └── vectordb/        # Vector database
├── go.mod
└── go.sum
```

### Building

```bash
# Build server
go build -o test-copilot-server ./cmd/server

# Run tests
go test ./...

# Build with race detector
go build -race -o test-copilot-server ./cmd/server

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o test-copilot-server-linux ./cmd/server
GOOS=windows GOARCH=amd64 go build -o test-copilot-server.exe ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o test-copilot-server-mac ./cmd/server
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/database
go test ./pkg/embeddings
go test ./pkg/llm

# Run benchmarks
go test -bench=. ./pkg/llm
```

### Testing with OpenAI

Tests requiring OpenAI API key are skipped if `OPENAI_API_KEY` is not set:

```bash
export OPENAI_API_KEY="your-key"
go test ./pkg/embeddings
go test ./pkg/llm
```

## Cost Estimation

### OpenAI Pricing (as of 2024)

**GPT-4 Turbo:**
- Input: $0.01 per 1K tokens
- Output: $0.03 per 1K tokens

**Text Embedding 3 Small:**
- $0.00013 per 1K tokens

### Typical Usage Costs

**Light usage** (50 requests/day):
- ~150K tokens/month
- Cost: ~$10-15/month

**Medium usage** (200 requests/day):
- ~600K tokens/month
- Cost: ~$40-60/month

**Heavy usage** (500 requests/day):
- ~1.5M tokens/month
- Cost: ~$90-120/month

## Performance

### Indexing Performance

- **Java files**: ~100 files/second
- **Gherkin files**: ~150 files/second
- **XML files**: ~200 files/second

### Search Performance

- **Semantic search**: ~50-100ms per query
- **Database queries**: ~10-20ms per query

### Code Generation

- **Without streaming**: 5-15 seconds
- **With streaming**: First chunk in <1 second
- **Token generation**: ~40-60 tokens/second

## Troubleshooting

### Server won't start

```bash
# Check if port is in use
lsof -i :8080

# Use different port
./test-copilot-server --port 9090
```

### OpenAI API errors

```bash
# Verify API key is set
echo $OPENAI_API_KEY

# Check API key validity
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

### Database errors

```bash
# Remove corrupted database
rm -rf ./data/copilot.db
rm -rf ./data/vectordb

# Server will recreate on next start
```

### Memory issues

```bash
# Limit workspace size during indexing
# Index only specific directories
# Increase available memory
```

## Contributing

### Adding New Endpoints

1. Add handler in `pkg/server/handlers.go`
2. Register route in `pkg/server/server.go`
3. Update client in VS Code extension
4. Add tests
5. Update documentation

### Adding New Parsers

1. Create parser in `pkg/parser/`
2. Integrate with indexer
3. Add database schema updates
4. Add tests

## License

See LICENSE file in repository root.

## Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check documentation in `/docs`
- Review example code in `/examples`
