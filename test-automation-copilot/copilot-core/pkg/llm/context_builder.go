package llm

import (
	"context"
	"fmt"

	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/embeddings"
	"github.com/yourusername/copilot-core/pkg/prompts"
)

// ContextBuilder builds prompts with relevant code context
type ContextBuilder struct {
	db             *database.DB
	semanticSearch *embeddings.SemanticSearch
	promptBuilder  *prompts.PromptBuilder
}

// ContextBuilderConfig configures the context builder
type ContextBuilderConfig struct {
	DB             *database.DB
	SemanticSearch *embeddings.SemanticSearch
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(config ContextBuilderConfig) (*ContextBuilder, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database is required")
	}
	if config.SemanticSearch == nil {
		return nil, fmt.Errorf("semantic search is required")
	}

	return &ContextBuilder{
		db:             config.DB,
		semanticSearch: config.SemanticSearch,
		promptBuilder:  prompts.NewPromptBuilder(),
	}, nil
}

// BuildPageObjectPrompt builds a prompt for generating a page object
func (cb *ContextBuilder) BuildPageObjectPrompt(ctx context.Context, spec string, elements []map[string]string) (CompletionRequest, error) {
	// Search for similar page objects for context
	similarPages, err := cb.semanticSearch.SearchPageObjects(ctx, spec, 3)
	if err != nil {
		// Non-fatal: continue without examples
		similarPages = []embeddings.CodeSearchResult{}
	}

	// Add existing code style examples
	if len(similarPages) > 0 {
		for _, page := range similarPages {
			cb.promptBuilder.AddContext(map[string]interface{}{
				"type":      "pageObject",
				"className": page.Name,
				"code":      page.Content,
				"filePath":  page.FilePath,
			})
		}
	}

	// Build the prompt
	prompt := cb.promptBuilder.BuildPageObjectPrompt(spec, elements)

	return CompletionRequest{
		SystemPrompt: "You are an expert Senior QA Automation Engineer with 10+ years of experience.",
		UserPrompt:   prompt,
		Temperature:  0.3, // Lower temperature for more consistent code generation
		MaxTokens:    2000,
	}, nil
}

// BuildTestCasePrompt builds a prompt for generating a test case
func (cb *ContextBuilder) BuildTestCasePrompt(ctx context.Context, spec string) (CompletionRequest, error) {
	// Search for relevant page objects
	pageObjects, err := cb.semanticSearch.SearchCode(ctx, spec, 5)
	if err != nil {
		pageObjects = []embeddings.CodeSearchResult{}
	}

	// Search for similar tests for context
	similarTests, err := cb.semanticSearch.SearchTests(ctx, spec, 2)
	if err != nil {
		similarTests = []embeddings.CodeSearchResult{}
	}

	// Add similar tests to context
	for _, test := range similarTests {
		cb.promptBuilder.AddContext(map[string]interface{}{
			"type":      "test",
			"className": test.Name,
			"code":      test.Content,
			"filePath":  test.FilePath,
		})
	}

	// Convert page objects to format expected by prompt builder
	pageObjectsData := make([]map[string]interface{}, 0)
	for _, po := range pageObjects {
		if po.Type == "class" {
			// Get methods from database
			classRepo := database.NewClassRepository(cb.db)
			class, err := classRepo.GetByID(po.ID)
			if err != nil {
				continue
			}

			methodRepo := database.NewMethodRepository(cb.db)
			methods, err := methodRepo.GetByClassID(class.ID)
			if err != nil {
				continue
			}

			methodsList := make([]interface{}, 0)
			for _, m := range methods {
				methodsList = append(methodsList, map[string]interface{}{
					"name": m.Name,
				})
			}

			pageObjectsData = append(pageObjectsData, map[string]interface{}{
				"className": po.Name,
				"methods":   methodsList,
			})
		}
	}

	// Build the prompt
	prompt := cb.promptBuilder.BuildTestCasePrompt(spec, pageObjectsData)

	return CompletionRequest{
		SystemPrompt: "You are a Senior QA Automation Engineer with 10+ years of experience.",
		UserPrompt:   prompt,
		Temperature:  0.3,
		MaxTokens:    2500,
	}, nil
}

// BuildChatPrompt builds a prompt for chat/conversation
func (cb *ContextBuilder) BuildChatPrompt(ctx context.Context, userMessage string) (CompletionRequest, error) {
	// Search for relevant code based on the user's message
	relevantCode, err := cb.semanticSearch.SearchCode(ctx, userMessage, 5)
	if err != nil {
		relevantCode = []embeddings.CodeSearchResult{}
	}

	// Build context map
	contextMap := make(map[string]interface{})

	// Add relevant code to context
	if len(relevantCode) > 0 {
		relevantCodeList := make([]interface{}, 0)
		for _, code := range relevantCode {
			relevantCodeList = append(relevantCodeList, map[string]interface{}{
				"className": code.Name,
				"code":      code.Content,
				"filePath":  code.FilePath,
				"type":      code.Type,
			})
		}
		contextMap["relevantCode"] = relevantCodeList
	}

	// Get workspace analysis summary
	analysisMap := make(map[string]interface{})

	// Count page objects
	classRepo := database.NewClassRepository(cb.db)
	allClasses, err := classRepo.GetAll()
	if err == nil {
		pageObjectCount := 0
		for _, class := range allClasses {
			// Heuristic: classes ending with "Page" are page objects
			if len(class.Name) > 4 && class.Name[len(class.Name)-4:] == "Page" {
				pageObjectCount++
			}
		}
		analysisMap["pageObjects"] = make([]interface{}, pageObjectCount)
	}

	// Count test cases
	methodRepo := database.NewMethodRepository(cb.db)
	testMethods, err := methodRepo.GetTestMethods()
	if err == nil {
		analysisMap["testCases"] = make([]interface{}, len(testMethods))
	}

	contextMap["analysis"] = analysisMap

	// Build the prompt
	prompt := cb.promptBuilder.BuildChatPrompt(userMessage, contextMap)

	return CompletionRequest{
		SystemPrompt: "You are a Senior QA Automation Consultant and Architect with 15+ years of experience.",
		UserPrompt:   prompt,
		Temperature:  0.7, // Higher temperature for more creative responses
		MaxTokens:    2000,
	}, nil
}

// BuildFixPrompt builds a prompt for fixing broken code
func (cb *ContextBuilder) BuildFixPrompt(ctx context.Context, brokenCode string, errorMessage string) (CompletionRequest, error) {
	// Search for similar working code
	similarCode, err := cb.semanticSearch.FindSimilarCode(ctx, brokenCode, 3)
	if err != nil {
		similarCode = []embeddings.CodeSearchResult{}
	}

	// Build context map
	contextMap := make(map[string]interface{})

	if len(similarCode) > 0 {
		// Use the first similar code as working example
		contextMap["similarCode"] = similarCode[0].Content
	}

	// Build the prompt
	prompt := cb.promptBuilder.BuildFixPrompt(brokenCode, errorMessage, contextMap)

	return CompletionRequest{
		SystemPrompt: "You are a Senior QA Automation Engineer and Expert Debugger.",
		UserPrompt:   prompt,
		Temperature:  0.2, // Very low temperature for precise fixes
		MaxTokens:    2000,
	}, nil
}

// SetFramework sets the framework and test runner for the prompt builder
func (cb *ContextBuilder) SetFramework(framework, testRunner string) {
	cb.promptBuilder.SetFramework(framework, testRunner)
}

// SetCodingStyle sets coding style preferences
func (cb *ContextBuilder) SetCodingStyle(style map[string]interface{}) {
	cb.promptBuilder.SetCodingStyle(style)
}
