package ai

import (
	"regexp"
	"strings"
	"unicode"
)

// Tokenizer handles text tokenization for BM25 indexing
// Optimized for Java code with special handling for:
// - camelCase splitting
// - Method names
// - Annotations
// - Keywords
type Tokenizer struct {
	stopWords map[string]bool
}

// NewTokenizer creates a new tokenizer
func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		stopWords: buildStopWords(),
	}
}

// Tokenize converts text into tokens for BM25 indexing
func (t *Tokenizer) Tokenize(text string) []string {
	tokens := []string{}

	// Step 1: Extract Java-specific tokens first
	javaTokens := t.extractJavaTokens(text)
	tokens = append(tokens, javaTokens...)

	// Step 2: Basic word tokenization
	words := t.splitIntoWords(text)

	for _, word := range words {
		// Skip if already processed as Java token
		if t.isJavaToken(word) {
			continue
		}

		// Convert to lowercase
		word = strings.ToLower(word)

		// Skip stop words
		if t.stopWords[word] {
			continue
		}

		// Skip very short tokens
		if len(word) < 2 {
			continue
		}

		tokens = append(tokens, word)
	}

	// Step 3: Split camelCase into separate tokens
	camelCaseTokens := t.splitCamelCase(tokens)

	// Deduplicate while preserving order
	return t.deduplicate(camelCaseTokens)
}

// extractJavaTokens extracts Java-specific patterns
func (t *Tokenizer) extractJavaTokens(text string) []string {
	tokens := []string{}

	// Extract @Annotations
	annotationPattern := regexp.MustCompile(`@(\w+)`)
	annotations := annotationPattern.FindAllStringSubmatch(text, -1)
	for _, match := range annotations {
		tokens = append(tokens, strings.ToLower(match[1]))
	}

	// Extract method calls: methodName(
	methodPattern := regexp.MustCompile(`\b([a-z][a-zA-Z0-9]*)\s*\(`)
	methods := methodPattern.FindAllStringSubmatch(text, -1)
	for _, match := range methods {
		tokens = append(tokens, strings.ToLower(match[1]))
	}

	// Extract class names: public class ClassName
	classPattern := regexp.MustCompile(`\b(?:class|interface)\s+([A-Z][a-zA-Z0-9]*)`)
	classes := classPattern.FindAllStringSubmatch(text, -1)
	for _, match := range classes {
		tokens = append(tokens, strings.ToLower(match[1]))
	}

	// Extract field names with types: String fieldName
	fieldPattern := regexp.MustCompile(`\b(?:private|public|protected)\s+\w+\s+([a-z][a-zA-Z0-9]*)`)
	fields := fieldPattern.FindAllStringSubmatch(text, -1)
	for _, match := range fields {
		tokens = append(tokens, strings.ToLower(match[1]))
	}

	// Extract locator strategies: id="...", xpath="...", css="..."
	locatorStrategies := []string{"id", "name", "css", "xpath", "classname", "linktext", "tagname"}
	textLower := strings.ToLower(text)
	for _, strategy := range locatorStrategies {
		if strings.Contains(textLower, strategy) {
			tokens = append(tokens, strategy)
		}
	}

	return tokens
}

// splitIntoWords splits text into words
func (t *Tokenizer) splitIntoWords(text string) []string {
	// Replace non-alphanumeric characters with spaces (except @ for annotations)
	wordPattern := regexp.MustCompile(`[@a-zA-Z0-9]+`)
	matches := wordPattern.FindAllString(text, -1)

	words := []string{}
	for _, match := range matches {
		// Remove leading @ if present
		match = strings.TrimPrefix(match, "@")
		if len(match) > 0 {
			words = append(words, match)
		}
	}

	return words
}

// splitCamelCase splits camelCase words into separate tokens
// Example: "loginButton" -> ["loginbutton", "login", "button"]
func (t *Tokenizer) splitCamelCase(tokens []string) []string {
	result := []string{}

	for _, token := range tokens {
		result = append(result, token) // Keep original

		// Only split if it looks like camelCase
		if t.isCamelCase(token) {
			parts := t.splitCamelCaseWord(token)
			for _, part := range parts {
				if len(part) >= 2 {
					result = append(result, strings.ToLower(part))
				}
			}
		}
	}

	return result
}

// isCamelCase checks if a word is in camelCase format
func (t *Tokenizer) isCamelCase(word string) bool {
	if len(word) == 0 {
		return false
	}

	hasUpper := false
	hasLower := false

	for _, r := range word {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}

	return hasUpper && hasLower
}

// splitCamelCaseWord splits a camelCase word into parts
func (t *Tokenizer) splitCamelCaseWord(word string) []string {
	if len(word) == 0 {
		return []string{}
	}

	parts := []string{}
	currentPart := []rune{}

	runes := []rune(word)

	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			// Capital letter - start new part
			if len(currentPart) > 0 {
				parts = append(parts, string(currentPart))
				currentPart = []rune{}
			}
		}

		currentPart = append(currentPart, r)
	}

	// Add last part
	if len(currentPart) > 0 {
		parts = append(parts, string(currentPart))
	}

	return parts
}

// isJavaToken checks if a word is a Java-specific token
func (t *Tokenizer) isJavaToken(word string) bool {
	// Check if it's an annotation
	if strings.HasPrefix(word, "@") {
		return true
	}

	// Check if it's a method call
	if strings.Contains(word, "(") {
		return true
	}

	return false
}

// deduplicate removes duplicate tokens while preserving order
func (t *Tokenizer) deduplicate(tokens []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, token := range tokens {
		if !seen[token] {
			seen[token] = true
			result = append(result, token)
		}
	}

	return result
}

// buildStopWords returns common English stop words
func buildStopWords() map[string]bool {
	words := []string{
		// Common English stop words
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "from", "as", "is", "was", "are", "were", "be",
		"been", "being", "have", "has", "had", "do", "does", "did", "will",
		"would", "should", "could", "may", "might", "can", "this", "that",
		"these", "those", "i", "you", "he", "she", "it", "we", "they",

		// Common Java keywords (keep these - they're useful for search)
		// "public", "private", "protected", "class", "interface", "void",
		// "return", "new", "if", "else", "for", "while", "try", "catch",
	}

	stopWords := make(map[string]bool)
	for _, word := range words {
		stopWords[word] = true
	}

	return stopWords
}

// EstimateTokens estimates the number of tokens in text (for chunk sizing)
func (t *Tokenizer) EstimateTokens(text string) int {
	// Rough approximation: ~1.3 tokens per word for code
	words := strings.Fields(text)
	return int(float64(len(words)) * 1.3)
}
