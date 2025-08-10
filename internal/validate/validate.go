// Package validate handles validation of the parsed tree structure
package validate

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pyzamo/chassis-cli/internal/parse"
)

// ValidationError represents a validation error
type ValidationError struct {
	Path    string // Path where error occurred
	Message string // Error message
	Type    ErrorType
}

// ErrorType represents the type of validation error
type ErrorType int

const (
	ErrorDuplicatePath ErrorType = iota
	ErrorPathTooLong
	ErrorPathTraversal
	ErrorInvalidCharacters
	ErrorReservedName
)

func (e *ValidationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("validation error at '%s': %s", e.Path, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationResult contains all validation errors found
type ValidationResult struct {
	Errors []error
	Valid  bool
}

// Validate checks the tree for syntax and semantic errors
func Validate(nodes []*parse.Node) error {
	v := &validator{
		paths:  make(map[string]bool),
		errors: []error{},
	}

	for _, node := range nodes {
		v.validateNode(node, "")
	}

	if len(v.errors) > 0 {
		// Return first error for now, could be enhanced to return all
		return v.errors[0]
	}

	return nil
}

// validator holds validation state
type validator struct {
	paths  map[string]bool // Track paths for duplicate detection
	errors []error
}

// validateNode recursively validates a node and its children
func (v *validator) validateNode(node *parse.Node, parentPath string) {
	if node == nil {
		return
	}

	// Build full path
	fullPath := filepath.Join(parentPath, node.Name)
	node.Path = fullPath // Update node's path

	// Check for empty name
	if node.Name == "" {
		v.errors = append(v.errors, &ValidationError{
			Path:    parentPath,
			Message: "empty name not allowed",
			Type:    ErrorInvalidCharacters,
		})
		return
	}

	// Check for path traversal attempts
	// Reject if name is ".." or starts with "../"
	if node.Name == ".." || strings.HasPrefix(node.Name, "../") {
		v.errors = append(v.errors, &ValidationError{
			Path:    fullPath,
			Message: fmt.Sprintf("path traversal not allowed: '%s'", node.Name),
			Type:    ErrorPathTraversal,
		})
		return
	}

	// Check for invalid characters
	if err := v.validatePathCharacters(node.Name); err != nil {
		v.errors = append(v.errors, &ValidationError{
			Path:    fullPath,
			Message: err.Error(),
			Type:    ErrorInvalidCharacters,
		})
		return
	}

	// Check for reserved names (Windows)
	if runtime.GOOS == "windows" {
		if err := v.validateWindowsReservedNames(node.Name); err != nil {
			v.errors = append(v.errors, &ValidationError{
				Path:    fullPath,
				Message: err.Error(),
				Type:    ErrorReservedName,
			})
			return
		}
	}

	// Check path length limits
	if err := v.validatePathLength(fullPath); err != nil {
		v.errors = append(v.errors, &ValidationError{
			Path:    fullPath,
			Message: err.Error(),
			Type:    ErrorPathTooLong,
		})
		return
	}

	// Check for duplicates (case-insensitive on Windows)
	pathKey := fullPath
	if runtime.GOOS == "windows" {
		pathKey = strings.ToLower(fullPath)
	}

	if v.paths[pathKey] {
		v.errors = append(v.errors, &ValidationError{
			Path:    fullPath,
			Message: "duplicate path",
			Type:    ErrorDuplicatePath,
		})
		return
	}
	v.paths[pathKey] = true

	// Validate children
	for _, child := range node.Children {
		v.validateNode(child, fullPath)
	}
}

// validatePathCharacters checks for invalid characters in path
func (v *validator) validatePathCharacters(name string) error {
	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return fmt.Errorf("null bytes not allowed in path")
	}

	// Platform-specific checks
	if runtime.GOOS == "windows" {
		// Windows forbidden characters
		forbidden := `<>:"|?*`
		for _, char := range forbidden {
			if strings.ContainsRune(name, char) {
				return fmt.Errorf("character '%c' not allowed in Windows paths", char)
			}
		}
		// Check for trailing dots or spaces (Windows strips these)
		if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
			return fmt.Errorf("paths cannot end with dots or spaces on Windows")
		}
	}

	// Check for control characters
	for _, r := range name {
		if r < 32 {
			return fmt.Errorf("control characters not allowed in paths")
		}
	}

	return nil
}

// validateWindowsReservedNames checks for Windows reserved filenames
func (v *validator) validateWindowsReservedNames(name string) error {
	reserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	// Check base name without extension
	baseName := strings.ToUpper(strings.TrimSuffix(name, filepath.Ext(name)))
	for _, reserved := range reserved {
		if baseName == reserved {
			return fmt.Errorf("'%s' is a reserved name on Windows", name)
		}
	}

	return nil
}

// validatePathLength checks if path exceeds OS limits
func (v *validator) validatePathLength(path string) error {
	const (
		maxPathWindows = 260
		maxPathUnix    = 4096
	)

	maxLen := maxPathUnix
	if runtime.GOOS == "windows" {
		maxLen = maxPathWindows
	}

	if len(path) > maxLen {
		return fmt.Errorf("path length %d exceeds maximum of %d", len(path), maxLen)
	}

	return nil
}
