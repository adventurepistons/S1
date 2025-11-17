# Local Embedding Service

**100% FREE alternative to OpenAI/Anthropic embeddings**

## Overview

This service runs a local embedding model using `sentence-transformers`. No API costs, runs on your machine, fast and efficient.

## Model Details

- **Model**: `sentence-transformers/all-MiniLM-L6-v2`
- **Dimension**: 384
- **Speed**: ~1000 tokens/sec
- **Size**: 90MB (one-time download)
- **Quality**: Excellent for code/text similarity

## Setup

```bash
# Install dependencies and download model
make setup-embeddings

# Or manually:
pip3 install -r scripts/requirements.txt
python3 scripts/setup_embeddings.sh
```

## Usage

### Start Service

```bash
make start-embeddings

# Or manually:
python3 scripts/embedding_service.py
```

Service runs at: `http://localhost:5000`

### Test Service

```bash
make test-embeddings
```

### Stop Service

```bash
make stop-embeddings
```

## API Endpoints

### 1. Health Check

```bash
GET /health

curl http://localhost:5000/health
```

Response:
```json
{
  "status": "healthy",
  "model": "sentence-transformers/all-MiniLM-L6-v2",
  "dimension": 384,
  "model_loaded": true
}
```

### 2. Embed Single Text

```bash
POST /embed

curl -X POST http://localhost:5000/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "public void login(String username, String password)"}'
```

Response:
```json
{
  "embedding": [0.123, -0.456, 0.789, ...],
  "dimension": 384,
  "processing_time_ms": 45
}
```

### 3. Embed Batch (FASTER!)

```bash
POST /embed_batch

curl -X POST http://localhost:5000/embed_batch \
  -H "Content-Type: application/json" \
  -d '{"texts": ["test method", "login function", "page object"]}'
```

Response:
```json
{
  "embeddings": [[...], [...], [...]],
  "count": 3,
  "dimension": 384,
  "processing_time_ms": 120,
  "avg_time_per_text_ms": 40
}
```

### 4. Calculate Similarity

```bash
POST /similarity

curl -X POST http://localhost:5000/similarity \
  -H "Content-Type: application/json" \
  -d '{"text1": "login test", "text2": "authentication test"}'
```

Response:
```json
{
  "similarity": 0.85,
  "processing_time_ms": 50
}
```

## Performance

| Operation | Time | Cost |
|-----------|------|------|
| Single embedding | ~45ms | $0 |
| Batch (10 texts) | ~120ms | $0 |
| Batch (100 texts) | ~800ms | $0 |

**vs OpenAI text-embedding-3-small:**
- Cost: $0.02 per 1M tokens
- Latency: 100-300ms (network + processing)

## Integration with Go

See `internal/ai/embedding_client.go` for the Go client that calls this service.

Example:
```go
client := NewEmbeddingClient("http://localhost:5000")
embedding, err := client.Embed("public void login(...)")
// embedding is []float64 with 384 dimensions
```

## Troubleshooting

### Port already in use

```bash
# Find and kill process on port 5000
lsof -ti:5000 | xargs kill -9

# Or use a different port
python3 scripts/embedding_service.py --port 5001
```

### Model download fails

```bash
# Manual download
python3 -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')"
```

### Out of memory

The model uses ~500MB RAM. If you have memory constraints, you can:
1. Use a smaller model: `all-MiniLM-L6-v2` → `paraphrase-MiniLM-L3-v2`
2. Reduce batch size in requests

## Why Local Embeddings?

✅ **Zero Cost** - No API fees
✅ **Privacy** - Data never leaves your machine
✅ **Speed** - No network latency
✅ **Reliability** - No rate limits or outages
✅ **Offline** - Works without internet

## Alternative: Cloud Embeddings

If you prefer cloud embeddings for higher quality:

1. **OpenAI** (`text-embedding-3-small`)
   - Cost: $0.02 per 1M tokens
   - Dimension: 1536
   - Better quality

2. **Voyage AI** (`voyage-code-2`)
   - Cost: $0.12 per 1M tokens
   - Dimension: 1024
   - Optimized for code

See `CONTEXT_BUILDER_PLAN.md` for cloud embedding implementation.
