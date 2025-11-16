# Deployment Guide - Test Automation Copilot

## Overview

This guide explains how to deploy the complete Test Automation Copilot system with cloud-AI architecture.

## Architecture Summary

```
┌─────────────────┐
│  VS Code Ext    │ (Published to Marketplace)
└────────┬────────┘
         │ HTTP/WebSocket
┌────────┴────────┐
│ Local Backend   │ (Bundled with extension)
│ (Go Binary)     │
└────────┬────────┘
         │ HTTPS
┌────────┴────────┐
│ Cloud Backend   │ (Your proprietary server)
│ (Node.js/Go)    │
└────────┬────────┘
         │ API
┌────────┴────────┐
│  OpenAI GPT-4   │
└─────────────────┘
```

## Phase 1: Build Local Backend

### 1.1 Build Go Binary

```bash
cd test-automation-copilot/copilot-core

# For Linux/Mac
GOOS=linux GOARCH=amd64 go build -o bin/test-copilot-server-linux cmd/server/main.go

# For Mac (Intel)
GOOS=darwin GOARCH=amd64 go build -o bin/test-copilot-server-mac cmd/server/main.go

# For Mac (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/test-copilot-server-mac-arm cmd/server/main.go

# For Windows
GOOS=windows GOARCH=amd64 go build -o bin/test-copilot-server.exe cmd/server/main.go
```

### 1.2 Test Local Backend

```bash
# Set environment variables (temporary for testing)
export TESTCOPILOT_CLOUD_URL="http://localhost:3000"
export TESTCOPILOT_API_KEY="test_key_123"

# Start server
./bin/test-copilot-server-linux --port 8080

# Test health endpoint
curl http://localhost:8080/health

# Test indexing (in another terminal)
curl -X POST http://localhost:8080/api/v1/workspace/index \
  -H "Content-Type: application/json" \
  -d '{"workspacePath": "/path/to/test/project"}'
```

## Phase 2: Build VS Code Extension

### 2.1 Install Dependencies

```bash
cd test-automation-copilot
npm install
```

### 2.2 Build Extension

```bash
npm run build
```

### 2.3 Test Extension Locally

```bash
# Press F5 in VS Code to launch extension in debug mode
# Or package it:
npm run package
```

This creates `test-automation-copilot-0.0.1.vsix`

### 2.4 Install Locally for Testing

```bash
code --install-extension test-automation-copilot-0.0.1.vsix
```

## Phase 3: Implement Cloud Backend

### 3.1 Choose Technology Stack

**Recommended: Node.js + Express**

Pros:
- Fast development
- Great streaming support
- Easy deployment to serverless
- Good OpenAI SDK

**Alternative: Go + Gin**

Pros:
- Better performance
- Type safety
- Easier to manage alongside Go backend

**Alternative: Python + FastAPI**

Pros:
- Quick prototyping
- Excellent AI/ML ecosystem
- Easy prompt management

### 3.2 Implement Endpoints

Based on [CLOUD_BACKEND_SPEC.md](CLOUD_BACKEND_SPEC.md), implement:

1. **`GET /health`** - Health check
2. **`POST /v1/generate`** - Non-streaming generation
3. **`POST /v1/generate/stream`** - Streaming generation
4. **`GET /v1/usage`** - Usage statistics

### 3.3 Example: Node.js Implementation

```bash
# Create new project
mkdir testcopilot-cloud-backend
cd testcopilot-cloud-backend
npm init -y

# Install dependencies
npm install express openai dotenv express-rate-limit
npm install --save-dev typescript @types/express @types/node

# Create TypeScript config
npx tsc --init
```

**`src/server.ts`** (simplified example):

```typescript
import express from 'express';
import OpenAI from 'openai';
import rateLimit from 'express-rate-limit';

const app = express();
const openai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });

app.use(express.json());

// Rate limiting
const limiter = rateLimit({
  windowMs: 60 * 1000, // 1 minute
  max: 20 // 20 requests per minute
});
app.use('/v1/', limiter);

// Authentication middleware
function authenticate(req, res, next) {
  const apiKey = req.headers.authorization?.replace('Bearer ', '');

  // Validate API key against your database
  if (!apiKey || !isValidApiKey(apiKey)) {
    return res.status(401).json({ error: 'Invalid API key' });
  }

  req.userId = getUserIdFromApiKey(apiKey);
  next();
}

// Generate endpoint
app.post('/v1/generate', authenticate, async (req, res) => {
  const { action, userQuery, relevantCode, ...context } = req.body;

  try {
    // Build prompt based on action
    const prompt = buildPrompt(action, context);

    // Call OpenAI
    const response = await openai.chat.completions.create({
      model: 'gpt-4-turbo-preview',
      messages: [
        { role: 'system', content: getSystemPrompt(action) },
        { role: 'user', content: prompt }
      ],
      temperature: 0.7
    });

    const code = response.choices[0].message.content;

    // Track usage
    await trackUsage(req.userId, {
      action,
      tokensUsed: response.usage.total_tokens,
      cost: calculateCost(response.usage)
    });

    res.json({
      code,
      tokensUsed: response.usage.total_tokens,
      model: response.model,
      finishReason: response.choices[0].finish_reason
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// Streaming endpoint
app.post('/v1/generate/stream', authenticate, async (req, res) => {
  const { action, userQuery, relevantCode, ...context } = req.body;

  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');

  try {
    const prompt = buildPrompt(action, context);

    const stream = await openai.chat.completions.create({
      model: 'gpt-4-turbo-preview',
      messages: [
        { role: 'system', content: getSystemPrompt(action) },
        { role: 'user', content: prompt }
      ],
      stream: true,
      temperature: 0.7
    });

    let fullCode = '';
    let totalTokens = 0;

    for await (const chunk of stream) {
      const content = chunk.choices[0]?.delta?.content || '';
      if (content) {
        fullCode += content;
        res.write(`data: ${JSON.stringify({ type: 'chunk', content })}\n\n`);
      }
    }

    // Send completion
    res.write(`data: ${JSON.stringify({
      type: 'complete',
      result: {
        code: fullCode,
        tokensUsed: totalTokens,
        model: 'gpt-4-turbo-preview',
        finishReason: 'stop'
      }
    })}\n\n`);

    res.end();
  } catch (error) {
    res.write(`data: ${JSON.stringify({ type: 'error', error: error.message })}\n\n`);
    res.end();
  }
});

function buildPrompt(action: string, context: any): string {
  // YOUR PROPRIETARY PROMPT ENGINEERING GOES HERE
  // This is your competitive advantage - keep it secret!

  switch (action) {
    case 'pageobject':
      return `Generate a ${context.framework} page object class.

Context:
${context.relevantCode?.map(c => c.content).join('\n\n')}

User Request: ${context.userQuery}

Elements:
${context.elements?.map(e => `- ${e.name}: ${e.locatorType}=${e.locatorValue}`).join('\n')}

Generate complete Java class with PageFactory pattern.`;

    // Add cases for 'test', 'chat', 'fix'
    default:
      return context.userQuery;
  }
}

function getSystemPrompt(action: string): string {
  const basePrompt = 'You are an expert test automation engineer.';

  switch (action) {
    case 'pageobject':
      return `${basePrompt} Generate Selenium page object classes following best practices.`;
    case 'test':
      return `${basePrompt} Generate test cases using appropriate test frameworks.`;
    case 'chat':
      return `${basePrompt} Answer questions about test automation.`;
    case 'fix':
      return `${basePrompt} Fix broken test automation code.`;
    default:
      return basePrompt;
  }
}

app.listen(3000, () => {
  console.log('Cloud backend running on port 3000');
});
```

### 3.4 Implement Database for User Management

**Schema**:

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  api_key VARCHAR(64) UNIQUE NOT NULL,
  plan VARCHAR(50) DEFAULT 'free',
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE usage (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  action VARCHAR(50),
  tokens_used INTEGER,
  cost DECIMAL(10, 4),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE api_keys (
  key VARCHAR(64) PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW()
);
```

## Phase 4: Deploy Cloud Backend

### Option 1: AWS Lambda + API Gateway

```bash
# Install Serverless Framework
npm install -g serverless

# Create serverless.yml
serverless create --template aws-nodejs-typescript

# Deploy
serverless deploy --stage prod
```

### Option 2: Google Cloud Run

```bash
# Build Docker image
docker build -t gcr.io/your-project/testcopilot-backend .

# Push to GCR
docker push gcr.io/your-project/testcopilot-backend

# Deploy
gcloud run deploy testcopilot-backend \
  --image gcr.io/your-project/testcopilot-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### Option 3: Heroku

```bash
# Create Heroku app
heroku create testcopilot-api

# Set environment variables
heroku config:set OPENAI_API_KEY=sk-...

# Deploy
git push heroku main
```

### Option 4: DigitalOcean App Platform

- Push code to GitHub
- Connect to DigitalOcean App Platform
- Auto-deploys on push

## Phase 5: Configure Domain & SSL

### 5.1 Purchase Domain

Buy `testcopilot.ai` or similar from:
- Namecheap
- Google Domains
- Cloudflare

### 5.2 Set up DNS

Point to your deployment:
```
api.testcopilot.ai → [Your Cloud Backend IP/URL]
www.testcopilot.ai → [Your Marketing Site]
```

### 5.3 Enable SSL

Most cloud providers (Heroku, Cloud Run, etc.) provide free SSL certificates automatically.

## Phase 6: Publish Extension

### 6.1 Create Publisher Account

1. Go to https://marketplace.visualstudio.com/manage
2. Create publisher (e.g., "test-automation-tools")
3. Get Personal Access Token

### 6.2 Update package.json

```json
{
  "publisher": "test-automation-tools",
  "repository": {
    "type": "git",
    "url": "https://github.com/yourusername/test-automation-copilot"
  },
  "icon": "resources/icon.png",
  "galleryBanner": {
    "color": "#1e1e1e",
    "theme": "dark"
  }
}
```

### 6.3 Publish

```bash
# Install vsce
npm install -g @vscode/vsce

# Package extension
vsce package

# Publish
vsce publish
```

## Phase 7: Set Up Billing

### 7.1 Choose Payment Processor

- **Stripe**: Recommended for subscriptions
- **Paddle**: Good for global payments
- **LemonSqueezy**: Simple for SaaS

### 7.2 Implement Subscription Plans

**Example Plans**:

| Plan | Price | Requests/Month | Support |
|------|-------|----------------|---------|
| Free | $0 | 100 | Community |
| Pro | $20 | 1,000 | Email |
| Team | $50 | 5,000 | Priority |
| Enterprise | Custom | Unlimited | Dedicated |

### 7.3 Stripe Integration

```typescript
import Stripe from 'stripe';
const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);

// Create subscription
app.post('/api/subscribe', async (req, res) => {
  const { email, priceId } = req.body;

  const customer = await stripe.customers.create({ email });

  const subscription = await stripe.subscriptions.create({
    customer: customer.id,
    items: [{ price: priceId }]
  });

  // Generate API key
  const apiKey = generateApiKey();
  await saveUser({ email, apiKey, plan: 'pro', stripeCustomerId: customer.id });

  res.json({ apiKey, subscription });
});
```

## Phase 8: Monitoring & Analytics

### 8.1 Set Up Monitoring

**Tools**:
- **Sentry**: Error tracking
- **DataDog**: Performance monitoring
- **LogRocket**: User session replay

### 8.2 Track Metrics

```typescript
// Track usage
app.use((req, res, next) => {
  const start = Date.now();

  res.on('finish', () => {
    const duration = Date.now() - start;

    analytics.track({
      event: 'API Request',
      userId: req.userId,
      properties: {
        endpoint: req.path,
        method: req.method,
        duration,
        statusCode: res.statusCode
      }
    });
  });

  next();
});
```

### 8.3 Dashboard

Use Grafana or similar to visualize:
- Requests per second
- Average response time
- Token usage
- Error rate
- User growth

## Phase 9: Marketing & Launch

### 9.1 Create Landing Page

**Essential Pages**:
- Homepage (value proposition)
- Features
- Pricing
- Documentation
- Blog

**Tech Stack**:
- Next.js + Tailwind CSS
- Vercel for hosting
- MDX for docs

### 9.2 Launch Checklist

- [ ] Cloud backend deployed and tested
- [ ] Extension published to marketplace
- [ ] Landing page live
- [ ] Documentation complete
- [ ] Billing integration working
- [ ] Support email set up
- [ ] Privacy policy & terms of service
- [ ] Analytics configured

### 9.3 Promotion

- Post on Reddit (/r/selenium, /r/QualityAssurance)
- Product Hunt launch
- LinkedIn posts
- Twitter/X announcements
- YouTube demo video

## Cost Estimates

### Monthly Costs

**For 100 users @ $20/month = $2,000 revenue**

| Item | Cost | Notes |
|------|------|-------|
| Cloud hosting | $50-100 | Heroku/Cloud Run |
| OpenAI API | $800-1,000 | ~40K requests @ $0.02 each |
| Database | $25 | Managed PostgreSQL |
| Monitoring | $50 | Sentry + DataDog |
| Domain | $3 | Annual / 12 |
| Email | $10 | SendGrid/Mailgun |
| **Total** | **$938-1,188** | **Profit: $812-1,062** |

**Break-even**: ~50 paying users

## Security Checklist

- [ ] API keys stored encrypted
- [ ] Rate limiting implemented
- [ ] SQL injection protection
- [ ] XSS protection
- [ ] CORS configured properly
- [ ] HTTPS only
- [ ] Environment variables secured
- [ ] Regular security audits
- [ ] DDoS protection (Cloudflare)
- [ ] Backup database daily

## Support & Maintenance

### 9.4 Support Channels

- Email: support@testcopilot.ai
- Discord community
- GitHub issues (for bugs)
- Documentation site

### 9.5 Maintenance Tasks

**Weekly**:
- Check error logs
- Review support tickets
- Monitor usage trends

**Monthly**:
- Update dependencies
- Review and improve prompts
- Analyze user feedback
- Generate usage reports

**Quarterly**:
- Security audit
- Performance optimization
- Feature planning
- User surveys

## Success Metrics

Track these KPIs:

1. **User Growth**
   - New signups/week
   - Activation rate (% who generate code)
   - Churn rate

2. **Engagement**
   - Requests per user/month
   - Active users (daily/weekly)
   - Feature usage distribution

3. **Revenue**
   - MRR (Monthly Recurring Revenue)
   - LTV (Lifetime Value)
   - CAC (Customer Acquisition Cost)

4. **Technical**
   - API latency (p95)
   - Error rate
   - Uptime %

## Next Steps

1. **Week 1-2**: Implement cloud backend MVP
2. **Week 3**: Test end-to-end integration
3. **Week 4**: Deploy to staging environment
4. **Week 5**: Beta testing with 10 users
5. **Week 6**: Fix bugs, improve prompts
6. **Week 7**: Public launch
7. **Week 8+**: Iterate based on feedback

---

**Questions?** Create an issue on GitHub or email support@testcopilot.ai

**Last Updated**: 2025-11-16
