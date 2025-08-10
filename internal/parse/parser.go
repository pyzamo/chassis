// Package parse handles parsing of various layout file formats
package parse

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Node represents a single entry in the directory tree
type Node struct {
	Name     string  // Name of the file or directory
	IsDir    bool    // True if this is a directory
	Children []*Node // Child nodes (only for directories)
	Path     string  // Full path from root (for error reporting)
	Line     int     // Line number in source file (for error reporting)
}

// Parser is the interface that all format parsers must implement
type Parser interface {
	Parse(reader io.Reader) ([]*Node, error)
}

// Format represents the supported file formats
type Format int

const (
	FormatUnknown Format = iota
	FormatPlainText
	FormatYAML
	FormatJSON
)

// String returns the string representation of the format
func (f Format) String() string {
	switch f {
	case FormatPlainText:
		return "plain-text"
	case FormatYAML:
		return "YAML"
	case FormatJSON:
		return "JSON"
	default:
		return "unknown"
	}
}

// DetectFormat determines the format based on file extension
func DetectFormat(filename string) Format {
	if filename == "-" {
		// stdin defaults to plain-text
		return FormatPlainText
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".tree":
		return FormatPlainText
	case ".yaml", ".yml":
		return FormatYAML
	case ".json":
		return FormatJSON
	default:
		return FormatUnknown
	}
}

// Parse reads from the reader and returns the parsed tree structure
func Parse(reader io.Reader, format Format) ([]*Node, error) {
	var parser Parser

	switch format {
	case FormatPlainText:
		parser = NewPlainTextParser(2) // Default to 2-space indentation
	case FormatYAML:
		parser = NewYAMLParser()
	case FormatJSON:
		parser = NewJSONParser()
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return parser.Parse(reader)
}

// ParseWithIndent reads from the reader with a specific indent width for plain-text
func ParseWithIndent(reader io.Reader, format Format, indentWidth int) ([]*Node, error) {
	var parser Parser

	switch format {
	case FormatPlainText:
		parser = NewPlainTextParser(indentWidth)
	case FormatYAML:
		parser = NewYAMLParser()
	case FormatJSON:
		parser = NewJSONParser()
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return parser.Parse(reader)
}

// Helper methods for Node

// AddChild adds a child node and returns it
func (n *Node) AddChild(name string, isDir bool) *Node {
	child := &Node{
		Name:  name,
		IsDir: isDir,
		Path:  filepath.Join(n.Path, name),
	}
	n.Children = append(n.Children, child)
	return child
}

// FindChild finds a child by name
func (n *Node) FindChild(name string) *Node {
	for _, child := range n.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// Walk recursively walks the tree, calling fn for each node
func (n *Node) Walk(fn func(*Node) error) error {
	if err := fn(n); err != nil {
		return err
	}
	for _, child := range n.Children {
		if err := child.Walk(fn); err != nil {
			return err
		}
	}
	return nil
}

// CountNodes returns the total number of nodes in the tree
func (n *Node) CountNodes() int {
	count := 1
	for _, child := range n.Children {
		count += child.CountNodes()
	}
	return count
}
