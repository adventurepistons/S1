# Cloud-AI Architecture

## Overview

The Test Automation Copilot uses a **cloud-AI architecture** similar to Cursor IDE and GitHub Copilot. This design protects proprietary prompt engineering IP while keeping preprocessing local for performance and cost efficiency.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      VS Code Extension                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   UI/UX      │  │  CoreClient  │  │  Browser Recorder    │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────────────────┘  │
└─────────┼─────────────────┼──────────────────────────────────────┘
          │                 │ HTTP/WebSocket
┌─────────┼─────────────────┼──────────────────────────────────────┐
│         │  Local Go Backend (Port 8080)                          │
│  ┌──────┴─────────────────┴────────────────────────────────┐    │
│  │              Preprocessing Layer                        │    │
│  │  • Database (SQLite)       • Vector Store (chromem-go)  │    │
│  │  • Semantic Search         • Context Extraction         │    │
│  └──────┬──────────────────────────────────────────────────┘    │
└─────────┼───────────────────────────────────────────────────────┘
          │ HTTPS (ContextPayload)
┌─────────┼───────────────────────────────────────────────────────┐
│         │  Your Cloud Backend (Protected IP)                    │
│  ┌──────┴──────────────────────────────────────────────────┐   │
│  │              Prompt Engineering Layer                    │   │
│  │  • Prompt Templates (Secret)  • Few-shot Examples       │   │
│  │  • Context Assembly           • LLM Orchestration       │   │
│  └──────┬──────────────────────────────────────────────────┘   │
│         │                                                        │
│  ┌──────┴──────────────────────────────────────────────────┐   │
│  │              OpenAI GPT-4 Integration                    │   │
│  │  • Model Selection            • Token Management        │   │
│  │  • Streaming                  • Cost Optimization       │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
```

## Data Flow

### 1. Request Flow (Page Object Generation Example)

```
1. User: "Generate login page object"
   ↓
2. Extension: Capture request + UI state
   ↓
3. Local Backend:
   - Index workspace (if needed)
   - Extract similar page objects from vector DB
   - Extract relevant code snippets
   - Detect framework/test runner
   - Build ContextPayload (NO PROMPTS)
   ↓
4. Cloud Backend (YOUR SERVER):
   - Receive ContextPayload
   - Select appropriate prompt template (HIDDEN)
   - Assemble few-shot examples (HIDDEN)
   - Build final prompt to GPT-4 (HIDDEN)
   - Call OpenAI API
   - Stream response back
   ↓
5. Local Backend: Forward stream to extension
   ↓
6. Extension: Display code in editor
```

### 2. What Stays Local

**Local Backend** (`test-automation-copilot/copilot-core/`):
- ✅ Database (SQLite) - code structure
- ✅ Vector Database (chromem-go) - embeddings
- ✅ Semantic Search - finding relevant code
- ✅ Context Extraction - preprocessing user request
- ✅ Workspace Indexing - parsing Java/Gherkin/XML
- ❌ Prompt Templates - NOT LOCAL
- ❌ LLM Calls - NOT LOCAL

### 3. What Goes to Cloud

**Cloud Backend** (to be implemented separately):
- ✅ Prompt Engineering (proprietary)
- ✅ Few-shot Example Selection
- ✅ Context Assembly Strategies
- ✅ GPT-4 Integration
- ✅ Model Selection Logic
- ✅ Streaming Logic
- ✅ Cost Optimization
- ✅ Usage Tracking

## Implementation Details

### Local Backend Components

#### 1. Cloud Client (`pkg/cloud/client.go`)

```go
type CloudClient struct {
    baseURL    string  // https://api.testcopilot.ai
    apiKey     string  // User's API key
    httpClient *http.Client
}

// Non-streaming generation
func (c *CloudClient) Generate(ctx context.Context, payload ContextPayload) (*GenerationResult, error)

// Streaming generation
func (c *CloudClient) GenerateStream(ctx context.Context, payload ContextPayload, onChunk func(string) error) (*GenerationResult, error)
```

**Environment Variables:**
- `TESTCOPILOT_CLOUD_URL` - Your cloud backend URL
- `TESTCOPILOT_API_KEY` - User's API key

#### 2. Context Extractor (`pkg/context/extractor.go`)

```go
type ContextExtractor struct {
    db             *database.DB
    semanticSearch *embeddings.SemanticSearch
}

// Extract context for different actions
func (ce *ContextExtractor) ExtractPageObjectContext(ctx, spec, elements) (ContextPayload, error)
func (ce *ContextExtractor) ExtractTestContext(ctx, spec) (ContextPayload, error)
func (ce *ContextExtractor) ExtractChatContext(ctx, message) (ContextPayload, error)
func (ce *ContextExtractor) ExtractFixContext(ctx, code, error, errorType) (ContextPayload, error)
```

**What It Extracts:**
- Similar code from vector database
- Relevant page objects and test methods
- Workspace metadata (total classes, tests, etc.)
- Framework and test runner detection
- Element information (for page objects)
- Error context (for fixes)

#### 3. Context Payload Structure

```go
type ContextPayload struct {
    // User intent
    Action      string  // "pageobject", "test", "chat", "fix"
    UserQuery   string
    Spec        string

    // Code context (preprocessed)
    RelevantCode []CodeSnippet
    PageObjects  []PageObject
    TestMethods  []TestMethod

    // Workspace metadata
    Framework   string  // "selenium-java", "playwright-java"
    TestRunner  string  // "testng", "junit", "cucumber"
    Workspace   WorkspaceInfo

    // Action-specific data
    Elements    []Element       // For page object generation
    ErrorInfo   *ErrorContext   // For code fixing
}
```

### Cloud Backend API (To Implement)

#### Endpoints

**1. Generate Code (Non-streaming)**
```http
POST /v1/generate
Authorization: Bearer <user-api-key>
Content-Type: application/json

{
  "action": "pageobject",
  "userQuery": "Login page with username and password",
  "spec": "...",
  "relevantCode": [...],
  "pageObjects": [...],
  "framework": "selenium-java",
  "elements": [...]
}

Response:
{
  "code": "package pages;\n\npublic class LoginPage {...}",
  "tokensUsed": 450,
  "model": "gpt-4-turbo-preview",
  "finishReason": "stop"
}
```

**2. Generate Code (Streaming)**
```http
POST /v1/generate/stream
Authorization: Bearer <user-api-key>
Content-Type: application/json

[Same payload as above]

Response (Server-Sent Events):
data: {"type":"chunk","content":"package"}
data: {"type":"chunk","content":" pages;\n\n"}
data: {"type":"chunk","content":"public class"}
...
data: {"type":"complete","result":{"code":"...","tokensUsed":450,...}}
```

**3. Health Check**
```http
GET /health

Response:
{
  "status": "healthy",
  "timestamp": 1700000000
}
```

**4. Usage Statistics**
```http
GET /v1/usage
Authorization: Bearer <user-api-key>

Response:
{
  "totalRequests": 1250,
  "totalTokens": 450000,
  "estimatedCost": 18.50,
  "requestsThisMonth": 150,
  "tokensThisMonth": 52000
}
```

## Benefits of This Architecture

### 1. IP Protection
- ✅ **Prompt templates stay secret** - your competitive advantage
- ✅ **Few-shot examples hidden** - refined over time
- ✅ **Context assembly logic protected** - proprietary algorithms
- ✅ Users can't inspect network calls to see your prompts
- ✅ Similar to how Cursor and GitHub Copilot work

### 2. Easy Updates
- ✅ **Update prompts without client updates** - improve quality instantly
- ✅ **A/B test different prompt strategies** - optimize conversion
- ✅ **Roll out new models** - GPT-5 when available
- ✅ **Fix bugs in prompt logic** - without user intervention

### 3. Cost Optimization
- ✅ **Local preprocessing reduces cloud costs** - only send needed context
- ✅ **Centralized token management** - optimize across all users
- ✅ **Caching at cloud level** - reuse responses
- ✅ **Model selection logic** - use cheaper models when possible

### 4. Better User Experience
- ✅ **Faster indexing** - done locally with tree-sitter
- ✅ **Offline search** - semantic search works without internet
- ✅ **Reduced latency** - preprocessing doesn't hit network
- ✅ **Streaming responses** - incremental updates

## Security Considerations

### Authentication
- User authenticates with API key (stored in extension settings)
- Cloud backend validates API key on every request
- Rate limiting per API key

### Data Privacy
- Code context sent to cloud (necessary for generation)
- Cloud backend should NOT log code snippets
- Consider encryption in transit (HTTPS) and at rest
- Comply with data retention policies

### API Key Management
- Store in VS Code secure storage (not plain text)
- Support key rotation
- Provide key revocation mechanism

## Cost Model

### For Users
- **Free Tier**: 100 requests/month
- **Pro Tier**: $20/month - 1000 requests
- **Enterprise**: Custom pricing

### Your Costs (Cloud Backend)
- **OpenAI API**: ~$0.02-0.05 per request (GPT-4)
- **Hosting**: AWS/GCP - ~$100-500/month
- **Infrastructure**: Database, caching, monitoring
- **Total**: Estimate $0.05-0.10 per request all-in

### Profit Margins
- Free users: Loss leader for conversion
- Pro users: $20/month - 1000 requests = $0.02/request → ~$5-10/month profit
- Enterprise: 80%+ margin with volume

## Implementation Checklist

### ✅ Completed
- [x] Cloud client implementation
- [x] Context extractor
- [x] Update server handlers
- [x] Update WebSocket streaming
- [x] Health checks for cloud API
- [x] Environment variable configuration

### 🔲 TODO
- [ ] Cloud backend implementation (separate project)
- [ ] Prompt template library
- [ ] Few-shot example database
- [ ] User authentication system
- [ ] API key management
- [ ] Usage tracking and billing
- [ ] Extension configuration UI
- [ ] Integration tests
- [ ] Documentation updates

## Next Steps

1. **Set up Cloud Backend**
   - Choose hosting (AWS Lambda, GCP Cloud Run, etc.)
   - Implement `/v1/generate` endpoint
   - Implement streaming endpoint
   - Add authentication

2. **Create Prompt Library**
   - Page object generation templates
   - Test case generation templates
   - Code fixing templates
   - Chat conversation templates

3. **Build User Management**
   - API key generation
   - Usage tracking
   - Rate limiting
   - Billing integration

4. **Update Extension**
   - Add cloud URL configuration
   - Add API key input UI
   - Show usage statistics
   - Handle rate limits gracefully

## Comparison: Before vs After

### Before (Local LLM)
```
Extension
    ↓
Local Backend
    ├── Database ✓
    ├── Vector DB ✓
    ├── Prompt Templates ❌ (exposed)
    └── LLM Client → OpenAI
```

**Problems:**
- Prompts visible in local code
- Can't update without client update
- Hard to A/B test
- No usage tracking

### After (Cloud AI)
```
Extension
    ↓
Local Backend
    ├── Database ✓
    ├── Vector DB ✓
    └── Context Extraction ✓
           ↓
Cloud Backend (Your Server)
    ├── Prompt Templates ✓ (hidden)
    ├── Few-shot Examples ✓ (hidden)
    └── LLM Client → OpenAI ✓
```

**Benefits:**
- Prompts protected
- Easy updates
- A/B testing
- Usage tracking
- Cost optimization

## FAQ

**Q: Why not just call OpenAI directly from extension?**
A: You'd expose your API key and prompts. Users could extract and reuse them.

**Q: Can users see my prompts by inspecting network traffic?**
A: No - they only see ContextPayload (preprocessed data), not the final prompt.

**Q: How is this different from GitHub Copilot?**
A: Same architecture! Copilot also does preprocessing locally, sends to cloud.

**Q: What if cloud backend is down?**
A: Local preprocessing still works (indexing, search). Generation requires cloud.

**Q: Can I self-host the cloud backend?**
A: Yes! Deploy on your own infrastructure for full control.

## References

- **Cursor Architecture**: https://cursor.sh/
- **GitHub Copilot**: https://github.com/features/copilot
- **OpenAI Best Practices**: https://platform.openai.com/docs/guides/production-best-practices

---

**Status**: ✅ Architecture implemented in local backend. Cloud backend needs implementation.

**Last Updated**: 2025-11-16
