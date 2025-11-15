package parser

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

// PomParser parses Maven pom.xml files
type PomParser struct{}

type PomFile struct {
	XMLName      xml.Name     `xml:"project"`
	GroupID      string       `xml:"groupId"`
	ArtifactID   string       `xml:"artifactId"`
	Version      string       `xml:"version"`
	Name         string       `xml:"name"`
	Description  string       `xml:"description"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
	Plugins      []Plugin     `xml:"build>plugins>plugin"`
	Properties   Properties   `xml:"properties"`
	Profiles     []Profile    `xml:"profiles>profile"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope,omitempty"`
	Type       string `xml:"type,omitempty"`
}

type Plugin struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type Properties struct {
	// Common Maven properties
	JavaVersion          string `xml:"maven.compiler.source"`
	JavaTargetVersion    string `xml:"maven.compiler.target"`
	ProjectBuildEncoding string `xml:"project.build.sourceEncoding"`
	SeleniumVersion      string `xml:"selenium.version"`
	TestNGVersion        string `xml:"testng.version"`
	JUnitVersion         string `xml:"junit.version"`

	// Custom properties will be in RawXML
	RawXML []byte `xml:",innerxml"`
}

type Profile struct {
	ID           string       `xml:"id"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
	Properties   Properties   `xml:"properties"`
}

func NewPomParser() *PomParser {
	return &PomParser{}
}

// ParseFile parses a pom.xml file
func (p *PomParser) ParseFile(filePath string) (*PomFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var pom PomFile
	err = xml.Unmarshal(data, &pom)
	if err != nil {
		return nil, err
	}

	return &pom, nil
}

// GetDependencyByArtifact finds a specific dependency
func (p *PomFile) GetDependencyByArtifact(artifactID string) *Dependency {
	for _, dep := range p.Dependencies {
		if dep.ArtifactID == artifactID {
			return &dep
		}
	}
	return nil
}

// HasDependency checks if a dependency exists
func (p *PomFile) HasDependency(artifactID string) bool {
	return p.GetDependencyByArtifact(artifactID) != nil
}

// GetFramework detects the automation framework
func (p *PomFile) GetFramework() string {
	if p.HasDependency("selenium-java") {
		return "selenium-java"
	}
	if p.HasDependency("playwright") {
		return "playwright"
	}
	if p.HasDependency("rest-assured") {
		return "rest-assured"
	}
	return "unknown"
}

// GetTestRunner detects the test runner
func (p *PomFile) GetTestRunner() string {
	if p.HasDependency("testng") {
		return "testng"
	}
	if p.HasDependency("junit") || p.HasDependency("junit-jupiter") {
		return "junit"
	}
	if p.HasDependency("cucumber-java") {
		return "cucumber"
	}
	return "unknown"
}

// ListAllDependencies returns formatted dependency list
func (p *PomFile) ListAllDependencies() []string {
	var deps []string
	for _, dep := range p.Dependencies {
		depStr := dep.GroupID + ":" + dep.ArtifactID + ":" + dep.Version
		deps = append(deps, depStr)
	}
	return deps
}
