package generator

// This file bridges the scenario package with the generator package
// It converts scenario types to generator types

import (
	"github.com/yourusername/copilot-core/pkg/scenario"
)

// ConvertRecording converts scenario.Recording to generator.RecordingData
func ConvertRecording(rec *scenario.Recording) *RecordingData {
	if rec == nil {
		return nil
	}

	recording := &RecordingData{
		SessionID:    rec.SessionID,
		URL:          rec.URL,
		Title:        rec.Title,
		Elements:     []RecordedElement{},
		Interactions: []RecordedInteraction{},
	}

	// Convert elements
	for _, elem := range rec.Elements {
		recording.Elements = append(recording.Elements, RecordedElement{
			ID:          elem.ID,
			Type:        elem.Type,
			Label:       elem.Label,
			Placeholder: elem.Placeholder,
			Required:    elem.Required,
			Pattern:     elem.Pattern,
			MinLength:   elem.MinLength,
			MaxLength:   elem.MaxLength,
		})
	}

	// Convert interactions
	for _, interaction := range rec.Interactions {
		recording.Interactions = append(recording.Interactions, RecordedInteraction{
			Action:  interaction.Action,
			Element: interaction.Element,
			Value:   interaction.Value,
			URL:     interaction.URL,
		})
	}

	return recording
}

// ConvertAnalysis converts scenario.IntelligentAnalysis to generator.AnalysisData
func ConvertAnalysis(analysis *scenario.IntelligentAnalysis) *AnalysisData {
	if analysis == nil {
		return nil
	}

	analysisData := &AnalysisData{
		PageType:               analysis.PageAnalysis.Type,
		PagePurpose:            analysis.PageAnalysis.Purpose,
		ValidationRules:        make(map[string]ValidationRule),
		RecommendedPageObjects: []PageObjectSpec{},
	}

	// Convert validation rules
	for field, rule := range analysis.ValidationRules {
		analysisData.ValidationRules[field] = ValidationRule{
			Required:  rule.Required,
			Format:    rule.Format,
			MinLength: rule.MinLength,
			MaxLength: rule.MaxLength,
			Pattern:   rule.Pattern,
		}
	}

	// Extract success/failure URLs
	if analysis.SuccessCriteria.Navigation.ExpectedURL != "" {
		analysisData.SuccessURL = analysis.SuccessCriteria.Navigation.ExpectedURL
	}

	if analysis.FailureCriteria.Navigation.ExpectedURL != "" {
		analysisData.FailureURL = analysis.FailureCriteria.Navigation.ExpectedURL
	}

	// Convert page object specs
	for _, po := range analysis.RecommendedPageObjects {
		analysisData.RecommendedPageObjects = append(analysisData.RecommendedPageObjects, PageObjectSpec{
			Name:     po.Name,
			Elements: po.Elements,
			Methods:  po.Methods,
		})
	}

	return analysisData
}

// ConvertScenarios converts []scenario.TestScenario to []ScenarioData
func ConvertScenarios(scenarios []scenario.TestScenario) []ScenarioData {
	converted := []ScenarioData{}

	for _, s := range scenarios {
		converted = append(converted, ScenarioData{
			Category:       s.Category,
			Name:           s.Name,
			Description:    s.Description,
			Priority:       s.Priority,
			TestData:       s.TestData,
			ExpectedResult: s.ExpectedResult,
			Reasoning:      s.Reasoning,
		})
	}

	return converted
}

// BuildGenerationRequest creates a ScenarioGenerationRequest from scenario package types
func BuildGenerationRequest(
	recording *scenario.Recording,
	analysis *scenario.IntelligentAnalysis,
	scenarios []scenario.TestScenario,
) ScenarioGenerationRequest {
	return ScenarioGenerationRequest{
		Recording:          ConvertRecording(recording),
		Analysis:           ConvertAnalysis(analysis),
		SelectedScenarios:  ConvertScenarios(scenarios),
		GeneratePageObject: true,
		GenerateTests:      true,
		GenerateFeatures:   true,
		GenerateSteps:      true,
	}
}
