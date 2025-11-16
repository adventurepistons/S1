package models

// FrameworkConfig represents detected framework combination
type FrameworkConfig struct {
	// Primary framework
	TestFramework string `json:"test_framework"` // TestNG, JUnit5

	// Additional frameworks
	BDDFramework string `json:"bdd_framework"` // Cucumber, Serenity, none
	APIFramework string `json:"api_framework"` // RestAssured, none

	// Detected from dependencies
	SeleniumVersion string `json:"selenium_version"`
	JavaVersion     string `json:"java_version"`

	// Architecture patterns
	UsesPageObjectModel bool `json:"uses_pom"`
	UsesPageFactory     bool `json:"uses_page_factory"`
	UsesScreenplay      bool `json:"uses_screenplay"`

	// Configuration files
	TestNGXMLPath    string `json:"testng_xml_path"`
	POMFilePath      string `json:"pom_file_path"`
	CucumberFeatures string `json:"cucumber_features_path"`
}

// CodingPatterns represents project coding style
type CodingPatterns struct {
	// Naming conventions
	TestMethodNaming string `json:"test_method_naming"` // testActionWithCondition, givenWhenThen
	PageObjectNaming string `json:"page_object_naming"` // LoginPage, LoginPageObject

	// Test structure
	UsesAAAPattern    bool `json:"uses_aaa_pattern"`     // Arrange-Act-Assert
	UsesGivenWhenThen bool `json:"uses_given_when_then"` // Given-When-Then

	// Waits
	PreferredWaitType  string `json:"preferred_wait_type"`  // explicit, implicit, fluent
	DefaultWaitTimeout int    `json:"default_wait_timeout"` // in seconds

	// Assertions
	AssertionLibrary   string `json:"assertion_library"`    // TestNG, JUnit, AssertJ, Hamcrest
	UsesAssertMessages bool   `json:"uses_assert_messages"` // true if assertions have messages

	// Page objects
	WebElementNaming string `json:"web_element_naming"` // usernameField, username_field
	MethodNaming     string `json:"method_naming"`      // enterUsername, setUsername, typeUsername

	// Data patterns
	UsesDataProviders bool `json:"uses_data_providers"`
	UsesCSVFiles      bool `json:"uses_csv_files"`
	UsesExcelFiles    bool `json:"uses_excel_files"`
}

// ProjectMetadata represents overall project information
type ProjectMetadata struct {
	ProjectRoot    string           `json:"project_root"`
	Framework      *FrameworkConfig `json:"framework"`
	CodingPatterns *CodingPatterns  `json:"coding_patterns"`
	TotalFiles     int              `json:"total_files"`
	TotalClasses   int              `json:"total_classes"`
	TotalTests     int              `json:"total_tests"`
	LastIndexed    string           `json:"last_indexed"`
}
