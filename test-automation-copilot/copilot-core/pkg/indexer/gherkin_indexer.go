package indexer

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/parser"
)

// GherkinIndexer indexes Gherkin feature files
type GherkinIndexer struct {
	db     *database.DB
	parser *parser.GherkinParser
}

// NewGherkinIndexer creates a new Gherkin indexer
func NewGherkinIndexer(db *database.DB) *GherkinIndexer {
	return &GherkinIndexer{
		db:     db,
		parser: parser.NewGherkinParser(),
	}
}

// IndexFile indexes a Gherkin feature file
func (idx *GherkinIndexer) IndexFile(file *database.File, content string) error {
	// Parse Gherkin file
	parseResult, err := idx.parser.Parse(content, file.Path)
	if err != nil {
		return fmt.Errorf("failed to parse Gherkin file: %w", err)
	}

	// Use transaction
	return idx.db.WithTransaction(func(tx *sqlx.Tx) error {
		// Delete old feature data for this file
		_, err := tx.Exec("DELETE FROM feature_files WHERE file_id = ?", file.ID)
		if err != nil {
			return err
		}

		// Create feature record
		tagsJSON, _ := json.Marshal(parseResult.Tags)
		tagsStr := string(tagsJSON)

		result, err := tx.Exec(`
			INSERT INTO feature_files (file_id, feature_name, description, tags, language)
			VALUES (?, ?, ?, ?, ?)`,
			file.ID,
			parseResult.FeatureName,
			database.StringPtr(parseResult.Description),
			&tagsStr,
			"en",
		)
		if err != nil {
			return fmt.Errorf("failed to insert feature: %w", err)
		}

		featureID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Index scenarios
		for _, scenario := range parseResult.Scenarios {
			if err := idx.indexScenario(tx, featureID, scenario); err != nil {
				return fmt.Errorf("failed to index scenario %s: %w", scenario.Name, err)
			}
		}

		return nil
	})
}

// indexScenario indexes a single scenario
func (idx *GherkinIndexer) indexScenario(tx *sqlx.Tx, featureID int64, scenario parser.Scenario) error {
	// Convert tags to JSON
	tagsJSON, _ := json.Marshal(scenario.Tags)
	tagsStr := string(tagsJSON)

	// Insert scenario
	result, err := tx.Exec(`
		INSERT INTO scenarios (feature_id, name, type, description, tags, start_line, end_line)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		featureID,
		scenario.Name,
		scenario.Type, // "scenario" or "scenario_outline"
		database.StringPtr(scenario.Description),
		&tagsStr,
		scenario.StartLine,
		scenario.EndLine,
	)
	if err != nil {
		return fmt.Errorf("failed to insert scenario: %w", err)
	}

	scenarioID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Index steps
	for _, step := range scenario.Steps {
		if err := idx.indexStep(tx, scenarioID, step); err != nil {
			return fmt.Errorf("failed to index step: %w", err)
		}
	}

	return nil
}

// indexStep indexes a single step
func (idx *GherkinIndexer) indexStep(tx *sqlx.Tx, scenarioID int64, step parser.Step) error {
	// Insert step
	_, err := tx.Exec(`
		INSERT INTO steps (scenario_id, keyword, text, argument, line_number)
		VALUES (?, ?, ?, ?, ?)`,
		scenarioID,
		step.Keyword,
		step.Text,
		database.StringPtr(step.Argument), // DataTable or DocString
		step.LineNumber,
	)

	return err
}
