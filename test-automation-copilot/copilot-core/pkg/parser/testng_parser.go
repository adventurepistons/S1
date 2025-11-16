package parser

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

// TestNGParser parses TestNG suite XML files
type TestNGParser struct{}

type TestNGSuite struct {
	XMLName      xml.Name      `xml:"suite"`
	Name         string        `xml:"name,attr"`
	Parallel     string        `xml:"parallel,attr,omitempty"`
	ThreadCount  int           `xml:"thread-count,attr,omitempty"`
	Verbose      int           `xml:"verbose,attr,omitempty"`
	Tests        []TestNGTest  `xml:"test"`
	Listeners    []Listener    `xml:"listeners>listener"`
	Parameters   []Parameter   `xml:"parameter"`
}

type TestNGTest struct {
	Name       string          `xml:"name,attr"`
	Parallel   string          `xml:"parallel,attr,omitempty"`
	Enabled    string          `xml:"enabled,attr,omitempty"`
	Classes    []TestNGClass   `xml:"classes>class"`
	Groups     *Groups         `xml:"groups,omitempty"`
	Parameters []Parameter     `xml:"parameter"`
}

type TestNGClass struct {
	Name    string           `xml:"name,attr"`
	Methods []TestNGMethod   `xml:"methods>include"`
}

type TestNGMethod struct {
	Name string `xml:"name,attr"`
}

type Groups struct {
	Run      []GroupRun `xml:"run>include"`
	Define   []GroupDef `xml:"define>group"`
}

type GroupRun struct {
	Name string `xml:"name,attr"`
}

type GroupDef struct {
	Name     string   `xml:"name,attr"`
	Includes []string `xml:"include>name,attr"`
}

type Listener struct {
	ClassName string `xml:"class-name,attr"`
}

type Parameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func NewTestNGParser() *TestNGParser {
	return &TestNGParser{}
}

// ParseFile parses a testng.xml file
func (p *TestNGParser) ParseFile(filePath string) (*TestNGSuite, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var suite TestNGSuite
	err = xml.Unmarshal(data, &suite)
	if err != nil {
		return nil, err
	}

	return &suite, nil
}

// GetAllTestClasses returns all test classes from the suite
func (s *TestNGSuite) GetAllTestClasses() []string {
	var classes []string
	for _, test := range s.Tests {
		for _, class := range test.Classes {
			classes = append(classes, class.Name)
		}
	}
	return classes
}

// GetTestByName finds a test by name
func (s *TestNGSuite) GetTestByName(name string) *TestNGTest {
	for _, test := range s.Tests {
		if test.Name == name {
			return &test
		}
	}
	return nil
}

// IsParallelExecution checks if parallel execution is enabled
func (s *TestNGSuite) IsParallelExecution() bool {
	return s.Parallel != "" && s.Parallel != "false"
}

// GetThreadCount returns the thread count for parallel execution
func (s *TestNGSuite) GetThreadCount() int {
	if s.ThreadCount > 0 {
		return s.ThreadCount
	}
	return 1
}

// GetAllGroups returns all group names
func (s *TestNGSuite) GetAllGroups() []string {
	var groups []string
	for _, test := range s.Tests {
		if test.Groups != nil {
			for _, run := range test.Groups.Run {
				groups = append(groups, run.Name)
			}
		}
	}
	return groups
}
