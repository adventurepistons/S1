# Test Automation Copilot - VSCode Extension

**AI-powered code generation for Selenium/Java test automation**

Like Cursor, but specialized for test automation! 🚀

## ✨ Features

- 🤖 **AI Code Generation** - Generate tests, page objects, and utilities
- 💬 **Interactive Chat** - Multi-turn conversations for iterative refinement
- 🧠 **Framework-Aware** - Understands TestNG, JUnit, Cucumber, Serenity
- 📊 **Pattern Learning** - Learns from YOUR codebase patterns
- 💰 **Cost Tracking** - Real-time cost display in status bar
- ⚡ **Fast & Local** - FREE local embeddings, no embedding API costs
- 🎯 **Context-Aware** - Right-click → "Generate test for this class"

## 📦 Installation

### Prerequisites

1. **VSCode** 1.80 or higher
2. **Go** 1.21 or higher (for building LSP server)
3. **Anthropic API Key** ([Get one here](https://console.anthropic.com/))

### Install from VSIX

1. Download the latest `.vsix` file from releases
2. Open VSCode
3. Press `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux)
4. Type "Install from VSIX"
5. Select the downloaded `.vsix` file

### Build from Source

```bash
# Clone repository
cd vscode-extension

# Install dependencies
npm install

# Build Go LSP server
npm run build-server

# Compile TypeScript
npm run compile

# Package extension
npm run package
```

This creates `test-automation-copilot-1.0.0.vsix`

## ⚙️ Configuration

### 1. Set API Key

**Option A: VSCode Settings**
```
Cmd+Shift+P → "Preferences: Open Settings"
Search for "Test Copilot"
Set "Test Copilot: Api Key"
```

**Option B: Environment Variable**
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. Build Index (First Time)

```
Cmd+Shift+P → "Test Copilot: Build Index"
```

This analyzes your Java codebase and builds:
- Code chunks (cAST-aware)
- BM25 keyword index
- Semantic embeddings (FREE local)
- Pattern examples

Takes ~2-5 minutes for a medium project.

## 🚀 Usage

### Method 1: Command Palette

```
Cmd+Shift+P → "Test Copilot: Generate Test Code"
Enter: "Create test for login with valid credentials"
```

### Method 2: Right-Click Menu

1. Open any Java file
2. Right-click in editor
3. Select **"Generate Test for This Class"**
4. Enter details

### Method 3: Chat Mode (Recommended!)

```
Cmd+Shift+P → "Test Copilot: Open Chat"
```

Interactive chat panel appears:
```
You: create test for login
🤖: [Generates LoginTest.java]
    💰 Cost: $0.012 | 📊 6100 tokens
    [Save to File] [Insert at Cursor]

You: add test for invalid password
🤖: [Updates LoginTest with both tests]
    💰 Cost: $0.008 | 📊 3500 tokens (cached!)
```

## 💡 Example Requests

### Generate Tests
```
"Create test for login with valid credentials"
"Add negative test for invalid password"
"Generate data-driven test using @DataProvider"
```

### Generate Page Objects
```
"Create page object for login page"
"Add page object for dashboard with username and logout button"
```

### Refactor & Improve
```
"Convert this test to use explicit waits"
"Add assertions to verify page title"
"Refactor to use Page Factory pattern"
```

## 📊 Status Bar

The status bar shows:
```
$(robot) Test Copilot | $0.45
```

- **Robot icon** - Extension active
- **Cost** - Total session cost
- **Click** - Show detailed statistics

## ⌨️ Keyboard Shortcuts

| Command | Shortcut |
|---------|----------|
| Generate Code | `Cmd+Shift+G` (Mac) / `Ctrl+Shift+G` (Win) |
| Open Chat | `Cmd+Shift+C` (Mac) / `Ctrl+Shift+C` (Win) |

*(Configure in Keyboard Shortcuts)*

## 🎯 How It Works

```
1. User Request
   ↓
2. Context Builder (Go)
   - Analyzes intent
   - Retrieves relevant code (Hybrid BM25 + Semantic)
   - Selects examples from YOUR codebase
   - Loads conversation history
   ↓
3. Claude API
   - Sends optimized prompt (90% cached!)
   - Receives generated code
   ↓
4. Response Parser (Go)
   - Extracts Java code
   - Validates (checks anti-patterns)
   - Suggests file path
   ↓
5. VSCode Extension (TypeScript)
   - Displays code
   - Offers save/insert
   - Updates cost tracker
```

## 💰 Cost

**Extremely cost-efficient:**

- Context building: **$0** (FREE local embeddings)
- First generation: **~$0.018**
- Subsequent generations: **~$0.008** (prompt caching!)
- Chat mode (turn 5+): **~$0.006** (everything cached)

**Typical session (10 generations):** ~$0.10

## 🏗️ Architecture

```
┌─────────────────────────┐
│  VSCode Extension (TS)  │
│  - Chat Panel           │
│  - Commands             │
│  - Status Bar           │
└──────────┬──────────────┘
           │ JSON-RPC
           ↓
┌─────────────────────────┐
│  Go LSP Server          │
│  - Context Builder      │
│  - Claude API           │
│  - Parser               │
│  - Knowledge Graph      │
│  - Embeddings (local)   │
└─────────────────────────┘
```

## 🤝 vs. Cursor/Copilot

| Feature | **Test Copilot** | Cursor/Copilot |
|---------|------------------|----------------|
| Test Domain Knowledge | ✅ Expert | ⚠️ Generic |
| Framework Detection | ✅ TestNG, JUnit, etc. | ❌ No |
| Pattern Learning | ✅ From YOUR code | ⚠️ Generic |
| Embedding Cost | ✅ FREE (local) | 💰 Paid |
| Test Validation | ✅ Anti-patterns | ❌ No |
| Knowledge Graph | ✅ Yes | ❌ No |

## 🐛 Troubleshooting

### "Copilot not initialized"

1. Check API key is set in settings
2. Restart VSCode
3. Check LSP server logs: `copilot-lsp.log`

### "Build index failed"

1. Ensure workspace has Java files
2. Check disk space
3. Try: `Cmd+Shift+P → "Test Copilot: Build Index"`

### "Generation is slow"

1. First generation is slow (no cache)
2. Subsequent generations are 3-5x faster (caching)
3. Check internet connection

## 📝 Requirements

- VSCode 1.80+
- Java 11+
- Anthropic API key
- Internet connection (for Claude API)

## 🔒 Privacy

- All code analysis happens **locally**
- Only generated prompts sent to Claude API
- No code stored on external servers
- Database stored in `.copilot/` folder

## 📄 License

MIT License - See LICENSE file

## 🙋 Support

- [GitHub Issues](https://github.com/adventurepistons/test-automation-copilot/issues)
- [Documentation](https://github.com/adventurepistons/test-automation-copilot/wiki)

---

**Made with ❤️ for test automation engineers**
