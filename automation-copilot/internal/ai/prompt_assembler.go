package ai

import (
	"database/sql"
	"fmt"
	"strings"
)

// PromptAssembler builds the complete prompt for the LLM
// Optimized for Claude's prompt caching (90% cost reduction!)
type PromptAssembler struct {
	db              *sql.DB
	frameworkConfig *FrameworkConfig
	codingPatterns  *CodingPatterns
	enableCaching   bool
}

// FrameworkConfig represents detected framework configuration
type FrameworkConfig struct {
	TestFramework       string
	BDDFramework        string
	APIFramework        string
	UsesPageObjectModel bool
	UsesPageFactory     bool
}

// CodingPatterns represents detected coding patterns
type CodingPatterns struct {
	TestMethodNaming   string
	PageObjectNaming   string
	WebElementNaming   string
	PreferredWaitType  string
	DefaultWaitTimeout int
	AssertionLibrary   string
	UsesAssertMessages bool
}

// NewPromptAssembler creates a new prompt assembler
func NewPromptAssembler(db *sql.DB, config Config) (*PromptAssembler, error) {
	// Load framework config
	frameworkConfig, err := loadFrameworkConfig(db)
	if err != nil {
		frameworkConfig = &FrameworkConfig{} // Use empty config
	}

	// Load coding patterns
	codingPatterns, err := loadCodingPatterns(db)
	if err != nil {
		codingPatterns = &CodingPatterns{} // Use empty config
	}

	return &PromptAssembler{
		db:              db,
		frameworkConfig: frameworkConfig,
		codingPatterns:  codingPatterns,
		enableCaching:   config.EnableCaching,
	}, nil
}

// AssemblePrompt builds the complete prompt with caching optimization
func (pa *PromptAssembler) AssemblePrompt(
	userRequest *UserRequest,
	retrievedContext []*SearchResult,
	examples []*Example,
	conversationHistory string,
) (*AssembledPrompt, error) {

	systemPrompt := pa.buildSystemPrompt(conversationHistory)
	userPrompt := pa.buildUserPrompt(userRequest, retrievedContext, examples)

	// Estimate tokens
	systemTokens := estimateTokenCount(systemPrompt)
	userTokens := estimateTokenCount(userPrompt)

	// Calculate cache breakdown
	cachedTokens := systemTokens // Entire system prompt is cached
	if conversationHistory != "" {
		cachedTokens += estimateTokenCount(conversationHistory)
	}

	return &AssembledPrompt{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		CacheBreakpoints: []string{
			"--- STATIC CACHE BREAKPOINT ---",
			"--- CONVERSATION CACHE BREAKPOINT ---",
		},
		TotalTokens:   systemTokens + userTokens,
		CachedTokens:  cachedTokens,
		DynamicTokens: userTokens,
	}, nil
}

// buildSystemPrompt builds the static system prompt (will be cached)
func (pa *PromptAssembler) buildSystemPrompt(conversationHistory string) string {
	var sb strings.Builder

	// ═══════════════════════════════════════════════════════════
	// CACHED SECTION 1: Static System Instructions
	// ═══════════════════════════════════════════════════════════

	sb.WriteString("You are an expert Selenium + Java test automation engineer.\n")
	sb.WriteString("Your task is to generate high-quality, production-ready test automation code ")
	sb.WriteString("that perfectly matches the existing codebase style and patterns.\n\n")

	// Framework Configuration
	sb.WriteString("## Project Framework Configuration\n\n")

	if pa.frameworkConfig.TestFramework != "" {
		sb.WriteString(fmt.Sprintf("- **Test Framework**: %s\n", pa.frameworkConfig.TestFramework))
	}
	if pa.frameworkConfig.BDDFramework != "" {
		sb.WriteString(fmt.Sprintf("- **BDD Framework**: %s\n", pa.frameworkConfig.BDDFramework))
	}
	if pa.frameworkConfig.APIFramework != "" {
		sb.WriteString(fmt.Sprintf("- **API Framework**: %s\n", pa.frameworkConfig.APIFramework))
	}

	sb.WriteString(fmt.Sprintf("- **Uses Page Object Model**: %v\n", pa.frameworkConfig.UsesPageObjectModel))
	sb.WriteString(fmt.Sprintf("- **Uses PageFactory**: %v\n\n", pa.frameworkConfig.UsesPageFactory))

	// Coding Patterns (MUST FOLLOW)
	sb.WriteString("## Project Coding Patterns (MUST FOLLOW)\n\n")

	if pa.codingPatterns.TestMethodNaming != "" {
		sb.WriteString(fmt.Sprintf("- **Test Method Naming**: %s\n", pa.codingPatterns.TestMethodNaming))
	}
	if pa.codingPatterns.PageObjectNaming != "" {
		sb.WriteString(fmt.Sprintf("- **Page Object Naming**: %s\n", pa.codingPatterns.PageObjectNaming))
	}
	if pa.codingPatterns.WebElementNaming != "" {
		sb.WriteString(fmt.Sprintf("- **WebElement Naming**: %s\n", pa.codingPatterns.WebElementNaming))
	}
	if pa.codingPatterns.PreferredWaitType != "" {
		sb.WriteString(fmt.Sprintf("- **Preferred Wait Type**: %s\n", pa.codingPatterns.PreferredWaitType))
	}
	if pa.codingPatterns.DefaultWaitTimeout > 0 {
		sb.WriteString(fmt.Sprintf("- **Default Wait Timeout**: %d seconds\n", pa.codingPatterns.DefaultWaitTimeout))
	}
	if pa.codingPatterns.AssertionLibrary != "" {
		sb.WriteString(fmt.Sprintf("- **Assertion Library**: %s\n", pa.codingPatterns.AssertionLibrary))
	}

	sb.WriteString(fmt.Sprintf("- **Uses Assert Messages**: %v\n\n", pa.codingPatterns.UsesAssertMessages))

	// Quality Requirements
	sb.WriteString("## Code Quality Requirements\n\n")
	sb.WriteString("1. **Style Consistency**: Match the EXACT coding style shown in examples\n")
	sb.WriteString("2. **Best Practices**: Use explicit waits, NEVER Thread.sleep()\n")
	sb.WriteString("3. **Maintainability**: Clear method names, proper JavaDoc comments\n")
	sb.WriteString("4. **Reliability**: Robust locators, proper error handling\n")
	sb.WriteString("5. **Completeness**: Include all imports, annotations, and setup code\n\n")

	// Self-RAG Instructions (Chain-of-Thought + Chain-of-Verification)
	sb.WriteString("## Generation Process (REQUIRED - Self-RAG)\n\n")
	sb.WriteString("Use this systematic approach to ensure high-quality code:\n\n")

	sb.WriteString("**Step 1: Plan** (Think before coding)\n")
	sb.WriteString("- What are the key components needed?\n")
	sb.WriteString("- Which patterns from the examples apply?\n")
	sb.WriteString("- What edge cases need handling?\n\n")

	sb.WriteString("**Step 2: Generate** (Write the code)\n")
	sb.WriteString("- Follow the EXACT patterns from examples\n")
	sb.WriteString("- Match the coding style precisely\n")
	sb.WriteString("- Include all necessary annotations and imports\n\n")

	sb.WriteString("**Step 3: Verify** (Check your work)\n")
	sb.WriteString("- Does it match the coding patterns?\n")
	sb.WriteString("- Are all WebElements properly defined with @FindBy?\n")
	sb.WriteString("- Are assertions following the project style?\n")
	sb.WriteString("- Are waits implemented correctly (explicit, not Thread.sleep)?\n")
	sb.WriteString("- Are there any anti-patterns?\n\n")

	sb.WriteString("**Step 4: Refine** (Fix any issues)\n")
	sb.WriteString("- Correct any style mismatches\n")
	sb.WriteString("- Remove any anti-patterns\n")
	sb.WriteString("- Ensure complete imports and annotations\n\n")

	if pa.enableCaching {
		sb.WriteString("--- STATIC CACHE BREAKPOINT ---\n\n")
	}

	// ═══════════════════════════════════════════════════════════
	// CACHED SECTION 2: Conversation History (if exists)
	// ═══════════════════════════════════════════════════════════

	if conversationHistory != "" {
		sb.WriteString(conversationHistory)
		sb.WriteString("\n")

		if pa.enableCaching {
			sb.WriteString("--- CONVERSATION CACHE BREAKPOINT ---\n\n")
		}
	}

	return sb.String()
}

// buildUserPrompt builds the dynamic user prompt (not cached)
func (pa *PromptAssembler) buildUserPrompt(
	userRequest *UserRequest,
	retrievedContext []*SearchResult,
	examples []*Example,
) string {
	var sb strings.Builder

	// Retrieved Context
	sb.WriteString("## Relevant Code from Codebase\n\n")

	if len(retrievedContext) == 0 {
		sb.WriteString("_No specific relevant code found. Generate based on best practices and examples below._\n\n")
	} else {
		for i, ctx := range retrievedContext {
			sb.WriteString(fmt.Sprintf("### Context %d: %s (Relevance: %.2f)\n\n",
				i+1, ctx.ChunkType, ctx.FinalScore))

			sb.WriteString("```java\n")
			sb.WriteString(ctx.Content)
			sb.WriteString("\n```\n\n")
		}
	}

	// Few-Shot Examples
	sb.WriteString("## Examples to Follow\n\n")

	// Positive examples
	positiveExamples := filterByType(examples, "positive")
	if len(positiveExamples) > 0 {
		sb.WriteString("### ✅ Good Examples (Follow These Patterns)\n\n")

		for i, ex := range positiveExamples {
			sb.WriteString(fmt.Sprintf("#### Example %d: %s", i+1, ex.Description))

			if ex.Similarity > 0 {
				sb.WriteString(fmt.Sprintf(" (Similarity: %.2f)", ex.Similarity))
			}
			sb.WriteString("\n\n")

			if ex.Explanation != "" {
				sb.WriteString(fmt.Sprintf("**Why this is good**: %s\n\n", ex.Explanation))
			}

			sb.WriteString("```java\n")
			sb.WriteString(ex.Code)
			sb.WriteString("\n```\n\n")
		}
	}

	// Negative examples (anti-patterns)
	negativeExamples := filterByType(examples, "negative")
	if len(negativeExamples) > 0 {
		sb.WriteString("### ❌ Anti-Patterns (AVOID These)\n\n")

		for i, ex := range negativeExamples {
			sb.WriteString(fmt.Sprintf("#### Anti-Pattern %d: %s\n\n", i+1, ex.Description))

			if ex.Explanation != "" {
				sb.WriteString(fmt.Sprintf("**Why to avoid**: %s\n\n", ex.Explanation))
			}

			sb.WriteString("```java\n")
			sb.WriteString(ex.Code)
			sb.WriteString("\n```\n\n")
		}
	}

	// User Request
	sb.WriteString("## Your Task\n\n")
	sb.WriteString(userRequest.RawRequest)
	sb.WriteString("\n\n")

	// Output Instructions
	sb.WriteString("## Output Format\n\n")
	sb.WriteString("Provide your response in the following format:\n\n")

	sb.WriteString("**Step 1: Plan**\n")
	sb.WriteString("[Your planning thoughts here]\n\n")

	sb.WriteString("**Step 2: Generate**\n")
	sb.WriteString("```java\n")
	sb.WriteString("// Your generated code here\n")
	sb.WriteString("```\n\n")

	sb.WriteString("**Step 3: Verify**\n")
	sb.WriteString("[Your verification checklist here]\n\n")

	sb.WriteString("**Final Code**\n")
	sb.WriteString("```java\n")
	sb.WriteString("// Final verified code here\n")
	sb.WriteString("```\n\n")

	return sb.String()
}

// loadFrameworkConfig loads framework configuration from database
func loadFrameworkConfig(db *sql.DB) (*FrameworkConfig, error) {
	config := &FrameworkConfig{}

	err := db.QueryRow(`
		SELECT test_framework, bdd_framework, api_framework,
		       uses_page_object_model, uses_page_factory
		FROM framework_config
		ORDER BY last_detected DESC
		LIMIT 1
	`).Scan(
		&config.TestFramework,
		&config.BDDFramework,
		&config.APIFramework,
		&config.UsesPageObjectModel,
		&config.UsesPageFactory,
	)

	return config, err
}

// loadCodingPatterns loads coding patterns from database
func loadCodingPatterns(db *sql.DB) (*CodingPatterns, error) {
	patterns := &CodingPatterns{}

	err := db.QueryRow(`
		SELECT test_method_naming, page_object_naming, web_element_naming,
		       preferred_wait_type, default_wait_timeout,
		       assertion_library, uses_assert_messages
		FROM coding_patterns
		ORDER BY last_analyzed DESC
		LIMIT 1
	`).Scan(
		&patterns.TestMethodNaming,
		&patterns.PageObjectNaming,
		&patterns.WebElementNaming,
		&patterns.PreferredWaitType,
		&patterns.DefaultWaitTimeout,
		&patterns.AssertionLibrary,
		&patterns.UsesAssertMessages,
	)

	return patterns, err
}

// filterByType filters examples by type (positive/negative)
func filterByType(examples []*Example, exampleType string) []*Example {
	filtered := []*Example{}
	for _, ex := range examples {
		if ex.Type == exampleType {
			filtered = append(filtered, ex)
		}
	}
	return filtered
}

// estimateTokenCount estimates the number of tokens in text
// Rough approximation: 1 token ≈ 4 characters for English/code
func estimateTokenCount(text string) int {
	return len(text) / 4
}
