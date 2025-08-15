// Package analyze provides functionality for analyzing directory structures
package analyze

import (
	"github.com/pyzamo/chassis/internal/parse"
)

// Result contains the analysis results
type Result struct {
	Nodes         []*parse.Node // The analyzed structure
	DirCount      int           // Number of directories found
	FileCount     int           // Number of files found
	FilteredCount int           // Number of items filtered out
	TotalScanned  int           // Total items scanned before filtering
}

// Analyzer is the interface for analyzing sources
type Analyzer interface {
	Analyze() (*Result, error)
}
