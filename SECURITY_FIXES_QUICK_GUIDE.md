# 🔒 SECURITY FIXES - QUICK IMPLEMENTATION GUIDE

## ⚠️ CRITICAL FIXES (Do These NOW - 5 hours total)

### Fix #1: SSRF Vulnerability (2 hours)

**File:** `backend/src/api-testing/api-testing.service.ts`

**Add at top of file:**
```typescript
// URL validation for SSRF prevention
const BLOCKED_HOSTS = [
  'localhost', '127.0.0.1', '0.0.0.0', '::1',
  '169.254.169.254',  // AWS/GCP metadata
  'metadata.google.internal',
  /^10\./,
  /^172\.(1[6-9]|2[0-9]|3[01])\./,
  /^192\.168\./,
  /^fd[0-9a-f]{2}:/i  // IPv6 private
];

function validateApiTestUrl(url: string): void {
  let parsed: URL;

  try {
    parsed = new URL(url);
  } catch (e) {
    throw new Error('Invalid URL format');
  }

  // Block non-HTTP(S) protocols
  if (!['http:', 'https:'].includes(parsed.protocol)) {
    throw new Error('Only HTTP and HTTPS protocols are allowed');
  }

  // Block private IPs and localhost
  const hostname = parsed.hostname.toLowerCase();
  for (const blocked of BLOCKED_HOSTS) {
    if (typeof blocked === 'string') {
      if (hostname === blocked) {
        throw new Error(`Access to ${hostname} is forbidden (SSRF protection)`);
      }
    } else if (blocked.test(hostname)) {
      throw new Error(`Access to private IP ranges is forbidden (SSRF protection)`);
    }
  }

  // Require HTTPS in production
  if (process.env.NODE_ENV === 'production' && parsed.protocol !== 'https:') {
    throw new Error('Only HTTPS URLs are allowed in production');
  }
}
```

**Update method (line 54):**
```typescript
async executeRestTest(test: ApiTestDefinition): Promise<ApiTestResult> {
  const startTime = Date.now();

  // ✅ VALIDATE URL BEFORE MAKING REQUEST
  validateApiTestUrl(test.url);

  try {
    // ... rest of method
  }
}
```

**Update method (line 242):**
```typescript
async generateTestsFromOpenAPI(specUrl: string): Promise<ApiTestDefinition[]> {
  // ✅ VALIDATE URL BEFORE FETCHING
  validateApiTestUrl(specUrl);

  try {
    // ... rest of method
  }
}
```

---

### Fix #2: XSS Vulnerability (1 hour)

**Step 1:** Install DOMPurify
```bash
cd test-automation-copilot
npm install dompurify @types/dompurify
```

**Step 2:** Update `src/ui/ChatPanelProvider.ts`

**Add import at top:**
```typescript
import DOMPurify from 'dompurify';
```

**Replace line 231:**
```typescript
// OLD (vulnerable):
contentDiv.innerHTML = marked.parse(content);

// NEW (secure):
contentDiv.innerHTML = DOMPurify.sanitize(marked.parse(content), {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'code', 'pre', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'blockquote'],
  ALLOWED_ATTR: ['href', 'class', 'id'],
  FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input'],
  FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover']
});
```

**Replace line 276:**
```typescript
// OLD:
contentDiv.innerHTML = marked.parse(streamingContent);

// NEW:
contentDiv.innerHTML = DOMPurify.sanitize(marked.parse(streamingContent), {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'code', 'pre', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'blockquote'],
  ALLOWED_ATTR: ['href', 'class', 'id'],
  FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input'],
  FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover']
});
```

**Replace line 301:**
```typescript
// OLD:
contentDiv.innerHTML = marked.parse(streamingContent);

// NEW:
contentDiv.innerHTML = DOMPurify.sanitize(marked.parse(streamingContent), {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'code', 'pre', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'blockquote'],
  ALLOWED_ATTR: ['href', 'class', 'id'],
  FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input'],
  FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover']
});
```

---

### Fix #3: CORS Configuration (1 hour)

**File 1:** `test-automation-copilot/copilot-core/pkg/server/server.go`

**Replace lines 71-77:**
```go
// OLD (insecure):
upgrader: websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true  // ⚠️ DANGER
    },
},

// NEW (secure):
upgrader: websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")

        // Allow VSCode webview origins
        if strings.HasPrefix(origin, "vscode-webview://") {
            return true
        }

        // Allow localhost for development
        allowedOrigins := map[string]bool{
            "http://localhost:3000": true,
            "http://127.0.0.1:3000": true,
            "http://localhost:8080": true,
        }

        if allowedOrigins[origin] {
            return true
        }

        log.Printf("⚠️  Rejected WebSocket origin: %s", origin)
        return false
    },
},
```

**File 2:** `cloud-backend/src/server.ts`

**Replace lines 27-30:**
```typescript
// OLD (insecure):
app.use(cors({
  origin: process.env.CORS_ORIGIN || '*',
  credentials: true
}));

// NEW (secure):
const allowedOrigins = process.env.ALLOWED_ORIGINS
  ? process.env.ALLOWED_ORIGINS.split(',')
  : ['http://localhost:3000', 'http://127.0.0.1:3000'];

app.use(cors({
  origin: (origin, callback) => {
    // Allow requests with no origin (mobile apps, Postman, etc.)
    if (!origin) {
      return callback(null, true);
    }

    if (allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      logger.warn(`⚠️  Rejected CORS origin: ${origin}`);
      callback(new Error('Not allowed by CORS'));
    }
  },
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-API-Key']
}));
```

**Update .env.example:**
```bash
# Add to cloud-backend/.env.example
ALLOWED_ORIGINS=http://localhost:3000,https://your-production-domain.com
```

---

### Fix #4: JWT Secret (30 minutes)

**File:** `cloud-backend/src/config/index.ts`

**Replace lines 29-32:**
```typescript
// OLD (insecure):
jwt: {
  secret: process.env.JWT_SECRET || 'your-secret-key-change-in-production',
  expiresIn: process.env.JWT_EXPIRES_IN || '30d'
}

// NEW (secure):
jwt: {
  secret: (() => {
    const secret = process.env.JWT_SECRET;

    if (!secret) {
      if (process.env.NODE_ENV === 'production') {
        throw new Error('❌ JWT_SECRET environment variable must be set in production!');
      }

      // Development only: generate random secret
      const devSecret = require('crypto').randomBytes(64).toString('hex');
      console.warn('⚠️  Using auto-generated JWT secret - FOR DEVELOPMENT ONLY!');
      console.warn('⚠️  Set JWT_SECRET in production: openssl rand -base64 64');
      return devSecret;
    }

    if (secret.length < 32) {
      throw new Error('❌ JWT_SECRET must be at least 32 characters long');
    }

    return secret;
  })(),
  accessTokenExpiry: process.env.JWT_ACCESS_EXPIRY || '1h',  // Shorter lived
  refreshTokenExpiry: process.env.JWT_REFRESH_EXPIRY || '7d'
}
```

**Generate production secret:**
```bash
# Run this command and add to .env
openssl rand -base64 64

# Add to cloud-backend/.env
JWT_SECRET=<paste-generated-secret-here>
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=7d
```

---

## 🟠 HIGH PRIORITY FIXES (Do This Week - 8 hours)

### Fix #5: Input Validation (3 hours)

**File:** `test-automation-copilot/copilot-core/pkg/server/handlers.go`

**Add constants at top of file:**
```go
const (
    MaxMessageLength    = 10000   // 10KB
    MaxCodeLength       = 100000  // 100KB
    MaxUrlLength        = 2048
    MaxNameLength       = 255
    MaxDescriptionLength = 1000
)
```

**Add validation function:**
```go
// validateInput checks input length and returns error if invalid
func validateInput(value string, maxLength int, fieldName string) error {
    if len(value) == 0 {
        return fmt.Errorf("%s cannot be empty", fieldName)
    }
    if len(value) > maxLength {
        return fmt.Errorf("%s exceeds maximum length of %d characters", fieldName, maxLength)
    }
    return nil
}
```

**Update chatMessage handler (~line 298):**
```go
func (s *Server) chatMessage(c *gin.Context) {
    var req ChatMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ✅ VALIDATE INPUT LENGTH
    if err := validateInput(req.Message, MaxMessageLength, "message"); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ... rest of method
}
```

**Update generatePageObject handler (~line 205):**
```go
func (s *Server) generatePageObject(c *gin.Context) {
    var req GeneratePageObjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ✅ VALIDATE INPUT
    if err := validateInput(req.Spec, MaxCodeLength, "spec"); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ... rest of method
}
```

**Update saveApiTest handler:**
```go
func (s *Server) saveApiTest(c *gin.Context) {
    var req SaveApiTestRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ✅ VALIDATE INPUTS
    if err := validateInput(req.Test.Name, MaxNameLength, "name"); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := validateInput(req.Test.URL, MaxUrlLength, "URL"); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if req.Test.Description != nil {
        if err := validateInput(*req.Test.Description, MaxDescriptionLength, "description"); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
    }

    // ... rest of method
}
```

---

### Fix #6: Path Traversal (1 hour)

**File:** `test-automation-copilot/src/ui/ChatPanelProvider.ts`

**Update requestPageObjectGeneration method (around line 508):**
```typescript
// Add validation function at class level
private validatePath(workspacePath: string, filePath: string): string {
    const resolvedWorkspace = path.resolve(workspacePath);
    const resolvedFile = path.resolve(workspacePath, filePath);

    // Ensure file path is within workspace
    if (!resolvedFile.startsWith(resolvedWorkspace + path.sep)) {
        throw new Error(`Path traversal detected: ${filePath}`);
    }

    // Block hidden files and system files
    if (filePath.includes('/.') || filePath.includes('\\.')) {
        throw new Error(`Access to hidden files is forbidden`);
    }

    return resolvedFile;
}

// Update method around line 518
for (const file of generatedFiles) {
    // ✅ VALIDATE PATH
    const fullPath = this.validatePath(workspacePath, file.filePath);

    // Create directory if doesn't exist
    const dir = path.dirname(fullPath);
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }

    // Write file
    fs.writeFileSync(fullPath, file.content, 'utf8');
    createdCount++;
    // ...
}
```

---

### Fix #7: Insecure Random (1 hour)

**File:** `test-automation-copilot/src/browser/BrowserRecorder.ts`

**Update line 443-444:**
```typescript
// OLD (insecure):
private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

// NEW (secure):
import { randomUUID } from 'crypto';

private generateSessionId(): string {
    return `session_${Date.now()}_${randomUUID()}`;
}
```

**Update line 447-448:**
```typescript
// OLD:
private generatePageId(): string {
    return `page_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

// NEW:
private generatePageId(): string {
    return `page_${Date.now()}_${randomUUID()}`;
}
```

---

### Fix #8: Add Authentication to Go Backend (2 hours)

**File:** `test-automation-copilot/copilot-core/pkg/server/server.go`

**Add middleware function:**
```go
// apiKeyAuth validates API key for local backend
func (s *Server) apiKeyAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get API key from header
        apiKey := c.GetHeader("X-API-Key")

        // For local backend, use environment variable or generate one
        expectedKey := os.Getenv("COPILOT_API_KEY")
        if expectedKey == "" {
            // Development mode: allow without key but log warning
            if os.Getenv("COPILOT_ENV") != "production" {
                log.Println("⚠️  No API key configured - running in development mode")
                c.Next()
                return
            }

            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "API key not configured"
            })
            c.Abort()
            return
        }

        if apiKey != expectedKey {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid API key"
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Update setupRoutes method (~line 92):**
```go
// Apply authentication to API routes
v1 := s.router.Group("/api/v1")

// ✅ ADD AUTHENTICATION MIDDLEWARE
if os.Getenv("COPILOT_ENV") == "production" {
    v1.Use(s.apiKeyAuth())
}

{
    // ... existing routes
}
```

**Update TypeScript client to send API key:**

**File:** `test-automation-copilot/src/api/CoreClient.ts` (~line 81)

```typescript
this.axiosClient = axios.create({
    baseURL: `http://localhost:${this.port}`,
    timeout: 60000,
    headers: {
        'Content-Type': 'application/json',
        // ✅ ADD API KEY
        'X-API-Key': config.get<string>('apiKey') || ''
    }
});
```

---

## 📝 TESTING YOUR FIXES

### Test SSRF Fix:
```bash
# This should be blocked:
curl -X POST http://localhost:8080/api/v1/api-tests/save \
  -H "Content-Type: application/json" \
  -d '{
    "test": {
      "id": "test1",
      "type": "REST",
      "url": "http://169.254.169.254/latest/meta-data/",
      "method": "GET"
    }
  }'

# Should return: "Access to internal resources is forbidden"
```

### Test XSS Fix:
1. Send malicious message to chat: `<img src=x onerror=alert(1)>`
2. Check that script doesn't execute in webview
3. Verify HTML is sanitized in output

### Test CORS Fix:
```bash
# This should be blocked:
curl -H "Origin: http://evil.com" \
  http://localhost:8080/api/v1/health

# Should return CORS error or no CORS headers
```

---

## 🚀 DEPLOYMENT CHECKLIST

Before deploying to production:

- [ ] All P0 fixes implemented
- [ ] JWT_SECRET generated and set
- [ ] ALLOWED_ORIGINS configured
- [ ] API keys generated
- [ ] Input validation tested
- [ ] SSRF protection tested
- [ ] XSS protection tested
- [ ] Error handling reviewed
- [ ] Logs reviewed (no sensitive data)
- [ ] Dependencies updated
- [ ] Security scan passed (`npm audit`)

---

## 📞 SUPPORT

If you need help implementing these fixes:
- Create issue on GitHub
- Contact: security@testcopilot.ai
- See: SECURITY_AUDIT_REPORT.md for details

**Remember:** Security is not optional. Fix P0 issues before any production deployment!
