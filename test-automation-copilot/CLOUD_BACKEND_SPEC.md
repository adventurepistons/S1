# Cloud Backend API Specification

## Overview

This document specifies the API that the cloud backend must implement to work with the Test Automation Copilot local backend.

The cloud backend is responsible for:
- Receiving preprocessed context from local backend
- Applying proprietary prompt engineering
- Calling OpenAI GPT-4 with assembled prompts
- Streaming responses back to local backend
- Tracking usage and billing

## Base URL

Production: `https://api.testcopilot.ai`
Staging: `https://staging-api.testcopilot.ai`

## Authentication

All requests (except `/health`) require authentication via Bearer token:

```http
Authorization: Bearer <user-api-key>
```

**API Key Format**: `tc_prod_<random-32-chars>`

Example: `tc_prod_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`

## Endpoints

### 1. Health Check

Check if the API is available.

```http
GET /health
```

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": 1700000000,
  "version": "1.0.0"
}
```

**Response** (503 Service Unavailable):
```json
{
  "status": "unhealthy",
  "timestamp": 1700000000,
  "error": "OpenAI API unavailable"
}
```

---

### 2. Generate Code (Non-Streaming)

Generate code based on preprocessed context.

```http
POST /v1/generate
Authorization: Bearer <api-key>
Content-Type: application/json
```

**Request Body**:
```json
{
  "action": "pageobject",
  "userQuery": "Login page with username and password fields",
  "spec": "Login page with username and password fields",
  "relevantCode": [
    {
      "name": "HomePage",
      "type": "class",
      "content": "public class HomePage {...}",
      "filePath": "src/test/java/pages/HomePage.java",
      "similarity": 0.85
    }
  ],
  "pageObjects": [
    {
      "name": "HomePage",
      "filePath": "src/test/java/pages/HomePage.java",
      "methods": ["clickLogin", "isLoaded"]
    }
  ],
  "testMethods": [],
  "workspace": {
    "totalClasses": 50,
    "totalTests": 120,
    "pageObjectCount": 15
  },
  "framework": "selenium-java",
  "testRunner": "testng",
  "elements": [
    {
      "name": "usernameField",
      "locatorType": "id",
      "locatorValue": "username"
    },
    {
      "name": "passwordField",
      "locatorType": "id",
      "locatorValue": "password"
    }
  ]
}
```

**Response** (200 OK):
```json
{
  "code": "package pages;\n\nimport org.openqa.selenium.WebDriver;\nimport org.openqa.selenium.WebElement;\nimport org.openqa.selenium.support.FindBy;\nimport org.openqa.selenium.support.PageFactory;\n\npublic class LoginPage {\n    private WebDriver driver;\n\n    @FindBy(id = \"username\")\n    private WebElement usernameField;\n\n    @FindBy(id = \"password\")\n    private WebElement passwordField;\n\n    public LoginPage(WebDriver driver) {\n        this.driver = driver;\n        PageFactory.initElements(driver, this);\n    }\n\n    public void enterUsername(String username) {\n        usernameField.sendKeys(username);\n    }\n\n    public void enterPassword(String password) {\n        passwordField.sendKeys(password);\n    }\n}",
  "tokensUsed": 450,
  "model": "gpt-4-turbo-preview",
  "finishReason": "stop"
}
```

**Response** (400 Bad Request):
```json
{
  "error": "Missing required field: action"
}
```

**Response** (401 Unauthorized):
```json
{
  "error": "Invalid API key"
}
```

**Response** (429 Too Many Requests):
```json
{
  "error": "Rate limit exceeded. Try again in 60 seconds.",
  "retryAfter": 60
}
```

**Response** (500 Internal Server Error):
```json
{
  "error": "Failed to generate code: OpenAI API error"
}
```

---

### 3. Generate Code (Streaming)

Generate code with real-time streaming responses.

```http
POST /v1/generate/stream
Authorization: Bearer <api-key>
Content-Type: application/json
Accept: text/event-stream
```

**Request Body**: Same as non-streaming endpoint

**Response** (200 OK, streaming):

```
data: {"type":"chunk","content":"package"}

data: {"type":"chunk","content":" pages;\n\n"}

data: {"type":"chunk","content":"import org"}

data: {"type":"chunk","content":".openqa.selenium"}

...

data: {"type":"complete","result":{"code":"package pages;...","tokensUsed":450,"model":"gpt-4-turbo-preview","finishReason":"stop"}}
```

**Message Types**:

1. **Chunk**:
```json
{"type": "chunk", "content": "text chunk"}
```

2. **Complete**:
```json
{
  "type": "complete",
  "result": {
    "code": "full generated code",
    "tokensUsed": 450,
    "model": "gpt-4-turbo-preview",
    "finishReason": "stop"
  }
}
```

3. **Error**:
```json
{"type": "error", "error": "error message"}
```

---

### 4. Get Usage Statistics

Get usage statistics for the authenticated user.

```http
GET /v1/usage
Authorization: Bearer <api-key>
```

**Response** (200 OK):
```json
{
  "userId": "user_123456",
  "plan": "pro",
  "period": {
    "start": 1698796800,
    "end": 1701388800,
    "current": 1700000000
  },
  "usage": {
    "totalRequests": 1250,
    "totalTokens": 450000,
    "estimatedCost": 18.50,
    "requestsThisMonth": 150,
    "tokensThisMonth": 52000,
    "requestsByAction": {
      "pageobject": 80,
      "test": 45,
      "chat": 20,
      "fix": 5
    }
  },
  "limits": {
    "requestsPerMonth": 1000,
    "tokensPerMonth": 500000,
    "requestsRemaining": 850,
    "tokensRemaining": 448000
  }
}
```

---

### 5. List Models (Optional)

Get available models and their capabilities.

```http
GET /v1/models
Authorization: Bearer <api-key>
```

**Response** (200 OK):
```json
{
  "models": [
    {
      "id": "gpt-4-turbo-preview",
      "name": "GPT-4 Turbo",
      "contextWindow": 128000,
      "maxOutputTokens": 4096,
      "costPer1kInputTokens": 0.01,
      "costPer1kOutputTokens": 0.03,
      "capabilities": ["code", "chat", "reasoning"],
      "recommended": true
    },
    {
      "id": "gpt-3.5-turbo",
      "name": "GPT-3.5 Turbo",
      "contextWindow": 16384,
      "maxOutputTokens": 4096,
      "costPer1kInputTokens": 0.0005,
      "costPer1kOutputTokens": 0.0015,
      "capabilities": ["code", "chat"],
      "recommended": false
    }
  ]
}
```

## Request/Response Formats

### ContextPayload (Input to Cloud)

The local backend sends this preprocessed payload:

```typescript
interface ContextPayload {
  // User intent
  action: "pageobject" | "test" | "chat" | "fix";
  userQuery: string;
  spec?: string;

  // Preprocessed context
  relevantCode?: CodeSnippet[];
  pageObjects?: PageObject[];
  testMethods?: TestMethod[];

  // Workspace metadata
  framework?: string;
  testRunner?: string;
  workspace?: WorkspaceInfo;

  // Action-specific data
  elements?: Element[];       // For page object generation
  errorInfo?: ErrorContext;   // For code fixing
}

interface CodeSnippet {
  name: string;
  type: string;        // "class", "method", "field"
  content: string;
  filePath: string;
  similarity?: number; // 0.0 to 1.0
}

interface PageObject {
  name: string;
  filePath: string;
  methods: string[];
}

interface TestMethod {
  name: string;
  className: string;
  annotations?: string[];
}

interface WorkspaceInfo {
  totalClasses: number;
  totalTests: number;
  pageObjectCount: number;
}

interface Element {
  name: string;
  locatorType: string;  // "id", "css", "xpath", "name"
  locatorValue: string;
}

interface ErrorContext {
  brokenCode: string;
  errorMessage: string;
  errorType?: string;
}
```

### GenerationResult (Output from Cloud)

```typescript
interface GenerationResult {
  code: string;
  tokensUsed: number;
  model: string;
  finishReason: string; // "stop", "length", "content_filter"
}
```

## Rate Limits

| Plan       | Requests/Minute | Requests/Month | Tokens/Month |
|------------|-----------------|----------------|--------------|
| Free       | 5               | 100            | 50,000       |
| Pro        | 20              | 1,000          | 500,000      |
| Enterprise | 100             | Unlimited      | Unlimited    |

**Rate Limit Headers**:
```http
X-RateLimit-Limit: 20
X-RateLimit-Remaining: 15
X-RateLimit-Reset: 1700000060
```

## Error Codes

| Code | Meaning                    | Resolution                          |
|------|----------------------------|-------------------------------------|
| 400  | Bad Request                | Check request format                |
| 401  | Unauthorized               | Provide valid API key               |
| 403  | Forbidden                  | Upgrade plan or wait for reset      |
| 429  | Too Many Requests          | Respect rate limits                 |
| 500  | Internal Server Error      | Retry after delay                   |
| 503  | Service Unavailable        | Backend or OpenAI API down          |

## Webhooks (Optional)

For enterprise customers, support webhooks for events:

```http
POST <customer-webhook-url>
Content-Type: application/json
X-Webhook-Signature: <hmac-signature>
```

**Event Types**:
- `generation.completed`
- `usage.limit.reached`
- `subscription.updated`

**Example Payload**:
```json
{
  "event": "generation.completed",
  "timestamp": 1700000000,
  "userId": "user_123456",
  "data": {
    "action": "pageobject",
    "tokensUsed": 450,
    "success": true
  }
}
```

## Implementation Recommendations

### Backend Stack Options

**1. Serverless (Recommended for Scalability)**
- AWS Lambda + API Gateway
- GCP Cloud Functions + Cloud Run
- Azure Functions

**2. Traditional Server**
- Node.js + Express
- Python + FastAPI
- Go + Gin

### Database
- **User Data**: PostgreSQL or MongoDB
- **Usage Tracking**: Time-series DB (InfluxDB, TimescaleDB)
- **Caching**: Redis for response caching

### Prompt Management
- Store prompt templates in database
- Version control for prompts
- A/B testing framework

### Example Prompt Template

```
You are an expert test automation engineer. Generate a Selenium Java page object class.

Context:
- Framework: {framework}
- Test Runner: {testRunner}
- Total Page Objects in Workspace: {pageObjectCount}

Similar Examples:
{relevantCode[0].content}

User Request:
{userQuery}

Elements:
{elements.map(e => `- ${e.name}: ${e.locatorType}=${e.locatorValue}`).join('\n')}

Requirements:
1. Use PageFactory pattern
2. Follow naming conventions: {pageObjectNamingConvention}
3. Include constructor with WebDriver parameter
4. Add methods for each element interaction
5. Include JavaDoc comments

Generate the complete Java class:
```

### Streaming Implementation

**Node.js Example**:
```javascript
app.post('/v1/generate/stream', async (req, res) => {
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');

  const prompt = buildPrompt(req.body);

  const stream = await openai.chat.completions.create({
    model: 'gpt-4-turbo-preview',
    messages: [{role: 'user', content: prompt}],
    stream: true
  });

  for await (const chunk of stream) {
    const content = chunk.choices[0]?.delta?.content || '';
    if (content) {
      res.write(`data: ${JSON.stringify({type: 'chunk', content})}\n\n`);
    }
  }

  res.write(`data: ${JSON.stringify({
    type: 'complete',
    result: {code: fullCode, tokensUsed: tokens, ...}
  })}\n\n`);

  res.end();
});
```

## Testing

### Test API Keys
- Test Key: `tc_test_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
- Sandbox Mode: No real OpenAI calls, returns mock data

### Example Test Request

```bash
curl -X POST https://api.testcopilot.ai/v1/generate \
  -H "Authorization: Bearer tc_test_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "pageobject",
    "userQuery": "Login page",
    "spec": "Login page with username and password",
    "framework": "selenium-java",
    "elements": [
      {"name": "username", "locatorType": "id", "locatorValue": "username"}
    ]
  }'
```

## Security Checklist

- [ ] HTTPS only (TLS 1.2+)
- [ ] API key validation on every request
- [ ] Rate limiting per API key
- [ ] Request size limits (max 1MB)
- [ ] Timeout after 30 seconds
- [ ] No logging of user code snippets
- [ ] SQL injection protection
- [ ] CORS configuration
- [ ] DDoS protection
- [ ] Regular security audits

## Monitoring

Track these metrics:
- Request rate per endpoint
- Average response time
- OpenAI API latency
- Error rate by type
- Token usage per user
- Cache hit rate
- Streaming connection duration

## SLA

- **Uptime**: 99.9% (excluding scheduled maintenance)
- **Response Time**: p95 < 3 seconds (non-streaming)
- **Support**: Enterprise customers get 24/7 support

---

**Version**: 1.0.0
**Last Updated**: 2025-11-16
**Maintained By**: Test Automation Copilot Team
