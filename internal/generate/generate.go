// Package generate handles creation of the actual filesystem structure
package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyzamo/chassis/internal/fsutil"
	"github.com/pyzamo/chassis/internal/parse"
)

// Result contains statistics about the generation process
type Result struct {
	Created      int      // Number of files/directories created
	Skipped      int      // Number of files/directories skipped (already exist)
	Errors       []string // Any errors encountered
	CreatedPaths []string // List of successfully created paths
	SkippedPaths []string // List of skipped paths
}

// Options configures the generation process
type Options struct {
	TargetDir string // Target directory for generation
	Verbose   bool   // Print detailed output
	DryRun    bool   // Preview changes without creating files (future enhancement)
	Force     bool   // Overwrite existing files (future enhancement)
}

// Generator handles the filesystem generation
type Generator struct {
	options Options
	result  *Result
	logger  Logger
}

// Logger interface for output
type Logger interface {
	Info(format string, args ...interface{})
	Verbose(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// ConsoleLogger implements Logger for console output
type ConsoleLogger struct {
	VerboseMode bool
}

func (l *ConsoleLogger) Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (l *ConsoleLogger) Verbose(format string, args ...interface{}) {
	if l.VerboseMode {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

func (l *ConsoleLogger) Warning(format string, args ...interface{}) {
	fmt.Printf("[WARNING] "+format+"\n", args...)
}

func (l *ConsoleLogger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// Generate creates the filesystem structure from the parsed nodes
func Generate(nodes []*parse.Node, targetDir string, verbose bool) (*Result, error) {
	gen := &Generator{
		options: Options{
			TargetDir: targetDir,
			Verbose:   verbose,
		},
		result: &Result{
			Errors:       []string{},
			CreatedPaths: []string{},
			SkippedPaths: []string{},
		},
		logger: &ConsoleLogger{VerboseMode: verbose},
	}

	return gen.Generate(nodes)
}

// Generate creates the filesystem structure
func (g *Generator) Generate(nodes []*parse.Node) (*Result, error) {
	// Ensure target directory exists
	targetAbs, err := filepath.Abs(g.options.TargetDir)
	if err != nil {
		return g.result, fmt.Errorf("invalid target directory: %w", err)
	}

	if err := fsutil.SafeMkdir(targetAbs, fsutil.DirPerm); err != nil {
		return g.result, fmt.Errorf("failed to create target directory: %w", err)
	}

	g.logger.Verbose("Target directory: %s", targetAbs)

	// Process each root node
	for _, node := range nodes {
		if err := g.generateNode(node, targetAbs); err != nil {
			g.result.Errors = append(g.result.Errors, err.Error())
			// Continue processing other nodes even if one fails
		}
	}

	// Check if there were any critical errors
	if len(g.result.Errors) > 0 {
		return g.result, fmt.Errorf("generation completed with %d errors", len(g.result.Errors))
	}

	return g.result, nil
}

// generateNode recursively generates a node and its children
func (g *Generator) generateNode(node *parse.Node, parentPath string) error {
	fullPath := filepath.Join(parentPath, node.Name)

	// Check if path exists
	if fsutil.PathExists(fullPath) {
		g.result.Skipped++
		g.result.SkippedPaths = append(g.result.SkippedPaths, fullPath)
		g.logger.Warning("SKIP: %s (already exists)", fullPath)

		// If it's a directory and it exists, still process children
		if node.IsDir {
			for _, child := range node.Children {
				if err := g.generateNode(child, fullPath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Create the path
	if node.IsDir {
		// Create directory
		if err := fsutil.SafeMkdir(fullPath, fsutil.DirPerm); err != nil {
			g.result.Errors = append(g.result.Errors, fmt.Sprintf("failed to create directory %s: %v", fullPath, err))
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
		g.logger.Verbose("CREATE: %s/", fullPath)
		g.result.Created++
		g.result.CreatedPaths = append(g.result.CreatedPaths, fullPath)

		// Process children
		for _, child := range node.Children {
			if err := g.generateNode(child, fullPath); err != nil {
				return err
			}
		}
	} else {
		// Ensure parent directory exists
		if err := fsutil.EnsureDir(fullPath); err != nil {
			g.result.Errors = append(g.result.Errors, fmt.Sprintf("failed to create parent directory for %s: %v", fullPath, err))
			return fmt.Errorf("failed to create parent directory for %s: %w", fullPath, err)
		}

		// Create empty file
		file, err := fsutil.SafeCreateFile(fullPath, fsutil.FilePerm)
		if err != nil {
			// Check if it's because the file exists (race condition)
			if strings.Contains(err.Error(), "already exists") {
				g.result.Skipped++
				g.result.SkippedPaths = append(g.result.SkippedPaths, fullPath)
				g.logger.Warning("SKIP: %s (already exists)", fullPath)
				return nil
			}
			g.result.Errors = append(g.result.Errors, fmt.Sprintf("failed to create file %s: %v", fullPath, err))
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
		file.Close()

		g.logger.Verbose("CREATE: %s", fullPath)
		g.result.Created++
		g.result.CreatedPaths = append(g.result.CreatedPaths, fullPath)
	}

	return nil
}

// PrintSummary prints a summary of the generation results
func (r *Result) PrintSummary() {
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Created: %d\n", r.Created)
	fmt.Printf("  Skipped: %d\n", r.Skipped)
	if len(r.Errors) > 0 {
		fmt.Printf("  Errors:  %d\n", len(r.Errors))
		for _, err := range r.Errors {
			fmt.Printf("    - %s\n", err)
		}
	}
}
