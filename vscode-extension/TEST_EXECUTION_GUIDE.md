# Test Execution Guide - VSCode Extension

Run your tests directly from VSCode with real-time results! 🧪

---

## ✨ Features

- ✅ **Execute tests from VSCode** - No need to switch to terminal
- ✅ **Real-time log streaming** - Watch tests run live via WebSocket
- ✅ **Visual test results** - See pass/fail/skip counts instantly
- ✅ **Screenshot capture** - Auto-capture screenshots on failures
- ✅ **Test history** - View past executions and trends
- ✅ **Flaky test detection** - Identify unreliable tests
- ✅ **Multiple browsers** - Chrome, Firefox, Safari, Edge
- ✅ **Parallel execution** - Run tests faster

---

## 🚀 Quick Start

### 1. Open Test Execution Panel

Press `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux) and type:

```
Test Copilot: Execute Tests
```

### 2. Configure Test Execution

Fill in the form:
- **Project ID**: Your project ID (e.g., `proj_123`)
- **Tests**: Comma-separated test names or `*` for all tests
- **Environment**: `dev`, `staging`, or `production`
- **Browser**: `chrome`, `firefox`, `safari`, or `edge`
- **Headless**: Run browser in headless mode (no UI)
- **Parallel**: Run tests in parallel
- **Max Workers**: Number of parallel workers (1-10)
- **Retry Failed**: Number of retries for failed tests (0-5)

### 3. Click "Execute Tests"

Watch your tests run in real-time! 🎉

---

## 📊 Understanding Test Results

### Test Summary Cards

After execution, you'll see:

```
┌─────────┬─────────┬─────────┬─────────┐
│ Total   │ Passed  │ Failed  │ Skipped │
│   10    │    9    │    1    │    0    │
└─────────┴─────────┴─────────┴─────────┘
```

### Real-Time Logs

```
[10:30:01] 🚀 Test execution started
[10:30:02] ▶️  Starting: LoginTest
[10:30:05] ✅ LoginTest: passed (3200ms)
[10:30:06] ▶️  Starting: CheckoutTest
[10:30:14] ❌ CheckoutTest: failed (8500ms)
[10:30:15] 📸 Screenshot captured: CheckoutTest
[10:30:15] ✨ Execution completed
```

### Status Badges

- **🟦 Queued** - Test is in queue
- **🔵 Running** - Test is executing
- **🟢 Completed** - All tests finished
- **🔴 Failed** - Execution failed
- **🟡 Cancelled** - User cancelled

---

## 🔧 Backend Setup (Required)

The VSCode extension connects to your backend API.

### 1. Start Backend Server

```bash
cd backend
npm install
npm run docker:up
npm run prisma:migrate
npm run start:dev
```

Backend runs at `http://localhost:3000`

### 2. Configure Backend URL (Optional)

If your backend is not at localhost:3000, set environment variable:

```bash
export BACKEND_URL=http://your-backend-url:3000
```

---

## 🎨 Features in Detail

### Real-Time Streaming

Uses **WebSocket** to stream logs as tests execute:

```javascript
// Connection happens automatically
ws://localhost:3000/tests
```

Messages you'll receive:
- `status` - Execution status changes
- `log` - Console logs from tests
- `test_started` - Test begins
- `test_completed` - Test finishes
- `screenshot` - Screenshot captured
- `execution_completed` - All done!

### Screenshot Capture

Screenshots are automatically captured for **failed tests** and uploaded to S3.

You'll see:
```
📸 Screenshot captured: CheckoutTest
https://cdn.../screenshot_123.png
```

### Parallel Execution

Run multiple tests simultaneously:

```
Tests: LoginTest, CheckoutTest, PaymentTest
Parallel: ✅
Max Workers: 4

→ All 3 tests run at the same time!
```

### Retry Failed Tests

Automatically retry flaky tests:

```
Retry Failed: 2

→ If test fails, retry up to 2 times
→ Only marked failed if all retries fail
```

---

## 📋 API Integration

The VSCode extension uses these backend endpoints:

### Execute Tests
```http
POST /api/v1/tests/execute
Authorization: Bearer <token>

{
  "projectId": "proj_123",
  "tests": ["LoginTest"],
  "browser": "chrome",
  "headless": true
}
```

### Cancel Execution
```http
DELETE /api/v1/tests/cancel/:executionId
Authorization: Bearer <token>
```

### Get Results
```http
GET /api/v1/tests/reports/:executionId
Authorization: Bearer <token>
```

---

## 🔐 Authentication

The extension uses your backend JWT token stored securely in VSCode secrets.

### Sign In Flow

1. User clicks "Sign In" in VSCode
2. Extension opens Firebase auth
3. User signs in with Google
4. Token stored in VSCode secrets
5. All API requests include token

### Token Storage

```typescript
// Store token
await context.secrets.store('accessToken', token);

// Retrieve token
const token = await context.secrets.get('accessToken');
```

---

## 🐛 Troubleshooting

### "Failed to execute tests"

**Problem**: Can't connect to backend

**Solutions**:
1. Make sure backend is running: `npm run start:dev`
2. Check backend URL: `http://localhost:3000`
3. Verify you're signed in
4. Check console: `View > Output > Test Copilot`

### "Authentication failed"

**Problem**: Token expired or invalid

**Solution**:
1. Sign out: `Cmd+Shift+P` → `Test Copilot: Sign Out`
2. Sign in again: `Test Copilot: Sign In`

### "Quota exceeded"

**Problem**: Monthly usage limit reached

**Solution**:
1. Use your own API key (BYOK mode)
2. Or upgrade to Pro tier
3. Command: `Test Copilot: Switch API Mode`

### WebSocket not connecting

**Problem**: Real-time logs not showing

**Solutions**:
1. Check if Redis is running: `docker ps`
2. Restart backend: `npm run start:dev`
3. Check firewall settings

### Tests not executing

**Problem**: Tests queued but not running

**Solutions**:
1. Check worker is running: `docker ps | grep worker`
2. Check Redis queue: `redis-cli llen bull:test-execution:waiting`
3. View worker logs: `docker logs testcopilot-worker`

---

## 🎯 Examples

### Example 1: Run All Tests

```
Project ID: proj_abc123
Tests: *
Environment: staging
Browser: chrome
Headless: ✅
Parallel: ✅
Max Workers: 4
```

### Example 2: Run Specific Tests

```
Project ID: proj_abc123
Tests: LoginTest, CheckoutTest, PaymentTest
Environment: dev
Browser: firefox
Headless: ❌  (see browser)
Parallel: ❌
Max Workers: 1
```

### Example 3: Smoke Tests Only

```
Project ID: proj_abc123
Tests: *
Tags: smoke, critical
Environment: production
Browser: chrome
Headless: ✅
Parallel: ✅
Max Workers: 8
Retry Failed: 2
```

---

## 📈 Advanced Features

### View Test History

```typescript
const apiClient = getApiClient();
const response = await apiClient.get(`/tests/history/${projectId}?days=30`);

// See past executions
response.data.executions.forEach(exec => {
  console.log(`${exec.createdAt}: ${exec.passedTests}/${exec.totalTests} passed`);
});
```

### Get Flaky Tests

```typescript
const response = await apiClient.get(`/tests/flaky/${projectId}`);

// Tests that sometimes pass, sometimes fail
response.data.forEach(test => {
  console.log(`${test.testName}: ${test.flakyScore * 100}% flaky`);
});
```

### Test Coverage

```typescript
const response = await apiClient.get(`/tests/coverage/${projectId}`);

console.log(`Coverage: ${response.data.coverage}%`);
console.log(`Executed: ${response.data.executedTests}/${response.data.totalTests}`);
```

---

## 🔜 Coming Soon

- [ ] **Video Recording** - Record test execution as video
- [ ] **Visual Testing** - Screenshot comparison
- [ ] **API Testing** - REST/GraphQL test execution
- [ ] **CI/CD Integration** - Auto-run tests on PR
- [ ] **Test Scheduling** - Schedule recurring test runs
- [ ] **Mobile Testing** - iOS/Android support

---

## 💡 Tips & Best Practices

### 1. Use Headless Mode in CI/CD

```
Headless: ✅  (faster, no UI)
```

### 2. Parallel for Regression Suites

```
Tests: *
Parallel: ✅
Max Workers: 8
```

### 3. Retry for Flaky Tests

```
Retry Failed: 2  (give flaky tests a chance)
```

### 4. Use Tags for Selective Execution

```
Tags: smoke  (run only smoke tests)
Tags: critical, p0  (high-priority tests)
```

### 5. Different Browsers for Cross-Browser Testing

```
Run 1: Browser: chrome
Run 2: Browser: firefox
Run 3: Browser: safari
```

---

## 🤝 Contributing

Found a bug? Have a feature request?

1. Open an issue on GitHub
2. Or submit a pull request
3. Join our Discord community

---

## 📚 Additional Resources

- [Backend API Documentation](../backend/TEST_EXECUTION_API.md)
- [Architecture Overview](../BACKEND_API_SPECIFICATION.md)
- [Implementation Plan](../IMPLEMENTATION_PLAN.md)

---

**Enjoy running tests from VSCode!** 🎉

