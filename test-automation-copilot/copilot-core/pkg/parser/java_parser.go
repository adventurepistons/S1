package parser

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

type JavaParser struct {
	parser *sitter.Parser
}

type ParsedClass struct {
	ClassName   string        `json:"className"`
	FilePath    string        `json:"filePath"`
	Package     string        `json:"package"`
	Imports     []string      `json:"imports"`
	Fields      []FieldInfo   `json:"fields"`
	Methods     []MethodInfo  `json:"methods"`
	Annotations []string      `json:"annotations"`
	SuperClass  string        `json:"superClass,omitempty"`
	Type        string        `json:"type"` // pageObject, test, step, utility, unknown
}

type FieldInfo struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	LocatorType  string   `json:"locatorType,omitempty"`
	LocatorValue string   `json:"locatorValue,omitempty"`
	Annotations  []string `json:"annotations"`
}

type MethodInfo struct {
	Name        string          `json:"name"`
	ReturnType  string          `json:"returnType"`
	Parameters  []ParameterInfo `json:"parameters"`
	Annotations []string        `json:"annotations"`
	Body        string          `json:"body"`
}

type ParameterInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func NewJavaParser() *JavaParser {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	return &JavaParser{
		parser: parser,
	}
}

func (p *JavaParser) ParseFile(filePath string) (*ParsedClass, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	tree, err := p.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	result := p.extractClassInfo(tree.RootNode(), content, filePath)
	return result, nil
}

func (p *JavaParser) ParseDirectory(dirPath string) ([]*ParsedClass, error) {
	var results []*ParsedClass

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".java") {
			parsed, err := p.ParseFile(path)
			if err != nil {
				fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
				return nil // Continue with other files
			}
			results = append(results, parsed)
		}

		return nil
	})

	return results, err
}

func (p *JavaParser) extractClassInfo(node *sitter.Node, content []byte, filePath string) *ParsedClass {
	result := &ParsedClass{
		FilePath:    filePath,
		Imports:     []string{},
		Fields:      []FieldInfo{},
		Methods:     []MethodInfo{},
		Annotations: []string{},
	}

	// Extract package
	packageNode := findNode(node, "package_declaration")
	if packageNode != nil {
		pkgText := nodeText(packageNode, content)
		result.Package = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(pkgText, "package"), ";"))
	}

	// Extract imports
	importNodes := findAllNodes(node, "import_declaration")
	for _, importNode := range importNodes {
		importText := nodeText(importNode, content)
		importText = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(importText, "import"), ";"))
		result.Imports = append(result.Imports, importText)
	}

	// Extract class declaration
	classNode := findNode(node, "class_declaration")
	if classNode != nil {
		// Class name
		identifierNode := findNode(classNode, "identifier")
		if identifierNode != nil {
			result.ClassName = nodeText(identifierNode, content)
		}

		// Annotations
		result.Annotations = extractAnnotations(classNode, content)

		// Super class
		superclassNode := findNode(classNode, "superclass")
		if superclassNode != nil {
			typeNode := findNode(superclassNode, "type_identifier")
			if typeNode != nil {
				result.SuperClass = nodeText(typeNode, content)
			}
		}

		// Class body
		classBody := findNode(classNode, "class_body")
		if classBody != nil {
			result.Fields = extractFields(classBody, content)
			result.Methods = extractMethods(classBody, content)
		}
	}

	// Determine class type
	result.Type = determineClassType(result)

	return result
}

func extractFields(classBody *sitter.Node, content []byte) []FieldInfo {
	var fields []FieldInfo
	fieldNodes := findAllNodes(classBody, "field_declaration")

	for _, fieldNode := range fieldNodes {
		typeNode := findNode(fieldNode, "type")
		declaratorNode := findNode(fieldNode, "variable_declarator")

		if typeNode != nil && declaratorNode != nil {
			identifierNode := findNode(declaratorNode, "identifier")
			if identifierNode != nil {
				field := FieldInfo{
					Name:        nodeText(identifierNode, content),
					Type:        nodeText(typeNode, content),
					Annotations: extractAnnotations(fieldNode, content),
				}

				// Extract locator information
				locatorType, locatorValue := extractLocatorInfo(fieldNode, content)
				if locatorType != "" {
					field.LocatorType = locatorType
					field.LocatorValue = locatorValue
				}

				fields = append(fields, field)
			}
		}
	}

	return fields
}

func extractMethods(classBody *sitter.Node, content []byte) []MethodInfo {
	var methods []MethodInfo
	methodNodes := findAllNodes(classBody, "method_declaration")

	for _, methodNode := range methodNodes {
		nameNode := findNode(methodNode, "identifier")
		var typeNode *sitter.Node
		if tn := findNode(methodNode, "type"); tn != nil {
			typeNode = tn
		} else if vt := findNode(methodNode, "void_type"); vt != nil {
			typeNode = vt
		}

		if nameNode != nil && typeNode != nil {
			method := MethodInfo{
				Name:        nodeText(nameNode, content),
				ReturnType:  nodeText(typeNode, content),
				Parameters:  extractParameters(methodNode, content),
				Annotations: extractAnnotations(methodNode, content),
				Body:        getMethodBody(methodNode, content),
			}
			methods = append(methods, method)
		}
	}

	return methods
}

func extractParameters(methodNode *sitter.Node, content []byte) []ParameterInfo {
	var parameters []ParameterInfo
	paramsNode := findNode(methodNode, "formal_parameters")

	if paramsNode != nil {
		paramNodes := findAllNodes(paramsNode, "formal_parameter")
		for _, paramNode := range paramNodes {
			typeNode := findNode(paramNode, "type")
			identifierNode := findNode(paramNode, "identifier")

			if typeNode != nil && identifierNode != nil {
				parameters = append(parameters, ParameterInfo{
					Name: nodeText(identifierNode, content),
					Type: nodeText(typeNode, content),
				})
			}
		}
	}

	return parameters
}

func extractAnnotations(node *sitter.Node, content []byte) []string {
	var annotations []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "modifiers" {
			annotationNodes := findAllNodes(child, "annotation")
			for _, annotNode := range annotationNodes {
				annotations = append(annotations, nodeText(annotNode, content))
			}
		} else if child.Type() == "annotation" {
			annotations = append(annotations, nodeText(child, content))
		}
	}

	return annotations
}

func extractLocatorInfo(fieldNode *sitter.Node, content []byte) (string, string) {
	fieldText := nodeText(fieldNode, content)

	// Check for @FindBy annotation
	if strings.Contains(fieldText, "@FindBy") {
		// Pattern: @FindBy(id = "username")
		for _, locType := range []string{"id", "name", "css", "xpath", "className", "tagName", "linkText", "partialLinkText"} {
			pattern := fmt.Sprintf(`%s\s*=\s*"([^"]+)"`, locType)
			if idx := strings.Index(fieldText, locType+" ="); idx != -1 {
				start := strings.Index(fieldText[idx:], `"`) + idx + 1
				end := strings.Index(fieldText[start:], `"`) + start
				if start > idx && end > start {
					return fmt.Sprintf("@FindBy(%s)", locType), fieldText[start:end]
				}
			}
		}
	}

	// Check for By.xxx() pattern
	if strings.Contains(fieldText, "By.") {
		for _, locType := range []string{"id", "name", "cssSelector", "xpath", "className", "tagName", "linkText", "partialLinkText"} {
			pattern := fmt.Sprintf(`By\.%s\("([^"]+)"\)`, locType)
			if strings.Contains(fieldText, "By."+locType) {
				start := strings.Index(fieldText, "By."+locType+"(\"") + len("By."+locType+"(\"")
				end := strings.Index(fieldText[start:], `"`) + start
				if end > start {
					return fmt.Sprintf("By.%s", locType), fieldText[start:end]
				}
			}
		}
	}

	return "", ""
}

func getMethodBody(methodNode *sitter.Node, content []byte) string {
	bodyNode := findNode(methodNode, "block")
	if bodyNode != nil {
		return nodeText(bodyNode, content)
	}
	return ""
}

func determineClassType(class *ParsedClass) string {
	className := strings.ToLower(class.ClassName)
	filePath := strings.ToLower(class.FilePath)

	// Check for Page Object
	if strings.Contains(className, "page") || strings.Contains(filePath, "/pages/") {
		return "pageObject"
	}

	// Check for Test class
	if strings.Contains(className, "test") || strings.Contains(filePath, "/tests/") || strings.Contains(filePath, "/test/") {
		return "test"
	}

	// Check for Step definition (Cucumber)
	if strings.Contains(className, "step") || strings.Contains(filePath, "/steps/") {
		return "step"
	}
	for _, ann := range class.Annotations {
		if strings.Contains(ann, "@Given") || strings.Contains(ann, "@When") || strings.Contains(ann, "@Then") {
			return "step"
		}
	}

	// Check for utility class
	if strings.Contains(className, "util") ||
		strings.Contains(className, "helper") ||
		strings.Contains(className, "factory") ||
		strings.Contains(filePath, "/utils/") ||
		strings.Contains(filePath, "/utilities/") {
		return "utility"
	}

	return "unknown"
}

// Helper functions
func findNode(node *sitter.Node, nodeType string) *sitter.Node {
	if node.Type() == nodeType {
		return node
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if found := findNode(node.Child(i), nodeType); found != nil {
			return found
		}
	}

	return nil
}

func findAllNodes(node *sitter.Node, nodeType string) []*sitter.Node {
	var results []*sitter.Node

	if node.Type() == nodeType {
		results = append(results, node)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		results = append(results, findAllNodes(node.Child(i), nodeType)...)
	}

	return results
}

func nodeText(node *sitter.Node, content []byte) string {
	return string(content[node.StartByte():node.EndByte()])
}
