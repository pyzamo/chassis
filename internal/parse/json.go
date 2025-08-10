package parse

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// JSONParser parses JSON format layout files
type JSONParser struct{}

// NewJSONParser creates a new JSON parser
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Parse implements the Parser interface
func (p *JSONParser) Parse(reader io.Reader) ([]*Node, error) {
	// Read the entire input
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	// Parse JSON into a generic structure
	var content interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	// Handle empty JSON
	if content == nil {
		return []*Node{}, nil
	}

	// Convert to nodes
	switch v := content.(type) {
	case map[string]interface{}:
		// Root is an object - each key becomes a root node
		return p.parseObject(v, "")
	case []interface{}:
		// Root is an array - not supported for layout
		return nil, fmt.Errorf("JSON root must be an object, not an array")
	default:
		return nil, fmt.Errorf("unexpected JSON root type: %T", content)
	}
}

// parseObject converts a JSON object to nodes
func (p *JSONParser) parseObject(obj map[string]interface{}, parentPath string) ([]*Node, error) {
	var nodes []*Node

	// Sort keys for consistent ordering
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := obj[name]
		node := &Node{
			Name: name,
			Path: name,
		}

		if parentPath != "" {
			node.Path = parentPath + "/" + name
		}

		// Determine if it's a directory or file based on value
		switch v := value.(type) {
		case map[string]interface{}:
			// Object means directory
			node.IsDir = true

			// Check if it's an empty object
			if len(v) == 0 {
				// Empty object means empty directory
				node.Children = []*Node{}
			} else {
				// Parse children
				children, err := p.parseObject(v, node.Path)
				if err != nil {
					return nil, err
				}
				node.Children = children
			}

		case nil:
			// Null means file
			node.IsDir = false

		case string:
			// Empty string also means file
			if v == "" {
				node.IsDir = false
			} else {
				return nil, fmt.Errorf("unexpected string value for '%s': files should have null or empty string value, got %q", name, v)
			}

		case float64, bool, []interface{}:
			// These types are not valid for our layout
			return nil, fmt.Errorf("unexpected value type for '%s': %T (use null for files, {} for directories)", name, value)

		default:
			return nil, fmt.Errorf("unexpected value type for '%s': %T", name, value)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
