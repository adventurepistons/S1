# Backend API Specification
## Shared Backend for Test Copilot, Code Review AI, and Documentation AI

**Version**: 2.0.0
**Date**: 2025-11-17
**Authors**: Engineering Team
**Status**: Design Phase

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Technology Stack Recommendations](#2-technology-stack-recommendations)
3. [Database Schema Design](#3-database-schema-design)
4. [Core API Endpoints (Detailed)](#4-core-api-endpoints-detailed)
5. [Authentication & Authorization](#5-authentication--authorization)
6. [Rate Limiting & Quotas](#6-rate-limiting--quotas)
7. [WebSocket Protocols](#7-websocket-protocols)
8. [Error Handling Standards](#8-error-handling-standards)
9. [Performance Considerations](#9-performance-considerations)
10. [Security Best Practices](#10-security-best-practices)
11. [Deployment Architecture](#11-deployment-architecture)

---

## 1. Architecture Overview

### 1.1 System Components

```
┌──────────────────────────────────────────────────────────────────┐
│                        FRONTEND LAYER                             │
├─────────────┬────────────────┬───────────────┬───────────────────┤
│ VSCode Ext  │ VSCode Ext     │ Web Dashboard │ Mobile App        │
│ (Test       │ (Code Review)  │ (Admin Panel) │ (Future)          │
│  Copilot)   │                │               │                   │
└─────────────┴────────────────┴───────────────┴───────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                     API GATEWAY (Kong / NGINX)                    │
│  • Authentication (JWT validation)                                │
│  • Rate limiting (per user/tier)                                  │
│  • Request routing                                                │
│  • SSL termination                                                │
│  • API versioning                                                 │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                      BACKEND SERVICES                             │
├─────────────┬────────────────┬──────────────┬────────────────────┤
│ Auth        │ AI/LLM         │ Test Exec    │ Storage            │
│ Service     │ Service        │ Service      │ Service            │
│             │                │              │                    │
│ - Login     │ - Context      │ - Runner     │ - Database         │
│ - Register  │   Builder      │ - Streams    │ - File Store       │
│ - Tokens    │ - RAG          │ - Reports    │ - Cache            │
│ - Quotas    │ - LLM Calls    │ - Debug      │                    │
└─────────────┴────────────────┴──────────────┴────────────────────┘
       │              │                │               │
       └──────────────┼────────────────┼───────────────┘
                      ▼                ▼
┌─────────────────────────────┐  ┌──────────────────────────────┐
│   PostgreSQL / Firestore     │  │  Redis (Cache & Queue)       │
│   • Users & Auth             │  │  • Session cache             │
│   • Projects & Code          │  │  • Job queue (test execution)│
│   • Test Results             │  │  • Rate limit counters       │
│   • Conversations            │  │  • Real-time pub/sub         │
└─────────────────────────────┘  └──────────────────────────────┘
                      │
                      ▼
┌──────────────────────────────────────────────────────────────────┐
│          EXTERNAL SERVICES                                        │
├─────────────┬────────────────┬──────────────┬────────────────────┤
│ Claude API  │ S3 / Azure     │ Stripe       │ SendGrid           │
│ (Anthropic) │ Blob Storage   │ (Billing)    │ (Email)            │
└─────────────┴────────────────┴──────────────┴────────────────────┘
```

### 1.2 Service Responsibilities

#### **Auth Service** (Port 3001)
- User registration, login, logout
- JWT token generation and validation
- OAuth integration (Google, GitHub)
- Tier management (free, pro, team)
- Usage quota tracking

#### **AI/LLM Service** (Port 3002)
- Context building (RAG pipeline)
- Claude API integration
- Prompt caching management
- Conversation history
- Code generation

#### **Test Execution Service** (Port 3003)
- Test runner orchestration
- Real-time log streaming (WebSocket)
- Browser management (Selenium Grid, Playwright)
- Screenshot/video capture
- Result storage

#### **Storage Service** (Port 3004)
- Database operations (PostgreSQL/Firestore)
- File uploads (S3/Azure Blob)
- Cache management (Redis)
- Backup and recovery

#### **Notification Service** (Port 3005)
- Email notifications (SendGrid)
- Slack/Teams webhooks
- In-app notifications
- CI/CD webhooks (GitHub, GitLab)

---

## 2. Technology Stack Recommendations

### Option 1: Node.js + TypeScript (Recommended)

**Pros**:
- Fast development
- Excellent for real-time features (WebSocket)
- Large ecosystem (npm)
- TypeScript = type safety
- Easy VSCode extension integration

**Stack**:
```yaml
Backend Framework: NestJS (enterprise-grade, modular)
Database: PostgreSQL + Prisma ORM
Cache: Redis
File Storage: AWS S3 or Azure Blob
Real-time: Socket.io
Testing: Jest
API Docs: Swagger/OpenAPI
```

**Sample Directory Structure**:
```
backend/
├── src/
│   ├── auth/                   # Auth service
│   │   ├── auth.controller.ts
│   │   ├── auth.service.ts
│   │   ├── jwt.strategy.ts
│   │   └── dto/
│   ├── ai/                     # AI/LLM service
│   │   ├── ai.controller.ts
│   │   ├── context-builder.ts
│   │   ├── claude-client.ts
│   │   └── rag/
│   ├── tests/                  # Test execution service
│   │   ├── tests.controller.ts
│   │   ├── runner.service.ts
│   │   ├── websocket.gateway.ts
│   │   └── reporting/
│   ├── storage/                # Storage service
│   │   ├── database.service.ts
│   │   ├── s3.service.ts
│   │   └── cache.service.ts
│   ├── shared/                 # Shared utilities
│   │   ├── guards/
│   │   ├── interceptors/
│   │   └── filters/
│   └── main.ts
├── prisma/
│   └── schema.prisma           # Database schema
├── package.json
└── tsconfig.json
```

---

### Option 2: Python + FastAPI

**Pros**:
- Excellent for AI/ML (if you add features later)
- FastAPI is fast and modern
- Great for data processing
- Easy integration with Pytest

**Stack**:
```yaml
Backend Framework: FastAPI
Database: PostgreSQL + SQLAlchemy
Cache: Redis
Real-time: FastAPI WebSocket
Testing: Pytest
API Docs: Auto-generated by FastAPI
```

---

### Option 3: Go (High Performance)

**Pros**:
- Extremely fast
- Compiled binaries (easy deployment)
- Great concurrency (goroutines)
- Low memory footprint

**Stack**:
```yaml
Backend Framework: Gin or Fiber
Database: PostgreSQL + GORM
Cache: Redis
Real-time: Gorilla WebSocket
Testing: Go testing
```

**Cons**:
- Slower development compared to Node.js/Python
- Smaller ecosystem for some integrations

---

### **Recommendation**: Node.js + NestJS + TypeScript

**Why?**
1. **Fast development** - Get to market quickly
2. **TypeScript** - Type safety + VSCode integration
3. **NestJS** - Enterprise architecture (modules, DI, guards)
4. **WebSocket** - Built-in for real-time features
5. **Ecosystem** - Huge library ecosystem
6. **Team familiarity** - Most developers know JavaScript/TypeScript

---

## 3. Database Schema Design

### 3.1 PostgreSQL Schema (Recommended)

#### **Users Table**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),  -- NULL if OAuth
    display_name VARCHAR(255),
    avatar_url TEXT,
    tier VARCHAR(20) DEFAULT 'free',  -- 'free', 'pro', 'team'
    api_key_mode VARCHAR(20) DEFAULT 'byok',  -- 'byok', 'local', 'managed'
    stripe_customer_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login_at TIMESTAMP,
    email_verified BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,

    CONSTRAINT tier_check CHECK (tier IN ('free', 'pro', 'team')),
    CONSTRAINT api_mode_check CHECK (api_key_mode IN ('byok', 'local', 'managed'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tier ON users(tier);
```

#### **OAuth Providers Table**

```sql
CREATE TABLE oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,  -- 'google', 'github'
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(provider, provider_user_id)
);
```

#### **User Quotas Table**

```sql
CREATE TABLE user_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tool VARCHAR(50) NOT NULL,  -- 'test-copilot', 'code-review', 'docs'
    monthly_limit INTEGER DEFAULT -1,  -- -1 = unlimited
    monthly_usage INTEGER DEFAULT 0,
    last_reset_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, tool)
);
```

#### **Projects Table**

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    repository_url TEXT,
    framework VARCHAR(50),  -- 'selenium-java', 'playwright-ts', etc.
    language VARCHAR(50),   -- 'java', 'typescript', 'python'
    indexed BOOLEAN DEFAULT false,
    indexed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, name)
);

CREATE INDEX idx_projects_user ON projects(user_id);
CREATE INDEX idx_projects_framework ON projects(framework);
```

#### **Code Index Table** (Indexed codebase)

```sql
CREATE TABLE code_index (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    class_name VARCHAR(255),
    method_name VARCHAR(255),
    code_type VARCHAR(50),  -- 'test', 'page_object', 'helper', 'model'
    code_content TEXT NOT NULL,
    embedding VECTOR(384),  -- pgvector extension for semantic search
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_code_project ON code_index(project_id);
CREATE INDEX idx_code_class ON code_index(class_name);
CREATE INDEX idx_code_type ON code_index(code_type);
-- pgvector index for fast similarity search
CREATE INDEX idx_code_embedding ON code_index USING ivfflat (embedding vector_cosine_ops);
```

#### **Conversations Table**

```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    tool VARCHAR(50),  -- 'test-copilot', 'code-review', 'docs'
    title VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_conversations_user ON conversations(user_id);
CREATE INDEX idx_conversations_project ON conversations(project_id);
```

#### **Messages Table**

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,  -- 'user', 'assistant', 'system'
    content TEXT NOT NULL,
    metadata JSONB,  -- {tokens: {input: 100, output: 50}, cost: 0.012, etc.}
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT role_check CHECK (role IN ('user', 'assistant', 'system'))
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created ON messages(created_at);
```

#### **Test Executions Table**

```sql
CREATE TABLE test_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'queued',  -- 'queued', 'running', 'completed', 'failed', 'cancelled'
    environment VARCHAR(50),  -- 'dev', 'staging', 'prod'
    browser VARCHAR(50),      -- 'chrome', 'firefox', 'safari'
    headless BOOLEAN DEFAULT true,
    parallel BOOLEAN DEFAULT false,
    max_workers INTEGER DEFAULT 1,
    retry_failed INTEGER DEFAULT 0,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration INTEGER,  -- in seconds
    total_tests INTEGER,
    passed_tests INTEGER,
    failed_tests INTEGER,
    skipped_tests INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled'))
);

CREATE INDEX idx_executions_project ON test_executions(project_id);
CREATE INDEX idx_executions_user ON test_executions(user_id);
CREATE INDEX idx_executions_status ON test_executions(status);
CREATE INDEX idx_executions_created ON test_executions(created_at DESC);
```

#### **Test Results Table**

```sql
CREATE TABLE test_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID REFERENCES test_executions(id) ON DELETE CASCADE,
    test_name VARCHAR(255) NOT NULL,
    class_name VARCHAR(255),
    status VARCHAR(20) NOT NULL,  -- 'passed', 'failed', 'skipped'
    duration INTEGER,  -- in milliseconds
    retries INTEGER DEFAULT 0,
    error_message TEXT,
    stack_trace TEXT,
    screenshot_url TEXT,
    video_url TEXT,
    logs TEXT,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT status_check CHECK (status IN ('passed', 'failed', 'skipped'))
);

CREATE INDEX idx_results_execution ON test_results(execution_id);
CREATE INDEX idx_results_status ON test_results(status);
CREATE INDEX idx_results_test_name ON test_results(test_name);
```

#### **Visual Baselines Table**

```sql
CREATE TABLE visual_baselines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    test_name VARCHAR(255) NOT NULL,
    viewport_width INTEGER,
    viewport_height INTEGER,
    browser VARCHAR(50),
    image_url TEXT NOT NULL,
    image_hash VARCHAR(64),  -- SHA-256 hash
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(project_id, test_name, viewport_width, viewport_height, browser)
);

CREATE INDEX idx_baselines_project ON visual_baselines(project_id);
```

#### **Visual Comparisons Table**

```sql
CREATE TABLE visual_comparisons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    baseline_id UUID REFERENCES visual_baselines(id) ON DELETE CASCADE,
    execution_id UUID REFERENCES test_executions(id) ON DELETE CASCADE,
    screenshot_url TEXT NOT NULL,
    diff_url TEXT,
    pixel_diff INTEGER,
    percentage_diff NUMERIC(5,2),
    threshold NUMERIC(5,2) DEFAULT 0.05,
    passed BOOLEAN,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_comparisons_baseline ON visual_comparisons(baseline_id);
CREATE INDEX idx_comparisons_execution ON visual_comparisons(execution_id);
```

#### **Flaky Tests Table**

```sql
CREATE TABLE flaky_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    test_name VARCHAR(255) NOT NULL,
    class_name VARCHAR(255),
    flaky_score NUMERIC(3,2),  -- 0.00 to 1.00 (1.00 = always flaky)
    total_runs INTEGER DEFAULT 0,
    passed_runs INTEGER DEFAULT 0,
    failed_runs INTEGER DEFAULT 0,
    last_failed_at TIMESTAMP,
    first_detected_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(project_id, test_name)
);

CREATE INDEX idx_flaky_project ON flaky_tests(project_id);
CREATE INDEX idx_flaky_score ON flaky_tests(flaky_score DESC);
```

---

### 3.2 Prisma Schema (if using Prisma ORM)

```prisma
// prisma/schema.prisma

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id               String    @id @default(uuid())
  email            String    @unique
  passwordHash     String?   @map("password_hash")
  displayName      String?   @map("display_name")
  avatarUrl        String?   @map("avatar_url")
  tier             Tier      @default(FREE)
  apiKeyMode       ApiMode   @default(BYOK) @map("api_key_mode")
  stripeCustomerId String?   @map("stripe_customer_id")
  createdAt        DateTime  @default(now()) @map("created_at")
  updatedAt        DateTime  @updatedAt @map("updated_at")
  lastLoginAt      DateTime? @map("last_login_at")
  emailVerified    Boolean   @default(false) @map("email_verified")
  isActive         Boolean   @default(true) @map("is_active")

  projects      Project[]
  conversations Conversation[]
  executions    TestExecution[]
  quotas        UserQuota[]
  oauthProviders OAuthProvider[]

  @@index([email])
  @@index([tier])
  @@map("users")
}

enum Tier {
  FREE
  PRO
  TEAM
}

enum ApiMode {
  BYOK
  LOCAL
  MANAGED
}

model Project {
  id            String    @id @default(uuid())
  userId        String    @map("user_id")
  name          String
  description   String?
  repositoryUrl String?   @map("repository_url")
  framework     String?
  language      String?
  indexed       Boolean   @default(false)
  indexedAt     DateTime? @map("indexed_at")
  createdAt     DateTime  @default(now()) @map("created_at")
  updatedAt     DateTime  @updatedAt @map("updated_at")

  user           User             @relation(fields: [userId], references: [id], onDelete: Cascade)
  codeIndex      CodeIndex[]
  conversations  Conversation[]
  executions     TestExecution[]
  visualBaselines VisualBaseline[]
  flakyTests     FlakyTest[]

  @@unique([userId, name])
  @@index([userId])
  @@index([framework])
  @@map("projects")
}

model TestExecution {
  id           String   @id @default(uuid())
  projectId    String   @map("project_id")
  userId       String?  @map("user_id")
  status       ExecutionStatus @default(QUEUED)
  environment  String?
  browser      String?
  headless     Boolean  @default(true)
  parallel     Boolean  @default(false)
  maxWorkers   Int      @default(1) @map("max_workers")
  retryFailed  Int      @default(0) @map("retry_failed")
  startedAt    DateTime? @map("started_at")
  completedAt  DateTime? @map("completed_at")
  duration     Int?
  totalTests   Int?     @map("total_tests")
  passedTests  Int?     @map("passed_tests")
  failedTests  Int?     @map("failed_tests")
  skippedTests Int?     @map("skipped_tests")
  createdAt    DateTime @default(now()) @map("created_at")

  project    Project       @relation(fields: [projectId], references: [id], onDelete: Cascade)
  user       User?         @relation(fields: [userId], references: [id], onDelete: SetNull)
  results    TestResult[]
  comparisons VisualComparison[]

  @@index([projectId])
  @@index([userId])
  @@index([status])
  @@index([createdAt(sort: Desc)])
  @@map("test_executions")
}

enum ExecutionStatus {
  QUEUED
  RUNNING
  COMPLETED
  FAILED
  CANCELLED
}

model TestResult {
  id           String   @id @default(uuid())
  executionId  String   @map("execution_id")
  testName     String   @map("test_name")
  className    String?  @map("class_name")
  status       TestStatus
  duration     Int?
  retries      Int      @default(0)
  errorMessage String?  @map("error_message")
  stackTrace   String?  @map("stack_trace")
  screenshotUrl String? @map("screenshot_url")
  videoUrl     String?  @map("video_url")
  logs         String?
  createdAt    DateTime @default(now()) @map("created_at")

  execution TestExecution @relation(fields: [executionId], references: [id], onDelete: Cascade)

  @@index([executionId])
  @@index([status])
  @@index([testName])
  @@map("test_results")
}

enum TestStatus {
  PASSED
  FAILED
  SKIPPED
}

// ... (other models similar to SQL schema)
```

---

## 4. Core API Endpoints (Detailed)

### 4.1 Authentication Endpoints

#### **POST /api/v1/auth/register**

**Description**: Create new user account

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "displayName": "John Doe"
}
```

**Response** (201 Created):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "displayName": "John Doe",
    "tier": "free",
    "emailVerified": false
  },
  "message": "Verification email sent to user@example.com"
}
```

**Errors**:
- `400 Bad Request`: Invalid email or weak password
- `409 Conflict`: Email already registered

---

#### **POST /api/v1/auth/login**

**Description**: Login with email/password

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response** (200 OK):
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "refresh_token_here",
  "expiresIn": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "displayName": "John Doe",
    "tier": "pro",
    "apiKeyMode": "managed"
  }
}
```

**JWT Payload**:
```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "tier": "pro",
  "iat": 1700000000,
  "exp": 1700003600
}
```

**Errors**:
- `401 Unauthorized`: Invalid credentials
- `403 Forbidden`: Account disabled

---

#### **POST /api/v1/auth/oauth/google**

**Description**: OAuth login with Google

**Request Body**:
```json
{
  "idToken": "google_id_token_from_frontend"
}
```

**Response** (200 OK):
```json
{
  "accessToken": "...",
  "refreshToken": "...",
  "user": { ... },
  "isNewUser": false
}
```

---

### 4.2 Test Execution Endpoints

#### **POST /api/v1/tests/execute**

**Description**: Execute tests

**Headers**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "projectId": "proj_123",
  "tests": ["LoginTest", "CheckoutTest"],  // or ["*"] for all
  "environment": "staging",
  "browser": "chrome",
  "headless": true,
  "parallel": true,
  "maxWorkers": 4,
  "retryFailedTests": 2,
  "timeout": 30000,
  "tags": ["smoke", "critical"]
}
```

**Response** (202 Accepted):
```json
{
  "executionId": "exec_550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "queuePosition": 2,
  "estimatedStartTime": "2025-11-17T10:35:00Z",
  "webhookUrl": "wss://api.testcopilot.dev/api/v1/tests/stream/exec_550e8400"
}
```

**Implementation Notes**:
```typescript
// NestJS Controller
@Controller('api/v1/tests')
export class TestsController {
  constructor(
    private readonly testRunner: TestRunnerService,
    private readonly queue: BullQueue,
  ) {}

  @Post('execute')
  @UseGuards(JwtAuthGuard, QuotaGuard)
  async executeTests(@Body() dto: ExecuteTestsDto, @User() user: UserEntity) {
    // 1. Validate user quota
    await this.checkQuota(user.id, 'test-execution');

    // 2. Create execution record
    const execution = await this.testRunner.createExecution({
      ...dto,
      userId: user.id,
      status: 'queued',
    });

    // 3. Add to queue (Redis Bull)
    await this.queue.add('run-tests', {
      executionId: execution.id,
      ...dto,
    });

    // 4. Increment usage
    await this.incrementQuota(user.id, 'test-execution');

    return {
      executionId: execution.id,
      status: execution.status,
      webhookUrl: `wss://api.../tests/stream/${execution.id}`,
    };
  }
}
```

---

#### **WebSocket: /api/v1/tests/stream/{executionId}**

**Description**: Real-time test execution logs

**Connection**:
```javascript
const ws = new WebSocket('wss://api.testcopilot.dev/api/v1/tests/stream/exec_123');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data);
};
```

**Message Format**:
```json
{
  "type": "status",
  "executionId": "exec_123",
  "status": "running",
  "timestamp": "2025-11-17T10:30:00Z"
}

{
  "type": "log",
  "executionId": "exec_123",
  "testName": "LoginTest",
  "level": "info",
  "message": "Test started",
  "timestamp": "2025-11-17T10:30:01Z"
}

{
  "type": "test_started",
  "executionId": "exec_123",
  "testName": "LoginTest",
  "timestamp": "2025-11-17T10:30:01Z"
}

{
  "type": "test_completed",
  "executionId": "exec_123",
  "testName": "LoginTest",
  "status": "passed",
  "duration": 3200,
  "timestamp": "2025-11-17T10:30:04Z"
}

{
  "type": "screenshot",
  "executionId": "exec_123",
  "testName": "CheckoutTest",
  "url": "https://cdn.../screenshot_123.png",
  "timestamp": "2025-11-17T10:30:05Z"
}

{
  "type": "execution_completed",
  "executionId": "exec_123",
  "status": "completed",
  "summary": {
    "total": 10,
    "passed": 9,
    "failed": 1,
    "skipped": 0,
    "duration": 45000
  },
  "timestamp": "2025-11-17T10:31:00Z"
}
```

**Implementation**:
```typescript
// NestJS WebSocket Gateway
@WebSocketGateway({
  namespace: '/api/v1/tests',
  cors: { origin: '*' },
})
export class TestExecutionGateway {
  @WebSocketServer()
  server: Server;

  @SubscribeMessage('stream')
  async handleStream(
    @MessageBody() data: { executionId: string },
    @ConnectedSocket() client: Socket,
  ) {
    const { executionId } = data;

    // Join room for this execution
    client.join(`execution:${executionId}`);

    // Subscribe to Redis pub/sub for logs
    this.redisClient.subscribe(`logs:${executionId}`, (message) => {
      this.server.to(`execution:${executionId}`).emit('log', JSON.parse(message));
    });
  }

  // Emit from test runner
  async emitLog(executionId: string, log: any) {
    await this.redisClient.publish(`logs:${executionId}`, JSON.stringify(log));
  }
}
```

---

#### **GET /api/v1/reports/{executionId}**

**Description**: Get execution report

**Response** (200 OK):
```json
{
  "executionId": "exec_123",
  "projectId": "proj_456",
  "status": "completed",
  "environment": "staging",
  "browser": "chrome",
  "startedAt": "2025-11-17T10:30:00Z",
  "completedAt": "2025-11-17T10:31:00Z",
  "duration": 60,
  "summary": {
    "total": 10,
    "passed": 9,
    "failed": 1,
    "skipped": 0,
    "passRate": 90.0
  },
  "tests": [
    {
      "id": "result_789",
      "testName": "LoginTest",
      "className": "com.example.LoginTest",
      "status": "passed",
      "duration": 3200,
      "retries": 0
    },
    {
      "id": "result_790",
      "testName": "CheckoutTest",
      "className": "com.example.CheckoutTest",
      "status": "failed",
      "duration": 8500,
      "retries": 2,
      "errorMessage": "Element not found: #submit-btn",
      "stackTrace": "org.openqa.selenium.NoSuchElementException...",
      "screenshot": "https://cdn.../screenshot_790.png",
      "video": "https://cdn.../video_790.mp4"
    }
  ],
  "flakyTests": [
    {
      "testName": "PaymentTest",
      "flakyScore": 0.67,
      "totalRuns": 3,
      "passedRuns": 1,
      "failedRuns": 2
    }
  ]
}
```

---

### 4.3 AI/LLM Endpoints

#### **POST /api/v1/ai/generate**

**Description**: Generate code with AI

**Request Body**:
```json
{
  "projectId": "proj_123",
  "request": "Create a test for login with valid credentials",
  "codeType": "test",  // or "page_object", "helper"
  "framework": "selenium-java",
  "conversationId": null  // or existing conversation ID
}
```

**Response** (200 OK):
```json
{
  "code": "public class LoginTest {\n    @Test\n    public void testLoginWithValidCredentials() {\n        ...\n    }\n}",
  "className": "LoginTest",
  "packageName": "com.example.tests",
  "suggestedPath": "src/test/java/com/example/tests/LoginTest.java",
  "imports": [
    "org.testng.annotations.Test",
    "org.openqa.selenium.WebDriver"
  ],
  "metadata": {
    "conversationId": "conv_123",
    "messageId": "msg_456",
    "tokens": {
      "input": 6100,
      "output": 452,
      "cached": 3500
    },
    "cost": 0.012,
    "model": "claude-sonnet-4.5",
    "duration": 2.3
  }
}
```

**Implementation**:
```typescript
@Controller('api/v1/ai')
export class AIController {
  constructor(
    private readonly contextBuilder: ContextBuilderService,
    private readonly claudeClient: ClaudeClientService,
  ) {}

  @Post('generate')
  @UseGuards(JwtAuthGuard, QuotaGuard)
  async generate(@Body() dto: GenerateCodeDto, @User() user: UserEntity) {
    // 1. Build context (RAG pipeline)
    const context = await this.contextBuilder.buildContext({
      projectId: dto.projectId,
      userRequest: dto.request,
      codeType: dto.codeType,
    });

    // 2. Call Claude API
    const response = await this.claudeClient.generate({
      systemPrompt: context.systemPrompt,
      userPrompt: context.userPrompt,
      cacheBreakpoints: context.cacheBreakpoints,
    });

    // 3. Parse response
    const parsed = this.parseCode(response.content);

    // 4. Save to conversation
    await this.saveConversation({
      userId: user.id,
      projectId: dto.projectId,
      conversationId: dto.conversationId,
      userMessage: dto.request,
      assistantMessage: response.content,
      metadata: {
        tokens: response.usage,
        cost: this.calculateCost(response.usage),
      },
    });

    // 5. Increment usage quota
    await this.incrementQuota(user.id, 'ai-generation');

    return parsed;
  }
}
```

---

#### **POST /api/v1/debug/analyze-failure**

**Description**: AI analyzes test failure

**Request Body**:
```json
{
  "testName": "CheckoutTest",
  "executionId": "exec_123",
  "error": "Element not found: #submit-btn",
  "screenshot": "https://cdn.../screenshot.png",
  "stackTrace": "...",
  "code": "driver.findElement(By.id(\"submit-btn\")).click();"
}
```

**Response** (200 OK):
```json
{
  "analysis": "The test failed because the submit button selector changed from '#submit-btn' to '#checkout-submit'. The screenshot shows the button is present but with a different ID.",
  "rootCause": "UI_CHANGE",
  "confidence": 0.95,
  "suggestedFix": {
    "oldCode": "driver.findElement(By.id(\"submit-btn\")).click();",
    "newCode": "driver.findElement(By.id(\"checkout-submit\")).click();",
    "explanation": "Updated selector to match current DOM structure"
  },
  "alternatives": [
    {
      "code": "driver.findElement(By.cssSelector(\"[data-testid='submit-checkout']\")).click();",
      "reason": "Using data-testid is more stable"
    }
  ],
  "preventionTips": [
    "Use data-testid attributes for critical elements",
    "Implement self-healing locators",
    "Add visual regression tests"
  ],
  "relatedFailures": [
    {
      "testName": "PaymentTest",
      "similarity": 0.87,
      "reason": "Same selector issue"
    }
  ]
}
```

---

## 5. Authentication & Authorization

### 5.1 JWT Token Structure

**Access Token** (expires in 1 hour):
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "tier": "pro",
  "iat": 1700000000,
  "exp": 1700003600,
  "type": "access"
}
```

**Refresh Token** (expires in 30 days):
```json
{
  "sub": "user_id",
  "type": "refresh",
  "iat": 1700000000,
  "exp": 1702592000
}
```

### 5.2 Authorization Rules

| Tier | Features | Quotas |
|------|----------|--------|
| **Free** | • Selenium/Java only<br>• BYOK or Local LLM<br>• Unlimited generations (BYOK) | • Max 3 projects<br>• No team features |
| **Pro** | • All 7 frameworks<br>• Managed API<br>• 500 generations/month<br>• Visual testing<br>• API testing | • Max 10 projects<br>• 500 test executions/month<br>• 10 GB storage |
| **Team** | • Everything in Pro<br>• Unlimited generations<br>• Team workspaces<br>• Priority support | • Unlimited projects<br>• Unlimited executions<br>• 100 GB storage |

### 5.3 Guards Implementation

```typescript
// quota.guard.ts
@Injectable()
export class QuotaGuard implements CanActivate {
  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest();
    const user = request.user;

    // Check quota
    const quota = await this.quotaService.getUserQuota(user.id, 'test-execution');

    if (quota.monthly_limit !== -1 && quota.monthly_usage >= quota.monthly_limit) {
      throw new ForbiddenException('Monthly quota exceeded. Upgrade to Pro for more.');
    }

    return true;
  }
}
```

---

## 6. Rate Limiting & Quotas

### 6.1 Rate Limiting Strategy

**Using Redis + Token Bucket Algorithm**

| Tier | Rate Limit | Burst |
|------|-----------|-------|
| **Free** | 10 requests/minute | 20 |
| **Pro** | 60 requests/minute | 100 |
| **Team** | 300 requests/minute | 500 |

**Implementation**:
```typescript
// rate-limiter.middleware.ts
import rateLimit from 'express-rate-limit';
import RedisStore from 'rate-limit-redis';

const limiter = rateLimit({
  store: new RedisStore({
    client: redisClient,
    prefix: 'rl:',
  }),
  windowMs: 60 * 1000, // 1 minute
  max: async (req) => {
    const user = req.user;
    switch (user.tier) {
      case 'team': return 300;
      case 'pro': return 60;
      default: return 10;
    }
  },
  message: 'Too many requests, please try again later.',
});
```

---

## 7. WebSocket Protocols

### 7.1 Connection Flow

```
Client                      Server
  |                           |
  |-- WS Connect -----------→ |
  |                           |
  |← Auth Challenge --------- |
  |                           |
  |-- JWT Token -----------→  |
  |                           |
  |← Connected (200) -------- |
  |                           |
  |-- Subscribe(execId) ----→ |
  |                           |
  |← Stream logs ------------ |
  |← Stream logs ------------ |
  |← Execution complete ----- |
  |                           |
  |-- Disconnect -----------→ |
```

### 7.2 Message Types

```typescript
type WSMessage =
  | { type: 'status'; executionId: string; status: ExecutionStatus; timestamp: string }
  | { type: 'log'; executionId: string; testName: string; level: string; message: string; timestamp: string }
  | { type: 'test_started'; executionId: string; testName: string; timestamp: string }
  | { type: 'test_completed'; executionId: string; testName: string; status: TestStatus; duration: number; timestamp: string }
  | { type: 'screenshot'; executionId: string; testName: string; url: string; timestamp: string }
  | { type: 'execution_completed'; executionId: string; status: ExecutionStatus; summary: Summary; timestamp: string }
  | { type: 'error'; executionId: string; error: string; timestamp: string };
```

---

## 8. Error Handling Standards

### 8.1 Error Response Format

```json
{
  "error": {
    "code": "AUTH_FAILED",
    "message": "Invalid API key",
    "details": "The API key 'sk-xxx' is invalid or expired",
    "timestamp": "2025-11-17T10:30:00Z",
    "requestId": "req_abc123",
    "path": "/api/v1/tests/execute",
    "suggestion": "Please check your API key in settings"
  }
}
```

### 8.2 Error Codes

| HTTP | Error Code | Description |
|------|-----------|-------------|
| 400 | `INVALID_REQUEST` | Malformed request body |
| 401 | `AUTH_FAILED` | Invalid credentials |
| 403 | `QUOTA_EXCEEDED` | Monthly limit reached |
| 403 | `FRAMEWORK_NOT_ALLOWED` | Framework requires Pro tier |
| 404 | `RESOURCE_NOT_FOUND` | Resource doesn't exist |
| 409 | `CONFLICT` | Resource already exists |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Claude API down |

---

## 9. Performance Considerations

### 9.1 Caching Strategy

**Redis Cache Layers**:

1. **User Session Cache** (TTL: 1 hour)
   - JWT validation results
   - User profile data

2. **Project Index Cache** (TTL: 24 hours)
   - Code index embeddings
   - RAG retrieval results

3. **LLM Response Cache** (TTL: 7 days)
   - Cache identical requests
   - Key: hash(project_id + user_request + code_type)

4. **Rate Limit Cache** (TTL: 1 minute)
   - Request counters per user

**Implementation**:
```typescript
@Injectable()
export class CacheService {
  async getOrSet<T>(
    key: string,
    factory: () => Promise<T>,
    ttl: number,
  ): Promise<T> {
    // Try cache first
    const cached = await this.redis.get(key);
    if (cached) {
      return JSON.parse(cached);
    }

    // Cache miss - call factory
    const value = await factory();
    await this.redis.setex(key, ttl, JSON.stringify(value));

    return value;
  }
}
```

### 9.2 Database Optimization

**Indexes**:
- All foreign keys
- Frequently queried columns (email, tier, status, created_at)
- Composite indexes for common queries

**Connection Pooling**:
```typescript
// Prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
  pool_size = 20
  connection_limit = 50
}
```

**Query Optimization**:
- Use `select` to fetch only needed columns
- Implement pagination for large result sets
- Use database views for complex reports

---

## 10. Security Best Practices

### 10.1 Password Security

- **Bcrypt** with cost factor 12
- Minimum 8 characters, require uppercase, lowercase, number
- Check against leaked password databases (HaveIBeenPwned API)

### 10.2 API Key Storage

- **User API keys** (BYOK): Encrypted with AES-256-GCM
- **Encryption key**: Stored in AWS Secrets Manager / Azure Key Vault
- Never log API keys

### 10.3 SQL Injection Prevention

- Use Prisma ORM (parameterized queries)
- Validate all input
- Escape special characters

### 10.4 XSS Prevention

- Sanitize user input
- Use Content Security Policy headers
- Encode output

### 10.5 CORS Configuration

```typescript
app.enableCors({
  origin: [
    'https://testcopilot.dev',
    'https://app.testcopilot.dev',
    'vscode://testcopilot',  // VSCode extension
  ],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Authorization', 'Content-Type'],
});
```

---

## 11. Deployment Architecture

### 11.1 Infrastructure

```
                      ┌─────────────────┐
                      │ CloudFlare CDN  │
                      │ (Static Assets) │
                      └────────┬────────┘
                               │
                      ┌────────▼────────┐
                      │  AWS ALB / NGINX │
                      │  Load Balancer   │
                      └────────┬────────┘
                               │
                ┌──────────────┼──────────────┐
                │              │              │
        ┌───────▼──────┐ ┌────▼─────┐ ┌─────▼──────┐
        │ API Server 1 │ │API Server│ │API Server 3│
        │ (Docker)     │ │    2     │ │            │
        └──────────────┘ └──────────┘ └────────────┘
                │              │              │
                └──────────────┼──────────────┘
                               │
                ┌──────────────┼──────────────┐
                │              │              │
        ┌───────▼──────┐ ┌────▼─────┐ ┌─────▼──────┐
        │ PostgreSQL   │ │  Redis   │ │   S3       │
        │ (RDS)        │ │ (Cluster)│ │ (Storage)  │
        └──────────────┘ └──────────┘ └────────────┘
```

### 11.2 Docker Compose (Development)

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgresql://postgres:password@db:5432/testcopilot
      - REDIS_URL=redis://redis:6379
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    depends_on:
      - db
      - redis

  db:
    image: postgres:15
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=testcopilot
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  worker:
    build: .
    command: npm run worker
    environment:
      - DATABASE_URL=postgresql://postgres:password@db:5432/testcopilot
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

volumes:
  postgres_data:
  redis_data:
```

### 11.3 Environment Variables

```.env
# Database
DATABASE_URL=postgresql://user:pass@host:5432/testcopilot
REDIS_URL=redis://localhost:6379

# Authentication
JWT_SECRET=your-jwt-secret-here
JWT_EXPIRES_IN=3600

# External Services
ANTHROPIC_API_KEY=sk-ant-xxx
STRIPE_SECRET_KEY=sk_test_xxx
SENDGRID_API_KEY=SG.xxx
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
AWS_S3_BUCKET=testcopilot-screenshots

# Application
NODE_ENV=production
PORT=3000
CORS_ORIGIN=https://testcopilot.dev
```

---

## Summary

This backend API specification provides:

1. **Complete architecture** for shared backend across 3 tools
2. **Technology recommendations** (Node.js + NestJS + PostgreSQL)
3. **Database schema** with all tables and relationships
4. **Detailed API endpoints** with request/response examples
5. **Authentication & authorization** with JWT and tier-based access
6. **Real-time features** via WebSocket
7. **Error handling** standards
8. **Performance optimizations** (caching, indexing)
9. **Security best practices**
10. **Deployment architecture**

**Next Steps**:
1. Review and approve architecture
2. Set up development environment
3. Build Phase 1 features (Test Execution + Reporting)
4. Deploy to staging for testing
5. Launch v2.0 with complete testing platform

