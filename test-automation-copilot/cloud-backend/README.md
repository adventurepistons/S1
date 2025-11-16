# Test Copilot Cloud Backend

Production-ready cloud backend for Test Automation Copilot with proprietary prompt engineering.

## 🏗️ Architecture

```
Local Backend (Go)
    ↓ HTTPS
Cloud Backend (This Service)
    ├── Prompt Engineering (SECRET)
    ├── OpenAI Integration
    └── Usage Tracking
        ↓
    OpenAI GPT-4
```

## ✨ Features

- **Proprietary Prompt Engineering**: Your competitive advantage stays secret
- **PostgreSQL Database**: User management and usage tracking
- **Authentication**: API key-based authentication
- **Rate Limiting**: Per-user and global rate limits
- **Usage Tracking**: Track tokens, costs, and requests
- **Streaming Support**: Real-time code generation
- **Production Ready**: Logging, error handling, security

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ and npm
- PostgreSQL 14+
- OpenAI API key

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

**Required**:
- `OPENAI_API_KEY`: Your OpenAI API key
- `JWT_SECRET`: Random secret for JWT tokens
- `DB_PASSWORD`: PostgreSQL password

### 3. Set Up Database

```bash
# Create database
createdb testcopilot

# Run migrations
npm run migrate
```

### 4. Start Server

```bash
# Development
npm run dev

# Production
npm run build
npm start
```

Server runs on http://localhost:3000

## 📋 API Endpoints

### Health Check

```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": 1700000000,
  "services": {
    "database": true,
    "openai": true
  }
}
```

### Generate Code (Non-streaming)

```bash
POST /v1/generate
Authorization: Bearer tc_pro_xxxxx
Content-Type: application/json

{
  "action": "pageobject",
  "userQuery": "Login page with username and password",
  "spec": "...",
  "elements": [
    {
      "name": "usernameField",
      "locatorType": "id",
      "locatorValue": "username"
    }
  ],
  "framework": "selenium-java",
  "testRunner": "testng"
}
```

Response:
```json
{
  "code": "package pages;\n\npublic class LoginPage {...}",
  "tokensUsed": 450,
  "model": "gpt-4-turbo-preview",
  "finishReason": "stop"
}
```

### Generate Code (Streaming)

```bash
POST /v1/generate/stream
Authorization: Bearer tc_pro_xxxxx
Content-Type: application/json

# Same request body as above
```

Response (Server-Sent Events):
```
data: {"type":"chunk","content":"package"}
data: {"type":"chunk","content":" pages;\n"}
...
data: {"type":"complete","result":{...}}
```

### Get Usage Statistics

```bash
GET /v1/usage
Authorization: Bearer tc_pro_xxxxx
```

Response:
```json
{
  "userId": "user_123",
  "plan": "pro",
  "usage": {
    "totalRequests": 45,
    "totalTokens": 15000,
    "estimatedCost": 0.30,
    "requestsByAction": {
      "pageobject": 20,
      "test": 15,
      "chat": 8,
      "fix": 2
    }
  },
  "limits": {
    "requestsPerMonth": 1000,
    "tokensPerMonth": 500000,
    "requestsRemaining": 955,
    "tokensRemaining": 485000
  }
}
```

## 🗄️ Database Schema

### Users Table

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  api_key VARCHAR(255) UNIQUE NOT NULL,
  plan VARCHAR(20) DEFAULT 'free',
  stripe_customer_id VARCHAR(255),
  stripe_subscription_id VARCHAR(255),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

### Usage Table

```sql
CREATE TABLE usage (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  action VARCHAR(50) NOT NULL,
  tokens_used INTEGER NOT NULL,
  cost DECIMAL(10, 4) NOT NULL,
  model VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
```

## 🔐 Creating Users

### Programmatically

```typescript
import { UserModel } from './src/models/User';

// Create a new user
const user = await UserModel.create(
  'user@example.com',
  'password123',
  'pro' // plan: 'free', 'pro', or 'team'
);

console.log('API Key:', user.api_key);
// tc_pro_abcd1234...
```

### Manual (for testing)

```bash
# Connect to database
psql testcopilot

# Insert test user
INSERT INTO users (id, email, password_hash, api_key, plan)
VALUES (
  gen_random_uuid(),
  'test@example.com',
  '$2b$10$...', -- bcrypt hash of password
  'tc_test_123456789abcdef',
  'pro'
);
```

## 📊 Plans & Pricing

| Plan | Requests/Month | Tokens/Month | Price |
|------|----------------|--------------|-------|
| Free | 100 | 50,000 | $0 |
| Pro | 1,000 | 500,000 | $20 |
| Team | 5,000 | 2,500,000 | $50 |

## 🚀 Deployment

### Docker

```dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]
```

Build and run:
```bash
docker build -t testcopilot-cloud .
docker run -p 3000:3000 --env-file .env testcopilot-cloud
```

### Heroku

```bash
# Create app
heroku create testcopilot-cloud

# Add PostgreSQL
heroku addons:create heroku-postgresql:mini

# Set environment variables
heroku config:set OPENAI_API_KEY=sk-...
heroku config:set JWT_SECRET=...

# Deploy
git push heroku main

# Run migrations
heroku run npm run migrate
```

### AWS (Elastic Beanstalk)

```bash
# Install EB CLI
pip install awsebcli

# Initialize
eb init -p node.js testcopilot-cloud

# Create environment
eb create testcopilot-prod

# Deploy
eb deploy
```

### DigitalOcean App Platform

1. Connect GitHub repository
2. Set environment variables in dashboard
3. Auto-deploys on push to main

## 🔧 Development

### Project Structure

```
cloud-backend/
├── src/
│   ├── config/          # Configuration
│   ├── models/          # Database models
│   ├── routes/          # API routes
│   ├── services/        # Business logic
│   │   ├── PromptService.ts  # SECRET PROMPTS
│   │   └── OpenAIService.ts
│   ├── middleware/      # Express middleware
│   ├── utils/           # Utilities
│   ├── scripts/         # DB migrations, etc.
│   └── server.ts        # Main entry point
├── tests/               # Tests
├── package.json
├── tsconfig.json
└── .env
```

### Testing

```bash
# Run tests
npm test

# Test with curl
curl http://localhost:3000/health

curl -X POST http://localhost:3000/v1/generate \
  -H "Authorization: Bearer tc_test_123" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "pageobject",
    "userQuery": "Login page",
    "elements": [
      {"name": "username", "locatorType": "id", "locatorValue": "username"}
    ]
  }'
```

### Logging

Logs are written to:
- Console (all environments)
- `logs/combined.log` (all logs)
- `logs/error.log` (errors only)

### Monitoring

Use environment variables to configure monitoring:

```bash
# Sentry
SENTRY_DSN=https://...

# DataDog
DD_API_KEY=...
DD_APP_KEY=...
```

## 🔒 Security

- **API Keys**: Stored hashed in database
- **Rate Limiting**: Per-user and global limits
- **Helmet**: Security headers
- **CORS**: Configurable origins
- **Input Validation**: All inputs validated
- **SQL Injection**: Parameterized queries
- **Secrets**: Environment variables only

## 💰 Cost Optimization

### OpenAI Costs

- GPT-4 Turbo: ~$0.02/1K tokens
- Average request: ~500 tokens = $0.01
- 1000 requests/month = ~$10

### Hosting

- Heroku: $7/month (basic)
- DigitalOcean: $5/month (basic droplet)
- AWS: ~$10-20/month (t3.micro)

### Database

- Heroku Postgres: $9/month (mini)
- DigitalOcean: $15/month (managed)
- AWS RDS: ~$15/month (db.t3.micro)

**Total**: ~$30-50/month for hosting + OpenAI costs

## 📈 Scaling

### Horizontal Scaling

Use load balancer + multiple instances:

```bash
# Heroku
heroku ps:scale web=3

# Docker Swarm
docker service scale testcopilot-cloud=3
```

### Caching

Add Redis for response caching:

```typescript
// Cache common generations
const cacheKey = `${action}:${hash(userPrompt)}`;
const cached = await redis.get(cacheKey);
if (cached) return JSON.parse(cached);

// ... generate ...

await redis.setex(cacheKey, 3600, JSON.stringify(result));
```

### Database Optimization

- Connection pooling (already configured)
- Read replicas for usage queries
- Partitioning usage table by month

## 🐛 Troubleshooting

### Database Connection Failed

```bash
# Check PostgreSQL is running
pg_isready

# Check connection
psql testcopilot

# Check environment variables
echo $DB_HOST $DB_PORT $DB_NAME
```

### OpenAI API Errors

```bash
# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Check rate limits
# OpenAI dashboard: https://platform.openai.com/usage
```

### Port Already in Use

```bash
# Find process using port 3000
lsof -i :3000

# Kill it
kill -9 <PID>
```

## 📚 Additional Resources

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Express.js Best Practices](https://expressjs.com/en/advanced/best-practice-performance.html)

## 🤝 Support

For issues or questions:
- GitHub Issues: https://github.com/yourusername/testcopilot
- Email: support@testcopilot.ai

---

**License**: MIT
**Version**: 1.0.0
