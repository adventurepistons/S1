#!/bin/bash

set -e

echo "🔨 Building Test Automation Copilot VSCode Extension"
echo ""

# Step 1: Build Go LSP server
echo "Step 1/4: Building Go LSP server..."
cd ../automation-copilot
go build -o ../vscode-extension/bin/copilot-lsp cmd/lsp-server/main.go
echo "✅ Go LSP server built"
echo ""

# Step 2: Install npm dependencies
echo "Step 2/4: Installing dependencies..."
cd ../vscode-extension
npm install
echo "✅ Dependencies installed"
echo ""

# Step 3: Compile TypeScript
echo "Step 3/4: Compiling TypeScript..."
npm run compile
echo "✅ TypeScript compiled"
echo ""

# Step 4: Package extension
echo "Step 4/4: Packaging extension..."
npm run package
echo "✅ Extension packaged"
echo ""

echo "════════════════════════════════════════════════════"
echo "✅ Build complete!"
echo "════════════════════════════════════════════════════"
echo ""
echo "📦 Extension package: test-automation-copilot-1.0.0.vsix"
echo ""
echo "To install:"
echo "  1. Open VSCode"
echo "  2. Cmd+Shift+P → 'Install from VSIX'"
echo "  3. Select test-automation-copilot-1.0.0.vsix"
echo ""
