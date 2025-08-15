package analyze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pyzamo/chassis/internal/parse"
	"gopkg.in/yaml.v3"
)

// Exporter handles exporting nodes to different formats
type Exporter struct {
	nodes []*parse.Node
}

// NewExporter creates a new exporter
func NewExporter(nodes []*parse.Node) *Exporter {
	return &Exporter{
		nodes: nodes,
	}
}

// ToTree exports nodes as plain-text tree format
func (e *Exporter) ToTree() (string, error) {
	var buf bytes.Buffer

	// Sort root nodes for consistent output
	sortNodes(e.nodes)

	for _, node := range e.nodes {
		if err := e.writeTreeNode(&buf, node, "", true); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// writeTreeNode recursively writes a node in tree format
func (e *Exporter) writeTreeNode(buf *bytes.Buffer, node *parse.Node, indent string, isLast bool) error {
	// Write the node name
	name := node.Name
	if node.IsDir {
		name += "/"
	}

	// Don't add tree symbols for root level
	if indent == "" {
		buf.WriteString(name + "\n")
	} else {
		// Add tree drawing characters
		if isLast {
			buf.WriteString(indent + "└── " + name + "\n")
		} else {
			buf.WriteString(indent + "├── " + name + "\n")
		}
	}

	// Sort children for consistent output
	sortNodes(node.Children)

	// Process children
	for i, child := range node.Children {
		childIndent := indent
		if indent == "" {
			childIndent = ""
		} else {
			if isLast {
				childIndent += "    "
			} else {
				childIndent += "│   "
			}
		}

		isChildLast := (i == len(node.Children)-1)
		if err := e.writeTreeNode(buf, child, childIndent, isChildLast); err != nil {
			return err
		}
	}

	return nil
}

// ToTreeSimple exports nodes as simple indented format (compatible with build command)
func (e *Exporter) ToTreeSimple() (string, error) {
	var buf bytes.Buffer

	// Sort root nodes for consistent output
	sortNodes(e.nodes)

	for _, node := range e.nodes {
		if err := e.writeSimpleTreeNode(&buf, node, 0); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// writeSimpleTreeNode writes a node in simple indented format
func (e *Exporter) writeSimpleTreeNode(buf *bytes.Buffer, node *parse.Node, depth int) error {
	// Write indentation
	indent := strings.Repeat("  ", depth) // 2 spaces per level

	// Write the node name
	name := node.Name
	if node.IsDir {
		name += "/"
	}
	buf.WriteString(indent + name + "\n")

	// Sort children for consistent output
	sortNodes(node.Children)

	// Process children
	for _, child := range node.Children {
		if err := e.writeSimpleTreeNode(buf, child, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// ToYAML exports nodes as YAML format
func (e *Exporter) ToYAML() (string, error) {
	// Convert nodes to YAML structure
	yamlData := e.nodesToMap(e.nodes)

	// Marshal to YAML
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(data), nil
}

// ToJSON exports nodes as JSON format
func (e *Exporter) ToJSON() (string, error) {
	// Convert nodes to JSON structure
	jsonData := e.nodesToMap(e.nodes)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data) + "\n", nil
}

// nodesToMap converts nodes to a map structure for YAML/JSON
func (e *Exporter) nodesToMap(nodes []*parse.Node) map[string]interface{} {
	result := make(map[string]interface{})

	// Sort nodes for consistent output
	sortNodes(nodes)

	for _, node := range nodes {
		if node.IsDir {
			if len(node.Children) > 0 {
				// Directory with children
				result[node.Name] = e.nodesToMap(node.Children)
			} else {
				// Empty directory
				result[node.Name] = map[string]interface{}{}
			}
		} else {
			// File
			result[node.Name] = nil
		}
	}

	return result
}

// sortNodes sorts nodes alphabetically (directories first, then files)
func sortNodes(nodes []*parse.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		// If one is dir and other is file, dir comes first
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		// Otherwise alphabetical
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}

// ToTree actually returns the simple format for build compatibility
func (e *Exporter) ToTreeForBuild() (string, error) {
	return e.ToTreeSimple()
}
