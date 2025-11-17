package ai

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
)

// BM25Index implements the BM25 ranking algorithm for code search
// BM25 is excellent for keyword-based retrieval (method names, class names, etc.)
type BM25Index struct {
	db              *sql.DB
	tokenizer       *Tokenizer
	k1              float64 // Term frequency saturation (default: 1.2)
	b               float64 // Length normalization (default: 0.75)
	avgDocLength    float64 // Average document length
	totalDocuments  int     // Total number of documents
	termIDF         map[string]float64 // Pre-calculated IDF scores
}

// NewBM25Index creates a new BM25 index
func NewBM25Index(db *sql.DB, config Config) *BM25Index {
	return &BM25Index{
		db:        db,
		tokenizer: NewTokenizer(),
		k1:        config.BM25_K1,
		b:         config.BM25_B,
		termIDF:   make(map[string]float64),
	}
}

// Document represents a chunk in the BM25 index
type Document struct {
	ChunkID    int
	ChunkType  string
	EntityID   int
	Tokens     []string
	TokenCount int
}

// BuildIndex builds the BM25 index from chunks
func (bm *BM25Index) BuildIndex() error {
	fmt.Println("🔍 Building BM25 index...")

	// Step 1: Load all documents
	documents, err := bm.loadDocuments()
	if err != nil {
		return fmt.Errorf("failed to load documents: %w", err)
	}

	bm.totalDocuments = len(documents)
	if bm.totalDocuments == 0 {
		return fmt.Errorf("no documents to index")
	}

	fmt.Printf("  📊 Loaded %d documents\n", bm.totalDocuments)

	// Step 2: Calculate average document length
	totalLength := 0
	for _, doc := range documents {
		totalLength += doc.TokenCount
	}
	bm.avgDocLength = float64(totalLength) / float64(bm.totalDocuments)

	fmt.Printf("  📏 Average document length: %.1f tokens\n", bm.avgDocLength)

	// Step 3: Calculate document frequencies for each term
	termDocFreq := make(map[string]int)

	for _, doc := range documents {
		// Get unique terms in this document
		uniqueTerms := make(map[string]bool)
		for _, token := range doc.Tokens {
			uniqueTerms[token] = true
		}

		// Increment document frequency for each unique term
		for term := range uniqueTerms {
			termDocFreq[term]++
		}
	}

	fmt.Printf("  📚 Found %d unique terms\n", len(termDocFreq))

	// Step 4: Calculate IDF for each term
	for term, df := range termDocFreq {
		// IDF = log((N - df + 0.5) / (df + 0.5) + 1)
		// This is the BM25 IDF formula (Robertson-Walker)
		idf := math.Log((float64(bm.totalDocuments-df) + 0.5) / (float64(df) + 0.5) + 1.0)
		bm.termIDF[term] = idf

		// Save to database
		_, err := bm.db.Exec(`
			INSERT OR REPLACE INTO bm25_stats (term, document_frequency, idf)
			VALUES (?, ?, ?)
		`, term, df, idf)

		if err != nil {
			fmt.Printf("  ⚠️  Warning: Failed to save term stats for '%s': %v\n", term, err)
		}
	}

	fmt.Printf("  ✅ BM25 index built successfully!\n")
	return nil
}

// Search performs BM25 search and returns top-K results
func (bm *BM25Index) Search(query string, topK int) ([]*SearchResult, error) {
	// Tokenize query
	queryTokens := bm.tokenizer.Tokenize(query)

	if len(queryTokens) == 0 {
		return []*SearchResult{}, nil
	}

	// Load term IDF scores if not in memory
	if len(bm.termIDF) == 0 {
		if err := bm.loadTermIDF(); err != nil {
			return nil, err
		}
	}

	// Load documents
	documents, err := bm.loadDocuments()
	if err != nil {
		return nil, err
	}

	// Calculate BM25 score for each document
	results := []*SearchResult{}

	for _, doc := range documents {
		score := bm.scoreBM25(queryTokens, doc)

		if score > 0 {
			results = append(results, &SearchResult{
				ChunkID:   doc.ChunkID,
				ChunkType: doc.ChunkType,
				EntityID:  doc.EntityID,
				BM25Score: score,
			})
		}
	}

	// Sort by score (descending)
	sortByBM25Score(results)

	// Return top-K
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// scoreBM25 calculates BM25 score for a document given query terms
func (bm *BM25Index) scoreBM25(queryTerms []string, doc Document) float64 {
	score := 0.0

	// Calculate term frequency for this document
	termFreq := make(map[string]int)
	for _, token := range doc.Tokens {
		termFreq[token]++
	}

	// BM25 formula:
	// score = Σ IDF(qi) × (f(qi, D) × (k1 + 1)) / (f(qi, D) + k1 × (1 - b + b × |D| / avgdl))
	//
	// Where:
	// - IDF(qi) = inverse document frequency of query term qi
	// - f(qi, D) = frequency of qi in document D
	// - |D| = length of document D
	// - avgdl = average document length
	// - k1 = term frequency saturation (typically 1.2)
	// - b = length normalization (typically 0.75)

	docLength := float64(doc.TokenCount)

	for _, term := range queryTerms {
		// Get term frequency in document
		tf := float64(termFreq[term])

		if tf == 0 {
			continue // Term not in document
		}

		// Get IDF
		idf, exists := bm.termIDF[term]
		if !exists {
			continue // Term not in index
		}

		// Calculate BM25 component for this term
		numerator := tf * (bm.k1 + 1)
		denominator := tf + bm.k1*(1-bm.b+bm.b*(docLength/bm.avgDocLength))

		score += idf * (numerator / denominator)
	}

	return score
}

// loadDocuments loads all chunks as documents
func (bm *BM25Index) loadDocuments() ([]Document, error) {
	rows, err := bm.db.Query(`
		SELECT id, chunk_type, entity_id, bm25_tokens, token_count
		FROM chunks
		WHERE bm25_tokens IS NOT NULL
		ORDER BY id
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := []Document{}

	for rows.Next() {
		var chunkID, entityID, tokenCount int
		var chunkType, tokensJSON string

		if err := rows.Scan(&chunkID, &chunkType, &entityID, &tokensJSON, &tokenCount); err != nil {
			continue
		}

		// Parse tokens from JSON
		var tokens []string
		if err := json.Unmarshal([]byte(tokensJSON), &tokens); err != nil {
			continue
		}

		documents = append(documents, Document{
			ChunkID:    chunkID,
			ChunkType:  chunkType,
			EntityID:   entityID,
			Tokens:     tokens,
			TokenCount: tokenCount,
		})
	}

	return documents, nil
}

// loadTermIDF loads pre-calculated IDF scores from database
func (bm *BM25Index) loadTermIDF() error {
	rows, err := bm.db.Query(`SELECT term, idf FROM bm25_stats`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var term string
		var idf float64

		if err := rows.Scan(&term, &idf); err != nil {
			continue
		}

		bm.termIDF[term] = idf
	}

	// Also get total documents and avg length
	err = bm.db.QueryRow(`
		SELECT COUNT(*), AVG(token_count)
		FROM chunks
		WHERE bm25_tokens IS NOT NULL
	`).Scan(&bm.totalDocuments, &bm.avgDocLength)

	return err
}

// GetStats returns index statistics
func (bm *BM25Index) GetStats() (map[string]interface{}, error) {
	if len(bm.termIDF) == 0 {
		if err := bm.loadTermIDF(); err != nil {
			return nil, err
		}
	}

	stats := map[string]interface{}{
		"total_documents":  bm.totalDocuments,
		"avg_doc_length":   bm.avgDocLength,
		"total_terms":      len(bm.termIDF),
		"k1":               bm.k1,
		"b":                bm.b,
	}

	return stats, nil
}

// sortByBM25Score sorts search results by BM25 score (descending)
func sortByBM25Score(results []*SearchResult) {
	// Simple bubble sort (fine for small result sets)
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].BM25Score < results[j+1].BM25Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// ClearIndex removes all BM25 statistics
func (bm *BM25Index) ClearIndex() error {
	_, err := bm.db.Exec(`DELETE FROM bm25_stats`)
	if err != nil {
		return err
	}

	bm.termIDF = make(map[string]float64)
	bm.totalDocuments = 0
	bm.avgDocLength = 0

	return nil
}
