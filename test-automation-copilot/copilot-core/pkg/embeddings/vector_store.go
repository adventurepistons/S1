package embeddings

import (
	"context"
	"fmt"

	chromem "github.com/philippgille/chromem-go"
)

// VectorStore manages vector embeddings using chromem-go
type VectorStore struct {
	db               *chromem.DB
	collection       *chromem.Collection
	embeddingService *EmbeddingService
	persistPath      string // Store path since DB doesn't export it
}

// VectorStoreConfig configures the vector store
type VectorStoreConfig struct {
	PersistPath      string            // Path to persist database
	CollectionName   string            // Name of the collection
	EmbeddingService *EmbeddingService // Embedding service to use
}

// Document represents a document to be stored
type Document struct {
	ID       string                 // Unique identifier
	Content  string                 // Text content
	Metadata map[string]interface{} // Additional metadata
}

// SearchResult represents a search result
type SearchResult struct {
	ID         string                 // Document ID
	Content    string                 // Document content
	Metadata   map[string]interface{} // Document metadata
	Similarity float32                // Cosine similarity score (0-1)
}

// NewVectorStore creates a new vector store
func NewVectorStore(config VectorStoreConfig) (*VectorStore, error) {
	if config.EmbeddingService == nil {
		return nil, fmt.Errorf("embedding service is required")
	}

	// Create data directory if it doesn't exist
	if config.PersistPath == "" {
		config.PersistPath = "./data/vectordb"
	}

	// Create chromem database
	db, err := chromem.NewPersistentDB(config.PersistPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create chromem database: %w", err)
	}

	// Collection name
	collectionName := config.CollectionName
	if collectionName == "" {
		collectionName = "test-copilot"
	}

	// Create or get collection
	collection, err := db.GetOrCreateCollection(
		collectionName,
		map[string]string{
			"description": "Test Automation Copilot vector embeddings",
		},
		nil, // We'll provide embeddings ourselves
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return &VectorStore{
		db:               db,
		collection:       collection,
		embeddingService: config.EmbeddingService,
		persistPath:      config.PersistPath,
	}, nil
}

// AddDocument adds a document to the vector store
func (vs *VectorStore) AddDocument(ctx context.Context, doc Document) error {
	// Generate embedding
	embedding, err := vs.embeddingService.Embed(ctx, doc.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert metadata to string map (chromem requirement)
	metadata := make(map[string]string)
	for k, v := range doc.Metadata {
		metadata[k] = fmt.Sprintf("%v", v)
	}

	// Add to collection
	err = vs.collection.AddDocument(ctx, chromem.Document{
		ID:        doc.ID,
		Content:   doc.Content,
		Metadata:  metadata,
		Embedding: embedding,
	})

	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

// AddDocuments adds multiple documents in batch
func (vs *VectorStore) AddDocuments(ctx context.Context, docs []Document) error {
	// Extract texts for batch embedding
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.Content
	}

	// Generate embeddings in batch
	embeddings, err := vs.embeddingService.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	// Add documents to collection
	chromemDocs := make([]chromem.Document, len(docs))
	for i, doc := range docs {
		// Convert metadata
		metadata := make(map[string]string)
		for k, v := range doc.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}

		chromemDocs[i] = chromem.Document{
			ID:        doc.ID,
			Content:   doc.Content,
			Metadata:  metadata,
			Embedding: embeddings[i],
		}
	}

	// Add documents one by one (chromem-go doesn't support batch add)
	for _, doc := range chromemDocs {
		err = vs.collection.AddDocument(ctx, doc)
		if err != nil {
			return fmt.Errorf("failed to add document %s: %w", doc.ID, err)
		}
	}

	return nil
}

// Search performs semantic search
func (vs *VectorStore) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	// Generate query embedding
	queryEmbedding, err := vs.embeddingService.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search collection
	results, err := vs.collection.QueryEmbedding(ctx, queryEmbedding, topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	// Convert results
	searchResults := make([]SearchResult, len(results))
	for i, result := range results {
		// Convert metadata back to map[string]interface{}
		metadata := make(map[string]interface{})
		for k, v := range result.Metadata {
			metadata[k] = v
		}

		searchResults[i] = SearchResult{
			ID:         result.ID,
			Content:    result.Content,
			Metadata:   metadata,
			Similarity: result.Similarity,
		}
	}

	return searchResults, nil
}

// SearchWithFilter performs semantic search with metadata filters
func (vs *VectorStore) SearchWithFilter(ctx context.Context, query string, topK int, filter map[string]string) ([]SearchResult, error) {
	// Generate query embedding
	queryEmbedding, err := vs.embeddingService.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search with filter
	results, err := vs.collection.QueryEmbedding(ctx, queryEmbedding, topK, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	// Convert results
	searchResults := make([]SearchResult, len(results))
	for i, result := range results {
		metadata := make(map[string]interface{})
		for k, v := range result.Metadata {
			metadata[k] = v
		}

		searchResults[i] = SearchResult{
			ID:         result.ID,
			Content:    result.Content,
			Metadata:   metadata,
			Similarity: result.Similarity,
		}
	}

	return searchResults, nil
}

// Delete removes a document by ID
// Note: chromem-go v0.5.0 doesn't support deleting individual documents.
// Use Clear() to remove all documents or recreate the collection.
func (vs *VectorStore) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("delete operation not supported in chromem-go v0.5.0")
}

// DeleteWhere removes documents matching a filter
// Note: chromem-go v0.5.0 doesn't support deleting individual documents.
// Use Clear() to remove all documents or recreate the collection.
func (vs *VectorStore) DeleteWhere(ctx context.Context, filter map[string]string) error {
	return fmt.Errorf("delete operation not supported in chromem-go v0.5.0")
}

// Count returns the number of documents in the collection
func (vs *VectorStore) Count() int {
	return vs.collection.Count()
}

// Clear removes all documents from the collection
func (vs *VectorStore) Clear(ctx context.Context) error {
	// chromem doesn't have a clear method, so we need to delete the collection and recreate
	collectionName := vs.collection.Name

	// Delete collection
	err := vs.db.DeleteCollection(collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	// Recreate collection
	collection, err := vs.db.CreateCollection(
		collectionName,
		map[string]string{
			"description": "Test Automation Copilot vector embeddings",
		},
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to recreate collection: %w", err)
	}

	vs.collection = collection
	return nil
}

// Close closes the vector store and persists data
func (vs *VectorStore) Close() error {
	// chromem-go auto-persists, but we can trigger it explicitly
	return nil
}

// GetCollection returns the underlying chromem collection
func (vs *VectorStore) GetCollection() *chromem.Collection {
	return vs.collection
}

// GetStats returns statistics about the vector store
type VectorStoreStats struct {
	DocumentCount int
	CollectionName string
	PersistPath    string
}

func (vs *VectorStore) GetStats() VectorStoreStats {
	return VectorStoreStats{
		DocumentCount:  vs.collection.Count(),
		CollectionName: vs.collection.Name,
		PersistPath:    vs.persistPath,
	}
}
