package ai

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// RequestAnalyzer parses user requests and extracts intent, entities, and task type
type RequestAnalyzer struct {
	db              *sql.DB
	embeddingClient *EmbeddingClient
}

// NewRequestAnalyzer creates a new request analyzer
func NewRequestAnalyzer(db *sql.DB, embeddingClient *EmbeddingClient) *RequestAnalyzer {
	return &RequestAnalyzer{
		db:              db,
		embeddingClient: embeddingClient,
	}
}

// Analyze parses the user request and returns a structured UserRequest
func (ra *RequestAnalyzer) Analyze(request string) (*UserRequest, error) {
	ur := &UserRequest{
		RawRequest: request,
	}

	// Step 1: Classify intent
	ur.Intent = ra.classifyIntent(request)

	// Step 2: Extract entities using pattern matching
	ur.Entities = ra.extractEntities(request)

	// Step 3: Determine task type
	ur.TaskType = ra.classifyTaskType(request, ur.Intent)

	// Step 4: Get framework from database
	framework, err := ra.getFrameworkConfig()
	if err == nil {
		ur.Framework = framework
	}

	// Step 5: Generate embedding for similarity search
	embedding, err := ra.embeddingClient.Embed(request)
	if err != nil {
		// Non-fatal error - we can still proceed without embeddings
		fmt.Printf("Warning: Failed to generate embedding: %v\n", err)
	} else {
		ur.Embedding = embedding
	}

	return ur, nil
}

// classifyIntent determines what the user wants to do
func (ra *RequestAnalyzer) classifyIntent(request string) string {
	requestLower := strings.ToLower(request)

	// Intent patterns (order matters - more specific first)
	intentPatterns := map[string][]string{
		"create_test": {
			`create.*test`,
			`add.*test.*case`,
			`write.*test.*for`,
			`test.*for.*verif`,
			`new.*test`,
			`generate.*test`,
		},
		"create_page_object": {
			`create.*page.*object`,
			`add.*page.*class`,
			`new.*page.*for`,
			`generate.*page`,
			`page.*object.*for`,
		},
		"create_method": {
			`add.*method`,
			`create.*function`,
			`implement.*action`,
			`new.*method`,
			`generate.*method`,
		},
		"refactor": {
			`refactor`,
			`improve`,
			`clean.*up`,
			`optimize`,
			`restructure`,
		},
		"add_assertion": {
			`add.*assert`,
			`verify.*that`,
			`check.*that`,
			`validate.*that`,
		},
		"add_wait": {
			`add.*wait`,
			`explicit.*wait`,
			`wait.*for.*element`,
			`wait.*until`,
		},
		"fix_bug": {
			`fix`,
			`debug`,
			`solve`,
			`correct`,
		},
	}

	for intent, patterns := range intentPatterns {
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, requestLower); matched {
				return intent
			}
		}
	}

	return "unknown"
}

// extractEntities extracts key entities from the request
func (ra *RequestAnalyzer) extractEntities(request string) []Entity {
	entities := []Entity{}
	requestLower := strings.ToLower(request)

	// Pattern 1: "for LoginPage" or "to LoginPage"
	pagePattern := regexp.MustCompile(`(?:for|to|in|on)\s+(\w*[Pp]age\w*)`)
	if matches := pagePattern.FindStringSubmatch(request); matches != nil {
		entities = append(entities, Entity{
			Type: "page",
			Name: matches[1],
		})
	}

	// Pattern 2: Action verbs with objects
	// "click submit button", "enter username", "verify error message"
	actionPattern := regexp.MustCompile(`(click|enter|type|select|verify|check|wait|navigate|scroll)\s+(?:the\s+)?(\w+(?:\s+\w+)?)`)
	if matches := actionPattern.FindStringSubmatch(requestLower); matches != nil {
		entities = append(entities, Entity{
			Action: matches[1],
			Name:   matches[2],
			Type:   "element",
		})
	}

	// Pattern 3: "test for login" or "test login functionality"
	testSubjectPattern := regexp.MustCompile(`test(?:\s+for)?\s+(\w+(?:\s+\w+)?)`)
	if matches := testSubjectPattern.FindStringSubmatch(requestLower); matches != nil {
		subject := matches[1]
		// Remove common stop words
		subject = strings.TrimSuffix(subject, " functionality")
		subject = strings.TrimSuffix(subject, " feature")

		entities = append(entities, Entity{
			Type: "test_subject",
			Name: subject,
		})
	}

	// Pattern 4: With/using specific elements
	// "with valid credentials", "using invalid password"
	conditionPattern := regexp.MustCompile(`(?:with|using)\s+(valid|invalid|empty|null)\s+(\w+)`)
	if matches := conditionPattern.FindStringSubmatch(requestLower); matches != nil {
		entities = append(entities, Entity{
			Type: "condition",
			Name: matches[1] + "_" + matches[2],
		})
	}

	// Pattern 5: Field names (camelCase or snake_case identifiers)
	fieldPattern := regexp.MustCompile(`\b([a-z][a-zA-Z0-9]*(?:Field|Element|Button|Link|Input))\b`)
	fieldMatches := fieldPattern.FindAllStringSubmatch(request, -1)
	for _, match := range fieldMatches {
		entities = append(entities, Entity{
			Type: "field",
			Name: match[1],
		})
	}

	return entities
}

// classifyTaskType determines what type of task this is
func (ra *RequestAnalyzer) classifyTaskType(request string, intent string) string {
	requestLower := strings.ToLower(request)

	// Check for API test keywords
	if strings.Contains(requestLower, "api") ||
		strings.Contains(requestLower, "rest") ||
		strings.Contains(requestLower, "endpoint") ||
		strings.Contains(requestLower, "request") ||
		strings.Contains(requestLower, "response") {
		return "api_test"
	}

	// Check for BDD keywords
	if strings.Contains(requestLower, "given") ||
		strings.Contains(requestLower, "when") ||
		strings.Contains(requestLower, "then") ||
		strings.Contains(requestLower, "scenario") ||
		strings.Contains(requestLower, "feature") {
		return "bdd_test"
	}

	// Check for page object keywords
	if intent == "create_page_object" ||
		strings.Contains(requestLower, "page object") ||
		strings.Contains(requestLower, "page class") {
		return "page_object"
	}

	// Check for utility/helper
	if strings.Contains(requestLower, "utility") ||
		strings.Contains(requestLower, "helper") ||
		strings.Contains(requestLower, "common") {
		return "utility"
	}

	// Default to functional test
	if intent == "create_test" {
		return "functional_test"
	}

	return "unknown"
}

// getFrameworkConfig retrieves framework configuration from database
func (ra *RequestAnalyzer) getFrameworkConfig() (string, error) {
	var testFramework, bddFramework, apiFramework string

	err := ra.db.QueryRow(`
		SELECT test_framework, bdd_framework, api_framework
		FROM framework_config
		ORDER BY last_detected DESC
		LIMIT 1
	`).Scan(&testFramework, &bddFramework, &apiFramework)

	if err != nil {
		return "", err
	}

	// Build framework string
	frameworks := []string{}
	if testFramework != "" {
		frameworks = append(frameworks, testFramework)
	}
	if bddFramework != "" {
		frameworks = append(frameworks, bddFramework)
	}
	if apiFramework != "" {
		frameworks = append(frameworks, apiFramework)
	}

	return strings.Join(frameworks, " + "), nil
}

// ExtractPageName attempts to extract a page name from the request
func (ra *RequestAnalyzer) ExtractPageName(request string) string {
	// Pattern: "LoginPage", "HomePage", etc.
	pagePattern := regexp.MustCompile(`(\w*[Pp]age\w*)`)
	if matches := pagePattern.FindStringSubmatch(request); matches != nil {
		return matches[1]
	}
	return ""
}

// ExtractMethodName attempts to extract a method name from the request
func (ra *RequestAnalyzer) ExtractMethodName(request string) string {
	// Pattern: camelCase method name
	methodPattern := regexp.MustCompile(`\b([a-z][a-zA-Z0-9]*)\s*\(`)
	if matches := methodPattern.FindStringSubmatch(request); matches != nil {
		return matches[1]
	}
	return ""
}

// HasKeyword checks if the request contains a specific keyword
func (ra *RequestAnalyzer) HasKeyword(request string, keyword string) bool {
	return strings.Contains(strings.ToLower(request), strings.ToLower(keyword))
}
