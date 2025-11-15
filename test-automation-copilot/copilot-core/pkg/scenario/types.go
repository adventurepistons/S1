package scenario

import "time"

// Recording represents a browser recording session
type Recording struct {
	SessionID    string                `json:"sessionId"`
	URL          string                `json:"url"`
	Title        string                `json:"title"`
	Elements     []RecordedElement     `json:"elements"`
	Interactions []RecordedInteraction `json:"interactions"`
	Timestamp    time.Time             `json:"timestamp"`
}

// RecordedElement represents a DOM element captured during recording
type RecordedElement struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name,omitempty"`
	Placeholder string            `json:"placeholder,omitempty"`
	Required    bool              `json:"required,omitempty"`
	Pattern     string            `json:"pattern,omitempty"`
	MinLength   int               `json:"minLength,omitempty"`
	MaxLength   int               `json:"maxLength,omitempty"`
	Label       string            `json:"label,omitempty"`
	Text        string            `json:"text,omitempty"`
	Value       string            `json:"value,omitempty"`
	Href        string            `json:"href,omitempty"`
	ClassName   string            `json:"className,omitempty"`
	TagName     string            `json:"tagName,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// RecordedInteraction represents a user interaction during recording
type RecordedInteraction struct {
	Action    string `json:"action"` // "input", "click", "navigate", "select"
	Element   string `json:"element"`
	Value     string `json:"value,omitempty"`
	URL       string `json:"url,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// IntelligentAnalysis represents GPT's comprehensive analysis of a recording
type IntelligentAnalysis struct {
	PageAnalysis           PageAnalysis           `json:"pageAnalysis"`
	ValidationRules        map[string]Rule        `json:"inferredValidationRules"`
	SuccessCriteria        Criteria               `json:"successCriteria"`
	FailureCriteria        Criteria               `json:"failureCriteria"`
	TestScenarios          []TestScenario         `json:"testScenarios"`
	RelatedFeatures        map[string]Feature     `json:"relatedFeatures"`
	RecommendedPageObjects []PageObjectSpec       `json:"recommendedPageObjects"`
}

// PageAnalysis contains GPT's understanding of the page type and purpose
type PageAnalysis struct {
	Type       string  `json:"type"`       // "login", "registration", "checkout", etc.
	Purpose    string  `json:"purpose"`    // Human-readable description
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
}

// Rule represents an inferred validation rule for a field
type Rule struct {
	Required   bool   `json:"required"`
	Format     string `json:"format,omitempty"`     // "email", "url", "number", etc.
	MinLength  int    `json:"minLength,omitempty"`  // Minimum length
	MaxLength  int    `json:"maxLength,omitempty"`  // Maximum length
	Pattern    string `json:"pattern,omitempty"`    // Regex pattern
	Min        int    `json:"min,omitempty"`        // Minimum value (for numbers)
	Max        int    `json:"max,omitempty"`        // Maximum value (for numbers)
	Reasoning  string `json:"reasoning"`            // Why this rule was inferred
}

// Criteria defines success or failure conditions
type Criteria struct {
	Navigation      NavigationCheck `json:"navigation"`
	ElementsToCheck []ElementCheck  `json:"elementsToCheck"`
}

// NavigationCheck defines expected URL navigation
type NavigationCheck struct {
	ExpectedURL string `json:"expectedUrl"`
	Reasoning   string `json:"reasoning"`
}

// ElementCheck defines expected element state
type ElementCheck struct {
	Selector      string `json:"selector"`
	ExpectedState string `json:"expectedState"` // "visible", "present", "hidden"
	Text          string `json:"text,omitempty"`
	Reasoning     string `json:"reasoning"`
}

// TestScenario represents a single test case scenario
type TestScenario struct {
	Category       string                 `json:"category"` // "positive", "negative", "edgeCase", "security"
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Priority       string                 `json:"priority"` // "high", "medium", "low"
	TestData       map[string]interface{} `json:"testData"`
	ExpectedResult string                 `json:"expectedResult"` // "success", "failure"
	Reasoning      string                 `json:"reasoning"`
}

// Feature represents a related feature detected in the page
type Feature struct {
	Detected  bool   `json:"detected"`
	Link      string `json:"link,omitempty"`
	Reasoning string `json:"reasoning"`
}

// PageObjectSpec represents a recommended Page Object class
type PageObjectSpec struct {
	Name      string   `json:"name"`
	Elements  []string `json:"elements"`
	Methods   []string `json:"methods"`
	Reasoning string   `json:"reasoning,omitempty"`
}

// GenerateRequest contains parameters for scenario generation
type GenerateRequest struct {
	Recording   *Recording
	UserIntent  string // Optional: "I want to test login"
	AppContext  string // Optional: "e-commerce", "banking", etc.
	Constraints struct {
		MaxScenarios    int  // Limit number of scenarios (0 = unlimited)
		IncludeSecurity bool // Include security tests
		PriorityFilter  string // Filter by priority: "high", "medium", "low", "all"
	}
}

// GenerateResponse contains the analysis results
type GenerateResponse struct {
	Analysis          *IntelligentAnalysis
	TotalScenarios    int
	PositiveCount     int
	NegativeCount     int
	EdgeCaseCount     int
	SecurityTestCount int
	ProcessingTime    time.Duration
}

// ScenarioFilter allows filtering scenarios
type ScenarioFilter struct {
	Categories []string // Filter by category
	Priorities []string // Filter by priority
	MinCount   int      // Minimum scenarios to return
	MaxCount   int      // Maximum scenarios to return
}
