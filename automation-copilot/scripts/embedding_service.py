#!/usr/bin/env python3
"""
Local Embedding Service
=======================
FREE embedding service using sentence-transformers.
No API costs, runs locally, fast and efficient.

Model: sentence-transformers/all-MiniLM-L6-v2
- Dimension: 384
- Speed: ~1000 tokens/sec
- Size: 90MB
- Quality: Good for code/text similarity

Usage:
    python3 embedding_service.py

API Endpoints:
    POST /embed          - Embed single text
    POST /embed_batch    - Embed multiple texts (faster)
    GET  /health         - Health check
    GET  /model_info     - Model information
"""

from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer
import numpy as np
import time
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Global model instance (loaded once at startup)
model = None
MODEL_NAME = 'sentence-transformers/all-MiniLM-L6-v2'
MODEL_DIM = 384

def load_model():
    """Load the embedding model once at startup."""
    global model
    logger.info(f"Loading embedding model: {MODEL_NAME}")
    start_time = time.time()

    model = SentenceTransformer(MODEL_NAME)

    load_time = time.time() - start_time
    logger.info(f"Model loaded successfully in {load_time:.2f}s")
    logger.info(f"Model dimension: {MODEL_DIM}")

    return model

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint."""
    return jsonify({
        'status': 'healthy',
        'model': MODEL_NAME,
        'dimension': MODEL_DIM,
        'model_loaded': model is not None
    })

@app.route('/model_info', methods=['GET'])
def model_info():
    """Get model information."""
    if model is None:
        return jsonify({'error': 'Model not loaded'}), 500

    return jsonify({
        'model_name': MODEL_NAME,
        'dimension': MODEL_DIM,
        'max_seq_length': model.max_seq_length,
        'tokenizer': str(type(model.tokenizer).__name__)
    })

@app.route('/embed', methods=['POST'])
def embed():
    """
    Embed a single text.

    Request:
        {
            "text": "your text here"
        }

    Response:
        {
            "embedding": [0.123, -0.456, ...],
            "dimension": 384,
            "processing_time_ms": 45
        }
    """
    if model is None:
        return jsonify({'error': 'Model not loaded'}), 500

    data = request.json
    if not data or 'text' not in data:
        return jsonify({'error': 'Missing "text" field in request'}), 400

    text = data['text']
    if not text or len(text.strip()) == 0:
        return jsonify({'error': 'Empty text provided'}), 400

    start_time = time.time()

    try:
        # Generate embedding
        embedding = model.encode(text, convert_to_numpy=True)

        processing_time = (time.time() - start_time) * 1000  # Convert to ms

        return jsonify({
            'embedding': embedding.tolist(),
            'dimension': len(embedding),
            'processing_time_ms': round(processing_time, 2)
        })

    except Exception as e:
        logger.error(f"Error embedding text: {str(e)}")
        return jsonify({'error': str(e)}), 500

@app.route('/embed_batch', methods=['POST'])
def embed_batch():
    """
    Embed multiple texts in batch (much faster than individual calls).

    Request:
        {
            "texts": ["text 1", "text 2", ...]
        }

    Response:
        {
            "embeddings": [[0.1, 0.2, ...], [0.3, 0.4, ...]],
            "count": 2,
            "dimension": 384,
            "processing_time_ms": 120,
            "avg_time_per_text_ms": 60
        }
    """
    if model is None:
        return jsonify({'error': 'Model not loaded'}), 500

    data = request.json
    if not data or 'texts' not in data:
        return jsonify({'error': 'Missing "texts" field in request'}), 400

    texts = data['texts']
    if not isinstance(texts, list) or len(texts) == 0:
        return jsonify({'error': 'texts must be a non-empty list'}), 400

    # Filter out empty texts
    texts = [t for t in texts if t and len(t.strip()) > 0]
    if len(texts) == 0:
        return jsonify({'error': 'No valid texts provided'}), 400

    start_time = time.time()

    try:
        # Batch encoding is much faster!
        embeddings = model.encode(texts, convert_to_numpy=True, show_progress_bar=False)

        processing_time = (time.time() - start_time) * 1000  # Convert to ms
        avg_time_per_text = processing_time / len(texts)

        return jsonify({
            'embeddings': embeddings.tolist(),
            'count': len(embeddings),
            'dimension': embeddings.shape[1],
            'processing_time_ms': round(processing_time, 2),
            'avg_time_per_text_ms': round(avg_time_per_text, 2)
        })

    except Exception as e:
        logger.error(f"Error embedding batch: {str(e)}")
        return jsonify({'error': str(e)}), 500

@app.route('/similarity', methods=['POST'])
def similarity():
    """
    Calculate cosine similarity between two texts.

    Request:
        {
            "text1": "first text",
            "text2": "second text"
        }

    Response:
        {
            "similarity": 0.85,
            "processing_time_ms": 50
        }
    """
    if model is None:
        return jsonify({'error': 'Model not loaded'}), 500

    data = request.json
    if not data or 'text1' not in data or 'text2' not in data:
        return jsonify({'error': 'Missing "text1" or "text2" field'}), 400

    start_time = time.time()

    try:
        # Encode both texts
        embeddings = model.encode([data['text1'], data['text2']], convert_to_numpy=True)

        # Calculate cosine similarity
        embedding1 = embeddings[0]
        embedding2 = embeddings[1]

        # Cosine similarity = dot(a, b) / (norm(a) * norm(b))
        similarity = np.dot(embedding1, embedding2) / (
            np.linalg.norm(embedding1) * np.linalg.norm(embedding2)
        )

        processing_time = (time.time() - start_time) * 1000

        return jsonify({
            'similarity': float(similarity),
            'processing_time_ms': round(processing_time, 2)
        })

    except Exception as e:
        logger.error(f"Error calculating similarity: {str(e)}")
        return jsonify({'error': str(e)}), 500

def main():
    """Main entry point."""
    logger.info("=" * 60)
    logger.info("Local Embedding Service")
    logger.info("=" * 60)
    logger.info("FREE alternative to OpenAI/Anthropic embeddings")
    logger.info("")

    # Load model at startup
    load_model()

    logger.info("")
    logger.info("Starting Flask server...")
    logger.info("Endpoints:")
    logger.info("  POST http://localhost:5000/embed")
    logger.info("  POST http://localhost:5000/embed_batch")
    logger.info("  POST http://localhost:5000/similarity")
    logger.info("  GET  http://localhost:5000/health")
    logger.info("  GET  http://localhost:5000/model_info")
    logger.info("")

    # Start Flask server
    app.run(
        host='localhost',
        port=5000,
        debug=False,
        threaded=True
    )

if __name__ == '__main__':
    main()
