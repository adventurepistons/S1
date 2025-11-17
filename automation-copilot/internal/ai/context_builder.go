package ai

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ContextBuilder is the main orchestrator for AI-powered code generation
// Combines all components: retrieval, enrichment, examples, conversation, prompting
type ContextBuilder struct {
	db                  *sql.DB
	config              Config
	requestAnalyzer     *RequestAnalyzer
	hybridRetriever     *HybridRetriever
	contextEnricher     *ContextEnricher
	fewShotSelector     *FewShotSelector
	conversationManager *ConversationManager
	promptAssembler     *PromptAssembler
	embeddingClient     *EmbeddingClient
	llmClient           LLMClient      // LLM client (Claude/OpenAI)
	responseParser      *ResponseParser // Parses LLM responses
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(db *sql.DB, config Config) (*ContextBuilder, error) {
	// Initialize embedding client
	embeddingClient := NewEmbeddingClient(config.EmbeddingServiceURL)

	// Check if embedding service is available
	if err := embeddingClient.HealthCheck(); err != nil {
		return nil, fmt.Errorf("embedding service not available: %w\nPlease start it with: make start-embeddings", err)
	}

	// Initialize prompt assembler
	promptAssembler, err := NewPromptAssembler(db, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create prompt assembler: %w", err)
	}

	// Initialize LLM client (optional - only if API key provided)
	var llmClient LLMClient
	if config.LLMConfig.APIKey != "" {
		llmClient, err = NewLLMClient(config.LLMConfig)
		if err != nil {
			// Don't fail - just warn that LLM integration won't work
			fmt.Printf("⚠️  Warning: Failed to initialize LLM client: %v\n", err)
			fmt.Println("   Context building will work, but code generation requires LLM API key")
		}
	}

	return &ContextBuilder{
		db:                  db,
		config:              config,
		requestAnalyzer:     NewRequestAnalyzer(db, embeddingClient),
		hybridRetriever:     NewHybridRetriever(db, embeddingClient, config),
		contextEnricher:     NewContextEnricher(db),
		fewShotSelector:     NewFewShotSelector(db, embeddingClient, config),
		conversationManager: NewConversationManager(db, config),
		promptAssembler:     promptAssembler,
		embeddingClient:     embeddingClient,
		llmClient:           llmClient,
		responseParser:      NewResponseParser(),
	}, nil
}

// BuildContext builds complete context for LLM code generation
// This is the main entry point - orchestrates all components
func (cb *ContextBuilder) BuildContext(userRequest string) (*AssembledPrompt, *ContextBuilderMetrics, error) {
	metrics := &ContextBuilderMetrics{}
	startTime := time.Now()

	fmt.Println("🤖 Building context for code generation...")
	fmt.Printf("   Request: %s\n", userRequest)

	// Step 1: Analyze user request
	fmt.Println("   [1/6] Analyzing request...")
	analyzedRequest, err := cb.requestAnalyzer.Analyze(userRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze request: %w", err)
	}

	fmt.Printf("   ✓ Intent: %s, Task: %s\n", analyzedRequest.Intent, analyzedRequest.TaskType)

	// Step 2: Retrieve relevant context using hybrid search
	fmt.Println("   [2/6] Retrieving relevant code...")
	retrievalStart := time.Now()

	retrievedChunks, err := cb.hybridRetriever.Retrieve(userRequest, cb.config.TopK)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve context: %w", err)
	}

	metrics.RetrievalLatency = time.Since(retrievalStart)
	metrics.HybridResultCount = len(retrievedChunks)

	if len(retrievedChunks) > 0 {
		avgScore := 0.0
		for _, chunk := range retrievedChunks {
			avgScore += chunk.FinalScore
		}
		metrics.AverageRelevanceScore = avgScore / float64(len(retrievedChunks))
	}

	fmt.Printf("   ✓ Retrieved %d relevant chunks (avg score: %.2f)\n",
		len(retrievedChunks), metrics.AverageRelevanceScore)

	// Step 3: Select few-shot examples dynamically
	fmt.Println("   [3/6] Selecting examples...")
	exampleStart := time.Now()

	examples, err := cb.fewShotSelector.SelectExamples(analyzedRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Warning: Failed to select examples: %v\n", err)
		examples = []*Example{} // Continue without examples
	}

	metrics.ExampleSelectionTime = time.Since(exampleStart)

	positiveCount := 0
	negativeCount := 0
	for _, ex := range examples {
		if ex.Type == "positive" {
			positiveCount++
		} else {
			negativeCount++
		}
	}

	fmt.Printf("   ✓ Selected %d examples (%d positive, %d negative)\n",
		len(examples), positiveCount, negativeCount)

	// Step 4: Get conversation history
	fmt.Println("   [4/6] Loading conversation history...")

	conversationHistory, err := cb.conversationManager.BuildContextString()
	if err != nil {
		fmt.Printf("   ⚠️  Warning: Failed to load conversation: %v\n", err)
		conversationHistory = "" // Continue without history
	}

	messageCount, _ := cb.conversationManager.GetMessageCount()
	fmt.Printf("   ✓ Loaded %d previous messages\n", messageCount)

	// Step 5: Assemble final prompt
	fmt.Println("   [5/6] Assembling prompt...")

	prompt, err := cb.promptAssembler.AssemblePrompt(
		analyzedRequest,
		retrievedChunks,
		examples,
		conversationHistory,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to assemble prompt: %w", err)
	}

	metrics.TotalTokens = prompt.TotalTokens
	metrics.CachedTokens = prompt.CachedTokens
	metrics.DynamicTokens = prompt.DynamicTokens

	// Estimate cost (Claude Sonnet 4.5 pricing)
	cacheCost := float64(prompt.CachedTokens) * 0.30 / 1_000_000      // $0.30 per 1M tokens
	dynamicCost := float64(prompt.DynamicTokens) * 3.00 / 1_000_000   // $3.00 per 1M tokens
	metrics.EstimatedCost = cacheCost + dynamicCost

	fmt.Printf("   ✓ Prompt assembled: %d tokens (%d cached, %d dynamic)\n",
		prompt.TotalTokens, prompt.CachedTokens, prompt.DynamicTokens)
	fmt.Printf("   ✓ Estimated cost: $%.4f\n", metrics.EstimatedCost)

	// Step 6: Save user message to conversation
	fmt.Println("   [6/6] Saving to conversation...")

	if err := cb.conversationManager.AddUserMessage(userRequest); err != nil {
		fmt.Printf("   ⚠️  Warning: Failed to save message: %v\n", err)
	}

	totalTime := time.Since(startTime)
	fmt.Printf("   ✅ Context built in %v\n", totalTime)

	return prompt, metrics, nil
}

// SaveResponse saves the LLM's response to conversation history
func (cb *ContextBuilder) SaveResponse(response string, generatedCode *GeneratedCode) error {
	return cb.conversationManager.AddAssistantMessage(response, generatedCode)
}

// GetConversationManager returns the conversation manager
func (cb *ContextBuilder) GetConversationManager() *ConversationManager {
	return cb.conversationManager
}

// GenerateCode is the COMPLETE end-to-end code generation method
// Builds context, calls LLM, parses response, saves to conversation
func (cb *ContextBuilder) GenerateCode(ctx context.Context, userRequest string) (*CodeGenerationResult, error) {
	result := &CodeGenerationResult{}
	startTime := time.Now()

	// Check if LLM client is available
	if cb.llmClient == nil {
		return nil, fmt.Errorf("LLM client not initialized - please set ANTHROPIC_API_KEY environment variable")
	}

	// Step 1: Build context (uses all RAG components)
	fmt.Println("")
	prompt, metrics, err := cb.BuildContext(userRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	result.ContextMetrics = metrics

	// Step 2: Call LLM
	fmt.Println("")
	fmt.Println("🤖 Calling Claude API...")
	fmt.Printf("   Model: %s\n", cb.config.LLMConfig.Model)
	fmt.Printf("   Max tokens: %d\n", cb.config.LLMConfig.MaxTokens)

	llmStart := time.Now()

	llmRequest := &GenerateRequest{
		SystemPrompt:     prompt.SystemPrompt,
		UserPrompt:       prompt.UserPrompt,
		CacheBreakpoints: prompt.CacheBreakpoints,
		MaxTokens:        cb.config.LLMConfig.MaxTokens,
		Temperature:      cb.config.LLMConfig.Temperature,
		Model:            cb.config.LLMConfig.Model,
	}

	llmResponse, err := cb.llmClient.Generate(ctx, llmRequest)
	if err != nil {
		return nil, fmt.Errorf("LLM API call failed: %w", err)
	}

	fmt.Printf("   ✓ Response received in %v\n", llmResponse.Latency)
	fmt.Printf("   ✓ Tokens: %d input (%d cached), %d output\n",
		llmResponse.InputTokens, llmResponse.CacheReadTokens, llmResponse.OutputTokens)
	fmt.Printf("   ✓ Actual cost: $%.4f\n", llmResponse.Cost)

	result.LLMResponse = llmResponse

	// Step 3: Parse response
	fmt.Println("")
	fmt.Println("📝 Parsing response...")

	parsed, err := cb.responseParser.ParseResponse(llmResponse.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("   ✓ Extracted code: %s (%s)\n", parsed.GeneratedCode.ClassName, parsed.GeneratedCode.CodeType)
	fmt.Printf("   ✓ Suggested path: %s\n", parsed.GeneratedCode.SuggestedPath)

	result.ParsedResponse = parsed

	// Step 4: Validate code
	fmt.Println("")
	fmt.Println("✅ Validating code...")

	validationErrors := cb.responseParser.ValidateCode(parsed.GeneratedCode)
	result.ValidationErrors = validationErrors

	if len(validationErrors) > 0 {
		for _, verr := range validationErrors {
			icon := "⚠️"
			if verr.Severity == "error" {
				icon = "❌"
			}
			fmt.Printf("   %s %s: %s", icon, verr.Severity, verr.Message)
			if verr.Line > 0 {
				fmt.Printf(" (line %d)", verr.Line)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("   ✓ No validation issues found")
	}

	// Step 5: Save to conversation
	if err := cb.SaveResponse(llmResponse.Content, parsed.GeneratedCode); err != nil {
		fmt.Printf("   ⚠️  Warning: Failed to save to conversation: %v\n", err)
	}

	result.TotalTime = time.Since(startTime)

	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("✅ Code generation complete in %v\n", result.TotalTime)
	fmt.Println("════════════════════════════════════════════════════════")

	return result, nil
}

// GenerateCodePromptOnly builds context but doesn't call LLM
// Useful for debugging or manual LLM use
func (cb *ContextBuilder) GenerateCodePromptOnly(userRequest string) (*AssembledPrompt, *ContextBuilderMetrics, error) {
	return cb.BuildContext(userRequest)
}

// BuildIndexes builds all indexes (chunks, BM25, embeddings, examples)
func (cb *ContextBuilder) BuildIndexes() error {
	fmt.Println("🔨 Building indexes for context retrieval...")
	fmt.Println("")

	// Step 1: Create chunks
	fmt.Println("Step 1/5: Creating code chunks...")
	chunker := NewChunker(cb.db, cb.config)

	if err := chunker.ChunkProject(); err != nil {
		return fmt.Errorf("failed to create chunks: %w", err)
	}

	chunkCount, _ := chunker.GetChunkCount()
	fmt.Printf("✅ Created %d chunks\n\n", chunkCount)

	// Step 2: Build BM25 index
	fmt.Println("Step 2/5: Building BM25 index...")
	bm25Index := NewBM25Index(cb.db, cb.config)

	if err := bm25Index.BuildIndex(); err != nil {
		return fmt.Errorf("failed to build BM25 index: %w", err)
	}

	bm25Stats, _ := bm25Index.GetStats()
	fmt.Printf("✅ BM25 index built: %d documents, %d terms\n\n",
		bm25Stats["total_documents"], bm25Stats["total_terms"])

	// Step 3: Build semantic index
	fmt.Println("Step 3/5: Building semantic index (generating embeddings)...")
	fmt.Println("   This may take a few minutes...")

	semanticIndex := NewSemanticIndex(cb.db, cb.embeddingClient, cb.config.EmbeddingDimension)

	if err := semanticIndex.BuildIndex(); err != nil {
		return fmt.Errorf("failed to build semantic index: %w", err)
	}

	semanticStats, _ := semanticIndex.GetStats()
	fmt.Printf("✅ Semantic index built: %d/%d chunks have embeddings (%.1f%% coverage)\n\n",
		semanticStats["with_embeddings"], semanticStats["total_chunks"], semanticStats["coverage"])

	// Step 4: Enrich chunks with context
	fmt.Println("Step 4/5: Enriching chunks with contextual information...")

	if err := cb.contextEnricher.EnrichAllChunks(); err != nil {
		return fmt.Errorf("failed to enrich chunks: %w", err)
	}

	fmt.Println("")

	// Step 5: Extract pattern examples
	fmt.Println("Step 5/5: Extracting pattern examples for few-shot learning...")

	if err := cb.fewShotSelector.ExtractExamplesFromCodebase(); err != nil {
		return fmt.Errorf("failed to extract examples: %w", err)
	}

	exampleStats, _ := cb.fewShotSelector.GetStats()
	fmt.Printf("✅ Pattern examples: %d total (%d positive, %d negative)\n\n",
		exampleStats["total_examples"], exampleStats["positive_examples"], exampleStats["negative_examples"])

	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("✅ All indexes built successfully!")
	fmt.Println("════════════════════════════════════════════════════════")

	return nil
}

// GetStats returns comprehensive statistics about the context builder
func (cb *ContextBuilder) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Hybrid retrieval stats
	hybridStats, err := cb.hybridRetriever.GetStats()
	if err == nil {
		stats["hybrid_retrieval"] = hybridStats
	}

	// Few-shot selector stats
	exampleStats, err := cb.fewShotSelector.GetStats()
	if err == nil {
		stats["examples"] = exampleStats
	}

	// Conversation stats
	messageCount, err := cb.conversationManager.GetMessageCount()
	if err == nil {
		stats["conversation_messages"] = messageCount
	}

	stats["session_id"] = cb.conversationManager.GetSessionID()
	stats["config"] = cb.config

	return stats, nil
}

// ClearIndexes removes all indexes (useful for re-indexing)
func (cb *ContextBuilder) ClearIndexes() error {
	fmt.Println("🗑️  Clearing all indexes...")

	chunker := NewChunker(cb.db, cb.config)
	if err := chunker.ClearChunks(); err != nil {
		return err
	}

	bm25Index := NewBM25Index(cb.db, cb.config)
	if err := bm25Index.ClearIndex(); err != nil {
		return err
	}

	semanticIndex := NewSemanticIndex(cb.db, cb.embeddingClient, cb.config.EmbeddingDimension)
	if err := semanticIndex.ClearIndex(); err != nil {
		return err
	}

	if err := cb.contextEnricher.ClearEnrichedContent(); err != nil {
		return err
	}

	fmt.Println("✅ All indexes cleared")
	return nil
}
