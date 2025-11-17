package ai

import (
	"database/sql"
	"fmt"
	"math"
)

// HybridRetriever combines BM25 (keyword) and semantic (embedding) search
// Research shows hybrid retrieval achieves 19% better precision than either alone
type HybridRetriever struct {
	db              *sql.DB
	bm25Index       *BM25Index
	semanticIndex   *SemanticIndex
	embeddingClient *EmbeddingClient
	alpha           float64 // Weight for semantic score (default: 0.65)
}

// NewHybridRetriever creates a new hybrid retriever
func NewHybridRetriever(db *sql.DB, embeddingClient *EmbeddingClient, config Config) *HybridRetriever {
	return &HybridRetriever{
		db:              db,
		bm25Index:       NewBM25Index(db, config),
		semanticIndex:   NewSemanticIndex(db, embeddingClient, config.EmbeddingDimension),
		embeddingClient: embeddingClient,
		alpha:           config.SemanticWeight,
	}
}

// Retrieve performs hybrid retrieval combining BM25 and semantic search
func (hr *HybridRetriever) Retrieve(query string, topK int) ([]*SearchResult, error) {
	// Step 1: BM25 retrieval (keyword-based)
	bm25Results, err := hr.bm25Index.Search(query, topK*2) // Get 2x for reranking
	if err != nil {
		return nil, fmt.Errorf("BM25 search failed: %w", err)
	}

	// Step 2: Generate query embedding
	queryEmbedding, err := hr.embeddingClient.Embed(query)
	if err != nil {
		// If embedding fails, fall back to BM25 only
		fmt.Printf("Warning: Embedding failed, using BM25 only: %v\n", err)
		if len(bm25Results) > topK {
			return bm25Results[:topK], nil
		}
		return bm25Results, nil
	}

	// Step 3: Semantic retrieval (embedding-based)
	semanticResults, err := hr.semanticIndex.Search(queryEmbedding, topK*2)
	if err != nil {
		// If semantic search fails, fall back to BM25 only
		fmt.Printf("Warning: Semantic search failed, using BM25 only: %v\n", err)
		if len(bm25Results) > topK {
			return bm25Results[:topK], nil
		}
		return bm25Results, nil
	}

	// Step 4: Fuse scores using weighted combination
	fusedResults := hr.fuseResults(bm25Results, semanticResults)

	// Step 5: Sort by final score (descending)
	sortByFinalScore(fusedResults)

	// Step 6: Enrich with content
	for _, result := range fusedResults {
		if err := hr.enrichResult(result); err != nil {
			fmt.Printf("Warning: Failed to enrich result for chunk %d: %v\n", result.ChunkID, err)
		}
	}

	// Step 7: Return top-K
	if len(fusedResults) > topK {
		fusedResults = fusedResults[:topK]
	}

	return fusedResults, nil
}

// fuseResults combines BM25 and semantic results using score fusion
func (hr *HybridRetriever) fuseResults(bm25Results, semanticResults []*SearchResult) []*SearchResult {
	// Create a map to combine results by chunk ID
	resultMap := make(map[int]*SearchResult)

	// Normalize BM25 scores to [0, 1] range
	maxBM25 := 0.0
	for _, r := range bm25Results {
		if r.BM25Score > maxBM25 {
			maxBM25 = r.BM25Score
		}
	}

	// Add BM25 results to map
	for _, r := range bm25Results {
		normalizedBM25 := 0.0
		if maxBM25 > 0 {
			normalizedBM25 = r.BM25Score / maxBM25
		}

		resultMap[r.ChunkID] = &SearchResult{
			ChunkID:       r.ChunkID,
			ChunkType:     r.ChunkType,
			EntityID:      r.EntityID,
			BM25Score:     normalizedBM25,
			SemanticScore: 0.0,
		}
	}

	// Normalize semantic scores (already in [0, 1] from cosine similarity)
	// Add semantic results to map
	for _, r := range semanticResults {
		if existing, ok := resultMap[r.ChunkID]; ok {
			// Chunk exists in both - update semantic score
			existing.SemanticScore = r.SemanticScore
		} else {
			// Chunk only in semantic results
			resultMap[r.ChunkID] = &SearchResult{
				ChunkID:       r.ChunkID,
				ChunkType:     r.ChunkType,
				EntityID:      r.EntityID,
				BM25Score:     0.0,
				SemanticScore: r.SemanticScore,
			}
		}
	}

	// Calculate final scores using weighted fusion
	// final_score = alpha × semantic_score + (1 - alpha) × bm25_score
	results := make([]*SearchResult, 0, len(resultMap))

	for _, r := range resultMap {
		r.FinalScore = hr.alpha*r.SemanticScore + (1-hr.alpha)*r.BM25Score
		results = append(results, r)
	}

	return results
}

// enrichResult adds content and metadata to search result
func (hr *HybridRetriever) enrichResult(result *SearchResult) error {
	// Get chunk content
	var content string
	var enrichedContent sql.NullString

	err := hr.db.QueryRow(`
		SELECT content, enriched_content
		FROM chunks
		WHERE id = ?
	`, result.ChunkID).Scan(&content, &enrichedContent)

	if err != nil {
		return err
	}

	// Use enriched content if available, otherwise use regular content
	if enrichedContent.Valid && enrichedContent.String != "" {
		result.Content = enrichedContent.String
	} else {
		result.Content = content
	}

	// Add metadata
	result.Metadata = make(map[string]interface{})
	result.Metadata["chunk_type"] = result.ChunkType
	result.Metadata["entity_id"] = result.EntityID

	return nil
}

// GetStats returns statistics about the hybrid retrieval system
func (hr *HybridRetriever) GetStats() (map[string]interface{}, error) {
	bm25Stats, err := hr.bm25Index.GetStats()
	if err != nil {
		return nil, err
	}

	semanticStats, err := hr.semanticIndex.GetStats()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"bm25":     bm25Stats,
		"semantic": semanticStats,
		"alpha":    hr.alpha,
		"fusion":   "weighted_sum",
	}

	return stats, nil
}

// sortByFinalScore sorts search results by final score (descending)
func sortByFinalScore(results []*SearchResult) {
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].FinalScore < results[j+1].FinalScore {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// NormalizeScores normalizes all scores to [0, 1] range
func (hr *HybridRetriever) NormalizeScores(results []*SearchResult) {
	if len(results) == 0 {
		return
	}

	// Find max scores
	maxBM25 := 0.0
	maxSemantic := 0.0
	maxFinal := 0.0

	for _, r := range results {
		maxBM25 = math.Max(maxBM25, r.BM25Score)
		maxSemantic = math.Max(maxSemantic, r.SemanticScore)
		maxFinal = math.Max(maxFinal, r.FinalScore)
	}

	// Normalize
	for _, r := range results {
		if maxBM25 > 0 {
			r.BM25Score = r.BM25Score / maxBM25
		}
		if maxSemantic > 0 {
			r.SemanticScore = r.SemanticScore / maxSemantic
		}
		if maxFinal > 0 {
			r.FinalScore = r.FinalScore / maxFinal
		}
	}
}

// Explain returns an explanation of why a chunk was retrieved
func (hr *HybridRetriever) Explain(result *SearchResult, query string) string {
	explanation := fmt.Sprintf("Chunk %d retrieved with score %.3f\n", result.ChunkID, result.FinalScore)
	explanation += fmt.Sprintf("  BM25 score (keyword match): %.3f\n", result.BM25Score)
	explanation += fmt.Sprintf("  Semantic score (meaning match): %.3f\n", result.SemanticScore)
	explanation += fmt.Sprintf("  Fusion: %.2f × semantic + %.2f × BM25\n", hr.alpha, 1-hr.alpha)

	if result.BM25Score > result.SemanticScore {
		explanation += "  → Primarily keyword-driven match\n"
	} else {
		explanation += "  → Primarily semantic-driven match\n"
	}

	return explanation
}
