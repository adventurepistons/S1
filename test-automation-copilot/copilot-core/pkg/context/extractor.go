package context

import (
	"context"
	"fmt"

	"github.com/yourusername/copilot-core/pkg/cloud"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/embeddings"
)

// ContextExtractor extracts and prepares context for cloud AI processing
// It does NOT contain prompts - those stay on the cloud backend
type ContextExtractor struct {
	db             *database.DB
	semanticSearch *embeddings.SemanticSearch
	classRepo      *database.ClassRepository
	methodRepo     *database.MethodRepository
	fileRepo       *database.FileRepository
}

// ContextExtractorConfig configures the context extractor
type ContextExtractorConfig struct {
	DB             *database.DB
	SemanticSearch *embeddings.SemanticSearch
}

// NewContextExtractor creates a new context extractor
func NewContextExtractor(config ContextExtractorConfig) (*ContextExtractor, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database is required")
	}
	if config.SemanticSearch == nil {
		return nil, fmt.Errorf("semantic search is required")
	}

	return &ContextExtractor{
		db:             config.DB,
		semanticSearch: config.SemanticSearch,
		classRepo:      database.NewClassRepository(config.DB),
		methodRepo:     database.NewMethodRepository(config.DB),
		fileRepo:       database.NewFileRepository(config.DB),
	}, nil
}

// ExtractPageObjectContext extracts context for page object generation
func (ce *ContextExtractor) ExtractPageObjectContext(
	ctx context.Context,
	spec string,
	elements []cloud.Element,
) (cloud.ContextPayload, error) {
	payload := cloud.ContextPayload{
		Action:    "pageobject",
		UserQuery: spec,
		Spec:      spec,
		Elements:  elements,
	}

	// Search for similar page objects
	similarPages, err := ce.semanticSearch.SearchPageObjects(ctx, spec, 5)
	if err == nil && len(similarPages) > 0 {
		for _, result := range similarPages {
			payload.PageObjects = append(payload.PageObjects, cloud.PageObject{
				Name:     result.Name,
				FilePath: result.FilePath,
				Methods:  ce.extractMethodNames(ctx, result.ID),
			})

			payload.RelevantCode = append(payload.RelevantCode, cloud.CodeSnippet{
				Name:       result.Name,
				Type:       result.Type,
				Content:    result.Content,
				FilePath:   result.FilePath,
				Similarity: result.Similarity,
			})
		}
	}

	// Get workspace metadata
	payload.Workspace = ce.getWorkspaceInfo(ctx)

	// Detect framework and test runner from existing code
	framework, testRunner := ce.detectFramework(ctx)
	payload.Framework = framework
	payload.TestRunner = testRunner

	return payload, nil
}

// ExtractTestContext extracts context for test case generation
func (ce *ContextExtractor) ExtractTestContext(
	ctx context.Context,
	spec string,
) (cloud.ContextPayload, error) {
	payload := cloud.ContextPayload{
		Action:    "test",
		UserQuery: spec,
		Spec:      spec,
	}

	// Search for similar tests
	similarTests, err := ce.semanticSearch.SearchTests(ctx, spec, 5)
	if err == nil && len(similarTests) > 0 {
		for _, result := range similarTests {
			payload.TestMethods = append(payload.TestMethods, cloud.TestMethod{
				Name:      result.Name,
				ClassName: ce.getClassName(ctx, result.ID),
			})

			payload.RelevantCode = append(payload.RelevantCode, cloud.CodeSnippet{
				Name:       result.Name,
				Type:       result.Type,
				Content:    result.Content,
				FilePath:   result.FilePath,
				Similarity: result.Similarity,
			})
		}
	}

	// Search for relevant page objects that might be used in this test
	pageObjects, err := ce.semanticSearch.SearchPageObjects(ctx, spec, 3)
	if err == nil && len(pageObjects) > 0 {
		for _, result := range pageObjects {
			payload.PageObjects = append(payload.PageObjects, cloud.PageObject{
				Name:     result.Name,
				FilePath: result.FilePath,
				Methods:  ce.extractMethodNames(ctx, result.ID),
			})
		}
	}

	// Get workspace metadata
	payload.Workspace = ce.getWorkspaceInfo(ctx)

	// Detect framework and test runner
	framework, testRunner := ce.detectFramework(ctx)
	payload.Framework = framework
	payload.TestRunner = testRunner

	return payload, nil
}

// ExtractChatContext extracts context for chat/Q&A
func (ce *ContextExtractor) ExtractChatContext(
	ctx context.Context,
	userMessage string,
) (cloud.ContextPayload, error) {
	payload := cloud.ContextPayload{
		Action:    "chat",
		UserQuery: userMessage,
	}

	// Search for relevant code across entire codebase
	relevantCode, err := ce.semanticSearch.SearchCode(ctx, userMessage, 10)
	if err == nil && len(relevantCode) > 0 {
		for _, result := range relevantCode {
			payload.RelevantCode = append(payload.RelevantCode, cloud.CodeSnippet{
				Name:       result.Name,
				Type:       result.Type,
				Content:    result.Content,
				FilePath:   result.FilePath,
				Similarity: result.Similarity,
			})
		}
	}

	// Get workspace metadata for context
	payload.Workspace = ce.getWorkspaceInfo(ctx)

	// Detect framework
	framework, testRunner := ce.detectFramework(ctx)
	payload.Framework = framework
	payload.TestRunner = testRunner

	return payload, nil
}

// ExtractFixContext extracts context for code fixing
func (ce *ContextExtractor) ExtractFixContext(
	ctx context.Context,
	brokenCode string,
	errorMessage string,
	errorType string,
) (cloud.ContextPayload, error) {
	payload := cloud.ContextPayload{
		Action:    "fix",
		UserQuery: fmt.Sprintf("Fix error: %s", errorMessage),
		ErrorInfo: &cloud.ErrorContext{
			BrokenCode:   brokenCode,
			ErrorMessage: errorMessage,
			ErrorType:    errorType,
		},
	}

	// Search for similar working code that might help
	// Extract key terms from error message for better search
	searchQuery := ce.extractErrorSearchTerms(errorMessage)
	relevantCode, err := ce.semanticSearch.SearchCode(ctx, searchQuery, 5)
	if err == nil && len(relevantCode) > 0 {
		for _, result := range relevantCode {
			payload.RelevantCode = append(payload.RelevantCode, cloud.CodeSnippet{
				Name:       result.Name,
				Type:       result.Type,
				Content:    result.Content,
				FilePath:   result.FilePath,
				Similarity: result.Similarity,
			})
		}
	}

	// Get workspace metadata
	payload.Workspace = ce.getWorkspaceInfo(ctx)

	// Detect framework
	framework, testRunner := ce.detectFramework(ctx)
	payload.Framework = framework
	payload.TestRunner = testRunner

	return payload, nil
}

// Helper methods

func (ce *ContextExtractor) extractMethodNames(ctx context.Context, classID int64) []string {
	// Query methods by class ID
	methods, err := ce.methodRepo.GetByClassID(classID)
	if err != nil {
		return []string{}
	}

	names := make([]string, len(methods))
	for i, method := range methods {
		names[i] = method.Name
	}
	return names
}

func (ce *ContextExtractor) getClassName(ctx context.Context, id int64) string {
	// Get class by ID and return name
	class, err := ce.classRepo.GetByID(id)
	if err != nil {
		return ""
	}
	return class.Name
}

func (ce *ContextExtractor) getWorkspaceInfo(ctx context.Context) cloud.WorkspaceInfo {
	// Get workspace statistics from database
	classCount, _ := ce.classRepo.Count()

	// Count page objects (classes with "Page" suffix)
	pageObjectCount := 0
	classes, err := ce.classRepo.GetAll()
	if err == nil {
		for _, class := range classes {
			if len(class.Name) > 4 && class.Name[len(class.Name)-4:] == "Page" {
				pageObjectCount++
			}
		}
	}

	// Count test methods
	testMethods, err := ce.methodRepo.GetTestMethods()
	testMethodCount := 0
	if err == nil {
		testMethodCount = len(testMethods)
	}

	return cloud.WorkspaceInfo{
		TotalClasses:    int(classCount),
		TotalTests:      testMethodCount,
		PageObjectCount: pageObjectCount,
	}
}

func (ce *ContextExtractor) detectFramework(ctx context.Context) (framework string, testRunner string) {
	framework = "selenium-java"
	testRunner = "testng"

	// Try to detect from existing files
	files, err := ce.fileRepo.GetAll()
	if err != nil {
		return
	}

	// Look for framework indicators in file paths
	for _, file := range files {
		if contains(file.Path, "playwright") {
			framework = "playwright-java"
		} else if contains(file.Path, "cypress") {
			framework = "cypress"
		}

		if contains(file.Path, "junit") {
			testRunner = "junit"
		} else if contains(file.Path, "cucumber") {
			testRunner = "cucumber"
		}
	}

	return
}

func (ce *ContextExtractor) extractErrorSearchTerms(errorMessage string) string {
	// Simple extraction - in practice you'd use more sophisticated NLP
	// For now, just use the error message as-is
	return errorMessage
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" && substr != "" &&
		(s == substr || (len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
