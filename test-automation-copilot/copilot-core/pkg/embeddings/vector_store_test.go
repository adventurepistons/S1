package embeddings

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupTestVectorStore(t *testing.T) (*VectorStore, func()) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectordb")

	// Create embedding service
	embeddingService, err := NewEmbeddingService(EmbeddingConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create embedding service: %v", err)
	}

	// Create vector store
	vectorStore, err := NewVectorStore(VectorStoreConfig{
		PersistPath:      dbPath,
		CollectionName:   "test-collection",
		EmbeddingService: embeddingService,
	})
	if err != nil {
		t.Fatalf("Failed to create vector store: %v", err)
	}

	cleanup := func() {
		vectorStore.Close()
	}

	return vectorStore, cleanup
}

func TestNewVectorStore(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	if vs == nil {
		t.Fatal("Vector store is nil")
	}

	stats := vs.GetStats()
	if stats.CollectionName != "test-collection" {
		t.Errorf("Expected collection name 'test-collection', got %s", stats.CollectionName)
	}

	if stats.DocumentCount != 0 {
		t.Errorf("Expected 0 documents, got %d", stats.DocumentCount)
	}
}

func TestAddDocument(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	doc := Document{
		ID:      "doc1",
		Content: "This is a test document about login functionality",
		Metadata: map[string]interface{}{
			"type": "test",
			"id":   "1",
		},
	}

	err := vs.AddDocument(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to add document: %v", err)
	}

	// Check count
	count := vs.Count()
	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}
}

func TestAddDocuments_Batch(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	docs := []Document{
		{
			ID:      "doc1",
			Content: "Login page with username and password fields",
			Metadata: map[string]interface{}{
				"type": "page",
			},
		},
		{
			ID:      "doc2",
			Content: "Dashboard showing user statistics and graphs",
			Metadata: map[string]interface{}{
				"type": "page",
			},
		},
		{
			ID:      "doc3",
			Content: "Profile page with user information and settings",
			Metadata: map[string]interface{}{
				"type": "page",
			},
		},
	}

	err := vs.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Check count
	count := vs.Count()
	if count != 3 {
		t.Errorf("Expected 3 documents, got %d", count)
	}
}

func TestSearch(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add test documents
	docs := []Document{
		{
			ID:      "login",
			Content: "Login page with username and password authentication",
			Metadata: map[string]interface{}{
				"type": "page",
				"name": "LoginPage",
			},
		},
		{
			ID:      "dashboard",
			Content: "Dashboard with charts, graphs and statistics",
			Metadata: map[string]interface{}{
				"type": "page",
				"name": "DashboardPage",
			},
		},
		{
			ID:      "profile",
			Content: "User profile with personal information and settings",
			Metadata: map[string]interface{}{
				"type": "page",
				"name": "ProfilePage",
			},
		},
	}

	err := vs.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Search for login-related content
	results, err := vs.Search(ctx, "authentication and credentials", 2)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("No results returned")
	}

	// Top result should be login page
	topResult := results[0]
	if topResult.ID != "login" {
		t.Errorf("Expected top result to be 'login', got %s", topResult.ID)
	}

	// Check similarity score is reasonable
	if topResult.Similarity < 0.5 {
		t.Errorf("Expected similarity > 0.5, got %f", topResult.Similarity)
	}
}

func TestSearchWithFilter(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add documents with different types
	docs := []Document{
		{
			ID:      "login-page",
			Content: "Login page object",
			Metadata: map[string]interface{}{
				"type": "page",
			},
		},
		{
			ID:      "login-test",
			Content: "Login test case",
			Metadata: map[string]interface{}{
				"type": "test",
			},
		},
	}

	err := vs.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Search only for pages
	filter := map[string]string{
		"type": "page",
	}

	results, err := vs.SearchWithFilter(ctx, "login", 10, filter)
	if err != nil {
		t.Fatalf("Failed to search with filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].ID != "login-page" {
		t.Errorf("Expected 'login-page', got %s", results[0].ID)
	}
}

func TestDelete(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add document
	doc := Document{
		ID:      "test-doc",
		Content: "Test document to be deleted",
		Metadata: map[string]interface{}{
			"type": "test",
		},
	}

	err := vs.AddDocument(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to add document: %v", err)
	}

	// Verify it exists
	if vs.Count() != 1 {
		t.Error("Document not added")
	}

	// Delete it
	err = vs.Delete(ctx, "test-doc")
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// Verify it's gone
	if vs.Count() != 0 {
		t.Error("Document not deleted")
	}
}

func TestClear(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add multiple documents
	docs := []Document{
		{ID: "doc1", Content: "Document 1", Metadata: map[string]interface{}{}},
		{ID: "doc2", Content: "Document 2", Metadata: map[string]interface{}{}},
		{ID: "doc3", Content: "Document 3", Metadata: map[string]interface{}{}},
	}

	err := vs.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Verify they exist
	if vs.Count() != 3 {
		t.Errorf("Expected 3 documents, got %d", vs.Count())
	}

	// Clear all
	err = vs.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear: %v", err)
	}

	// Verify all gone
	if vs.Count() != 0 {
		t.Errorf("Expected 0 documents after clear, got %d", vs.Count())
	}
}

func TestGetStats(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	stats := vs.GetStats()

	if stats.CollectionName != "test-collection" {
		t.Errorf("Wrong collection name: %s", stats.CollectionName)
	}

	if stats.DocumentCount != 0 {
		t.Errorf("Expected 0 documents, got %d", stats.DocumentCount)
	}

	if stats.PersistPath == "" {
		t.Error("Persist path is empty")
	}
}

// Test semantic search accuracy
func TestSemanticSearchAccuracy(t *testing.T) {
	vs, cleanup := setupTestVectorStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add documents with clear semantic differences
	docs := []Document{
		{
			ID:      "java-class",
			Content: "public class LoginPage extends BasePage { private WebElement usernameField; }",
			Metadata: map[string]interface{}{
				"type": "code",
				"lang": "java",
			},
		},
		{
			ID:      "test-method",
			Content: "@Test public void testLoginWithValidCredentials() { loginPage.login(user, pass); }",
			Metadata: map[string]interface{}{
				"type": "code",
				"lang": "java",
			},
		},
		{
			ID:      "feature-file",
			Content: "Feature: Login\n  Scenario: User logs in with valid credentials",
			Metadata: map[string]interface{}{
				"type": "code",
				"lang": "gherkin",
			},
		},
	}

	err := vs.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Search for test-related content
	results, err := vs.Search(ctx, "test case with valid credentials", 3)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// The test method should rank highly
	found := false
	for i, result := range results {
		if result.ID == "test-method" && i < 2 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Test method not found in top 2 results for test-related query")
	}
}

// Benchmark vector search
func BenchmarkSearch(b *testing.B) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set")
	}

	// Setup
	tmpDir := b.TempDir()
	embeddingService, _ := NewEmbeddingService(EmbeddingConfig{APIKey: apiKey})
	vs, _ := NewVectorStore(VectorStoreConfig{
		PersistPath:      filepath.Join(tmpDir, "vectordb"),
		EmbeddingService: embeddingService,
	})

	ctx := context.Background()

	// Add 10 documents
	docs := make([]Document, 10)
	for i := 0; i < 10; i++ {
		docs[i] = Document{
			ID:       fmt.Sprintf("doc%d", i),
			Content:  fmt.Sprintf("Test document number %d with some content", i),
			Metadata: map[string]interface{}{"id": i},
		}
	}
	vs.AddDocuments(ctx, docs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.Search(ctx, "test content", 5)
	}
}
