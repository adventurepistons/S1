package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GherkinParser parses Cucumber .feature files
type GherkinParser struct{}

type FeatureFile struct {
	FilePath    string            `json:"filePath"`
	Feature     FeatureInfo       `json:"feature"`
	Scenarios   []ScenarioInfo    `json:"scenarios"`
	Background  *BackgroundInfo   `json:"background,omitempty"`
}

type FeatureInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type ScenarioInfo struct {
	Type        string         `json:"type"` // "Scenario" or "Scenario Outline"
	Name        string         `json:"name"`
	Tags        []string       `json:"tags"`
	Steps       []StepInfo     `json:"steps"`
	Examples    *ExamplesTable `json:"examples,omitempty"`
}

type BackgroundInfo struct {
	Name  string     `json:"name"`
	Steps []StepInfo `json:"steps"`
}

type StepInfo struct {
	Keyword string `json:"keyword"` // Given, When, Then, And, But
	Text    string `json:"text"`
	Line    int    `json:"line"`
}

type ExamplesTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

func NewGherkinParser() *GherkinParser {
	return &GherkinParser{}
}

// ParseFile parses a single .feature file
func (p *GherkinParser) ParseFile(filePath string) (*FeatureFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := &FeatureFile{
		FilePath:  filePath,
		Scenarios: []ScenarioInfo{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	var currentTags []string
	var currentScenario *ScenarioInfo
	var inExamples bool
	var examplesHeaders []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Tags
		if strings.HasPrefix(trimmed, "@") {
			tags := strings.Fields(trimmed)
			currentTags = append(currentTags, tags...)
			continue
		}

		// Feature
		if strings.HasPrefix(trimmed, "Feature:") {
			result.Feature.Name = strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))
			result.Feature.Tags = currentTags
			currentTags = []string{}
			continue
		}

		// Background
		if strings.HasPrefix(trimmed, "Background:") {
			result.Background = &BackgroundInfo{
				Name:  strings.TrimSpace(strings.TrimPrefix(trimmed, "Background:")),
				Steps: []StepInfo{},
			}
			continue
		}

		// Scenario or Scenario Outline
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			// Save previous scenario if exists
			if currentScenario != nil {
				result.Scenarios = append(result.Scenarios, *currentScenario)
			}

			scenarioType := "Scenario"
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "Scenario:"))

			if strings.HasPrefix(trimmed, "Scenario Outline:") {
				scenarioType = "Scenario Outline"
				name = strings.TrimSpace(strings.TrimPrefix(trimmed, "Scenario Outline:"))
			}

			currentScenario = &ScenarioInfo{
				Type:  scenarioType,
				Name:  name,
				Tags:  currentTags,
				Steps: []StepInfo{},
			}
			currentTags = []string{}
			inExamples = false
			continue
		}

		// Examples
		if strings.HasPrefix(trimmed, "Examples:") {
			inExamples = true
			if currentScenario != nil {
				currentScenario.Examples = &ExamplesTable{
					Headers: []string{},
					Rows:    [][]string{},
				}
			}
			continue
		}

		// Examples table headers and rows
		if inExamples && strings.HasPrefix(trimmed, "|") {
			cells := parseTableRow(trimmed)
			if currentScenario != nil && currentScenario.Examples != nil {
				if len(examplesHeaders) == 0 {
					// First row is headers
					examplesHeaders = cells
					currentScenario.Examples.Headers = cells
				} else {
					// Data rows
					currentScenario.Examples.Rows = append(currentScenario.Examples.Rows, cells)
				}
			}
			continue
		}

		// Steps (Given, When, Then, And, But)
		for _, keyword := range []string{"Given", "When", "Then", "And", "But"} {
			if strings.HasPrefix(trimmed, keyword+" ") {
				step := StepInfo{
					Keyword: keyword,
					Text:    strings.TrimSpace(strings.TrimPrefix(trimmed, keyword+" ")),
					Line:    lineNum,
				}

				// Add to current context (Background or Scenario)
				if result.Background != nil && currentScenario == nil {
					result.Background.Steps = append(result.Background.Steps, step)
				} else if currentScenario != nil {
					currentScenario.Steps = append(currentScenario.Steps, step)
				}
				break
			}
		}
	}

	// Don't forget the last scenario
	if currentScenario != nil {
		result.Scenarios = append(result.Scenarios, *currentScenario)
	}

	return result, scanner.Err()
}

// ParseDirectory parses all .feature files in a directory
func (p *GherkinParser) ParseDirectory(dirPath string) ([]*FeatureFile, error) {
	var results []*FeatureFile

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			parsed, err := p.ParseFile(path)
			if err != nil {
				// Log but continue
				return nil
			}
			results = append(results, parsed)
		}

		return nil
	})

	return results, err
}

// Helper function to parse table rows
func parseTableRow(line string) []string {
	var cells []string
	// Remove leading and trailing pipes
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	// Split by pipe and trim each cell
	parts := strings.Split(line, "|")
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}

	return cells
}

// Match steps to step definitions
func (p *GherkinParser) MatchStepsToDefinitions(featureFile *FeatureFile, stepDefs []StepDefinition) map[string][]string {
	matches := make(map[string][]string)

	for _, scenario := range featureFile.Scenarios {
		for _, step := range scenario.Steps {
			stepText := step.Text
			for _, stepDef := range stepDefs {
				if matchesPattern(stepText, stepDef.Pattern) {
					key := scenario.Name + " - " + step.Keyword + " " + step.Text
					matches[key] = append(matches[key], stepDef.MethodName)
				}
			}
		}
	}

	return matches
}

type StepDefinition struct {
	MethodName string
	Pattern    string
	Type       string // Given, When, Then
}

// Simple pattern matching (can be enhanced with regex)
func matchesPattern(text, pattern string) bool {
	// Remove regex characters for simple matching
	pattern = strings.ReplaceAll(pattern, "^", "")
	pattern = strings.ReplaceAll(pattern, "$", "")
	pattern = strings.ReplaceAll(pattern, ".*", "*")

	// Simple wildcard matching
	return strings.Contains(text, pattern) || pattern == "*"
}
