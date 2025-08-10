package parse

import (
	"fmt"
	"io"
	"sort"

	"gopkg.in/yaml.v3"
)

// YAMLParser parses YAML format layout files
type YAMLParser struct{}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// Parse implements the Parser interface
func (p *YAMLParser) Parse(reader io.Reader) ([]*Node, error) {
	// Read the entire input
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	// Parse YAML into a generic structure
	var content interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Handle empty YAML
	if content == nil {
		return []*Node{}, nil
	}

	// Convert to nodes
	switch v := content.(type) {
	case map[string]interface{}:
		// Root is a map - each key becomes a root node
		return p.parseMap(v, "")
	case []interface{}:
		// Root is an array - not supported for layout
		return nil, fmt.Errorf("YAML root must be an object, not an array")
	default:
		return nil, fmt.Errorf("unexpected YAML root type: %T", content)
	}
}

// parseMap converts a YAML map to nodes
func (p *YAMLParser) parseMap(m map[string]interface{}, parentPath string) ([]*Node, error) {
	var nodes []*Node

	// Sort keys for consistent ordering
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := m[name]
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
			// Map means directory
			node.IsDir = true

			// Parse children
			children, err := p.parseMap(v, node.Path)
			if err != nil {
				return nil, err
			}
			node.Children = children

		case nil:
			// Null means file
			node.IsDir = false

		case string:
			// Empty string also means file
			if v == "" {
				node.IsDir = false
			} else {
				return nil, fmt.Errorf("unexpected string value for '%s': files should have null or empty string value", name)
			}

		case map[interface{}]interface{}:
			// Sometimes YAML parses as map[interface{}]interface{}
			// Convert to map[string]interface{}
			converted := make(map[string]interface{})
			for k, val := range v {
				key, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("non-string key in YAML at '%s'", name)
				}
				converted[key] = val
			}

			node.IsDir = true
			children, err := p.parseMap(converted, node.Path)
			if err != nil {
				return nil, err
			}
			node.Children = children

		default:
			return nil, fmt.Errorf("unexpected value type for '%s': %T (use null for files, {} for empty directories)", name, value)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseYAMLNode is a helper to handle both map[string]interface{} and map[interface{}]interface{}
func (p *YAMLParser) parseYAMLNode(name string, value interface{}, parentPath string) (*Node, error) {
	node := &Node{
		Name: name,
		Path: name,
	}

	if parentPath != "" {
		node.Path = parentPath + "/" + name
	}

	switch v := value.(type) {
	case map[string]interface{}:
		node.IsDir = true
		children, err := p.parseMap(v, node.Path)
		if err != nil {
			return nil, err
		}
		node.Children = children

	case map[interface{}]interface{}:
		// Convert to map[string]interface{}
		converted := make(map[string]interface{})
		for k, val := range v {
			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("non-string key in YAML at '%s'", name)
			}
			converted[key] = val
		}

		node.IsDir = true
		children, err := p.parseMap(converted, node.Path)
		if err != nil {
			return nil, err
		}
		node.Children = children

	case nil:
		node.IsDir = false

	case string:
		// Empty string means file
		if v == "" {
			node.IsDir = false
		} else {
			return nil, fmt.Errorf("unexpected string value for '%s': files should have null or empty string value", name)
		}

	default:
		return nil, fmt.Errorf("unexpected value type for '%s': %T", name, value)
	}

	return node, nil
}
