package detector

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/adventurepistons/automation-copilot/pkg/models"
)

// PatternDetector analyzes code to learn project coding patterns
type PatternDetector struct {
	projectRoot string
	db          *sql.DB
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(projectRoot string, db *sql.DB) *PatternDetector {
	return &PatternDetector{
		projectRoot: projectRoot,
		db:          db,
	}
}

// DetectPatterns analyzes all code to learn project coding style
func (pd *PatternDetector) DetectPatterns() (*models.CodingPatterns, error) {
	patterns := &models.CodingPatterns{}

	// 1. Analyze test method naming
	patterns.TestMethodNaming = pd.detectTestNaming()

	// 2. Analyze page object naming
	patterns.PageObjectNaming = pd.detectPageObjectNaming()

	// 3. Analyze WebElement field naming
	patterns.WebElementNaming = pd.detectWebElementNaming()

	// 4. Analyze method naming in page objects
	patterns.MethodNaming = pd.detectMethodNaming()

	// 5. Detect test structure (AAA vs Given-When-Then)
	patterns.UsesAAAPattern = pd.detectAAAPattern()
	patterns.UsesGivenWhenThen = pd.detectGivenWhenThen()

	// 6. Detect wait strategies
	patterns.PreferredWaitType = pd.detectPreferredWait()
	patterns.DefaultWaitTimeout = pd.detectDefaultTimeout()

	// 7. Detect assertion style
	patterns.AssertionLibrary = pd.detectAssertionLibrary()
	patterns.UsesAssertMessages = pd.detectAssertMessages()

	// 8. Detect data patterns
	patterns.UsesDataProviders = pd.detectDataProviders()
	patterns.UsesCSVFiles = pd.detectCSVFiles()
	patterns.UsesExcelFiles = pd.detectExcelFiles()

	// 9. Extract pattern examples
	if err := pd.extractPatternExamples(); err != nil {
		fmt.Printf("Warning: Failed to extract pattern examples: %v\n", err)
	}

	// 10. Save to database
	if err := pd.saveCodingPatterns(patterns); err != nil {
		return nil, fmt.Errorf("failed to save coding patterns: %w", err)
	}

	return patterns, nil
}

// detectTestNaming analyzes test method names to find pattern
func (pd *PatternDetector) detectTestNaming() string {
	rows, err := pd.db.Query(`
		SELECT method_name FROM methods
		WHERE is_test = 1
		LIMIT 50
	`)
	if err != nil {
		return "unknown"
	}
	defer rows.Close()

	var methodNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		methodNames = append(methodNames, name)
	}

	if len(methodNames) == 0 {
		return "unknown"
	}

	// Count naming patterns
	patterns := map[string]int{
		"testActionWithCondition": 0,
		"givenWhenThen":            0,
		"shouldDoSomething":        0,
		"camelCase":                0,
		"snake_case":               0,
	}

	for _, name := range methodNames {
		// Test if it starts with "test"
		if strings.HasPrefix(name, "test") {
			patterns["testActionWithCondition"]++
		}

		// Test for Given-When-Then
		if strings.Contains(name, "given") || strings.Contains(name, "when") || strings.Contains(name, "then") {
			patterns["givenWhenThen"]++
		}

		// Test for "should"
		if strings.HasPrefix(name, "should") {
			patterns["shouldDoSomething"]++
		}

		// Test for snake_case
		if strings.Contains(name, "_") {
			patterns["snake_case"]++
		} else {
			patterns["camelCase"]++
		}
	}

	// Return most common pattern
	maxCount := 0
	maxPattern := "camelCase"
	for pattern, count := range patterns {
		if count > maxCount {
			maxCount = count
			maxPattern = pattern
		}
	}

	return maxPattern
}

// detectPageObjectNaming analyzes page object class names
func (pd *PatternDetector) detectPageObjectNaming() string {
	rows, err := pd.db.Query(`
		SELECT class_name FROM classes
		WHERE is_page_object = 1
		LIMIT 20
	`)
	if err != nil {
		return "unknown"
	}
	defer rows.Close()

	var classNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		classNames = append(classNames, name)
	}

	if len(classNames) == 0 {
		return "unknown"
	}

	// Check for common suffixes
	patterns := map[string]int{
		"Page":         0,
		"PageObject":   0,
		"PO":           0,
		"Screen":       0,
		"NoSuffix":     0,
	}

	for _, name := range classNames {
		switch {
		case strings.HasSuffix(name, "Page"):
			patterns["Page"]++
		case strings.HasSuffix(name, "PageObject"):
			patterns["PageObject"]++
		case strings.HasSuffix(name, "PO"):
			patterns["PO"]++
		case strings.HasSuffix(name, "Screen"):
			patterns["Screen"]++
		default:
			patterns["NoSuffix"]++
		}
	}

	// Return most common pattern
	maxCount := 0
	maxPattern := "Page"
	for pattern, count := range patterns {
		if count > maxCount {
			maxCount = count
			maxPattern = pattern
		}
	}

	if maxPattern == "NoSuffix" {
		return "no suffix"
	}
	return maxPattern
}

// detectWebElementNaming analyzes WebElement field names
func (pd *PatternDetector) detectWebElementNaming() string {
	rows, err := pd.db.Query(`
		SELECT field_name FROM fields
		WHERE is_web_element = 1
		LIMIT 50
	`)
	if err != nil {
		return "unknown"
	}
	defer rows.Close()

	var fieldNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		fieldNames = append(fieldNames, name)
	}

	if len(fieldNames) == 0 {
		return "unknown"
	}

	// Check for common suffixes and patterns
	patterns := map[string]int{
		"Field":      0,
		"Element":    0,
		"Button":     0,
		"Input":      0,
		"NoSuffix":   0,
		"snake_case": 0,
		"camelCase":  0,
	}

	for _, name := range fieldNames {
		// Check suffixes
		switch {
		case strings.HasSuffix(name, "Field"):
			patterns["Field"]++
		case strings.HasSuffix(name, "Element"):
			patterns["Element"]++
		case strings.HasSuffix(name, "Button"):
			patterns["Button"]++
		case strings.HasSuffix(name, "Input"):
			patterns["Input"]++
		default:
			patterns["NoSuffix"]++
		}

		// Check case style
		if strings.Contains(name, "_") {
			patterns["snake_case"]++
		} else {
			patterns["camelCase"]++
		}
	}

	// Build pattern description
	var parts []string

	// Determine case style
	if patterns["camelCase"] > patterns["snake_case"] {
		parts = append(parts, "camelCase")
	} else {
		parts = append(parts, "snake_case")
	}

	// Determine suffix preference
	maxSuffixCount := 0
	maxSuffix := ""
	for suffix := range patterns {
		if suffix == "snake_case" || suffix == "camelCase" {
			continue
		}
		if patterns[suffix] > maxSuffixCount {
			maxSuffixCount = patterns[suffix]
			maxSuffix = suffix
		}
	}

	if maxSuffix != "" && maxSuffix != "NoSuffix" {
		parts = append(parts, fmt.Sprintf("with '%s' suffix", maxSuffix))
	}

	return strings.Join(parts, " ")
}

// detectMethodNaming analyzes page object method names
func (pd *PatternDetector) detectMethodNaming() string {
	rows, err := pd.db.Query(`
		SELECT m.method_name
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		WHERE c.is_page_object = 1 AND m.is_test = 0
		LIMIT 50
	`)
	if err != nil {
		return "unknown"
	}
	defer rows.Close()

	var methodNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		methodNames = append(methodNames, name)
	}

	if len(methodNames) == 0 {
		return "unknown"
	}

	// Check for common verb prefixes
	patterns := map[string]int{
		"click":   0,
		"enter":   0,
		"set":     0,
		"type":    0,
		"get":     0,
		"is":      0,
		"verify":  0,
	}

	for _, name := range methodNames {
		nameLower := strings.ToLower(name)
		for prefix := range patterns {
			if strings.HasPrefix(nameLower, prefix) {
				patterns[prefix]++
			}
		}
	}

	// Find most common prefixes
	var commonPrefixes []string
	for prefix, count := range patterns {
		if count > 0 {
			commonPrefixes = append(commonPrefixes, prefix)
		}
	}

	if len(commonPrefixes) > 0 {
		return fmt.Sprintf("verb-based (%s, %s, etc.)", commonPrefixes[0], commonPrefixes[min(1, len(commonPrefixes)-1)])
	}

	return "camelCase"
}

// detectAAAPattern detects if Arrange-Act-Assert pattern is used
func (pd *PatternDetector) detectAAAPattern() bool {
	// Check for comments in test methods containing AAA keywords
	var count int
	err := pd.db.QueryRow(`
		SELECT COUNT(*) FROM methods
		WHERE is_test = 1 AND (
			body_source LIKE '%// Arrange%' OR
			body_source LIKE '%// Act%' OR
			body_source LIKE '%// Assert%'
		)
	`).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// detectGivenWhenThen detects if Given-When-Then pattern is used
func (pd *PatternDetector) detectGivenWhenThen() bool {
	// Check for Cucumber annotations or comments
	var count int
	err := pd.db.QueryRow(`
		SELECT COUNT(*) FROM method_annotations
		WHERE annotation_type IN ('Given', 'When', 'Then', 'And', 'But')
	`).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// detectPreferredWait detects preferred wait strategy
func (pd *PatternDetector) detectPreferredWait() string {
	var explicitCount, implicitCount int

	pd.db.QueryRow("SELECT COUNT(*) FROM methods WHERE uses_explicit_wait = 1").Scan(&explicitCount)
	pd.db.QueryRow("SELECT COUNT(*) FROM methods WHERE uses_implicit_wait = 1").Scan(&implicitCount)

	if explicitCount > implicitCount {
		return "explicit"
	} else if implicitCount > 0 {
		return "implicit"
	}

	return "none"
}

// detectDefaultTimeout detects most common wait timeout
func (pd *PatternDetector) detectDefaultTimeout() int {
	rows, err := pd.db.Query(`
		SELECT wait_timeout, COUNT(*) as count
		FROM methods
		WHERE wait_timeout > 0
		GROUP BY wait_timeout
		ORDER BY count DESC
		LIMIT 1
	`)
	if err != nil {
		return 10
	}
	defer rows.Close()

	if rows.Next() {
		var timeout, count int
		if err := rows.Scan(&timeout, &count); err == nil {
			return timeout
		}
	}

	return 10 // Default
}

// detectAssertionLibrary detects which assertion library is used
func (pd *PatternDetector) detectAssertionLibrary() string {
	rows, err := pd.db.Query(`
		SELECT assertion_type FROM assertions
		LIMIT 100
	`)
	if err != nil {
		return "unknown"
	}
	defer rows.Close()

	libraries := map[string]int{
		"TestNG":   0,
		"JUnit":    0,
		"AssertJ":  0,
		"Hamcrest": 0,
	}

	for rows.Next() {
		var assertType string
		if err := rows.Scan(&assertType); err != nil {
			continue
		}

		// Classify assertion
		assertLower := strings.ToLower(assertType)
		switch {
		case strings.Contains(assertLower, "assert."):
			libraries["TestNG"]++
		case strings.Contains(assertLower, "assertions."):
			libraries["JUnit"]++
		case strings.Contains(assertLower, "assertthat"):
			if strings.Contains(assertLower, "hamcrest") {
				libraries["Hamcrest"]++
			} else {
				libraries["AssertJ"]++
			}
		}
	}

	// Return most common
	maxCount := 0
	maxLibrary := "TestNG"
	for lib, count := range libraries {
		if count > maxCount {
			maxCount = count
			maxLibrary = lib
		}
	}

	return maxLibrary
}

// detectAssertMessages detects if assertions have custom messages
func (pd *PatternDetector) detectAssertMessages() bool {
	var countWithMessage, countTotal int

	pd.db.QueryRow("SELECT COUNT(*) FROM assertions WHERE message != ''").Scan(&countWithMessage)
	pd.db.QueryRow("SELECT COUNT(*) FROM assertions").Scan(&countTotal)

	if countTotal == 0 {
		return false
	}

	// If more than 50% have messages, consider it a pattern
	return float64(countWithMessage)/float64(countTotal) > 0.5
}

// detectDataProviders detects if data providers are used
func (pd *PatternDetector) detectDataProviders() bool {
	var count int
	err := pd.db.QueryRow(`
		SELECT COUNT(*) FROM method_annotations
		WHERE annotation_type IN ('DataProvider', 'ParameterizedTest')
	`).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// detectCSVFiles detects if CSV files are used for test data
func (pd *PatternDetector) detectCSVFiles() bool {
	var count int
	err := pd.db.QueryRow(`
		SELECT COUNT(*) FROM method_annotations
		WHERE annotation_type = 'CsvSource' OR annotation_type = 'CsvFileSource'
	`).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// detectExcelFiles detects if Excel files are used for test data
func (pd *PatternDetector) detectExcelFiles() bool {
	var count int
	err := pd.db.QueryRow(`
		SELECT COUNT(*) FROM imports
		WHERE import_path LIKE '%apache.poi%'
	`).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// extractPatternExamples extracts real code examples to use as templates
func (pd *PatternDetector) extractPatternExamples() error {
	// Extract test method examples
	if err := pd.extractTestMethodExamples(); err != nil {
		return err
	}

	// Extract assertion examples
	if err := pd.extractAssertionExamples(); err != nil {
		return err
	}

	// Extract wait examples
	if err := pd.extractWaitExamples(); err != nil {
		return err
	}

	return nil
}

// extractTestMethodExamples extracts example test methods
func (pd *PatternDetector) extractTestMethodExamples() error {
	rows, err := pd.db.Query(`
		SELECT body_source, line_start, line_end
		FROM methods m
		JOIN classes c ON m.class_id = c.id
		JOIN files f ON c.file_id = f.id
		WHERE m.is_test = 1
		ORDER BY RANDOM()
		LIMIT 5
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := pd.db.Prepare(`
		INSERT INTO pattern_examples (pattern_type, example_code, line_start, line_end, frequency)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT DO UPDATE SET frequency = frequency + 1
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var code string
		var lineStart, lineEnd int
		if err := rows.Scan(&code, &lineStart, &lineEnd); err != nil {
			continue
		}

		stmt.Exec("test_method", code, lineStart, lineEnd)
	}

	return nil
}

// extractAssertionExamples extracts assertion code patterns
func (pd *PatternDetector) extractAssertionExamples() error {
	rows, err := pd.db.Query(`
		SELECT assertion_type, expected, actual, message
		FROM assertions
		WHERE message != ''
		ORDER BY RANDOM()
		LIMIT 10
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := pd.db.Prepare(`
		INSERT INTO pattern_examples (pattern_type, example_code, frequency)
		VALUES (?, ?, 1)
		ON CONFLICT DO UPDATE SET frequency = frequency + 1
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var assertType, expected, actual, message string
		if err := rows.Scan(&assertType, &expected, &actual, &message); err != nil {
			continue
		}

		// Construct example code
		exampleCode := fmt.Sprintf("%s(%s, %s, \"%s\")", assertType, expected, actual, message)
		stmt.Exec("assertion", exampleCode)
	}

	return nil
}

// extractWaitExamples extracts wait code patterns
func (pd *PatternDetector) extractWaitExamples() error {
	rows, err := pd.db.Query(`
		SELECT body_source
		FROM methods
		WHERE uses_explicit_wait = 1
		LIMIT 5
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := pd.db.Prepare(`
		INSERT INTO pattern_examples (pattern_type, example_code, frequency)
		VALUES (?, ?, 1)
		ON CONFLICT DO UPDATE SET frequency = frequency + 1
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Regular expression to extract wait statements
	waitRegex := regexp.MustCompile(`(?m)^\s*wait\.until\([^)]+\);?`)

	for rows.Next() {
		var bodySource string
		if err := rows.Scan(&bodySource); err != nil {
			continue
		}

		// Find wait statements
		matches := waitRegex.FindAllString(bodySource, -1)
		for _, match := range matches {
			stmt.Exec("wait", strings.TrimSpace(match))
		}
	}

	return nil
}

// saveCodingPatterns saves coding patterns to database
func (pd *PatternDetector) saveCodingPatterns(patterns *models.CodingPatterns) error {
	_, err := pd.db.Exec(`
		INSERT INTO coding_patterns (
			project_root, test_method_naming, page_object_naming,
			web_element_naming, method_naming, uses_aaa_pattern,
			uses_given_when_then, preferred_wait_type, default_wait_timeout,
			assertion_library, uses_assert_messages, uses_data_providers,
			uses_csv_files, uses_excel_files, last_analyzed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(project_root) DO UPDATE SET
			test_method_naming = excluded.test_method_naming,
			page_object_naming = excluded.page_object_naming,
			web_element_naming = excluded.web_element_naming,
			method_naming = excluded.method_naming,
			uses_aaa_pattern = excluded.uses_aaa_pattern,
			uses_given_when_then = excluded.uses_given_when_then,
			preferred_wait_type = excluded.preferred_wait_type,
			default_wait_timeout = excluded.default_wait_timeout,
			assertion_library = excluded.assertion_library,
			uses_assert_messages = excluded.uses_assert_messages,
			uses_data_providers = excluded.uses_data_providers,
			uses_csv_files = excluded.uses_csv_files,
			uses_excel_files = excluded.uses_excel_files,
			last_analyzed = datetime('now')
	`,
		pd.projectRoot, patterns.TestMethodNaming, patterns.PageObjectNaming,
		patterns.WebElementNaming, patterns.MethodNaming, patterns.UsesAAAPattern,
		patterns.UsesGivenWhenThen, patterns.PreferredWaitType, patterns.DefaultWaitTimeout,
		patterns.AssertionLibrary, patterns.UsesAssertMessages, patterns.UsesDataProviders,
		patterns.UsesCSVFiles, patterns.UsesExcelFiles,
	)

	return err
}

// GetCodingPatterns retrieves coding patterns from database
func (pd *PatternDetector) GetCodingPatterns() (*models.CodingPatterns, error) {
	patterns := &models.CodingPatterns{}

	err := pd.db.QueryRow(`
		SELECT test_method_naming, page_object_naming, web_element_naming,
		       method_naming, uses_aaa_pattern, uses_given_when_then,
		       preferred_wait_type, default_wait_timeout, assertion_library,
		       uses_assert_messages, uses_data_providers, uses_csv_files,
		       uses_excel_files
		FROM coding_patterns
		WHERE project_root = ?
	`, pd.projectRoot).Scan(
		&patterns.TestMethodNaming, &patterns.PageObjectNaming, &patterns.WebElementNaming,
		&patterns.MethodNaming, &patterns.UsesAAAPattern, &patterns.UsesGivenWhenThen,
		&patterns.PreferredWaitType, &patterns.DefaultWaitTimeout, &patterns.AssertionLibrary,
		&patterns.UsesAssertMessages, &patterns.UsesDataProviders, &patterns.UsesCSVFiles,
		&patterns.UsesExcelFiles,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no coding patterns found for project: %s", pd.projectRoot)
	}

	if err != nil {
		return nil, err
	}

	return patterns, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
