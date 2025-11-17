package ai

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
)

// SemanticIndex handles embedding-based semantic search
// Uses local embedding service (FREE!) to find conceptually similar code
type SemanticIndex struct {
	db              *sql.DB
	embeddingClient *EmbeddingClient
	dimension       int // Embedding dimension (default: 384)
}

// NewSemanticIndex creates a new semantic index
func NewSemanticIndex(db *sql.DB, embeddingClient *EmbeddingClient, dimension int) *SemanticIndex {
	return &SemanticIndex{
		db:              db,
		embeddingClient: embeddingClient,
		dimension:       dimension,
	}
}

// BuildIndex generates embeddings for all chunks
func (si *SemanticIndex) BuildIndex() error {
	fmt.Println("🧠 Building semantic index...")

	// Check if embedding service is available
	if err := si.embeddingClient.HealthCheck(); err != nil {
		return fmt.Errorf("embedding service not available: %w", err)
	}

	// Get all chunks without embeddings
	rows, err := si.db.Query(`
		SELECT id, content
		FROM chunks
		WHERE embedding IS NULL OR embedding = ''
		ORDER BY id
	`)

	if err != nil {
		return fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	// Collect chunks for batch processing
	type chunkToEmbed struct {
		id      int
		content string
	}

	chunks := []chunkToEmbed{}

	for rows.Next() {
		var id int
		var content string

		if err := rows.Scan(&id, &content); err != nil {
			continue
		}

		chunks = append(chunks, chunkToEmbed{id: id, content: content})
	}

	if len(chunks) == 0 {
		fmt.Println("  ✅ All chunks already have embeddings")
		return nil
	}

	fmt.Printf("  📦 Generating embeddings for %d chunks...\n", len(chunks))

	// Process in batches for better performance
	batchSize := 10
	processedCount := 0

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[i:end]

		// Extract content for batch
		texts := make([]string, len(batch))
		for j, chunk := range batch {
			texts[j] = chunk.content
		}

		// Generate embeddings for batch
		embeddings, err := si.embeddingClient.EmbedBatch(texts)
		if err != nil {
			fmt.Printf("  ⚠️  Warning: Failed to embed batch: %v\n", err)
			continue
		}

		// Save embeddings to database
		for j, chunk := range batch {
			if j >= len(embeddings) {
				break
			}

			embeddingBytes := floatsToBytes(embeddings[j])

			_, err := si.db.Exec(`
				UPDATE chunks
				SET embedding = ?
				WHERE id = ?
			`, embeddingBytes, chunk.id)

			if err != nil {
				fmt.Printf("  ⚠️  Warning: Failed to save embedding for chunk %d: %v\n", chunk.id, err)
			} else {
				processedCount++
			}
		}

		// Progress update
		if processedCount%50 == 0 {
			fmt.Printf("  ⏳ Processed %d/%d chunks...\n", processedCount, len(chunks))
		}
	}

	fmt.Printf("  ✅ Generated %d embeddings\n", processedCount)
	return nil
}

// Search performs semantic search using cosine similarity
func (si *SemanticIndex) Search(queryEmbedding []float64, topK int) ([]*SearchResult, error) {
	if len(queryEmbedding) == 0 {
		return []*SearchResult{}, nil
	}

	// Get all chunks with embeddings
	rows, err := si.db.Query(`
		SELECT id, chunk_type, entity_id, embedding
		FROM chunks
		WHERE embedding IS NOT NULL AND embedding != ''
		ORDER BY id
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	results := []*SearchResult{}

	for rows.Next() {
		var chunkID, entityID int
		var chunkType string
		var embeddingBytes []byte

		if err := rows.Scan(&chunkID, &chunkType, &entityID, &embeddingBytes); err != nil {
			continue
		}

		// Convert bytes back to float64 array
		chunkEmbedding := bytesToFloats(embeddingBytes)

		if len(chunkEmbedding) != len(queryEmbedding) {
			continue // Dimension mismatch
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryEmbedding, chunkEmbedding)

		if similarity > 0 {
			results = append(results, &SearchResult{
				ChunkID:       chunkID,
				ChunkType:     chunkType,
				EntityID:      entityID,
				SemanticScore: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	sortBySemanticScore(results)

	// Return top-K
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// GetStats returns index statistics
func (si *SemanticIndex) GetStats() (map[string]interface{}, error) {
	var totalChunks, withEmbeddings int

	err := si.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN embedding IS NOT NULL AND embedding != '' THEN 1 END) as with_embeddings
		FROM chunks
	`).Scan(&totalChunks, &withEmbeddings)

	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_chunks":    totalChunks,
		"with_embeddings": withEmbeddings,
		"dimension":       si.dimension,
		"coverage":        float64(withEmbeddings) / float64(totalChunks) * 100,
	}

	return stats, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
// Returns a value between 0 and 1 (1 = identical, 0 = completely different)
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// floatsToBytes converts []float64 to []byte for storage
func floatsToBytes(floats []float64) []byte {
	bytes := make([]byte, len(floats)*8)

	for i, f := range floats {
		bits := math.Float64bits(f)
		binary.LittleEndian.PutUint64(bytes[i*8:(i+1)*8], bits)
	}

	return bytes
}

// bytesToFloats converts []byte back to []float64
func bytesToFloats(bytes []byte) []float64 {
	if len(bytes)%8 != 0 {
		return []float64{}
	}

	floats := make([]float64, len(bytes)/8)

	for i := 0; i < len(floats); i++ {
		bits := binary.LittleEndian.Uint64(bytes[i*8 : (i+1)*8])
		floats[i] = math.Float64frombits(bits)
	}

	return floats
}

// sortBySemanticScore sorts search results by semantic score (descending)
func sortBySemanticScore(results []*SearchResult) {
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].SemanticScore < results[j+1].SemanticScore {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// ClearIndex removes all embeddings
func (si *SemanticIndex) ClearIndex() error {
	_, err := si.db.Exec(`UPDATE chunks SET embedding = NULL`)
	return err
}
