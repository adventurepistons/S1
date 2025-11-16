# Test Automation Copilot - Quick Start Guide

Get up and running with the Test Automation Copilot in minutes.

## Prerequisites

- **Node.js** 18+ (for VS Code extension)
- **Go** 1.23+ (for backend server)
- **OpenAI API Key** (for GPT-4 and embeddings)
- **VS Code** 1.80+ (for extension development)

## Step 1: Setup Backend Server

### 1.1 Install Dependencies

```bash
cd test-automation-copilot/copilot-core
go mod download
```

### 1.2 Set Environment Variables

```bash
export OPENAI_API_KEY="sk-your-api-key-here"
```

**Windows (PowerShell):**
```powershell
$env:OPENAI_API_KEY="sk-your-api-key-here"
```

### 1.3 Build the Server

```bash
go build -o test-copilot-server ./cmd/server
```

**Windows:**
```bash
go build -o test-copilot-server.exe ./cmd/server
```

### 1.4 Run the Server

```bash
./test-copilot-server --port 8080
```

You should see:
```
🚀 Starting Test Automation Copilot Server...
📊 Initializing database...
🔗 Initializing OpenAI embedding service...
💾 Initializing vector database...
🔍 Initializing semantic search...
🤖 Initializing OpenAI GPT-4 client...
✅ LLM client configured successfully
🧠 Initializing context builder...
🌐 Creating HTTP/WebSocket server on port 8080...
✨ Server started successfully!
```

### 1.5 Verify Server is Running

```bash
curl http://localhost:8080/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "timestamp": 1700000000,
  "services": {
    "database": true,
    "vectorStore": true,
    "llm": true
  }
}
```

## Step 2: Setup VS Code Extension

### 2.1 Install Dependencies

```bash
cd test-automation-copilot
npm install
```

### 2.2 Open in VS Code

```bash
code .
```

### 2.3 Run Extension in Development Mode

1. Press `F5` or click "Run > Start Debugging"
2. This opens a new "Extension Development Host" window
3. The backend server will start automatically

### 2.4 Verify Extension is Running

In the Extension Development Host window:
1. Open Command Palette (`Cmd+Shift+P` or `Ctrl+Shift+P`)
2. Type "Test Copilot"
3. You should see available commands

## Step 3: Index Your Workspace

### 3.1 Using the Extension

1. Open your test automation project in VS Code
2. Open Command Palette (`Cmd+Shift+P`)
3. Run: `Test Copilot: Index Workspace`
4. Wait for indexing to complete

### 3.2 Using API Directly

```bash
curl -X POST http://localhost:8080/api/v1/workspace/index \
  -H "Content-Type: application/json" \
  -d '{"workspacePath": "/path/to/your/test/project"}'
```

### 3.3 Check Index Status

```bash
curl http://localhost:8080/api/v1/workspace/stats
```

You should see indexed files, classes, and page objects.

## Step 4: Try Code Generation

### 4.1 Generate a Page Object

1. Open Command Palette
2. Run: `Test Copilot: Generate Page Object`
3. Enter specification: "Login page with username and password"
4. Watch as code streams in real-time!

### 4.2 Using API Directly

```bash
curl -X POST http://localhost:8080/api/v1/generate/pageobject \
  -H "Content-Type: application/json" \
  -d '{
    "spec": "Login page with username, password and submit button",
    "elements": [
      {"name": "usernameField", "locatorType": "id", "locatorValue": "username"},
      {"name": "passwordField", "locatorType": "id", "locatorValue": "password"},
      {"name": "submitButton", "locatorType": "css", "locatorValue": "button[type=submit]"}
    ]
  }'
```

## Step 5: Use Semantic Search

### 5.1 Search for Page Objects

```bash
curl -X POST http://localhost:8080/api/v1/search/pageobjects \
  -H "Content-Type: application/json" \
  -d '{
    "query": "login page with credentials",
    "limit": 5
  }'
```

### 5.2 Search for Tests

```bash
curl -X POST http://localhost:8080/api/v1/search/tests \
  -H "Content-Type: application/json" \
  -d '{
    "query": "successful login test",
    "limit": 5
  }'
```

### 5.3 Search All Code

```bash
curl -X POST http://localhost:8080/api/v1/search/code \
  -H "Content-Type: application/json" \
  -d '{
    "query": "page factory pattern",
    "limit": 10
  }'
```

## Step 6: Chat with AI

### 6.1 Using Extension

1. Click the "Test Copilot Chat" icon in the sidebar
2. Type your question: "How do I create a data-driven test?"
3. Get context-aware answers based on your codebase

### 6.2 Using API

```bash
curl -X POST http://localhost:8080/api/v1/chat/message \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How do I create a page object for the checkout page?"
  }'
```

## Step 7: Fix Broken Code

### 7.1 Using API

```bash
curl -X POST http://localhost:8080/api/v1/generate/fix \
  -H "Content-Type: application/json" \
  -d '{
    "code": "public void login() { driver.findElement(By.id(\"user\")).click(); }",
    "error": "ElementNotInteractableException: element not visible"
  }'
```

The AI will analyze the error and provide a fix with proper waits.

## Common Issues

### Issue: Server won't start

**Solution:**
```bash
# Check if port 8080 is in use
lsof -i :8080  # Mac/Linux
netstat -ano | findstr :8080  # Windows

# Use different port
./test-copilot-server --port 9090
```

### Issue: OpenAI API errors

**Solution:**
```bash
# Verify API key
echo $OPENAI_API_KEY

# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

### Issue: Extension can't find server binary

**Solution:**
```bash
# Build in correct location
cd copilot-core
go build -o ../bin/test-copilot-server ./cmd/server

# Or update CoreClient.ts to point to your build location
```

### Issue: Slow indexing

**Solution:**
- Index only specific directories
- Exclude build/target folders
- Use SSD for better performance

### Issue: Out of memory during indexing

**Solution:**
- Index in smaller batches
- Increase Node.js memory: `NODE_OPTIONS="--max-old-space-size=4096"`
- Close other applications

## Next Steps

### 1. Explore the UI

- **Sidebar**: Browse indexed classes and tests
- **Chat Panel**: Ask questions about your codebase
- **Recording**: Record browser interactions

### 2. Customize Configuration

Create `.testcopilot` in your project root:

```json
{
  "naming_conventions": {
    "pageObject": "PascalCase ending with 'Page'",
    "testClass": "PascalCase ending with 'Test'"
  },
  "framework": "selenium-java",
  "testRunner": "testng",
  "wait_strategy": "Explicit waits with WebDriverWait, timeout: 10s",
  "locator_priority": ["id", "data-testid", "name", "css", "xpath"]
}
```

### 3. Generate Complete Test Suites

1. Index your workspace
2. Use chat to discuss test scenarios
3. Generate page objects
4. Generate test cases
5. Review and refine

### 4. Integrate with CI/CD

```bash
# Example: Generate tests in CI pipeline
./test-copilot-server &
sleep 5  # Wait for server to start

curl -X POST http://localhost:8080/api/v1/workspace/index \
  -d '{"workspacePath": "'$WORKSPACE'"}'

# Generate tests
curl -X POST http://localhost:8080/api/v1/generate/test \
  -d '{"spec": "Smoke test suite"}'
```

## Learn More

- **[Backend README](copilot-core/README.md)** - Complete API documentation
- **[Architecture](VS_CODE_EXTENSION_ARCHITECTURE.md)** - System design
- **[Implementation Plan](IMPLEMENTATION_PLAN.md)** - Development roadmap
- **[Cursor-Style Features](CURSOR_STYLE_ARCHITECTURE.md)** - Advanced capabilities

## Getting Help

### Documentation
- Check the `/docs` folder for detailed guides
- Review example code in `/examples`
- Read API documentation in backend README

### Troubleshooting
- Enable verbose logging in extension settings
- Check server logs for errors
- Verify OpenAI API quota

### Community
- Open issues on GitHub
- Join discussions
- Contribute improvements

## Cost Management

### Monitor Usage

Check token usage in responses:
```json
{
  "code": "...",
  "tokensUsed": 450,  // Track this
  "model": "gpt-4-turbo-preview"
}
```

### Optimize Costs

1. **Use streaming** for better UX without extra cost
2. **Cache responses** for repeated queries
3. **Adjust temperature** (lower = cheaper, more deterministic)
4. **Limit context** to essential code only
5. **Use batch operations** when possible

### Typical Costs

- **Per page object**: ~$0.02-0.05
- **Per test case**: ~$0.03-0.08
- **Per chat message**: ~$0.01-0.03
- **Per workspace index**: ~$0.10-0.50

**Monthly estimates:**
- Light usage (10 req/day): ~$10
- Medium usage (50 req/day): ~$40
- Heavy usage (200 req/day): ~$150

## Success Tips

1. **Start small** - Index a single module first
2. **Use semantic search** - Find examples before generating
3. **Review generated code** - AI is a tool, not replacement
4. **Iterate** - Refine specifications for better results
5. **Learn patterns** - Understand what prompts work best
6. **Keep context** - Well-indexed code = better generation

## Ready to Go!

You now have a fully functional AI-powered test automation assistant. Start generating code, searching semantically, and chatting with AI about your test framework!

Happy testing! 🚀
