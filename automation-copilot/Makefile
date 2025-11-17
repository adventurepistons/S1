.PHONY: help build test clean setup-embeddings start-embeddings stop-embeddings

help:
	@echo "Automation Copilot - Makefile Commands"
	@echo "======================================="
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              - Build the Go binary"
	@echo "  make test               - Run all tests"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Embedding Service (FREE):"
	@echo "  make setup-embeddings   - Install dependencies & download model"
	@echo "  make start-embeddings   - Start local embedding service"
	@echo "  make stop-embeddings    - Stop local embedding service"
	@echo "  make test-embeddings    - Test embedding service"
	@echo ""
	@echo "Development:"
	@echo "  make fmt                - Format Go code"
	@echo "  make lint               - Run linter"
	@echo ""

build:
	@echo "Building automation-copilot..."
	@cd cmd/copilot && go build -o ../../copilot
	@echo "✅ Build complete: ./copilot"

test:
	@echo "Running tests..."
	@go test ./... -v

clean:
	@echo "Cleaning build artifacts..."
	@rm -f copilot
	@rm -f copilot.db
	@echo "✅ Clean complete"

setup-embeddings:
	@echo "Setting up local embedding service..."
	@bash scripts/setup_embeddings.sh

start-embeddings:
	@echo "Starting local embedding service..."
	@python3 scripts/embedding_service.py &
	@sleep 2
	@curl -s http://localhost:5000/health | grep -q "healthy" && echo "✅ Embedding service ready at http://localhost:5000" || echo "❌ Failed to start embedding service"

stop-embeddings:
	@echo "Stopping local embedding service..."
	@pkill -f embedding_service.py || echo "Service not running"
	@echo "✅ Embedding service stopped"

test-embeddings:
	@echo "Testing embedding service..."
	@curl -s http://localhost:5000/health || (echo "❌ Service not running. Start with: make start-embeddings" && exit 1)
	@echo ""
	@echo "Testing /embed endpoint..."
	@curl -s -X POST http://localhost:5000/embed \
		-H "Content-Type: application/json" \
		-d '{"text": "public void login(String username, String password)"}' \
		| python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"✅ Embedding generated! Dimension: {data['dimension']}, Time: {data['processing_time_ms']}ms\")"
	@echo ""
	@echo "Testing /embed_batch endpoint..."
	@curl -s -X POST http://localhost:5000/embed_batch \
		-H "Content-Type: application/json" \
		-d '{"texts": ["test method", "login function", "page object"]}' \
		| python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"✅ Batch embedding generated! Count: {data['count']}, Avg time: {data['avg_time_per_text_ms']}ms\")"

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "✅ Formatting complete"

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
