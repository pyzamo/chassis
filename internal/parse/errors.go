package parse

import (
	"fmt"
)

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Line    int    // Line number where error occurred (1-indexed)
	Column  int    // Column number where error occurred (1-indexed)
	Message string // Error message
	File    string // File being parsed (optional)
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.File != "" {
		if e.Line > 0 && e.Column > 0 {
			return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
		} else if e.Line > 0 {
			return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
		}
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}

	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Message)
	} else if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return e.Message
}

// IndentationError represents an inconsistent indentation error
type IndentationError struct {
	ParseError
	Expected int // Expected indentation
	Got      int // Actual indentation
}

// Error implements the error interface
func (e *IndentationError) Error() string {
	expectedChar := "spaces"
	if e.Expected == '\t' {
		expectedChar = "tabs"
	}
	gotChar := "spaces"
	if e.Got == '\t' {
		gotChar = "tabs"
	}

	// Special case for mixed tabs/spaces
	if (e.Expected == ' ' && e.Got == '\t') || (e.Expected == '\t' && e.Got == ' ') {
		e.Message = fmt.Sprintf("inconsistent indentation: cannot mix %s and %s", expectedChar, gotChar)
	} else {
		e.Message = fmt.Sprintf("inconsistent indentation: expected %d %s, got %d %s", e.Expected, expectedChar, e.Got, gotChar)
	}
	return e.ParseError.Error()
}

// SyntaxError represents a syntax error in the input
type SyntaxError struct {
	ParseError
	Context string // Additional context about what was expected
}

// Error implements the error interface
func (e *SyntaxError) Error() string {
	if e.Context != "" {
		e.Message = fmt.Sprintf("syntax error: %s (expected %s)", e.Message, e.Context)
	} else {
		e.Message = fmt.Sprintf("syntax error: %s", e.Message)
	}
	return e.ParseError.Error()
}

// NewParseError creates a new ParseError
func NewParseError(line int, message string) *ParseError {
	return &ParseError{
		Line:    line,
		Message: message,
	}
}

// NewIndentationError creates a new IndentationError
func NewIndentationError(line, column, expected, got int) *IndentationError {
	return &IndentationError{
		ParseError: ParseError{
			Line:   line,
			Column: column,
		},
		Expected: expected,
		Got:      got,
	}
}

// NewSyntaxError creates a new SyntaxError
func NewSyntaxError(line int, message, context string) *SyntaxError {
	return &SyntaxError{
		ParseError: ParseError{
			Line:    line,
			Message: message,
		},
		Context: context,
	}
}
