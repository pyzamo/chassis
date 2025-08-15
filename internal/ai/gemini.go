// Package ai provides AI integration for analyzing and extracting project skeletons
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GeminiClient handles communication with Google Gemini API
type GeminiClient struct {
	apiKey string
	client *http.Client
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient() (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set. Please set it with your API key from https://aistudio.google.com/app/apikey")
	}

	return &GeminiClient{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ExtractSkeleton sends the directory structure to Gemini and gets back a generalized skeleton
func (c *GeminiClient) ExtractSkeleton(treeStructure string, projectType string) (string, error) {
	// Prepare the prompt
	prompt := c.buildPrompt(treeStructure, projectType)

	// Call Gemini API
	response, err := c.callGeminiAPI(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}

	// Extract and clean the skeleton from response
	skeleton := c.extractSkeletonFromResponse(response)
	return skeleton, nil
}

// buildPrompt creates the prompt for Gemini
func (c *GeminiClient) buildPrompt(treeStructure string, projectType string) string {
	return fmt.Sprintf(`Analyze this project structure and extract a generalized, reusable scaffolding template.

IMPORTANT RULES:
1. Replace specific file names with generic descriptive names
2. Keep folder structure but make it template-like
3. Group similar files into representative examples
4. Use comments (lines starting with #) to explain sections
5. Output MUST be in plain-text tree format with 2-space indentation
6. Directories end with /
7. Keep only the essential scaffolding structure
8. Remove project-specific names (like "UserController" -> just "controllers/")
9. If you see multiple similar files, represent them with one or two examples
10. Focus on the architectural pattern, not specific implementations

Project type (if identifiable): %s

Current structure:
%s

Return ONLY the generalized skeleton in tree format, nothing else. Start directly with the root folder name.`, projectType, treeStructure)
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content represents content in the request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a part of the content
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// callGeminiAPI makes the actual API call to Gemini
func (c *GeminiClient) callGeminiAPI(prompt string) (string, error) {
	// Gemini API endpoint for Gemini 2.0 Flash
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

	// Prepare request body
	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", c.apiKey)

	// Make the request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	// Extract text from response
	if len(geminiResp.Candidates) > 0 &&
		len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response text from Gemini")
}

// extractSkeletonFromResponse cleans and extracts the skeleton from Gemini's response
func (c *GeminiClient) extractSkeletonFromResponse(response string) string {
	// Clean up the response
	lines := strings.Split(response, "\n")
	var cleanedLines []string
	inCodeBlock := false

	for _, line := range lines {
		// Skip markdown code blocks
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip empty lines at the beginning and end
		if !inCodeBlock && strings.TrimSpace(line) != "" {
			// Remove any markdown formatting
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			cleanedLines = append(cleanedLines, line)
		}
	}

	// Join the lines back
	result := strings.Join(cleanedLines, "\n")

	// Trim any leading/trailing whitespace
	result = strings.TrimSpace(result)

	return result
}

// DetectProjectType attempts to identify the project type from the structure
func DetectProjectType(treeStructure string) string {
	lower := strings.ToLower(treeStructure)

	// Check for various project indicators
	switch {
	case strings.Contains(lower, "package.json"):
		if strings.Contains(lower, "react") || strings.Contains(lower, "components/") {
			return "React/JavaScript application"
		}
		if strings.Contains(lower, "vue") {
			return "Vue.js application"
		}
		if strings.Contains(lower, "angular") {
			return "Angular application"
		}
		return "Node.js/JavaScript project"
	case strings.Contains(lower, "go.mod"):
		return "Go project"
	case strings.Contains(lower, "cargo.toml"):
		return "Rust project"
	case strings.Contains(lower, "requirements.txt") || strings.Contains(lower, "setup.py"):
		if strings.Contains(lower, "django") || strings.Contains(lower, "manage.py") {
			return "Django project"
		}
		if strings.Contains(lower, "flask") {
			return "Flask project"
		}
		return "Python project"
	case strings.Contains(lower, "pom.xml"):
		return "Java/Maven project"
	case strings.Contains(lower, "build.gradle"):
		return "Java/Gradle project"
	case strings.Contains(lower, "gemfile"):
		if strings.Contains(lower, "rails") {
			return "Ruby on Rails project"
		}
		return "Ruby project"
	case strings.Contains(lower, ".csproj"):
		return "C#/.NET project"
	case strings.Contains(lower, "composer.json"):
		if strings.Contains(lower, "laravel") {
			return "Laravel project"
		}
		return "PHP project"
	default:
		return "Unknown project type"
	}
}
