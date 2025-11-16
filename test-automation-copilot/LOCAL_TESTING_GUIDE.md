# 🧪 Local Testing Guide - Test Automation Copilot

## 📋 Prerequisites

```bash
✅ VS Code installed
✅ Node.js 18+ installed
✅ Go 1.23+ installed
✅ OpenAI API key (optional - for AI chat)
```

---

## 🚀 Quick Start (5 Minutes)

### Step 1: Build Everything

```bash
cd test-automation-copilot

# Install extension dependencies
npm install

# Compile TypeScript
npm run compile

# Build Go backend
cd copilot-core
go build -o bin/server cmd/server/main.go
cd ..
```

### Step 2: Open in VS Code

```bash
# Open the extension project in VS Code
code .
```

### Step 3: Launch Extension

1. Press `F5` (or Run → Start Debugging)
2. A new VS Code window opens: "Extension Development Host"
3. Extension is now running!

---

## 🎯 Complete User Journey

### Journey 1: Index a Workspace (WORKS WITHOUT CLOUD BACKEND)

**What happens:**
1. Open a Java/Selenium project in the Extension Development Host window
2. Open Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`)
3. Type: **"Test Copilot: Analyze Framework"**
4. Watch real-time progress:
   ```
   Starting indexing... (0%)
   Indexing files... (45%)
   Indexing complete! (100%)

   ✅ Workspace indexed!
   📁 150 files
   📦 45 classes
   📄 12 page objects
   ✅ 30 test methods
   🔍 500 code chunks indexed
   ```

**What's happening behind the scenes:**
- Go backend starts on port 8080
- Parses all `.java`, `.feature`, `pom.xml` files
- Stores in SQLite database
- Creates vector embeddings
- Shows real-time progress via WebSocket

---

### Journey 2: Semantic Search (WORKS WITHOUT CLOUD BACKEND)

**What happens:**
1. After indexing, open Command Palette
2. Type: **"Test Copilot: Test Semantic Search"**
3. Enter query: `"login page object"`
4. See results in Output Channel:
   ```
   🔍 Semantic Search Results
   ============================================================
   Query: "login page object"

   📄 Code Results:
   ------------------------------------------------------------
     1. LoginPage (class)
        Similarity: 95.2%
        File: src/main/java/pages/LoginPage.java:15

     2. clickLoginButton() (method)
        Similarity: 87.3%
        File: src/main/java/pages/LoginPage.java:42

   📦 Page Object Results:
   ------------------------------------------------------------
     1. LoginPage
        Similarity: 95.2%
        File: src/main/java/pages/LoginPage.java:15

   ✅ Test Results:
   ------------------------------------------------------------
     1. testLogin()
        Similarity: 82.1%
        File: src/test/java/tests/LoginTest.java:23
   ```

**What's happening:**
- Uses OpenAI embeddings for semantic matching
- Searches vector database
- Ranks by similarity
- NO LLM/GPT needed - just embeddings!

---

### Journey 3: Chat with AI (REQUIRES CLOUD BACKEND)

**Status:** ⚠️ Partially implemented

**What should happen:**
1. Click Test Copilot icon in Activity Bar
2. Chat panel opens
3. Type: `"How do I create a login test?"`
4. AI streams response token-by-token with:
   - Code blocks highlighted
   - Markdown formatting
   - Blinking cursor during streaming

**Current state:**
- ✅ Chat UI works
- ✅ Streaming UI works
- ✅ Syntax highlighting works
- ❌ Cloud backend needs OpenAI key
- ❌ Prompts need configuration

**To make it work:**
```bash
cd cloud-backend

# 1. Create .env file
cp .env.example .env

# 2. Edit .env and add your OpenAI API key
OPENAI_API_KEY=sk-your-key-here

# 3. Install and run
npm install
npm run dev

# Cloud backend now running on port 3000
```

---

## 🛠️ What Works NOW (No Cloud Backend)

### ✅ Ready to Test Immediately

1. **Workspace Indexing** ✅
   - Parses Java, Gherkin, XML
   - SQLite database
   - Vector embeddings
   - Real-time progress

2. **Semantic Search** ✅
   - Search all code
   - Search page objects
   - Search tests
   - Similarity scoring

3. **Chat UI** ✅
   - Beautiful interface
   - Markdown rendering
   - Syntax highlighting
   - Streaming cursor animation

4. **Element Recording** ✅
   - Browser automation
   - Element picking
   - Locator generation

---

## ⚠️ What Needs Cloud Backend

### ❌ Not Working Without Cloud Setup

1. **AI Chat Responses**
   - Needs OpenAI API key
   - Needs cloud backend running
   - Needs prompt configuration

2. **Code Generation**
   - Generate Page Objects from prompts
   - Generate Tests from descriptions
   - Fix broken code

3. **Smart Context**
   - AI-powered context assembly
   - Few-shot examples
   - Intelligent suggestions

---

## 📊 Testing Scenarios

### Scenario 1: Test Indexing (5 min)

```bash
# 1. Create a simple test project
mkdir -p test-project/src/main/java/pages
mkdir -p test-project/src/test/java/tests

# 2. Create LoginPage.java
cat > test-project/src/main/java/pages/LoginPage.java << 'EOF'
package pages;

import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;

public class LoginPage {
    private WebDriver driver;
    private By usernameField = By.id("username");
    private By passwordField = By.id("password");
    private By loginButton = By.id("login-btn");

    public LoginPage(WebDriver driver) {
        this.driver = driver;
    }

    public void login(String username, String password) {
        driver.findElement(usernameField).sendKeys(username);
        driver.findElement(passwordField).sendKeys(password);
        driver.findElement(loginButton).click();
    }
}
EOF

# 3. Open test-project in Extension Development Host
# 4. Run: Test Copilot: Analyze Framework
# 5. See: "1 file, 1 class, 1 page object"
```

### Scenario 2: Test Search (2 min)

```bash
# After indexing above project:
# 1. Run: Test Copilot: Test Semantic Search
# 2. Query: "login"
# 3. Should find: LoginPage class, login() method
```

### Scenario 3: Test Chat UI (1 min)

```bash
# 1. Run: Test Copilot: Open Chat
# 2. See beautiful chat interface
# 3. Type a message (will fail gracefully without cloud backend)
# 4. See error: "Copilot Core is not running" or "Cloud API not configured"
```

---

## 🐛 Troubleshooting

### Issue: "Go binary not found"

```bash
cd copilot-core
go build -o bin/server cmd/server/main.go

# Verify binary exists
ls -lh bin/server
# Should see: bin/server (34MB)
```

### Issue: "TypeScript errors"

```bash
# Reinstall dependencies
rm -rf node_modules dist
npm install
npm run compile
```

### Issue: "Extension not activating"

```bash
# Check VS Code developer tools:
# Help → Toggle Developer Tools
# Look for errors in Console tab
```

### Issue: "Can't connect to Go backend"

```bash
# Check if server is running:
lsof -i :8080

# If not running, check logs in VS Code Debug Console
```

---

## 📁 File Structure for Testing

```
test-automation-copilot/
├── .vscode/
│   ├── launch.json          ← F5 to debug
│   └── tasks.json           ← Build tasks
├── copilot-core/
│   └── bin/
│       └── server           ← Go binary (run this)
├── dist/
│   └── extension.js         ← Compiled TypeScript
├── package.json             ← Extension manifest
└── LOCAL_TESTING_GUIDE.md   ← This file
```

---

## ✅ Success Checklist

After setup, you should be able to:

- [ ] Press F5 and see "Extension Development Host" window
- [ ] See "Test Copilot" icon in Activity Bar
- [ ] Run "Analyze Framework" and see progress
- [ ] Run "Test Semantic Search" and see results
- [ ] Open Chat and see UI (even if chat doesn't respond)
- [ ] See Go backend logs in Debug Console

---

## 🎯 Next Steps After Testing

### To get full AI features working:

1. **Set up Cloud Backend**
   ```bash
   cd cloud-backend
   cp .env.example .env
   # Add OPENAI_API_KEY=sk-...
   npm install
   npm run dev
   ```

2. **Configure Extension Settings**
   - Open VS Code Settings
   - Search "Test Copilot"
   - Set API key and cloud URL

3. **Test Full Flow**
   - Index workspace ✅
   - Search code ✅
   - Chat with AI ✅ (now works!)
   - Generate code ✅ (now works!)

---

## 🚀 Quick Commands Reference

```bash
# Extension Development
F5                                  # Launch extension
Cmd/Ctrl + Shift + F5              # Reload extension
Cmd/Ctrl + Shift + P               # Command Palette

# Test Copilot Commands
Test Copilot: Analyze Framework    # Index workspace
Test Copilot: Test Semantic Search # Search code
Test Copilot: Open Chat           # Open chat UI

# Building
npm run compile                    # Build TypeScript
npm run watch                      # Auto-rebuild on changes
cd copilot-core && go build ...   # Build Go backend
```

---

## 💡 Tips

1. **Use "watch" task** - Automatically rebuilds on file changes:
   ```bash
   npm run watch
   ```

2. **Check logs** - All Go backend logs appear in VS Code Debug Console

3. **Test incrementally** - Start with indexing, then search, then chat

4. **Use real project** - Test with your actual Selenium projects for best results

5. **Hot reload** - After code changes, press `Cmd/Ctrl+Shift+F5` to reload

---

## 📞 Need Help?

Check logs in:
- VS Code Debug Console (Go backend logs)
- VS Code Developer Tools (Extension logs)
- Terminal output (Build errors)

---

**Built with ❤️ for QA Engineers**

Ready to test! Press F5 and start exploring! 🚀
