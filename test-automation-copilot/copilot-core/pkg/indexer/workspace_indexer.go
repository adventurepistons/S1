package indexer

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/copilot-core/pkg/database"
)

// WorkspaceIndexer indexes an entire workspace
type WorkspaceIndexer struct {
	db              *database.DB
	workspacePath   string
	fileRepo        *database.FileRepository
	classRepo       *database.ClassRepository
	methodRepo      *database.MethodRepository
	depRepo         *database.DependencyRepository

	javaIndexer     *JavaIndexer
	gherkinIndexer  *GherkinIndexer
	xmlIndexer      *XMLIndexer

	mu               sync.Mutex
	indexedFiles     map[string]string // path -> hash
	indexingStats    *IndexingStats
	progressCallback func(stats *IndexingStats)
}

// IndexingStats tracks indexing progress
type IndexingStats struct {
	TotalFiles       int
	IndexedFiles     int
	SkippedFiles     int
	FailedFiles      int
	JavaFiles        int
	GherkinFiles     int
	XMLFiles         int
	TotalClasses     int
	TotalMethods     int
	TotalFields      int
	TotalDependencies int
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
}

// IndexerConfig configures the workspace indexer
type IndexerConfig struct {
	WorkspacePath     string
	IncludePatterns   []string // e.g., ["*.java", "*.feature", "pom.xml"]
	ExcludePatterns   []string // e.g., ["target/*", "build/*"]
	Parallel          bool
	MaxWorkers        int
	ProgressCallback  func(stats *IndexingStats)
}

// NewWorkspaceIndexer creates a new workspace indexer
func NewWorkspaceIndexer(db *database.DB, config IndexerConfig) *WorkspaceIndexer {
	return &WorkspaceIndexer{
		db:               db,
		workspacePath:    config.WorkspacePath,
		fileRepo:         database.NewFileRepository(db),
		classRepo:        database.NewClassRepository(db),
		methodRepo:       database.NewMethodRepository(db),
		depRepo:          database.NewDependencyRepository(db),
		javaIndexer:      NewJavaIndexer(db),
		gherkinIndexer:   NewGherkinIndexer(db),
		xmlIndexer:       NewXMLIndexer(db),
		indexedFiles:     make(map[string]string),
		indexingStats:    &IndexingStats{},
		progressCallback: config.ProgressCallback, // Store the callback
	}
}

// IndexWorkspace indexes the entire workspace
func (idx *WorkspaceIndexer) IndexWorkspace() (*IndexingStats, error) {
	idx.indexingStats = &IndexingStats{
		StartTime: time.Now(),
	}

	fmt.Printf("🔍 Indexing workspace: %s\n", idx.workspacePath)

	// Find all relevant files
	files, err := idx.findFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	idx.indexingStats.TotalFiles = len(files)
	fmt.Printf("📁 Found %d files to index\n", len(files))

	// Call progress callback at start
	if idx.progressCallback != nil {
		idx.progressCallback(idx.indexingStats)
	}

	// Index each file
	for i, filePath := range files {
		if err := idx.indexFile(filePath); err != nil {
			fmt.Printf("⚠️  Failed to index %s: %v\n", filePath, err)
			idx.indexingStats.FailedFiles++
		} else {
			idx.indexingStats.IndexedFiles++
		}

		// Progress update every 10 files
		if (i+1)%10 == 0 {
			fmt.Printf("📊 Progress: %d/%d files indexed\n", i+1, len(files))

			// Call progress callback if provided
			if idx.progressCallback != nil {
				idx.progressCallback(idx.indexingStats)
			}
		}
	}

	// Build dependency graph
	fmt.Println("🔗 Building dependency graph...")
	if err := idx.buildDependencyGraph(); err != nil {
		fmt.Printf("⚠️  Failed to build dependency graph: %v\n", err)
	}

	idx.indexingStats.EndTime = time.Now()
	idx.indexingStats.Duration = idx.indexingStats.EndTime.Sub(idx.indexingStats.StartTime)

	// Get final stats from database
	dbStats, _ := idx.db.GetStats()
	idx.indexingStats.TotalClasses = int(dbStats.ClassCount)
	idx.indexingStats.TotalMethods = int(dbStats.MethodCount)
	idx.indexingStats.TotalFields = int(dbStats.FieldCount)
	idx.indexingStats.TotalDependencies = int(dbStats.DependencyCount)

	idx.printSummary()

	// Call progress callback at completion
	if idx.progressCallback != nil {
		idx.progressCallback(idx.indexingStats)
	}

	return idx.indexingStats, nil
}

// indexFile indexes a single file
func (idx *WorkspaceIndexer) indexFile(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	hash := computeHash(contentStr)

	// Check if file changed
	if existingHash, exists := idx.indexedFiles[filePath]; exists && existingHash == hash {
		idx.indexingStats.SkippedFiles++
		return nil // File unchanged, skip
	}

	// Determine file type and index accordingly
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".java":
		return idx.indexJavaFile(filePath, contentStr, hash)
	case ".feature":
		return idx.indexGherkinFile(filePath, contentStr, hash)
	case ".xml":
		return idx.indexXMLFile(filePath, contentStr, hash)
	default:
		idx.indexingStats.SkippedFiles++
		return nil
	}
}

// indexJavaFile indexes a Java file
func (idx *WorkspaceIndexer) indexJavaFile(filePath, content, hash string) error {
	idx.indexingStats.JavaFiles++

	// Create or update file record
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file := &database.File{
		Path:         filePath,
		Type:         "java",
		ContentHash:  hash,
		Content:      content,
		LastModified: fileInfo.ModTime().Unix(),
		IndexedAt:    database.Now(),
		SizeBytes:    fileInfo.Size(),
		LineCount:    countLines(content),
	}

	// Upsert file
	if err := idx.fileRepo.Upsert(file); err != nil {
		return fmt.Errorf("failed to upsert file: %w", err)
	}

	// Parse and index Java content
	if err := idx.javaIndexer.IndexFile(file, content); err != nil {
		return fmt.Errorf("failed to index Java file: %w", err)
	}

	idx.indexedFiles[filePath] = hash
	return nil
}

// indexGherkinFile indexes a Gherkin feature file
func (idx *WorkspaceIndexer) indexGherkinFile(filePath, content, hash string) error {
	idx.indexingStats.GherkinFiles++

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file := &database.File{
		Path:         filePath,
		Type:         "gherkin",
		ContentHash:  hash,
		Content:      content,
		LastModified: fileInfo.ModTime().Unix(),
		IndexedAt:    database.Now(),
		SizeBytes:    fileInfo.Size(),
		LineCount:    countLines(content),
	}

	if err := idx.fileRepo.Upsert(file); err != nil {
		return fmt.Errorf("failed to upsert file: %w", err)
	}

	if err := idx.gherkinIndexer.IndexFile(file, content); err != nil {
		return fmt.Errorf("failed to index Gherkin file: %w", err)
	}

	idx.indexedFiles[filePath] = hash
	return nil
}

// indexXMLFile indexes an XML file (pom.xml, testng.xml)
func (idx *WorkspaceIndexer) indexXMLFile(filePath, content, hash string) error {
	idx.indexingStats.XMLFiles++

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file := &database.File{
		Path:         filePath,
		Type:         "xml",
		ContentHash:  hash,
		Content:      content,
		LastModified: fileInfo.ModTime().Unix(),
		IndexedAt:    database.Now(),
		SizeBytes:    fileInfo.Size(),
		LineCount:    countLines(content),
	}

	if err := idx.fileRepo.Upsert(file); err != nil {
		return fmt.Errorf("failed to upsert file: %w", err)
	}

	if err := idx.xmlIndexer.IndexFile(file, content); err != nil {
		return fmt.Errorf("failed to index XML file: %w", err)
	}

	idx.indexedFiles[filePath] = hash
	return nil
}

// findFiles finds all relevant files in the workspace
func (idx *WorkspaceIndexer) findFiles() ([]string, error) {
	var files []string

	// Patterns to include
	includeExts := map[string]bool{
		".java":    true,
		".feature": true,
		".xml":     true,
	}

	// Patterns to exclude
	excludeDirs := map[string]bool{
		"target":       true,
		"build":        true,
		".git":         true,
		".idea":        true,
		"node_modules": true,
		".vscode":      true,
	}

	err := filepath.WalkDir(idx.workspacePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if d.IsDir() {
			if excludeDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be included
		ext := strings.ToLower(filepath.Ext(path))
		if includeExts[ext] {
			// For XML, only include pom.xml and testng.xml
			if ext == ".xml" {
				baseName := strings.ToLower(filepath.Base(path))
				if baseName != "pom.xml" && baseName != "testng.xml" {
					return nil
				}
			}
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// buildDependencyGraph builds relationships between classes
func (idx *WorkspaceIndexer) buildDependencyGraph() error {
	// This will be implemented by JavaIndexer
	// It will analyze imports and create dependency edges
	return idx.javaIndexer.BuildDependencyGraph()
}

// printSummary prints indexing summary
func (idx *WorkspaceIndexer) printSummary() {
	stats := idx.indexingStats
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Indexing Complete!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏱️  Duration: %v\n", stats.Duration)
	fmt.Printf("📁 Total Files: %d\n", stats.TotalFiles)
	fmt.Printf("✅ Indexed: %d\n", stats.IndexedFiles)
	fmt.Printf("⏭️  Skipped: %d\n", stats.SkippedFiles)
	fmt.Printf("❌ Failed: %d\n", stats.FailedFiles)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("☕ Java Files: %d\n", stats.JavaFiles)
	fmt.Printf("🥒 Gherkin Files: %d\n", stats.GherkinFiles)
	fmt.Printf("📄 XML Files: %d\n", stats.XMLFiles)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("📦 Classes: %d\n", stats.TotalClasses)
	fmt.Printf("⚡ Methods: %d\n", stats.TotalMethods)
	fmt.Printf("🔧 Fields: %d\n", stats.TotalFields)
	fmt.Printf("🔗 Dependencies: %d\n", stats.TotalDependencies)
	fmt.Println(strings.Repeat("=", 60))
}

// ReindexFile re-indexes a single file (useful for file watchers)
func (idx *WorkspaceIndexer) ReindexFile(filePath string) error {
	return idx.indexFile(filePath)
}

// GetStats returns current indexing statistics
func (idx *WorkspaceIndexer) GetStats() *IndexingStats {
	return idx.indexingStats
}

// Helper functions

func computeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func countLines(content string) int {
	return strings.Count(content, "\n") + 1
}
