package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ApiTestType represents the type of API test
type ApiTestType string

const (
	ApiTestTypeREST    ApiTestType = "REST"
	ApiTestTypeGraphQL ApiTestType = "GRAPHQL"
)

// ApiTestDefinition represents an API test configuration
type ApiTestDefinition struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"projectId"`
	Name        string       `json:"name"`
	Description *string      `json:"description,omitempty"`
	Type        ApiTestType  `json:"type"`
	Method      *string      `json:"method,omitempty"`      // HTTP method for REST
	URL         string       `json:"url"`
	Headers     *string      `json:"headers,omitempty"`     // JSON string
	Body        *string      `json:"body,omitempty"`        // JSON string
	Query       *string      `json:"query,omitempty"`       // GraphQL query
	Variables   *string      `json:"variables,omitempty"`   // JSON string
	Assertions  *string      `json:"assertions,omitempty"`  // JSON string
	Timeout     int          `json:"timeout"`
	IsActive    bool         `json:"isActive"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// ApiTestResult represents an API test execution result
type ApiTestResult struct {
	ID              string     `json:"id"`
	TestID          string     `json:"testId"`
	ExecutionID     *string    `json:"executionId,omitempty"`
	Passed          bool       `json:"passed"`
	Duration        int        `json:"duration"`        // milliseconds
	StatusCode      *int       `json:"statusCode,omitempty"`
	ResponseBody    *string    `json:"responseBody,omitempty"`    // JSON string
	ResponseHeaders *string    `json:"responseHeaders,omitempty"` // JSON string
	ErrorMessage    *string    `json:"errorMessage,omitempty"`
	Assertions      *string    `json:"assertions,omitempty"`      // JSON string
	CreatedAt       time.Time  `json:"createdAt"`
}

// SaveApiTestDefinition saves an API test definition
func (db *DB) SaveApiTestDefinition(test *ApiTestDefinition) error {
	// First, check if api_test_definitions table exists, if not create it
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS api_test_definitions (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			method TEXT,
			url TEXT NOT NULL,
			headers TEXT,
			body TEXT,
			query TEXT,
			variables TEXT,
			assertions TEXT,
			timeout INTEGER DEFAULT 30000,
			is_active INTEGER DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_api_tests_project ON api_test_definitions(project_id);
		CREATE INDEX IF NOT EXISTS idx_api_tests_type ON api_test_definitions(type);
	`

	if _, err := db.conn.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create api_test_definitions table: %w", err)
	}

	query := `
		INSERT INTO api_test_definitions (
			id, project_id, name, description, type, method, url,
			headers, body, query, variables, assertions, timeout,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			type = excluded.type,
			method = excluded.method,
			url = excluded.url,
			headers = excluded.headers,
			body = excluded.body,
			query = excluded.query,
			variables = excluded.variables,
			assertions = excluded.assertions,
			timeout = excluded.timeout,
			is_active = excluded.is_active,
			updated_at = excluded.updated_at
	`

	createdAt := test.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := test.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	_, err := db.conn.Exec(query,
		test.ID,
		test.ProjectID,
		test.Name,
		test.Description,
		test.Type,
		test.Method,
		test.URL,
		test.Headers,
		test.Body,
		test.Query,
		test.Variables,
		test.Assertions,
		test.Timeout,
		test.IsActive,
		createdAt.Unix(),
		updatedAt.Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to save API test definition: %w", err)
	}

	return nil
}

// GetApiTestDefinition retrieves an API test definition by ID
func (db *DB) GetApiTestDefinition(id string) (*ApiTestDefinition, error) {
	query := `
		SELECT id, project_id, name, description, type, method, url,
		       headers, body, query, variables, assertions, timeout,
		       is_active, created_at, updated_at
		FROM api_test_definitions
		WHERE id = ?
	`

	test := &ApiTestDefinition{}
	var createdAt, updatedAt int64
	var isActive int

	err := db.conn.QueryRow(query, id).Scan(
		&test.ID,
		&test.ProjectID,
		&test.Name,
		&test.Description,
		&test.Type,
		&test.Method,
		&test.URL,
		&test.Headers,
		&test.Body,
		&test.Query,
		&test.Variables,
		&test.Assertions,
		&test.Timeout,
		&isActive,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API test definition not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API test definition: %w", err)
	}

	test.IsActive = isActive == 1
	test.CreatedAt = time.Unix(createdAt, 0)
	test.UpdatedAt = time.Unix(updatedAt, 0)

	return test, nil
}

// GetApiTestDefinitionsByProject retrieves all API test definitions for a project
func (db *DB) GetApiTestDefinitionsByProject(projectID string) ([]*ApiTestDefinition, error) {
	query := `
		SELECT id, project_id, name, description, type, method, url,
		       headers, body, query, variables, assertions, timeout,
		       is_active, created_at, updated_at
		FROM api_test_definitions
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API test definitions: %w", err)
	}
	defer rows.Close()

	var tests []*ApiTestDefinition
	for rows.Next() {
		test := &ApiTestDefinition{}
		var createdAt, updatedAt int64
		var isActive int

		err := rows.Scan(
			&test.ID,
			&test.ProjectID,
			&test.Name,
			&test.Description,
			&test.Type,
			&test.Method,
			&test.URL,
			&test.Headers,
			&test.Body,
			&test.Query,
			&test.Variables,
			&test.Assertions,
			&test.Timeout,
			&isActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API test definition: %w", err)
		}

		test.IsActive = isActive == 1
		test.CreatedAt = time.Unix(createdAt, 0)
		test.UpdatedAt = time.Unix(updatedAt, 0)

		tests = append(tests, test)
	}

	return tests, nil
}

// DeleteApiTestDefinition deletes an API test definition
func (db *DB) DeleteApiTestDefinition(id string) error {
	query := `DELETE FROM api_test_definitions WHERE id = ?`

	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API test definition: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API test definition not found: %s", id)
	}

	return nil
}

// SaveApiTestResult saves an API test execution result
func (db *DB) SaveApiTestResult(result *ApiTestResult) error {
	// First, check if api_test_results table exists, if not create it
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS api_test_results (
			id TEXT PRIMARY KEY,
			test_id TEXT NOT NULL,
			execution_id TEXT,
			passed INTEGER NOT NULL,
			duration INTEGER NOT NULL,
			status_code INTEGER,
			response_body TEXT,
			response_headers TEXT,
			error_message TEXT,
			assertions TEXT,
			created_at INTEGER NOT NULL,
			FOREIGN KEY (test_id) REFERENCES api_test_definitions(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_api_results_test ON api_test_results(test_id);
		CREATE INDEX IF NOT EXISTS idx_api_results_execution ON api_test_results(execution_id);
		CREATE INDEX IF NOT EXISTS idx_api_results_created ON api_test_results(created_at DESC);
	`

	if _, err := db.conn.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create api_test_results table: %w", err)
	}

	query := `
		INSERT INTO api_test_results (
			id, test_id, execution_id, passed, duration, status_code,
			response_body, response_headers, error_message, assertions, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	createdAt := result.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	var passed int
	if result.Passed {
		passed = 1
	}

	_, err := db.conn.Exec(query,
		result.ID,
		result.TestID,
		result.ExecutionID,
		passed,
		result.Duration,
		result.StatusCode,
		result.ResponseBody,
		result.ResponseHeaders,
		result.ErrorMessage,
		result.Assertions,
		createdAt.Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to save API test result: %w", err)
	}

	return nil
}

// GetApiTestResultsByTestID retrieves all results for a specific API test
func (db *DB) GetApiTestResultsByTestID(testID string, limit int) ([]*ApiTestResult, error) {
	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT id, test_id, execution_id, passed, duration, status_code,
		       response_body, response_headers, error_message, assertions, created_at
		FROM api_test_results
		WHERE test_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, testID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query API test results: %w", err)
	}
	defer rows.Close()

	var results []*ApiTestResult
	for rows.Next() {
		result := &ApiTestResult{}
		var createdAt int64
		var passed int

		err := rows.Scan(
			&result.ID,
			&result.TestID,
			&result.ExecutionID,
			&passed,
			&result.Duration,
			&result.StatusCode,
			&result.ResponseBody,
			&result.ResponseHeaders,
			&result.ErrorMessage,
			&result.Assertions,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API test result: %w", err)
		}

		result.Passed = passed == 1
		result.CreatedAt = time.Unix(createdAt, 0)

		results = append(results, result)
	}

	return results, nil
}

// GetApiTestStats returns statistics for API tests
func (db *DB) GetApiTestStats(projectID string) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_tests,
			SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END) as active_tests,
			SUM(CASE WHEN type = 'REST' THEN 1 ELSE 0 END) as rest_tests,
			SUM(CASE WHEN type = 'GRAPHQL' THEN 1 ELSE 0 END) as graphql_tests
		FROM api_test_definitions
		WHERE project_id = ?
	`

	var totalTests, activeTests, restTests, graphqlTests int

	err := db.conn.QueryRow(query, projectID).Scan(
		&totalTests,
		&activeTests,
		&restTests,
		&graphqlTests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get API test stats: %w", err)
	}

	stats := map[string]interface{}{
		"totalTests":    totalTests,
		"activeTests":   activeTests,
		"restTests":     restTests,
		"graphqlTests":  graphqlTests,
	}

	return stats, nil
}

// ParseApiTestHeaders parses JSON headers string
func ParseApiTestHeaders(headersJSON string) (map[string]string, error) {
	var headers map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}
	return headers, nil
}

// ParseApiTestBody parses JSON body string
func ParseApiTestBody(bodyJSON string) (map[string]interface{}, error) {
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}
	return body, nil
}
