#!/bin/bash
set -e

echo "🚀 Test Automation Copilot - Quick Setup"
echo "=========================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cd "$(dirname "$0")"

echo -e "\n${YELLOW}Step 1: Installing TypeScript dependencies${NC}"
npm install

echo -e "\n${YELLOW}Step 2: Compiling TypeScript${NC}"
npm run compile

echo -e "\n${YELLOW}Step 3: Building Go backend${NC}"
cd copilot-core
go build -o bin/server cmd/server/main.go
cd ..

echo -e "\n${YELLOW}Step 4: Setting up Cloud Backend (Optional)${NC}"
cd cloud-backend

if [ ! -f .env ]; then
    echo "Creating cloud-backend/.env from example..."
    cp .env.example .env
    echo -e "${YELLOW}⚠️  IMPORTANT: Edit cloud-backend/.env and add your OPENAI_API_KEY${NC}"
else
    echo ".env already exists"
fi

npm install

cd ..

echo -e "\n${GREEN}✅ Setup Complete!${NC}"
echo ""
echo "To test WITHOUT AI (infrastructure only):"
echo "  1. Open in VS Code: code ."
echo "  2. Press F5 to launch extension"
echo "  3. Run: Test Copilot: Analyze Framework"
echo "  4. Run: Test Copilot: Test Semantic Search"
echo ""
echo "To enable AI features:"
echo "  1. Edit cloud-backend/.env"
echo "  2. Add: OPENAI_API_KEY=sk-your-key"
echo "  3. Run: cd cloud-backend && npm run dev"
echo "  4. In another terminal: cd copilot-core && ./bin/server"
echo "  5. Press F5 in VS Code"
echo ""
echo "Documentation:"
echo "  - LOCAL_TESTING_GUIDE.md - Complete testing guide"
echo "  - QUICK_START.md - Feature overview"
echo "  - CLOUD_ARCHITECTURE.md - System design"
echo ""
echo "Have fun! 🎉"
