package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CodeWriter writes generated code to files
type CodeWriter struct {
	baseDir         string
	overwritePolicy string // "ask", "always", "never"
	createDirs      bool
}

// NewCodeWriter creates a new code writer
func NewCodeWriter(baseDir string) *CodeWriter {
	return &CodeWriter{
		baseDir:         baseDir,
		overwritePolicy: "ask",
		createDirs:      true,
	}
}

// WriteCode writes generated code to a file
func (cw *CodeWriter) WriteCode(code *GeneratedCode, outputPath string) (*WriteResult, error) {
	result := &WriteResult{}

	// Determine full path
	fullPath := outputPath
	if !filepath.IsAbs(outputPath) {
		fullPath = filepath.Join(cw.baseDir, outputPath)
	}

	result.FilePath = fullPath

	// Check if file exists
	fileExists := false
	if _, err := os.Stat(fullPath); err == nil {
		fileExists = true
	}

	result.FileExists = fileExists

	// Handle overwrite policy
	if fileExists {
		switch cw.overwritePolicy {
		case "never":
			return result, fmt.Errorf("file already exists: %s (use --force to overwrite)", fullPath)
		case "ask":
			// Will be handled by caller (CLI will prompt user)
			result.RequiresConfirmation = true
			return result, nil
		}
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if cw.createDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return result, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(code.Code), 0644); err != nil {
		return result, fmt.Errorf("failed to write file: %w", err)
	}

	result.Success = true
	result.BytesWritten = len(code.Code)

	return result, nil
}

// WriteCodeInteractive writes code with user prompts for path
func (cw *CodeWriter) WriteCodeInteractive(code *GeneratedCode) (*WriteResult, error) {
	// Use suggested path or prompt user
	outputPath := code.SuggestedPath

	if outputPath == "" {
		// Fallback to class name
		if code.ClassName != "" {
			outputPath = fmt.Sprintf("%s.java", code.ClassName)
		} else {
			return nil, fmt.Errorf("cannot determine output path - no class name found")
		}
	}

	return cw.WriteCode(code, outputPath)
}

// BuildFileTree builds a tree of files to be created
func (cw *CodeWriter) BuildFileTree(codes []*GeneratedCode) string {
	var sb strings.Builder

	sb.WriteString("Files to be created:\n")
	for i, code := range codes {
		path := code.SuggestedPath
		if path == "" {
			path = fmt.Sprintf("Unknown_%d.java", i+1)
		}

		sb.WriteString(fmt.Sprintf("  %d. %s (%d lines)\n",
			i+1, path, countLines(code.Code)))
	}

	return sb.String()
}

// GetOutputPath determines the best output path for generated code
func (cw *CodeWriter) GetOutputPath(code *GeneratedCode, userSpecified string) string {
	// Priority:
	// 1. User-specified path
	// 2. Suggested path from parser
	// 3. Class name + .java
	// 4. Fallback to GeneratedCode.java

	if userSpecified != "" {
		return userSpecified
	}

	if code.SuggestedPath != "" {
		return code.SuggestedPath
	}

	if code.ClassName != "" {
		return fmt.Sprintf("%s.java", code.ClassName)
	}

	return "GeneratedCode.java"
}

// SetOverwritePolicy sets the policy for handling existing files
func (cw *CodeWriter) SetOverwritePolicy(policy string) {
	cw.overwritePolicy = policy
}

// FormatCodePreview formats code for display before writing
func (cw *CodeWriter) FormatCodePreview(code *GeneratedCode, maxLines int) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("📄 Generated Code: %s\n", code.ClassName))
	sb.WriteString("═══════════════════════════════════════════════════════════\n\n")

	lines := strings.Split(code.Code, "\n")

	if len(lines) <= maxLines {
		// Show full code
		sb.WriteString(code.Code)
	} else {
		// Show truncated
		for i := 0; i < maxLines; i++ {
			sb.WriteString(fmt.Sprintf("%3d | %s\n", i+1, lines[i]))
		}
		sb.WriteString(fmt.Sprintf("\n... (%d more lines)\n", len(lines)-maxLines))
	}

	sb.WriteString("\n═══════════════════════════════════════════════════════════\n")

	return sb.String()
}

// WriteResult contains the result of a write operation
type WriteResult struct {
	Success              bool
	FilePath             string
	FileExists           bool
	RequiresConfirmation bool
	BytesWritten         int
}

// countLines counts the number of lines in code
func countLines(code string) int {
	return len(strings.Split(code, "\n"))
}
