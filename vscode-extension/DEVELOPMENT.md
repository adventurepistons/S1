# Development Guide - Test Automation Copilot VSCode Extension

This guide explains how to develop and test the extension locally.

## Architecture Overview

```
vscode-extension/
├── src/
│   ├── extension.ts        # Main extension entry point
│   └── chatPanel.ts        # Chat UI component
├── bin/
│   └── copilot-lsp         # Compiled Go LSP server
├── out/                    # Compiled TypeScript
├── package.json            # Extension manifest
└── tsconfig.json           # TypeScript config

../automation-copilot/
└── cmd/lsp-server/
    └── main.go             # Go LSP server wrapper
```

## Prerequisites

- **VSCode** 1.80+
- **Node.js** 18+
- **Go** 1.21+
- **TypeScript** knowledge
- **Anthropic API Key**

## Setup Development Environment

### 1. Clone & Install

```bash
cd vscode-extension
npm install
```

### 2. Build Go LSP Server

```bash
npm run build-server
```

This compiles `../automation-copilot/cmd/lsp-server/main.go` → `bin/copilot-lsp`

### 3. Compile TypeScript

```bash
npm run compile
```

Or watch mode:
```bash
npm run watch
```

## Running the Extension

### 1. Open in VSCode

```bash
code .
```

### 2. Debug Extension

1. Press **F5** or Run → "Start Debugging"
2. New "Extension Development Host" window opens
3. Extension is now active in that window

### 3. Test Features

In the Extension Development Host window:

```
# Set API key
Cmd+Shift+P → "Preferences: Open Settings"
Search: "Test Copilot: Api Key"
Enter your Anthropic API key

# Build index
Cmd+Shift+P → "Test Copilot: Build Index"

# Test generation
Cmd+Shift+P → "Test Copilot: Generate Test Code"
Enter: "Create test for login"

# Test chat
Cmd+Shift+P → "Test Copilot: Open Chat"
Chat: "create test for login"
```

### 4. View Logs

**LSP Server logs:**
```bash
tail -f copilot-lsp.log
```

**VSCode Extension logs:**
- Help → Toggle Developer Tools → Console

## Code Structure

### extension.ts

Main extension file:

```typescript
export function activate(context: vscode.ExtensionContext) {
    // 1. Start LSP server
    startLanguageServer(context);

    // 2. Initialize copilot
    initializeCopilot();

    // 3. Register commands
    registerCommands(context);
}
```

**Key functions:**
- `startLanguageServer()` - Spawns Go LSP server
- `initializeCopilot()` - Sends init request to server
- `generateCode()` - Handles code generation
- `registerCommands()` - Registers all commands

### chatPanel.ts

Chat UI component:

```typescript
export class ChatPanel {
    // Shows webview panel
    public show() { }

    // Handles user messages
    private async handleUserMessage(text: string) { }

    // Updates webview
    private updateWebview() { }

    // Returns HTML for webview
    private getWebviewContent(): string { }
}
```

### cmd/lsp-server/main.go

Go LSP server:

```go
type Server struct {
    contextBuilder *ai.ContextBuilder
    db             *storage.Database
}

func (s *Server) handleMessage(msg JSONRPCMessage) *JSONRPCMessage {
    switch msg.Method {
    case "testCopilot/generate":
        return s.handleGenerate(msg)
    case "testCopilot/chat":
        return s.handleChat(msg)
    case "testCopilot/buildIndex":
        return s.handleBuildIndex(msg)
    }
}
```

## JSON-RPC Protocol

The extension communicates with the Go server via JSON-RPC over stdio.

### Initialize

**Request:**
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "testCopilot/initialize",
    "params": {
        "workspaceRoot": "/path/to/workspace",
        "apiKey": "sk-ant-..."
    }
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "success": true,
        "message": "Copilot initialized"
    }
}
```

### Generate Code

**Request:**
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "testCopilot/generate",
    "params": {
        "request": "Create test for login"
    }
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "result": {
        "code": "public class LoginTest { ... }",
        "className": "LoginTest",
        "packageName": "com.example.tests",
        "codeType": "test",
        "suggestedPath": "com/example/tests/LoginTest.java",
        "cost": 0.012,
        "tokens": {
            "input": 6100,
            "output": 452,
            "cached": 3500
        },
        "latency": 3200
    }
}
```

## Making Changes

### Adding a New Command

**1. Update package.json:**

```json
{
    "contributes": {
        "commands": [
            {
                "command": "testCopilot.myNewCommand",
                "title": "My New Command",
                "category": "Test Copilot"
            }
        ]
    }
}
```

**2. Register in extension.ts:**

```typescript
context.subscriptions.push(
    vscode.commands.registerCommand('testCopilot.myNewCommand', async () => {
        // Implementation
    })
);
```

### Adding New LSP Method

**1. Add handler in main.go:**

```go
case "testCopilot/myMethod":
    return s.handleMyMethod(msg)
```

**2. Implement handler:**

```go
func (s *Server) handleMyMethod(msg JSONRPCMessage) *JSONRPCMessage {
    // Parse params
    var params struct {
        Foo string `json:"foo"`
    }
    json.Unmarshal(msg.Params, &params)

    // Do something
    result := doSomething(params.Foo)

    // Return response
    return &JSONRPCMessage{
        JSONRPC: "2.0",
        ID:      msg.ID,
        Result:  result,
    }
}
```

**3. Call from TypeScript:**

```typescript
const result = await client.sendRequest('testCopilot/myMethod', {
    foo: 'bar'
});
```

## Testing

### Manual Testing

1. Make code changes
2. Press F5 (starts Extension Development Host)
3. Test in new window
4. Check logs for errors

### Debugging Go Server

```bash
# Add logging in Go code
log.Printf("Debug: %v", someValue)

# Watch logs
tail -f copilot-lsp.log
```

### Debugging TypeScript

1. Open Developer Tools in Extension Development Host
2. Set breakpoints in TypeScript code
3. Debugger will pause

## Building for Release

### 1. Update Version

Update `package.json`:
```json
{
    "version": "1.1.0"
}
```

### 2. Build

```bash
chmod +x build.sh
./build.sh
```

This creates `test-automation-copilot-1.1.0.vsix`

### 3. Test Installation

```
# Install locally
code --install-extension test-automation-copilot-1.1.0.vsix

# Test
code /path/to/test/project
```

### 4. Publish to Marketplace (Optional)

```bash
# Login
vsce login adventurepistons

# Publish
vsce publish
```

## Common Issues

### "LSP server not starting"

- Check `bin/copilot-lsp` exists and is executable
- Check logs: `copilot-lsp.log`
- Rebuild: `npm run build-server`

### "Copilot not initialized"

- Check API key in settings
- Check workspace has `.copilot/` folder
- Check LSP server logs

### "TypeScript errors"

- Run `npm install`
- Run `npm run compile`
- Restart VSCode

## File Watching

For active development:

**Terminal 1** (watch TypeScript):
```bash
npm run watch
```

**Terminal 2** (watch Go):
```bash
cd ../automation-copilot
while true; do
    go build -o ../vscode-extension/bin/copilot-lsp cmd/lsp-server/main.go
    sleep 2
done
```

**VSCode**: Press F5 to reload extension

## Performance Tips

- LSP server starts on extension activation (fast)
- Database loaded on first `testCopilot/initialize` (1-2s)
- Index building is async (doesn't block)
- Prompt caching kicks in after first generation (3-5x speedup)

## Next Steps

- Add more commands (refactor, explain, fix)
- Inline completions (like Copilot)
- Test execution integration
- Multi-file generation
- OpenAI support

Happy hacking! 🚀
