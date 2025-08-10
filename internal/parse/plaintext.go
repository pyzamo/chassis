package parse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// PlainTextParser parses plain-text tree format
type PlainTextParser struct {
	IndentWidth int  // Expected width of indentation (default 2)
	AutoDetect  bool // Auto-detect indent width from first indented line
}

// NewPlainTextParser creates a new plain-text parser
func NewPlainTextParser(indentWidth int) *PlainTextParser {
	return &PlainTextParser{
		IndentWidth: indentWidth,
		AutoDetect:  indentWidth <= 0, // Auto-detect if width is 0 or negative
	}
}

// Parse implements the Parser interface
func (p *PlainTextParser) Parse(reader io.Reader) ([]*Node, error) {
	scanner := bufio.NewScanner(reader)
	var lines []parsedLine
	lineNum := 0

	// First pass: read and parse all lines
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()

		// Skip empty lines
		if len(strings.TrimSpace(text)) == 0 {
			continue
		}

		// Parse the line
		parsed, err := p.parseLine(text, lineNum)
		if err != nil {
			return nil, err
		}

		// Skip comments
		if parsed.isComment {
			continue
		}

		lines = append(lines, parsed)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	if len(lines) == 0 {
		return []*Node{}, nil
	}

	// Auto-detect indentation if needed
	if p.AutoDetect {
		p.detectIndentation(lines)
	}

	// Build the tree structure
	return p.buildTree(lines)
}

// parsedLine represents a parsed line of input
type parsedLine struct {
	text       string // Original text
	content    string // Trimmed content
	indent     int    // Number of leading spaces/tabs
	indentChar rune   // ' ' or '\t'
	isDir      bool   // True if ends with /
	isComment  bool   // True if line is a comment
	lineNum    int    // Line number in source
}

// parseLine parses a single line of input
func (p *PlainTextParser) parseLine(text string, lineNum int) (parsedLine, error) {
	line := parsedLine{
		text:    text,
		lineNum: lineNum,
	}

	// Count leading whitespace
	for i, ch := range text {
		if ch == ' ' || ch == '\t' {
			if line.indentChar == 0 {
				line.indentChar = ch
			} else if line.indentChar != ch {
				return line, NewIndentationError(lineNum, i+1,
					int(line.indentChar), int(ch))
			}
			line.indent++
		} else {
			break
		}
	}

	// Get the content after indentation
	line.content = strings.TrimSpace(text)

	// Check for comments
	if strings.HasPrefix(line.content, "#") {
		line.isComment = true
		return line, nil
	}

	// Check if it's a directory
	if strings.HasSuffix(line.content, "/") {
		line.isDir = true
		line.content = strings.TrimSuffix(line.content, "/")
	}

	// Validate the name
	if line.content == "" {
		return line, NewParseError(lineNum, "empty name after trimming")
	}

	// Check for multiple slashes or invalid patterns
	if strings.Contains(line.content, "/") {
		return line, NewParseError(lineNum,
			fmt.Sprintf("invalid name '%s': names cannot contain '/' except at the end for directories",
				line.content))
	}

	return line, nil
}

// detectIndentation auto-detects the indentation width
func (p *PlainTextParser) detectIndentation(lines []parsedLine) {
	// Find the first indented line to detect indent width
	for i := 1; i < len(lines); i++ {
		if lines[i].indent > lines[i-1].indent {
			p.IndentWidth = lines[i].indent - lines[i-1].indent
			break
		}
	}

	// Default to 2 if we couldn't detect
	if p.IndentWidth <= 0 {
		p.IndentWidth = 2
	}
}

// buildTree builds the tree structure from parsed lines
func (p *PlainTextParser) buildTree(lines []parsedLine) ([]*Node, error) {
	if len(lines) == 0 {
		return []*Node{}, nil
	}

	// Check that first line has no indentation
	if lines[0].indent > 0 {
		return nil, NewParseError(lines[0].lineNum,
			"first line must not be indented")
	}

	// Build the tree using a stack-based approach
	var roots []*Node
	stack := []*stackItem{}

	for _, line := range lines {
		// Calculate depth based on indentation
		depth := 0
		if line.indent > 0 {
			if line.indent%p.IndentWidth != 0 {
				return nil, NewIndentationError(line.lineNum, line.indent,
					p.IndentWidth, line.indent%p.IndentWidth)
			}
			depth = line.indent / p.IndentWidth
		}

		// Create new node
		node := &Node{
			Name:  line.content,
			IsDir: line.isDir,
			Line:  line.lineNum,
		}

		// Pop stack to correct depth
		for len(stack) > depth {
			stack = stack[:len(stack)-1]
		}

		// Validate depth increase
		if depth > len(stack) {
			return nil, NewParseError(line.lineNum,
				fmt.Sprintf("invalid indentation: jumped from level %d to %d (can only increase by 1)",
					len(stack), depth))
		}

		// Add node to tree
		if depth == 0 {
			// Root level node
			roots = append(roots, node)
			stack = []*stackItem{{node: node, depth: 0}}
		} else {
			// Child node
			parent := stack[len(stack)-1].node
			if !parent.IsDir {
				return nil, NewParseError(line.lineNum,
					fmt.Sprintf("cannot add children to file '%s' (only directories can have children)",
						parent.Name))
			}
			parent.Children = append(parent.Children, node)
			node.Path = parent.Path + "/" + node.Name

			// Update stack
			if depth == len(stack)-1 {
				// Same level - replace last item
				stack[len(stack)-1] = &stackItem{node: node, depth: depth}
			} else {
				// Deeper level - push to stack
				stack = append(stack, &stackItem{node: node, depth: depth})
			}
		}
	}

	return roots, nil
}

// stackItem helps track the tree building state
type stackItem struct {
	node  *Node
	depth int
}

// ValidateIndentation checks if all lines use consistent indentation
func (p *PlainTextParser) ValidateIndentation(lines []string) error {
	indentChar := rune(0)

	for lineNum, text := range lines {
		// Skip empty lines and comments
		trimmed := strings.TrimSpace(text)
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check leading whitespace
		for i, ch := range text {
			if ch != ' ' && ch != '\t' {
				break
			}

			if indentChar == 0 {
				indentChar = ch
			} else if ch != indentChar {
				return NewIndentationError(lineNum+1, i+1,
					int(indentChar), int(ch))
			}
		}
	}

	return nil
}
