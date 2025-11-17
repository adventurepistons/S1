package ai

import (
	"database/sql"
	"fmt"
)

// FewShotSelector dynamically selects best examples for in-context learning
// Research shows 7.3% F1-score improvement with dynamic selection vs random
type FewShotSelector struct {
	db              *sql.DB
	embeddingClient *EmbeddingClient
	positiveCount   int // Number of positive examples (default: 3)
	negativeCount   int // Number of negative examples (default: 1)
}

// NewFewShotSelector creates a new few-shot selector
func NewFewShotSelector(db *sql.DB, embeddingClient *EmbeddingClient, config Config) *FewShotSelector {
	return &FewShotSelector{
		db:              db,
		embeddingClient: embeddingClient,
		positiveCount:   config.PositiveExamples,
		negativeCount:   config.NegativeExamples,
	}
}

// SelectExamples selects the best examples for a user request
func (fs *FewShotSelector) SelectExamples(userRequest *UserRequest) ([]*Example, error) {
	allExamples := []*Example{}

	// Step 1: Select positive examples (similar to user request)
	positiveExamples, err := fs.selectPositiveExamples(userRequest)
	if err != nil {
		fmt.Printf("Warning: Failed to select positive examples: %v\n", err)
	} else {
		allExamples = append(allExamples, positiveExamples...)
	}

	// Step 2: Select negative examples (anti-patterns to avoid)
	negativeExamples, err := fs.selectNegativeExamples(userRequest)
	if err != nil {
		fmt.Printf("Warning: Failed to select negative examples: %v\n", err)
	} else {
		allExamples = append(allExamples, negativeExamples...)
	}

	return allExamples, nil
}

// selectPositiveExamples selects examples similar to the user request
func (fs *FewShotSelector) selectPositiveExamples(userRequest *UserRequest) ([]*Example, error) {
	if len(userRequest.Embedding) == 0 {
		// Fallback: select by task type without semantic similarity
		return fs.selectByTaskType(userRequest.TaskType, false, fs.positiveCount)
	}

	// Get all positive examples for this task type
	rows, err := fs.db.Query(`
		SELECT id, pattern_type, example_code, explanation, embedding
		FROM pattern_examples
		WHERE task_type = ? AND is_anti_pattern = 0
		ORDER BY frequency DESC
	`, userRequest.TaskType)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Calculate similarity for each example
	examples := []*Example{}

	for rows.Next() {
		var id int
		var patternType, code, explanation string
		var embeddingBytes []byte

		if err := rows.Scan(&id, &patternType, &code, &explanation, &embeddingBytes); err != nil {
			continue
		}

		// Convert embedding bytes to floats
		exampleEmbedding := bytesToFloats(embeddingBytes)

		if len(exampleEmbedding) == 0 {
			// No embedding - use default similarity
			examples = append(examples, &Example{
				ID:          id,
				Type:        "positive",
				Description: patternType,
				Code:        code,
				Explanation: explanation,
				Similarity:  0.5, // Neutral similarity
			})
			continue
		}

		// Calculate similarity
		similarity := cosineSimilarity(userRequest.Embedding, exampleEmbedding)

		examples = append(examples, &Example{
			ID:          id,
			Type:        "positive",
			Description: patternType,
			Code:        code,
			Explanation: explanation,
			Similarity:  similarity,
		})
	}

	// Sort by similarity (descending)
	sortExamplesBySimilarity(examples)

	// Return top-K
	if len(examples) > fs.positiveCount {
		examples = examples[:fs.positiveCount]
	}

	return examples, nil
}

// selectNegativeExamples selects anti-pattern examples to show what NOT to do
func (fs *FewShotSelector) selectNegativeExamples(userRequest *UserRequest) ([]*Example, error) {
	// Get common anti-patterns for this task type
	rows, err := fs.db.Query(`
		SELECT id, pattern_type, example_code, explanation
		FROM pattern_examples
		WHERE task_type = ? AND is_anti_pattern = 1
		ORDER BY frequency DESC
		LIMIT ?
	`, userRequest.TaskType, fs.negativeCount)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	examples := []*Example{}

	for rows.Next() {
		var id int
		var patternType, code, explanation string

		if err := rows.Scan(&id, &patternType, &code, &explanation); err != nil {
			continue
		}

		examples = append(examples, &Example{
			ID:          id,
			Type:        "negative",
			Description: patternType,
			Code:        code,
			Explanation: explanation,
		})
	}

	return examples, nil
}

// selectByTaskType selects examples by task type (fallback when no embeddings)
func (fs *FewShotSelector) selectByTaskType(taskType string, isAntiPattern bool, limit int) ([]*Example, error) {
	rows, err := fs.db.Query(`
		SELECT id, pattern_type, example_code, explanation
		FROM pattern_examples
		WHERE task_type = ? AND is_anti_pattern = ?
		ORDER BY frequency DESC
		LIMIT ?
	`, taskType, isAntiPattern, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	examples := []*Example{}
	exampleType := "positive"
	if isAntiPattern {
		exampleType = "negative"
	}

	for rows.Next() {
		var id int
		var patternType, code, explanation string

		if err := rows.Scan(&id, &patternType, &code, &explanation); err != nil {
			continue
		}

		examples = append(examples, &Example{
			ID:          id,
			Type:        exampleType,
			Description: patternType,
			Code:        code,
			Explanation: explanation,
		})
	}

	return examples, nil
}

// AddExample adds a new example to the database
func (fs *FewShotSelector) AddExample(taskType, patternType, code, explanation string, isAntiPattern bool) error {
	// Generate embedding for the example
	var embeddingBytes []byte

	embedding, err := fs.embeddingClient.Embed(code)
	if err == nil && len(embedding) > 0 {
		embeddingBytes = floatsToBytes(embedding)
	}

	_, err = fs.db.Exec(`
		INSERT INTO pattern_examples
		(task_type, pattern_type, example_code, explanation, is_anti_pattern, embedding, frequency)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`, taskType, patternType, code, explanation, isAntiPattern, embeddingBytes)

	return err
}

// ExtractExamplesFromCodebase extracts pattern examples from parsed code
func (fs *FewShotSelector) ExtractExamplesFromCodebase() error {
	fmt.Println("📚 Extracting pattern examples from codebase...")

	count := 0

	// Extract test method patterns
	testCount, err := fs.extractTestPatterns()
	if err != nil {
		fmt.Printf("  ⚠️  Warning: Failed to extract test patterns: %v\n", err)
	} else {
		count += testCount
	}

	// Extract page object patterns
	poCount, err := fs.extractPageObjectPatterns()
	if err != nil {
		fmt.Printf("  ⚠️  Warning: Failed to extract page object patterns: %v\n", err)
	} else {
		count += poCount
	}

	// Extract assertion patterns
	assertCount, err := fs.extractAssertionPatterns()
	if err != nil {
		fmt.Printf("  ⚠️  Warning: Failed to extract assertion patterns: %v\n", err)
	} else {
		count += assertCount
	}

	// Add common anti-patterns
	antiCount := fs.addCommonAntiPatterns()
	count += antiCount

	fmt.Printf("  ✅ Extracted %d pattern examples\n", count)
	return nil
}

// extractTestPatterns extracts good test method examples
func (fs *FewShotSelector) extractTestPatterns() (int, error) {
	// Get well-formed test methods
	rows, err := fs.db.Query(`
		SELECT m.id, m.method_name, m.body_source, c.class_name
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		WHERE m.is_test = 1
		AND m.uses_explicit_wait = 1
		AND EXISTS (SELECT 1 FROM assertions WHERE method_id = m.id)
		ORDER BY m.id
		LIMIT 10
	`)

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		var methodID int
		var methodName, bodySource, className string

		if err := rows.Scan(&methodID, &methodName, &bodySource, &className); err != nil {
			continue
		}

		// Build full method code
		code := fmt.Sprintf("@Test\npublic void %s() {\n%s\n}", methodName, bodySource)

		explanation := fmt.Sprintf("Well-formed test method from %s using explicit waits and assertions", className)

		if err := fs.AddExample("functional_test", "complete_test_method", code, explanation, false); err == nil {
			count++
		}
	}

	return count, nil
}

// extractPageObjectPatterns extracts page object examples
func (fs *FewShotSelector) extractPageObjectPatterns() (int, error) {
	// Get well-formed page objects
	rows, err := fs.db.Query(`
		SELECT c.id, c.class_name
		FROM classes c
		WHERE c.is_page_object = 1
		AND EXISTS (SELECT 1 FROM fields WHERE class_id = c.id AND is_web_element = 1)
		ORDER BY c.id
		LIMIT 5
	`)

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		var classID int
		var className string

		if err := rows.Scan(&classID, &className); err != nil {
			continue
		}

		// Get class code from chunks
		var code string
		err := fs.db.QueryRow(`
			SELECT content FROM chunks WHERE chunk_type = 'class' AND entity_id = ?
		`, classID).Scan(&code)

		if err != nil {
			continue
		}

		explanation := fmt.Sprintf("Complete page object class %s with @FindBy annotations and action methods", className)

		if err := fs.AddExample("page_object", "complete_page_object", code, explanation, false); err == nil {
			count++
		}
	}

	return count, nil
}

// extractAssertionPatterns extracts assertion examples
func (fs *FewShotSelector) extractAssertionPatterns() (int, error) {
	// Get methods with good assertion practices
	rows, err := fs.db.Query(`
		SELECT a.assertion_type, a.expected, a.actual, a.message
		FROM assertions a
		WHERE a.message IS NOT NULL AND a.message != ''
		LIMIT 10
	`)

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		var assertionType, expected, actual string
		var message sql.NullString

		if err := rows.Scan(&assertionType, &expected, &actual, &message); err != nil {
			continue
		}

		code := fmt.Sprintf("Assert.%s(%s, %s, \"%s\");",
			assertionType, actual, expected, message.String)

		explanation := fmt.Sprintf("Assertion with descriptive message explaining what is being verified")

		if err := fs.AddExample("functional_test", "assertion_with_message", code, explanation, false); err == nil {
			count++
		}
	}

	return count, nil
}

// addCommonAntiPatterns adds well-known anti-patterns
func (fs *FewShotSelector) addCommonAntiPatterns() int {
	antiPatterns := []struct {
		taskType    string
		patternType string
		code        string
		explanation string
	}{
		{
			taskType:    "functional_test",
			patternType: "thread_sleep",
			code:        "Thread.sleep(5000); // Wait for element\nloginButton.click();",
			explanation: "AVOID Thread.sleep()! Use explicit waits instead (WebDriverWait). Thread.sleep is brittle and slows down tests.",
		},
		{
			taskType:    "functional_test",
			patternType: "no_assertions",
			code:        "@Test\npublic void testLogin() {\n    loginPage.login(\"user\", \"pass\");\n    // No assertions!\n}",
			explanation: "AVOID tests without assertions! Always verify expected outcomes with Assert statements.",
		},
		{
			taskType:    "page_object",
			patternType: "assertions_in_page_object",
			code:        "public void clickLogin() {\n    loginButton.click();\n    Assert.assertTrue(driver.getCurrentUrl().contains(\"dashboard\"));\n}",
			explanation: "AVOID assertions in page objects! Page objects should only perform actions. Put assertions in test methods.",
		},
		{
			taskType:    "functional_test",
			patternType: "hardcoded_data",
			code:        "loginPage.enterUsername(\"admin@test.com\");\nloginPage.enterPassword(\"password123\");",
			explanation: "AVOID hardcoded test data! Use test data from properties files, CSV, or DataProviders for maintainability.",
		},
	}

	count := 0
	for _, ap := range antiPatterns {
		err := fs.db.Exec(`
			INSERT OR IGNORE INTO pattern_examples
			(task_type, pattern_type, example_code, explanation, is_anti_pattern, frequency)
			VALUES (?, ?, ?, ?, 1, 1)
		`, ap.taskType, ap.patternType, ap.code, ap.explanation).Error

		if err == nil {
			count++
		}
	}

	return count
}

// sortExamplesBySimilarity sorts examples by similarity (descending)
func sortExamplesBySimilarity(examples []*Example) {
	n := len(examples)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if examples[j].Similarity < examples[j+1].Similarity {
				examples[j], examples[j+1] = examples[j+1], examples[j]
			}
		}
	}
}

// GetStats returns statistics about pattern examples
func (fs *FewShotSelector) GetStats() (map[string]interface{}, error) {
	var totalExamples, positiveExamples, negativeExamples int

	err := fs.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN is_anti_pattern = 0 THEN 1 END) as positive,
			COUNT(CASE WHEN is_anti_pattern = 1 THEN 1 END) as negative
		FROM pattern_examples
	`).Scan(&totalExamples, &positiveExamples, &negativeExamples)

	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_examples":    totalExamples,
		"positive_examples": positiveExamples,
		"negative_examples": negativeExamples,
		"positive_count":    fs.positiveCount,
		"negative_count":    fs.negativeCount,
	}

	return stats, nil
}
