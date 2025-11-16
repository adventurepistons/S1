# Quick Start Guide

Get the cloud backend running in 5 minutes!

## Option 1: Docker Compose (Recommended)

```bash
# 1. Set your OpenAI API key
export OPENAI_API_KEY=sk-your-key-here

# 2. Start everything (database + backend)
docker-compose up

# Server running on http://localhost:3000
```

That's it! The database migrations run automatically.

## Option 2: Local Development

### 1. Install PostgreSQL

```bash
# macOS
brew install postgresql@14
brew services start postgresql@14

# Ubuntu/Debian
sudo apt-get install postgresql-14

# Create database
createdb testcopilot
```

### 2. Install Dependencies

```bash
npm install
```

### 3. Configure Environment

```bash
# Copy example
cp .env.example .env

# Edit .env and set:
nano .env
```

Required:
- `OPENAI_API_KEY=sk-...` (your OpenAI API key)
- `DB_PASSWORD=...` (your PostgreSQL password)

### 4. Run Migrations

```bash
npm run migrate
```

### 5. Start Server

```bash
# Development (auto-reload)
npm run dev

# Production
npm run build
npm start
```

Server runs on http://localhost:3000

## Testing the API

### 1. Check Health

```bash
curl http://localhost:3000/health
```

Expected:
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

### 2. Create a Test User

```bash
# Connect to database
psql testcopilot

# Create user
INSERT INTO users (id, email, password_hash, api_key, plan, is_active)
VALUES (
  gen_random_uuid(),
  'test@example.com',
  '$2b$10$dummy',
  'tc_test_your_key_here_12345',
  'pro',
  true
);
```

### 3. Test Code Generation

```bash
curl -X POST http://localhost:3000/v1/generate \
  -H "Authorization: Bearer tc_test_your_key_here_12345" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "pageobject",
    "userQuery": "Login page with username and password fields",
    "spec": "Login page",
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
      },
      {
        "name": "loginButton",
        "locatorType": "css",
        "locatorValue": "button[type=\"submit\"]"
      }
    ]
  }'
```

### 4. Check Usage

```bash
curl http://localhost:3000/v1/usage \
  -H "Authorization: Bearer tc_test_your_key_here_12345"
```

## Connecting Local Backend

Update your local backend (Go) environment:

```bash
export TESTCOPILOT_CLOUD_URL="http://localhost:3000"
export TESTCOPILOT_API_KEY="tc_test_your_key_here_12345"

# Start local backend
./copilot-core/server --port 8080
```

Now the complete flow works:
1. Extension → Local Go Backend (port 8080)
2. Local Backend → Cloud Backend (port 3000)
3. Cloud Backend → OpenAI
4. Response flows back

## Next Steps

- [ ] Add your OpenAI API key to `.env`
- [ ] Create real users with proper passwords
- [ ] Set up Stripe for billing (optional)
- [ ] Deploy to production (see README.md)
- [ ] Configure monitoring (Sentry, DataDog)
- [ ] Set up domain and SSL

## Troubleshooting

**Database connection error?**
```bash
# Check PostgreSQL is running
pg_isready

# Check you can connect
psql testcopilot
```

**OpenAI API error?**
```bash
# Verify your API key
echo $OPENAI_API_KEY

# Test it
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

**Port already in use?**
```bash
# Find what's using port 3000
lsof -i :3000

# Kill it
kill -9 <PID>

# Or use a different port
PORT=3001 npm run dev
```

## Production Deployment

See [README.md](README.md) for deployment options:
- Heroku
- AWS Elastic Beanstalk
- DigitalOcean App Platform
- Docker on any cloud

---

Need help? Check [README.md](README.md) or create an issue.
