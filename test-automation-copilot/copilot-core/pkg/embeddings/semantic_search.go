package embeddings

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourusername/copilot-core/pkg/database"
)

// SemanticSearch provides semantic search capabilities over the codebase
type SemanticSearch struct {
	vectorStore *VectorStore
	db          *database.DB
	classRepo   *database.ClassRepository
	methodRepo  *database.MethodRepository
}

// SemanticSearchConfig configures semantic search
type SemanticSearchConfig struct {
	VectorStore *VectorStore
	DB          *database.DB
}

// NewSemanticSearch creates a new semantic search instance
func NewSemanticSearch(config SemanticSearchConfig) (*SemanticSearch, error) {
	if config.VectorStore == nil {
		return nil, fmt.Errorf("vector store is required")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &SemanticSearch{
		vectorStore: config.VectorStore,
		db:          config.DB,
		classRepo:   database.NewClassRepository(config.DB),
		methodRepo:  database.NewMethodRepository(config.DB),
	}, nil
}

// CodeSearchResult represents a code search result
type CodeSearchResult struct {
	Type       string  // "class", "method", "field"
	ID         int64   // Database ID
	Name       string  // Name of the entity
	FQN        string  // Fully qualified name
	FilePath   string  // File path
	Content    string  // Code content
	Javadoc    string  // Documentation
	Similarity float32 // Similarity score
	StartLine  int     // Start line number
	EndLine    int     // End line number
}

// SearchCode searches for relevant code based on a query
func (ss *SemanticSearch) SearchCode(ctx context.Context, query string, limit int) ([]CodeSearchResult, error) {
	// Search vector store
	results, err := ss.vectorStore.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert to code search results
	codeResults := make([]CodeSearchResult, 0, len(results))

	for _, result := range results {
		// Extract metadata
		entityType, ok := result.Metadata["type"].(string)
		if !ok {
			continue
		}

		entityIDStr, ok := result.Metadata["id"].(string)
		if !ok {
			continue
		}

		var entityID int64
		fmt.Sscanf(entityIDStr, "%d", &entityID)

		// Build code search result based on type
		switch entityType {
		case "class":
			class, err := ss.classRepo.GetByID(entityID)
			if err != nil {
				continue
			}

			// Get file path
			file, err := database.NewFileRepository(ss.db).GetByID(class.FileID)
			if err != nil {
				continue
			}

			javadoc := ""
			if class.Javadoc != nil {
				javadoc = *class.Javadoc
			}

			codeResults = append(codeResults, CodeSearchResult{
				Type:       "class",
				ID:         class.ID,
				Name:       class.Name,
				FQN:        class.FullyQualifiedName,
				FilePath:   file.Path,
				Content:    result.Content,
				Javadoc:    javadoc,
				Similarity: result.Similarity,
				StartLine:  class.StartLine,
				EndLine:    class.EndLine,
			})

		case "method":
			method, err := ss.methodRepo.GetByID(entityID)
			if err != nil {
				continue
			}

			// Get class and file
			class, err := ss.classRepo.GetByID(method.ClassID)
			if err != nil {
				continue
			}

			file, err := database.NewFileRepository(ss.db).GetByID(class.FileID)
			if err != nil {
				continue
			}

			javadoc := ""
			if method.Javadoc != nil {
				javadoc = *method.Javadoc
			}

			fqn := class.FullyQualifiedName + "." + method.Name

			codeResults = append(codeResults, CodeSearchResult{
				Type:       "method",
				ID:         method.ID,
				Name:       method.Name,
				FQN:        fqn,
				FilePath:   file.Path,
				Content:    result.Content,
				Javadoc:    javadoc,
				Similarity: result.Similarity,
				StartLine:  method.StartLine,
				EndLine:    method.EndLine,
			})
		}
	}

	return codeResults, nil
}

// SearchPageObjects searches for Page Object classes
func (ss *SemanticSearch) SearchPageObjects(ctx context.Context, query string, limit int) ([]CodeSearchResult, error) {
	// Add filter for Page Objects
	filter := map[string]string{
		"type": "class",
	}

	results, err := ss.vectorStore.SearchWithFilter(ctx, query, limit*2, filter)
	if err != nil {
		return nil, err
	}

	// Filter for Page Objects (classes ending with "Page")
	codeResults := make([]CodeSearchResult, 0)

	for _, result := range results {
		entityIDStr, ok := result.Metadata["id"].(string)
		if !ok {
			continue
		}

		var entityID int64
		fmt.Sscanf(entityIDStr, "%d", &entityID)

		class, err := ss.classRepo.GetByID(entityID)
		if err != nil {
			continue
		}

		// Check if it's a Page Object
		if !strings.HasSuffix(class.Name, "Page") {
			continue
		}

		file, err := database.NewFileRepository(ss.db).GetByID(class.FileID)
		if err != nil {
			continue
		}

		javadoc := ""
		if class.Javadoc != nil {
			javadoc = *class.Javadoc
		}

		codeResults = append(codeResults, CodeSearchResult{
			Type:       "class",
			ID:         class.ID,
			Name:       class.Name,
			FQN:        class.FullyQualifiedName,
			FilePath:   file.Path,
			Content:    result.Content,
			Javadoc:    javadoc,
			Similarity: result.Similarity,
			StartLine:  class.StartLine,
			EndLine:    class.EndLine,
		})

		if len(codeResults) >= limit {
			break
		}
	}

	return codeResults, nil
}

// SearchTests searches for test classes and methods
func (ss *SemanticSearch) SearchTests(ctx context.Context, query string, limit int) ([]CodeSearchResult, error) {
	results, err := ss.vectorStore.Search(ctx, query, limit*2)
	if err != nil {
		return nil, err
	}

	codeResults := make([]CodeSearchResult, 0)

	for _, result := range results {
		entityType, ok := result.Metadata["type"].(string)
		if !ok {
			continue
		}

		entityIDStr, ok := result.Metadata["id"].(string)
		if !ok {
			continue
		}

		var entityID int64
		fmt.Sscanf(entityIDStr, "%d", &entityID)

		if entityType == "method" {
			method, err := ss.methodRepo.GetByID(entityID)
			if err != nil {
				continue
			}

			// Check if it's a test method (has @Test annotation)
			if method.Annotations == nil || !strings.Contains(*method.Annotations, "@Test") {
				continue
			}

			class, err := ss.classRepo.GetByID(method.ClassID)
			if err != nil {
				continue
			}

			file, err := database.NewFileRepository(ss.db).GetByID(class.FileID)
			if err != nil {
				continue
			}

			javadoc := ""
			if method.Javadoc != nil {
				javadoc = *method.Javadoc
			}

			fqn := class.FullyQualifiedName + "." + method.Name

			codeResults = append(codeResults, CodeSearchResult{
				Type:       "method",
				ID:         method.ID,
				Name:       method.Name,
				FQN:        fqn,
				FilePath:   file.Path,
				Content:    result.Content,
				Javadoc:    javadoc,
				Similarity: result.Similarity,
				StartLine:  method.StartLine,
				EndLine:    method.EndLine,
			})

			if len(codeResults) >= limit {
				break
			}
		}
	}

	return codeResults, nil
}

// FindSimilarCode finds code similar to a given snippet
func (ss *SemanticSearch) FindSimilarCode(ctx context.Context, codeSnippet string, limit int) ([]CodeSearchResult, error) {
	return ss.SearchCode(ctx, codeSnippet, limit)
}

// SearchByIntent searches based on user intent (natural language)
func (ss *SemanticSearch) SearchByIntent(ctx context.Context, intent string, limit int) ([]CodeSearchResult, error) {
	// Enhance query with common test automation terms
	enhancedQuery := intent

	// Add context based on intent keywords
	intentLower := strings.ToLower(intent)
	if strings.Contains(intentLower, "login") {
		enhancedQuery += " authentication login page credentials"
	} else if strings.Contains(intentLower, "search") {
		enhancedQuery += " search functionality input"
	} else if strings.Contains(intentLower, "add") || strings.Contains(intentLower, "create") {
		enhancedQuery += " create add new form"
	}

	return ss.SearchCode(ctx, enhancedQuery, limit)
}
