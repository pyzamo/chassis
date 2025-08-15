package analyze

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pyzamo/chassis/internal/parse"
)

// LocalAnalyzer analyzes local directory structures
type LocalAnalyzer struct {
	sourcePath string
	maxDepth   int
	filter     *Filter
}

// NewLocalAnalyzer creates a new local directory analyzer
func NewLocalAnalyzer(sourcePath string, maxDepth int) *LocalAnalyzer {
	return &LocalAnalyzer{
		sourcePath: sourcePath,
		maxDepth:   maxDepth,
		filter:     NewFilter(),
	}
}

// Analyze performs the analysis of the local directory
func (a *LocalAnalyzer) Analyze() (*Result, error) {
	// Check if source exists
	info, err := os.Stat(a.sourcePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access source: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("source must be a directory: %s", a.sourcePath)
	}

	// Get the base name for the root node
	baseName := filepath.Base(a.sourcePath)
	if baseName == "." || baseName == "/" {
		// If analyzing current directory or root, use a meaningful name
		absPath, _ := filepath.Abs(a.sourcePath)
		baseName = filepath.Base(absPath)
	}

	result := &Result{
		Nodes: []*parse.Node{},
	}

	// Create root node
	rootNode := &parse.Node{
		Name:  baseName,
		IsDir: true,
		Path:  baseName,
	}

	// Walk the directory tree
	walkResult := &walkResult{
		filter: a.filter,
		result: result,
	}

	if err := a.walkDirectory(a.sourcePath, rootNode, 1, walkResult); err != nil {
		return nil, err
	}

	// Only add root if it has children or we're analyzing an empty directory
	if len(rootNode.Children) > 0 || (result.DirCount == 0 && result.FileCount == 0) {
		result.Nodes = append(result.Nodes, rootNode)
		result.DirCount++ // Count the root directory
	}

	return result, nil
}

// walkResult holds state during directory walking
type walkResult struct {
	filter *Filter
	result *Result
}

// walkDirectory recursively walks the directory tree
func (a *LocalAnalyzer) walkDirectory(dirPath string, parentNode *parse.Node, currentDepth int, walkResult *walkResult) error {
	if currentDepth > a.maxDepth {
		return nil // Stop at max depth
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Skip directories we can't read (permissions, etc.)
		return nil
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dirPath, name)
		walkResult.result.TotalScanned++

		// Check if should filter
		if walkResult.filter.ShouldFilter(name, entry.IsDir()) {
			walkResult.result.FilteredCount++
			continue
		}

		// Create node
		node := &parse.Node{
			Name:  name,
			IsDir: entry.IsDir(),
			Path:  filepath.Join(parentNode.Path, name),
		}

		// Add to parent
		parentNode.Children = append(parentNode.Children, node)

		if entry.IsDir() {
			walkResult.result.DirCount++
			// Recursively walk subdirectory
			if err := a.walkDirectory(fullPath, node, currentDepth+1, walkResult); err != nil {
				return err
			}
		} else {
			walkResult.result.FileCount++
		}
	}

	return nil
}
