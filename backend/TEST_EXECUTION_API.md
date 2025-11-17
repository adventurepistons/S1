# Test Execution API Documentation

**Phase 2**: Test Execution & Reporting

---

## 🧪 Test Execution Endpoints

### 1. Execute Tests

**POST** `/api/v1/tests/execute`

Run tests for a project.

**Headers**:
```
Authorization: Bearer <access_token>
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
  "websocketUrl": "ws://localhost:3000/tests/stream/exec_550e8400..."
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:3000/api/v1/tests/execute \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "projectId": "proj_123",
    "tests": ["*"],
    "environment": "staging",
    "browser": "chrome",
    "headless": true
  }'
```

---

### 2. Get Execution Status

**GET** `/api/v1/tests/status/:executionId`

Get current status of test execution.

**Response** (200 OK):
```json
{
  "executionId": "exec_123",
  "status": "running",
  "startedAt": "2025-11-17T10:30:00Z",
  "summary": {
    "total": 10,
    "passed": 7,
    "failed": 2,
    "skipped": 1
  },
  "tests": [
    {
      "id": "result_789",
      "testName": "LoginTest",
      "status": "passed",
      "duration": 3200
    }
  ]
}
```

---

### 3. Cancel Execution

**DELETE** `/api/v1/tests/cancel/:executionId`

Cancel a running or queued execution.

**Response** (200 OK):
```json
{
  "message": "Execution cancelled"
}
```

---

### 4. Get Execution Report

**GET** `/api/v1/tests/reports/:executionId`

Get detailed execution report with all test results.

**Response** (200 OK):
```json
{
  "executionId": "exec_123",
  "projectId": "proj_456",
  "projectName": "My Test Project",
  "framework": "selenium-java",
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
      "id": "result_790",
      "testName": "CheckoutTest",
      "className": "com.example.CheckoutTest",
      "status": "failed",
      "duration": 8500,
      "retries": 2,
      "errorMessage": "Element not found: #submit-btn",
      "stackTrace": "org.openqa.selenium.NoSuchElementException...",
      "screenshotUrl": "https://cdn.../screenshot_790.png",
      "videoUrl": "https://cdn.../video_790.mp4"
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

### 5. Get Execution History

**GET** `/api/v1/tests/history/:projectId?days=30`

Get execution history for a project.

**Query Parameters**:
- `days` (optional): Number of days to look back (default: 30)

**Response** (200 OK):
```json
{
  "projectId": "proj_123",
  "days": 30,
  "executions": [
    {
      "id": "exec_123",
      "status": "completed",
      "createdAt": "2025-11-17T10:30:00Z",
      "duration": 60,
      "totalTests": 10,
      "passedTests": 9,
      "failedTests": 1,
      "environment": "staging",
      "browser": "chrome"
    }
  ],
  "trends": {
    "totalExecutions": 45,
    "successfulExecutions": 40,
    "successRate": 88.9,
    "averagePassRate": 92.3,
    "averageDuration": 58
  }
}
```

---

### 6. Get Flaky Tests

**GET** `/api/v1/tests/flaky/:projectId`

Get list of flaky tests for a project.

**Response** (200 OK):
```json
[
  {
    "testName": "PaymentTest",
    "className": "com.example.PaymentTest",
    "flakyScore": 0.67,
    "totalRuns": 30,
    "passedRuns": 10,
    "failedRuns": 20,
    "lastFailedAt": "2025-11-17T10:25:00Z"
  },
  {
    "testName": "LoginTest",
    "className": "com.example.LoginTest",
    "flakyScore": 0.25,
    "totalRuns": 50,
    "passedRuns": 37,
    "failedRuns": 13,
    "lastFailedAt": "2025-11-16T14:30:00Z"
  }
]
```

**Flaky Score**: 0.0 to 1.0 (higher = more flaky)
- Score calculated as: `failedRuns / totalRuns`
- Tests with score between 0.1 and 0.9 are considered flaky
- Sorted by score descending

---

### 7. Get Test Coverage

**GET** `/api/v1/tests/coverage/:projectId`

Get test coverage information.

**Response** (200 OK):
```json
{
  "totalTests": 50,
  "executedTests": 35,
  "coverage": 70.0,
  "tests": [
    {
      "testName": "LoginTest",
      "executions": 15
    },
    {
      "testName": "CheckoutTest",
      "executions": 8
    }
  ]
}
```

---

## 🌐 WebSocket Real-Time Logs

### Connect to WebSocket

```javascript
const socket = io('ws://localhost:3000/tests', {
  auth: {
    token: 'YOUR_ACCESS_TOKEN'
  }
});

// Subscribe to execution logs
socket.emit('stream', { executionId: 'exec_123' });

// Listen for logs
socket.on('log', (message) => {
  console.log(message);
});

// Unsubscribe
socket.emit('unsubscribe', { executionId: 'exec_123' });
```

### WebSocket Message Types

#### 1. Status Update
```json
{
  "type": "status",
  "executionId": "exec_123",
  "status": "running",
  "timestamp": "2025-11-17T10:30:00Z"
}
```

#### 2. Log Message
```json
{
  "type": "log",
  "executionId": "exec_123",
  "testName": "LoginTest",
  "level": "info",
  "message": "Test started",
  "timestamp": "2025-11-17T10:30:01Z"
}
```

#### 3. Test Started
```json
{
  "type": "test_started",
  "executionId": "exec_123",
  "testName": "LoginTest",
  "timestamp": "2025-11-17T10:30:01Z"
}
```

#### 4. Test Completed
```json
{
  "type": "test_completed",
  "executionId": "exec_123",
  "testName": "LoginTest",
  "status": "passed",
  "duration": 3200,
  "timestamp": "2025-11-17T10:30:04Z"
}
```

#### 5. Screenshot Captured
```json
{
  "type": "screenshot",
  "executionId": "exec_123",
  "testName": "CheckoutTest",
  "url": "https://cdn.../screenshot_123.png",
  "timestamp": "2025-11-17T10:30:05Z"
}
```

#### 6. Execution Completed
```json
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

---

## 🎯 Usage Examples

### Complete Flow: Execute Tests + Monitor Logs

```javascript
// 1. Start execution
const response = await fetch('http://localhost:3000/api/v1/tests/execute', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    projectId: 'proj_123',
    tests: ['*'],
    environment: 'staging',
    browser: 'chrome',
  }),
});

const { executionId, websocketUrl } = await response.json();

// 2. Connect to WebSocket
const socket = io(websocketUrl, {
  auth: { token: 'YOUR_TOKEN' }
});

// 3. Subscribe to logs
socket.emit('stream', { executionId });

// 4. Listen for real-time updates
socket.on('log', (message) => {
  if (message.type === 'test_completed') {
    console.log(`${message.testName}: ${message.status} (${message.duration}ms)`);
  }
});

socket.on('log', (message) => {
  if (message.type === 'execution_completed') {
    console.log('Execution finished:', message.summary);
    socket.disconnect();
  }
});

// 5. Get final report
setTimeout(async () => {
  const report = await fetch(`http://localhost:3000/api/v1/tests/reports/${executionId}`, {
    headers: { 'Authorization': 'Bearer YOUR_TOKEN' }
  });
  const data = await report.json();
  console.log('Final report:', data);
}, 60000);
```

---

## 🔄 Test Execution Flow

```
User Triggers Test
      ↓
POST /api/v1/tests/execute
      ↓
Create Execution Record (status: QUEUED)
      ↓
Add to BullMQ Queue
      ↓
Worker Picks Up Job
      ↓
Update Status to RUNNING
      ↓
Execute Tests (Selenium/Playwright)
      ↓
Stream Logs via WebSocket
      ↓
Save Test Results to Database
      ↓
Update Execution Status to COMPLETED
      ↓
Detect Flaky Tests
      ↓
Emit completion event
```

---

## 🚧 Current Limitations (Will be added next)

1. **Real Selenium Grid Integration** - Currently simulates test execution
2. **Screenshot Capture** - Placeholder URLs (will integrate with S3)
3. **Video Recording** - Not yet implemented
4. **Parallel Execution** - Queue setup ready, needs worker scaling
5. **Browser Selection** - Chrome, Firefox, Safari (needs Selenium Grid setup)

---

## ✅ What's Working Now

- ✅ Test execution queueing
- ✅ Real-time WebSocket streaming
- ✅ Test result storage
- ✅ Execution reports
- ✅ Flaky test detection
- ✅ Test history & trends
- ✅ Test coverage analysis
- ✅ Cancellation support

---

## 🔜 Next Steps

1. Integrate with Selenium Grid for real test execution
2. Add screenshot capture using S3
3. Add video recording (Playwright)
4. Build VSCode UI for test execution panel
5. Add CI/CD webhook support

