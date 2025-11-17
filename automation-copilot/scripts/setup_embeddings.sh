#!/bin/bash

# Setup Local Embedding Service
# ==============================
# Installs dependencies and downloads the embedding model

set -e  # Exit on error

echo "============================================================"
echo "Setting up Local Embedding Service (FREE)"
echo "============================================================"
echo ""

# Check if Python 3 is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed. Please install Python 3.8 or higher."
    exit 1
fi

PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
echo "✅ Python version: $PYTHON_VERSION"
echo ""

# Check if pip is installed
if ! command -v pip3 &> /dev/null; then
    echo "❌ pip3 is not installed. Please install pip3."
    exit 1
fi

echo "📦 Installing Python dependencies..."
echo ""

# Install dependencies
pip3 install -r scripts/requirements.txt

echo ""
echo "⬇️  Downloading embedding model (one-time, ~90MB)..."
echo ""

# Download model by running a simple Python script
python3 << 'EOF'
from sentence_transformers import SentenceTransformer
import sys

try:
    print("Loading model: sentence-transformers/all-MiniLM-L6-v2")
    model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')
    print("✅ Model downloaded successfully!")

    # Test encoding
    test_embedding = model.encode("test", convert_to_numpy=True)
    print(f"✅ Model working! Embedding dimension: {len(test_embedding)}")

except Exception as e:
    print(f"❌ Error downloading model: {e}")
    sys.exit(1)
EOF

echo ""
echo "============================================================"
echo "✅ Setup Complete!"
echo "============================================================"
echo ""
echo "To start the embedding service, run:"
echo "  python3 scripts/embedding_service.py"
echo ""
echo "Or use the Makefile:"
echo "  make start-embeddings"
echo ""
