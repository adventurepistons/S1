# Test Automation Copilot - Backend API

Shared backend API for Test Copilot, Code Review AI, and Documentation AI.

**Version**: 2.0.0
**Stack**: Node.js + NestJS + TypeScript + PostgreSQL + Redis
**Status**: Phase 1 - Foundation Complete ✅

---

## 🏗️ Architecture

```
backend/
├── src/
│   ├── main.ts              # Application entry point
│   ├── app.module.ts        # Root module
│   ├── auth/                # Authentication & Authorization
│   │   ├── auth.controller.ts
│   │   ├── auth.service.ts
│   │   ├── users.service.ts
│   │   ├── quota.service.ts
│   │   ├── strategies/      # Passport strategies (JWT, Google)
│   │   ├── guards/          # Auth & Quota guards
│   │   └── decorators/      # Custom decorators
│   ├── ai/                  # AI/LLM Service (coming in Phase 3)
│   ├── tests/               # Test Execution Service (coming in Phase 2)
│   ├── storage/             # File storage (S3)
│   │   └── s3.service.ts
│   └── shared/              # Shared utilities
│       ├── database.service.ts  # Prisma wrapper
│       └── cache.service.ts     # Redis cache
├── prisma/
│   └── schema.prisma        # Database schema
├── docker-compose.yml       # Local development stack
└── package.json
```

---

## 🚀 Quick Start

### Prerequisites

- **Node.js** 20+ and npm
- **Docker** and Docker Compose
- **PostgreSQL** 15+ (or use Docker)
- **Redis** 7+ (or use Docker)

### Step 1: Install Dependencies

```bash
cd backend
npm install
```

### Step 2: Set Up Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env file with your configuration
nano .env
```

**Required environment variables**:
```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/testcopilot
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
ANTHROPIC_API_KEY=sk-ant-api03-your-key-here
```

### Step 3: Start Infrastructure with Docker

```bash
# Start PostgreSQL + Redis + API
npm run docker:up

# Or start database only
docker-compose up -d postgres redis
```

### Step 4: Run Database Migrations

```bash
# Generate Prisma Client
npm run prisma:generate

# Run migrations
npm run prisma:migrate

# (Optional) Open Prisma Studio to view database
npm run prisma:studio
```

### Step 5: Start Development Server

```bash
npm run start:dev
```

The API will be available at **http://localhost:3000/api/v1**

---

## 📋 Available Scripts

| Script | Description |
|--------|-------------|
| `npm run start:dev` | Start in development mode with hot-reload |
| `npm run start:prod` | Start in production mode |
| `npm run build` | Build for production |
| `npm run test` | Run unit tests |
| `npm run test:e2e` | Run end-to-end tests |
| `npm run lint` | Lint code |
| `npm run format` | Format code with Prettier |
| `npm run docker:up` | Start Docker services |
| `npm run docker:down` | Stop Docker services |
| `npm run prisma:generate` | Generate Prisma Client |
| `npm run prisma:migrate` | Run database migrations |
| `npm run prisma:studio` | Open Prisma Studio (database GUI) |

---

## 🔑 API Endpoints (Phase 1)

### Authentication

```http
POST   /api/v1/auth/register         # Register new user
POST   /api/v1/auth/login            # Login with email/password
POST   /api/v1/auth/refresh          # Refresh access token
POST   /api/v1/auth/logout           # Logout user
GET    /api/v1/auth/google           # Google OAuth (initiate)
GET    /api/v1/auth/google/callback  # Google OAuth (callback)
GET    /api/v1/auth/me               # Get current user profile
PUT    /api/v1/auth/me               # Update user profile
GET    /api/v1/auth/usage            # Get usage and quotas
PUT    /api/v1/auth/api-mode         # Update API key mode
```

### Example: Register User

**Request**:
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "displayName": "John Doe"
  }'
```

**Response**:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresIn": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "displayName": "John Doe",
    "tier": "FREE",
    "apiKeyMode": "BYOK",
    "emailVerified": false
  }
}
```

### Example: Login

**Request**:
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

### Example: Get Profile (Authenticated)

**Request**:
```bash
curl -X GET http://localhost:3000/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response**:
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "displayName": "John Doe",
    "tier": "FREE",
    "apiKeyMode": "BYOK",
    "createdAt": "2025-11-17T10:30:00.000Z"
  },
  "counts": {
    "projects": 0,
    "executions": 0,
    "conversations": 0
  },
  "quotas": [
    {
      "tool": "test-copilot",
      "monthlyLimit": -1,
      "monthlyUsage": 0,
      "isUnlimited": true
    }
  ]
}
```

---

## 🔒 Authentication & Authorization

### JWT Authentication

All protected endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

**Access Token**: Expires in 1 hour
**Refresh Token**: Expires in 30 days

### Tier-Based Access

| Tier | Frameworks | Monthly Limit (Managed API) | Quotas |
|------|-----------|----------------------------|--------|
| **FREE** | Selenium/Java only | Unlimited (BYOK/Local) | Max 3 projects |
| **PRO** | All 7 frameworks | 500 generations/month | Max 10 projects |
| **TEAM** | All frameworks | Unlimited | Unlimited projects |

### API Key Modes

- **BYOK** (Bring Your Own Key): User provides their own Claude API key - No usage limits
- **LOCAL**: User runs local LLM (Ollama) - No usage limits
- **MANAGED**: We provide API access - Tier-based limits apply

### Using Guards in Code

```typescript
// Protect endpoint with JWT
@Get('protected')
@UseGuards(JwtAuthGuard)
async protectedRoute(@CurrentUser() user: User) {
  return { user };
}

// Enforce quota check
@Post('generate')
@UseGuards(JwtAuthGuard, QuotaGuard)
@Tool('test-copilot')  // Specify tool for quota tracking
async generateCode(@CurrentUser() user: User) {
  // This will check quota before executing
  // Will throw ForbiddenException if quota exceeded
}

// Public endpoint (no auth required)
@Get('public')
@Public()
async publicRoute() {
  return { message: 'No auth required' };
}
```

---

## 🗄️ Database Schema

**Database**: PostgreSQL 15+
**ORM**: Prisma
**Tables**: 15+ tables

### Core Tables

- **users** - User accounts
- **oauth_providers** - OAuth credentials (Google, GitHub)
- **api_keys** - Encrypted user API keys (BYOK mode)
- **user_quotas** - Usage tracking per tool
- **projects** - User projects
- **code_index** - Indexed codebase for RAG
- **conversations** - Multi-turn chat history
- **messages** - Individual messages in conversations
- **test_executions** - Test run metadata
- **test_results** - Individual test results
- **flaky_tests** - Flaky test tracking
- **visual_baselines** - Visual regression baselines
- **visual_comparisons** - Screenshot comparisons

### View Schema in Prisma Studio

```bash
npm run prisma:studio
```

Opens browser at `http://localhost:5555` with full database GUI.

---

## 📦 Docker Setup

### Using Docker Compose (Recommended for Development)

**Start all services**:
```bash
docker-compose up -d
```

**Services**:
- **postgres** (port 5432) - PostgreSQL database
- **redis** (port 6379) - Redis cache & queue
- **api** (port 3000) - NestJS backend API
- **worker** - BullMQ worker for background jobs
- **prisma-studio** (port 5555) - Database GUI (optional)

**Stop all services**:
```bash
docker-compose down
```

**View logs**:
```bash
docker-compose logs -f api
```

---

## 🔧 Configuration

### Environment Variables

See `.env.example` for all available options.

**Database**:
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string

**Authentication**:
- `JWT_SECRET` - Secret for signing JWT tokens
- `JWT_EXPIRES_IN` - Access token expiration (seconds)
- `GOOGLE_CLIENT_ID` - Google OAuth client ID
- `GOOGLE_CLIENT_SECRET` - Google OAuth secret

**External Services**:
- `ANTHROPIC_API_KEY` - Claude API key
- `AWS_ACCESS_KEY_ID` - AWS S3 access key
- `AWS_SECRET_ACCESS_KEY` - AWS S3 secret
- `AWS_S3_BUCKET` - S3 bucket name
- `STRIPE_SECRET_KEY` - Stripe for billing

**CORS**:
- `CORS_ORIGIN` - Allowed origins (comma-separated)

---

## 🧪 Testing

### Run Tests

```bash
# Unit tests
npm run test

# E2E tests
npm run test:e2e

# Coverage
npm run test:cov
```

### Test User Accounts

For development, you can create test users:

```bash
# Using Prisma Studio
npm run prisma:studio

# Or via API
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123456!",
    "displayName": "Test User"
  }'
```

---

## 🚧 Roadmap

### ✅ Phase 1: Backend Foundation (Weeks 1-4) - COMPLETE

- [x] NestJS project setup
- [x] PostgreSQL database schema
- [x] Docker Compose development environment
- [x] Authentication service (JWT + OAuth)
- [x] User management
- [x] Tier and quota system
- [x] File storage (S3)

### 🔨 Phase 2: Test Execution (Weeks 5-8) - NEXT

- [ ] Test runner service
- [ ] WebSocket for real-time logs
- [ ] Test reporting API
- [ ] Screenshot/video capture
- [ ] Flaky test detection

### 📅 Phase 3: AI/LLM Service (Weeks 9-10)

- [ ] Migrate Go RAG system to TypeScript
- [ ] Context builder API
- [ ] Claude API integration
- [ ] Conversation management

### 🔮 Phase 4: Advanced Features (Weeks 11-18)

- [ ] Visual testing
- [ ] API testing
- [ ] Multi-language support
- [ ] CI/CD integration
- [ ] Team collaboration

---

## 📚 Additional Resources

- [NestJS Documentation](https://docs.nestjs.com/)
- [Prisma Documentation](https://www.prisma.io/docs)
- [Anthropic Claude API](https://docs.anthropic.com/)
- [Project Architecture](../BACKEND_API_SPECIFICATION.md)
- [Implementation Plan](../IMPLEMENTATION_PLAN.md)

---

## 🆘 Troubleshooting

### Database connection failed

```bash
# Check if PostgreSQL is running
docker-compose ps

# Restart PostgreSQL
docker-compose restart postgres

# View logs
docker-compose logs postgres
```

### Redis connection failed

```bash
# Check if Redis is running
docker-compose ps

# Restart Redis
docker-compose restart redis
```

### Prisma Client errors

```bash
# Regenerate Prisma Client
npm run prisma:generate

# Reset database (WARNING: deletes all data)
npx prisma migrate reset
```

### Port already in use

```bash
# Change port in .env
PORT=3001

# Or kill process using port 3000
lsof -ti:3000 | xargs kill -9
```

---

## 📝 License

MIT

---

**Built with ❤️ for the Test Automation Copilot Project**

