package detector

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adventurepistons/automation-copilot/pkg/models"
)

// FrameworkDetector detects framework combination from project structure
type FrameworkDetector struct {
	projectRoot string
	db          *sql.DB
}

// NewFrameworkDetector creates a new framework detector
func NewFrameworkDetector(projectRoot string, db *sql.DB) *FrameworkDetector {
	return &FrameworkDetector{
		projectRoot: projectRoot,
		db:          db,
	}
}

// Detect identifies framework combination from 6 possibilities:
// 1. Selenium + TestNG
// 2. Selenium + TestNG + Cucumber
// 3. Selenium + JUnit 5
// 4. Selenium + JUnit 5 + Cucumber
// 5. Selenium + Serenity BDD
// 6. Selenium + Rest-Assured
func (fd *FrameworkDetector) Detect() (*models.FrameworkConfig, error) {
	config := &models.FrameworkConfig{}

	// 1. Check pom.xml for dependencies
	pomPath := filepath.Join(fd.projectRoot, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		config.POMFilePath = pomPath
		if err := fd.detectFromPOM(pomPath, config); err != nil {
			fmt.Printf("Warning: Failed to parse pom.xml: %v\n", err)
		}
	}

	// 2. Check imports in Java files (from database)
	if err := fd.detectFromImports(config); err != nil {
		fmt.Printf("Warning: Failed to analyze imports: %v\n", err)
	}

	// 3. Check for TestNG XML
	testNGXML := fd.findTestNGXML(fd.projectRoot)
	if testNGXML != "" {
		config.TestNGXMLPath = testNGXML
		if config.TestFramework == "" {
			config.TestFramework = "TestNG"
		}
	}

	// 4. Check for Cucumber features
	featuresPath := fd.findCucumberFeatures(fd.projectRoot)
	if featuresPath != "" {
		config.CucumberFeatures = featuresPath
		config.BDDFramework = "Cucumber"
	}

	// 5. Detect architecture patterns
	if err := fd.detectArchitecturePatterns(config); err != nil {
		fmt.Printf("Warning: Failed to detect architecture patterns: %v\n", err)
	}

	// 6. Save to database
	if err := fd.saveFrameworkConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save framework config: %w", err)
	}

	return config, nil
}

// POM represents a simplified Maven POM file
type POM struct {
	XMLName      xml.Name     `xml:"project"`
	Dependencies Dependencies `xml:"dependencies"`
	Properties   Properties   `xml:"properties"`
}

type Dependencies struct {
	Dependency []Dependency `xml:"dependency"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type Properties struct {
	Property []Property `xml:",any"`
}

type Property struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// detectFromPOM parses pom.xml and detects framework from dependencies
func (fd *FrameworkDetector) detectFromPOM(pomPath string, config *models.FrameworkConfig) error {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return err
	}

	var pom POM
	if err := xml.Unmarshal(data, &pom); err != nil {
		return err
	}

	// Analyze dependencies
	for _, dep := range pom.Dependencies.Dependency {
		fd.analyzeDependency(dep, config)
	}

	return nil
}

// analyzeDependency analyzes a single dependency
func (fd *FrameworkDetector) analyzeDependency(dep Dependency, config *models.FrameworkConfig) {
	groupArtifact := dep.GroupID + ":" + dep.ArtifactID

	switch {
	// Test frameworks
	case strings.Contains(groupArtifact, "org.testng:testng"):
		config.TestFramework = "TestNG"

	case strings.Contains(groupArtifact, "org.junit.jupiter:junit-jupiter"):
		config.TestFramework = "JUnit5"

	case strings.Contains(groupArtifact, "junit:junit") && strings.HasPrefix(dep.Version, "5"):
		config.TestFramework = "JUnit5"

	// BDD frameworks
	case strings.Contains(groupArtifact, "io.cucumber:cucumber-java"):
		config.BDDFramework = "Cucumber"

	case strings.Contains(groupArtifact, "io.cucumber:cucumber-testng"):
		config.BDDFramework = "Cucumber"
		if config.TestFramework == "" {
			config.TestFramework = "TestNG"
		}

	case strings.Contains(groupArtifact, "io.cucumber:cucumber-junit"):
		config.BDDFramework = "Cucumber"
		if config.TestFramework == "" {
			config.TestFramework = "JUnit5"
		}

	case strings.Contains(groupArtifact, "net.serenity-bdd"):
		config.BDDFramework = "Serenity"

	// API testing
	case strings.Contains(groupArtifact, "io.rest-assured:rest-assured"):
		config.APIFramework = "RestAssured"

	// Selenium
	case strings.Contains(groupArtifact, "org.seleniumhq.selenium:selenium-java"):
		config.SeleniumVersion = dep.Version
	}
}

// detectFromImports analyzes imports from database
func (fd *FrameworkDetector) detectFromImports(config *models.FrameworkConfig) error {
	rows, err := fd.db.Query("SELECT DISTINCT import_path FROM imports")
	if err != nil {
		return err
	}
	defer rows.Close()

	var imports []string
	for rows.Next() {
		var importPath string
		if err := rows.Scan(&importPath); err != nil {
			return err
		}
		imports = append(imports, importPath)
	}

	// Analyze imports
	for _, imp := range imports {
		switch {
		case strings.Contains(imp, "org.testng"):
			if config.TestFramework == "" {
				config.TestFramework = "TestNG"
			}

		case strings.Contains(imp, "org.junit.jupiter"):
			if config.TestFramework == "" {
				config.TestFramework = "JUnit5"
			}

		case strings.Contains(imp, "io.cucumber"):
			if config.BDDFramework == "" {
				config.BDDFramework = "Cucumber"
			}

		case strings.Contains(imp, "net.serenity"):
			if config.BDDFramework == "" {
				config.BDDFramework = "Serenity"
			}

		case strings.Contains(imp, "net.serenitybdd.screenplay"):
			config.UsesScreenplay = true

		case strings.Contains(imp, "io.restassured"):
			if config.APIFramework == "" {
				config.APIFramework = "RestAssured"
			}

		case strings.Contains(imp, "org.openqa.selenium.support.PageFactory"):
			config.UsesPageFactory = true
		}
	}

	return nil
}

// findTestNGXML searches for testng.xml files
func (fd *FrameworkDetector) findTestNGXML(root string) string {
	var testngPath string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && (info.Name() == "testng.xml" || strings.HasSuffix(info.Name(), ".xml")) {
			// Check if it's actually a TestNG XML file
			data, err := os.ReadFile(path)
			if err == nil && strings.Contains(string(data), "<!DOCTYPE suite") {
				testngPath = path
				return filepath.SkipDir // Found it, stop searching
			}
		}

		return nil
	})

	return testngPath
}

// findCucumberFeatures searches for .feature files
func (fd *FrameworkDetector) findCucumberFeatures(root string) string {
	var featuresPath string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && (info.Name() == "features" || info.Name() == "feature") {
			featuresPath = path
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".feature") {
			featuresPath = filepath.Dir(path)
			return filepath.SkipDir
		}

		return nil
	})

	return featuresPath
}

// detectArchitecturePatterns detects architecture patterns from database
func (fd *FrameworkDetector) detectArchitecturePatterns(config *models.FrameworkConfig) error {
	// Check if Page Object Model is used
	var pageObjectCount int
	err := fd.db.QueryRow("SELECT COUNT(*) FROM classes WHERE is_page_object = 1").Scan(&pageObjectCount)
	if err != nil {
		return err
	}

	if pageObjectCount > 0 {
		config.UsesPageObjectModel = true
	}

	// Check if PageFactory is used (by looking for PageFactory.initElements calls)
	var pageFactoryCount int
	err = fd.db.QueryRow(`
		SELECT COUNT(*) FROM method_calls
		WHERE method_name = 'initElements' AND object_name LIKE '%PageFactory%'
	`).Scan(&pageFactoryCount)
	if err == nil && pageFactoryCount > 0 {
		config.UsesPageFactory = true
	}

	return nil
}

// saveFrameworkConfig saves framework configuration to database
func (fd *FrameworkDetector) saveFrameworkConfig(config *models.FrameworkConfig) error {
	_, err := fd.db.Exec(`
		INSERT INTO framework_config (
			project_root, test_framework, bdd_framework, api_framework,
			selenium_version, uses_page_object_model, uses_page_factory,
			uses_screenplay, testng_xml_path, pom_file_path, cucumber_features_path,
			last_detected
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(project_root) DO UPDATE SET
			test_framework = excluded.test_framework,
			bdd_framework = excluded.bdd_framework,
			api_framework = excluded.api_framework,
			selenium_version = excluded.selenium_version,
			uses_page_object_model = excluded.uses_page_object_model,
			uses_page_factory = excluded.uses_page_factory,
			uses_screenplay = excluded.uses_screenplay,
			testng_xml_path = excluded.testng_xml_path,
			pom_file_path = excluded.pom_file_path,
			cucumber_features_path = excluded.cucumber_features_path,
			last_detected = datetime('now')
	`,
		fd.projectRoot, config.TestFramework, config.BDDFramework, config.APIFramework,
		config.SeleniumVersion, config.UsesPageObjectModel, config.UsesPageFactory,
		config.UsesScreenplay, config.TestNGXMLPath, config.POMFilePath, config.CucumberFeatures,
	)

	return err
}

// GetFrameworkConfig retrieves framework configuration from database
func (fd *FrameworkDetector) GetFrameworkConfig() (*models.FrameworkConfig, error) {
	config := &models.FrameworkConfig{}

	err := fd.db.QueryRow(`
		SELECT test_framework, bdd_framework, api_framework, selenium_version,
		       uses_page_object_model, uses_page_factory, uses_screenplay,
		       testng_xml_path, pom_file_path, cucumber_features_path
		FROM framework_config
		WHERE project_root = ?
	`, fd.projectRoot).Scan(
		&config.TestFramework, &config.BDDFramework, &config.APIFramework,
		&config.SeleniumVersion, &config.UsesPageObjectModel, &config.UsesPageFactory,
		&config.UsesScreenplay, &config.TestNGXMLPath, &config.POMFilePath,
		&config.CucumberFeatures,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no framework config found for project: %s", fd.projectRoot)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetFrameworkCombination returns a human-readable framework combination string
func (fd *FrameworkDetector) GetFrameworkCombination(config *models.FrameworkConfig) string {
	parts := []string{}

	// Always starts with Selenium
	if config.SeleniumVersion != "" {
		parts = append(parts, fmt.Sprintf("Selenium %s", config.SeleniumVersion))
	} else {
		parts = append(parts, "Selenium")
	}

	// Add test framework
	if config.TestFramework != "" {
		parts = append(parts, config.TestFramework)
	}

	// Add BDD framework
	if config.BDDFramework != "" {
		parts = append(parts, config.BDDFramework)
	}

	// Add API framework
	if config.APIFramework != "" {
		parts = append(parts, config.APIFramework)
	}

	return strings.Join(parts, " + ")
}
