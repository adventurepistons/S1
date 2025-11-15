# 🎨 VS Code Extension Architecture - Test Automation Copilot

## 📋 Table of Contents
1. [User Journey](#user-journey)
2. [Extension UI Structure](#extension-ui-structure)
3. [Application Map](#application-map)
4. [Recording Flow](#recording-flow)
5. [Chat Interface](#chat-interface)
6. [State Management](#state-management)
7. [Backend Communication](#backend-communication)
8. [Component Architecture](#component-architecture)

---

## 🚀 User Journey (End-to-End)

### **Phase 1: Extension Activation**
```
User opens VS Code in test automation project
    ↓
Extension activates automatically (when Java/Selenium project detected)
    ↓
Shows "Test Copilot" icon in Activity Bar (left sidebar)
    ↓
User clicks icon → Opens Copilot Sidebar Panel
```

### **Phase 2: Initial State**
```
Sidebar Panel Shows:
┌─────────────────────────────────────┐
│   🤖 Test Automation Copilot        │
├─────────────────────────────────────┤
│                                     │
│   📍 Status: Not Indexed            │
│   ⚙️  Indexing workspace...         │
│   Progress: 45/100 files            │
│                                     │
├─────────────────────────────────────┤
│   📱 Application Map (0 pages)     │
│   [Empty State]                     │
│   "Start recording to build map"   │
│                                     │
│   [🎬 Start Recording]              │
│                                     │
├─────────────────────────────────────┤
│   💬 Chat (Disabled until indexed)  │
│                                     │
└─────────────────────────────────────┘
```

### **Phase 3: Workspace Indexed**
```
Indexing Complete (2-30 seconds depending on project size)
    ↓
Status changes to "Ready"
    ↓
Sidebar updates:
┌─────────────────────────────────────┐
│   🤖 Test Automation Copilot        │
├─────────────────────────────────────┤
│   ✅ Status: Ready                  │
│   📊 Indexed: 100 files             │
│       - 45 Java files               │
│       - 15 Feature files            │
│       - 5 Page Objects              │
│       - 8 Tests                     │
│                                     │
├─────────────────────────────────────┤
│   📱 Application Map (0 pages)     │
│   [Empty State]                     │
│                                     │
│   [🎬 Start Recording]              │
│                                     │
├─────────────────────────────────────┤
│   💬 Chat (Now Enabled)            │
│   Ask me anything about tests...   │
│   [Chat Input Field]                │
│                                     │
└─────────────────────────────────────┘
```

### **Phase 4: User Starts Recording**
```
User clicks "Start Recording" button
    ↓
Webview panel opens (shows browser + recording controls)
    ↓
Browser opens with recording overlay
    ↓
User navigates app + interacts
    ↓
Copilot captures everything in real-time
```

### **Phase 5: Recording in Progress**
```
Webview Panel:
┌─────────────────────────────────────────────────────────┐
│  🔴 Recording...                    [⏹ Stop Recording]  │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────┐  │
│  │  🌐 https://app.example.com/login                │  │
│  │                                                   │  │
│  │  [Actual browser embedded via Playwright]        │  │
│  │                                                   │  │
│  │  Login Page                                      │  │
│  │  Email: [____________]  👁️ Captured              │  │
│  │  Password: [________]   👁️ Captured              │  │
│  │  [Login Button]         👁️ Captured              │  │
│  │                                                   │  │
│  └──────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│  📊 Recording Stats                                     │
│  ├─ Pages visited: 3                                   │
│  ├─ Elements captured: 12                              │
│  ├─ Interactions: 8                                    │
│  └─ Duration: 02:34                                    │
├─────────────────────────────────────────────────────────┤
│  📝 Live Capture Log                                    │
│  ├─ ✅ Captured input#email (@id="email")              │
│  ├─ ✅ Captured input#password (@name="password")      │
│  ├─ ✅ Captured button.login (@class="btn-primary")    │
│  └─ ✅ Page navigated to /dashboard                    │
└─────────────────────────────────────────────────────────┘
```

### **Phase 6: Recording Complete**
```
User clicks "Stop Recording"
    ↓
Recording data saved to database
    ↓
Application Map updates automatically
    ↓
Sidebar shows updated map
    ↓
User can now chat about recorded flows
```

### **Phase 7: Application Map Populated**
```
Sidebar Panel after recording:
┌─────────────────────────────────────┐
│   🤖 Test Automation Copilot        │
├─────────────────────────────────────┤
│   ✅ Status: Ready                  │
│   📊 Last recording: 2 min ago      │
│                                     │
├─────────────────────────────────────┤
│   📱 Application Map (3 pages)     │
│   ┌───────────────────────────────┐ │
│   │ 📄 Login Page                 │ │
│   │    └─ 2 inputs, 1 button      │ │
│   │                               │ │
│   │ 📄 Dashboard                  │ │
│   │    └─ 5 elements              │ │
│   │                               │ │
│   │ 📄 Profile                    │ │
│   │    └─ 3 inputs, 2 buttons     │ │
│   └───────────────────────────────┘ │
│                                     │
│   [🎬 New Recording]                │
│   [📋 View All Recordings]          │
│                                     │
├─────────────────────────────────────┤
│   💬 Chat                           │
│   ┌─────────────────────────────┐   │
│   │ 🤖: I've learned 3 pages    │   │
│   │     from your recording!    │   │
│   │     What tests should I     │   │
│   │     create?                 │   │
│   └─────────────────────────────┘   │
│   [Chat Input Field]                │
└─────────────────────────────────────┘
```

### **Phase 8: User Chats with AI**
```
User types: "Create login tests"
    ↓
Message sent to Go backend
    ↓
Backend uses OpenAI GPT-4 to understand intent
    ↓
Backend searches recorded data + codebase
    ↓
GPT-4 generates test code
    ↓
Response streams back to UI
    ↓
Shows diff preview before applying
```

### **Phase 9: AI Response with Actions**
```
Chat Panel:
┌──────────────────────────────────────┐
│  You: Create login tests            │
│                                      │
│  🤖: I'll create comprehensive login │
│      tests based on your recording. │
│                                      │
│  📁 Files to be created:             │
│  ├─ pages/LoginPage.java ✨ NEW      │
│  ├─ tests/LoginTest.java ✨ NEW      │
│  └─ features/login.feature ✨ NEW    │
│                                      │
│  [👁️ Preview Changes]                │
│  [✅ Apply All]  [❌ Reject]          │
└──────────────────────────────────────┘
```

### **Phase 10: Code Preview & Apply**
```
User clicks "Preview Changes"
    ↓
Opens diff view in VS Code
    ↓
Shows side-by-side before/after
    ↓
User reviews changes
    ↓
Clicks "Apply All"
    ↓
Files created/modified
    ↓
Workspace re-indexed automatically
    ↓
Chat confirms completion
```

---

## 🎨 Extension UI Structure

### **1. Activity Bar Icon**
```
Location: Left sidebar (next to Explorer, Search, etc.)
Icon: 🤖 or robot icon
Tooltip: "Test Copilot"
Badge: Shows count of unread AI suggestions
```

### **2. Sidebar Panel (Primary View)**
```
Location: Replaces Explorer when clicked
Width: 300-400px (resizable)
Scrollable: Yes
Sections:
  1. Status Header
  2. Application Map (collapsible tree)
  3. Chat Interface
  4. Quick Actions
```

### **3. Webview Panel (Recording View)**
```
Location: Editor area (replaces/splits with code)
Type: Webview with embedded browser
Size: Full editor width
Features:
  - Embedded browser (Playwright)
  - Recording controls overlay
  - Live stats sidebar
  - Capture log
```

### **4. Command Palette Commands**
```
Ctrl+Shift+P → "Test Copilot: Start Recording"
Ctrl+Shift+P → "Test Copilot: Stop Recording"
Ctrl+Shift+P → "Test Copilot: Open Chat"
Ctrl+Shift+P → "Test Copilot: Analyze Workspace"
Ctrl+Shift+P → "Test Copilot: Generate Tests"
```

---

## 📱 Application Map Component

### **Visual Structure**
```
📱 Application Map
├─ 🏠 https://app.example.com
│   ├─ 📄 Login (/login)
│   │   ├─ 🔲 Email Input (id="email")
│   │   ├─ 🔒 Password Input (id="password")
│   │   └─ 🔘 Login Button (class="btn-primary")
│   │
│   ├─ 📄 Dashboard (/dashboard)
│   │   ├─ 📊 Stats Widget
│   │   ├─ 📋 User Table
│   │   └─ ➕ Add User Button
│   │
│   └─ 📄 Profile (/profile)
│       ├─ 🔲 Name Input
│       ├─ 📧 Email Input
│       └─ 💾 Save Button
│
├─ [🎬 New Recording]
└─ [📋 View All Recordings (3)]
```

### **Interactive Features**
```
✅ Click on page → Highlights in recording
✅ Click on element → Shows locator strategy
✅ Right-click → Context menu:
   - Generate Page Object
   - Generate Test
   - Copy Locator
   - Go to Implementation (if exists)
```

### **Data Source**
```typescript
interface ApplicationMap {
  url: string;
  pages: Page[];
  recordedAt: Date;
  sessionId: string;
}

interface Page {
  name: string;
  url: string;
  elements: Element[];
  interactions: Interaction[];
  screenshot?: string;
}

interface Element {
  id: string;
  type: string; // input, button, link, etc.
  locators: {
    id?: string;
    name?: string;
    css?: string;
    xpath?: string;
  };
  attributes: Record<string, string>;
}
```

---

## 🎬 Recording Flow (Detailed)

### **Step 1: User Clicks "Start Recording"**
```typescript
// src/commands/startRecording.ts
async function startRecording() {
  // 1. Show webview panel
  const panel = vscode.window.createWebviewPanel(
    'testCopilotRecorder',
    '🔴 Recording...',
    vscode.ViewColumn.One,
    { enableScripts: true }
  );

  // 2. Start backend recording session
  const sessionId = await coreClient.startRecordingSession();

  // 3. Launch browser with Playwright
  const recorder = new BrowserRecorder(sessionId);
  await recorder.start();

  // 4. Stream recording events to webview
  recorder.on('element-captured', (element) => {
    panel.webview.postMessage({
      type: 'element-captured',
      data: element
    });
  });

  // 5. Update UI in real-time
  recorder.on('page-navigated', (url) => {
    panel.webview.postMessage({
      type: 'page-navigated',
      data: { url }
    });
  });
}
```

### **Step 2: User Interacts with Browser**
```typescript
// src/browser/BrowserRecorder.ts
class BrowserRecorder {
  async captureInteraction(event: 'click' | 'input' | 'navigate') {
    // 1. Analyze element
    const element = await this.analyzer.analyzeElement(event.target);

    // 2. Generate locators
    const locators = await this.analyzer.generateLocators(element);

    // 3. Capture screenshot
    const screenshot = await this.page.screenshot({
      clip: element.boundingBox
    });

    // 4. Save to database via backend
    await this.coreClient.saveInteraction({
      sessionId: this.sessionId,
      type: event,
      element,
      locators,
      screenshot,
      timestamp: Date.now()
    });

    // 5. Emit event for UI update
    this.emit('element-captured', element);
  }
}
```

### **Step 3: User Clicks "Stop Recording"**
```typescript
async function stopRecording() {
  // 1. Stop recorder
  await recorder.stop();

  // 2. Save session to database
  const session = await coreClient.finishRecordingSession(sessionId);

  // 3. Build application map
  const appMap = await buildApplicationMap(session);

  // 4. Update sidebar UI
  sidebarProvider.updateApplicationMap(appMap);

  // 5. Close recording panel
  panel.dispose();

  // 6. Show success notification
  vscode.window.showInformationMessage(
    `✅ Recording saved! Captured ${session.pageCount} pages`
  );

  // 7. Trigger workspace re-indexing (if new pages)
  await coreClient.triggerReindex();
}
```

---

## 💬 Chat Interface (Detailed)

### **Chat UI Structure**
```
┌────────────────────────────────────┐
│  💬 Chat                           │
├────────────────────────────────────┤
│  ┌──────────────────────────────┐  │
│  │ 🧑 You:                      │  │
│  │ Create login tests           │  │
│  │                              │  │
│  │ 🤖 Copilot:                  │  │
│  │ I'll generate login tests    │  │
│  │ based on your recording...   │  │
│  │                              │  │
│  │ Generated 3 files:           │  │
│  │ ├─ LoginPage.java           │  │
│  │ ├─ LoginTest.java           │  │
│  │ └─ login.feature            │  │
│  │                              │  │
│  │ [👁️ Preview] [✅ Apply]      │  │
│  └──────────────────────────────┘  │
│                                    │
│  ┌──────────────────────────────┐  │
│  │ Type your message...         │  │
│  │ [Send]                       │  │
│  └──────────────────────────────┘  │
│                                    │
│  💡 Suggestions:                   │
│  • "Add negative test cases"      │
│  • "Generate data provider"        │
│  • "Create BDD scenarios"          │
└────────────────────────────────────┘
```

### **Chat Flow**
```typescript
// src/ui/ChatPanelProvider.ts
async handleChatMessage(message: string) {
  // 1. Add user message to UI
  this.addMessage({ role: 'user', content: message });

  // 2. Show "typing" indicator
  this.showTypingIndicator();

  // 3. Send to backend with context
  const response = await this.coreClient.chat({
    message,
    sessionId: this.sessionId,
    workspacePath: vscode.workspace.rootPath,
    context: {
      currentFile: vscode.window.activeTextEditor?.document.uri.fsPath,
      applicationMap: this.applicationMap,
      recentRecordings: this.getRecentRecordings()
    }
  });

  // 4. Hide typing indicator
  this.hideTypingIndicator();

  // 5. Stream response (for real-time feel)
  for await (const chunk of response.stream()) {
    this.appendToMessage(chunk);
  }

  // 6. If response has actions (code generation)
  if (response.actions) {
    this.showActionButtons(response.actions);
  }
}
```

### **Message Types**
```typescript
interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  actions?: Action[];
}

interface Action {
  type: 'create' | 'modify' | 'delete';
  filePath: string;
  before?: string; // For modify/delete
  after?: string;  // For create/modify
  diff?: string;   // Unified diff
}
```

---

## 🔄 State Management

### **Extension State**
```typescript
// src/state/ExtensionState.ts
class ExtensionState {
  // Workspace state
  isIndexed: boolean = false;
  indexingProgress: number = 0;
  fileCount: number = 0;

  // Recording state
  isRecording: boolean = false;
  currentSession: RecordingSession | null = null;

  // Application map
  applicationMap: ApplicationMap | null = null;
  recordings: RecordingSession[] = [];

  // Chat state
  chatHistory: ChatMessage[] = [];
  currentSessionId: string = '';

  // Backend connection
  isBackendConnected: boolean = false;
  backendPort: number = 8080;
}
```

### **State Persistence**
```typescript
// Save to workspace storage
context.workspaceState.update('appMap', applicationMap);
context.workspaceState.update('chatHistory', chatHistory);

// Save to global storage (across workspaces)
context.globalState.update('apiKey', openaiApiKey);
context.globalState.update('preferences', userPreferences);
```

### **State Sync**
```typescript
// Sync state with sidebar
stateManager.on('change', (key, value) => {
  sidebarProvider.updateState(key, value);
});

// Sync with webview panels
stateManager.on('change', (key, value) => {
  recordingPanel?.webview.postMessage({
    type: 'state-update',
    key,
    value
  });
});
```

---

## 🌐 Backend Communication

### **Connection Flow**
```
Extension starts
    ↓
Checks if Go binary is running (localhost:8080)
    ↓
If not running → Start Go binary
    ↓
Wait for health check (max 10s)
    ↓
Connected → Initialize workspace
```

### **API Client**
```typescript
// src/api/CoreClient.ts
class CoreClient {
  private baseURL = 'http://localhost:8080';
  private ws: WebSocket | null = null;

  // REST endpoints
  async analyzeWorkspace(path: string) {
    return this.post('/analyze', { workspacePath: path });
  }

  async chat(request: ChatRequest) {
    return this.post('/chat', request);
  }

  async generateFromRecording(recording: Recording) {
    return this.post('/generate', recording);
  }

  // WebSocket for real-time updates
  connectWebSocket() {
    this.ws = new WebSocket('ws://localhost:8080/ws');

    this.ws.on('message', (data) => {
      const event = JSON.parse(data);
      this.handleEvent(event);
    });
  }

  private handleEvent(event: Event) {
    switch (event.type) {
      case 'indexing-progress':
        this.emit('indexing-progress', event.data);
        break;
      case 'recording-update':
        this.emit('recording-update', event.data);
        break;
      case 'chat-response':
        this.emit('chat-response', event.data);
        break;
    }
  }
}
```

### **Error Handling**
```typescript
// Retry logic for network errors
async makeRequest(url: string, options: RequestOptions) {
  let retries = 3;
  while (retries > 0) {
    try {
      return await fetch(url, options);
    } catch (error) {
      if (error.code === 'ECONNREFUSED' && retries > 0) {
        retries--;
        await sleep(2000);
        continue;
      }
      throw error;
    }
  }
}
```

---

## 🏗️ Component Architecture

```
test-automation-copilot/
├── src/
│   ├── extension.ts                    # Entry point
│   │
│   ├── commands/                       # VS Code commands
│   │   ├── startRecording.ts
│   │   ├── stopRecording.ts
│   │   ├── openChat.ts
│   │   └── analyzeWorkspace.ts
│   │
│   ├── ui/                             # UI components
│   │   ├── SidebarProvider.ts          # Main sidebar
│   │   ├── ChatPanelProvider.ts        # Chat interface
│   │   ├── RecordingPanelProvider.ts   # Recording webview
│   │   └── DiffViewProvider.ts         # Code diff preview
│   │
│   ├── browser/                        # Browser automation
│   │   ├── BrowserRecorder.ts          # Playwright recorder
│   │   ├── ElementAnalyzer.ts          # Element analysis
│   │   └── LocatorGenerator.ts         # Locator strategies
│   │
│   ├── api/                            # Backend communication
│   │   └── CoreClient.ts               # HTTP/WebSocket client
│   │
│   ├── state/                          # State management
│   │   ├── ExtensionState.ts
│   │   └── StateManager.ts
│   │
│   └── utils/                          # Utilities
│       ├── logger.ts
│       └── helpers.ts
│
└── webview/                            # Webview UIs (React)
    ├── recording/
    │   ├── RecordingView.tsx
    │   └── BrowserEmbed.tsx
    │
    └── chat/
        ├── ChatView.tsx
        └── MessageList.tsx
```

---

## 📊 Data Flow

```
User Action (Recording)
    ↓
BrowserRecorder captures interaction
    ↓
ElementAnalyzer analyzes element
    ↓
CoreClient sends to Go backend (POST /recording/save)
    ↓
Go backend saves to SQLite
    ↓
WebSocket event sent back to extension
    ↓
SidebarProvider updates Application Map
    ↓
UI renders updated state
```

```
User Action (Chat)
    ↓
ChatPanelProvider captures message
    ↓
CoreClient sends to Go backend (POST /chat)
    ↓
Go backend:
    1. Semantic search (chromem-go + OpenAI embeddings)
    2. Context building (SQLite)
    3. LLM call (OpenAI GPT-4)
    4. Response parsing
    ↓
Streams response via WebSocket
    ↓
ChatPanelProvider displays streaming response
    ↓
If code generation → Shows diff preview
    ↓
User approves → Files created/modified
    ↓
Workspace re-indexed
```

---

## 🎯 Next Steps

1. **Design Review** ✅ (This document)
2. **Implement Sidebar UI** → Day 12
3. **Implement Recording Webview** → Day 13
4. **Integrate with Backend** → Day 14
5. **Add Chat Interface** → Day 15
6. **End-to-End Testing** → Day 16-17

---

**Total Lines:** 700+
**Complete User Journey:** ✅
**All Components Mapped:** ✅
**Ready for Implementation:** ✅
