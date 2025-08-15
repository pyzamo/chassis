// Package github provides GitHub repository analysis functionality
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pyzamo/chassis/internal/analyze"
	"github.com/pyzamo/chassis/internal/parse"
)

// GitHubAnalyzer analyzes GitHub repository structures
type GitHubAnalyzer struct {
	repoURL  string
	owner    string
	repo     string
	maxDepth int
}

// NewAnalyzer creates a new GitHub analyzer
func NewAnalyzer(repoURL string) *GitHubAnalyzer {
	// Parse the URL to extract owner and repo
	owner, repo := parseGitHubURL(repoURL)

	return &GitHubAnalyzer{
		repoURL:  repoURL,
		owner:    owner,
		repo:     repo,
		maxDepth: 5, // Default max depth
	}
}

// Analyze fetches and analyzes the GitHub repository structure
func (a *GitHubAnalyzer) Analyze() (*analyze.Result, error) {
	if a.owner == "" || a.repo == "" {
		return nil, fmt.Errorf("invalid GitHub URL: %s", a.repoURL)
	}

	// Fetch repository tree from GitHub API
	tree, err := a.fetchRepoTree()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Convert GitHub tree to our node structure
	result := &analyze.Result{
		Nodes: []*parse.Node{},
	}

	// Create root node with repo name
	rootNode := &parse.Node{
		Name:     a.repo,
		IsDir:    true,
		Path:     a.repo,
		Children: []*parse.Node{},
	}

	// Build node tree from GitHub response
	filter := analyze.NewFilter()
	for _, item := range tree.Tree {
		if a.shouldSkipItem(item, filter) {
			result.FilteredCount++
			continue
		}

		// Count the item
		result.TotalScanned++
		if item.Type == "tree" {
			result.DirCount++
		} else {
			result.FileCount++
		}

		// Add to node tree
		a.addItemToTree(rootNode, item)
	}

	if rootNode.Children != nil && len(rootNode.Children) > 0 {
		result.Nodes = append(result.Nodes, rootNode)
		result.DirCount++ // Count root
	}

	return result, nil
}

// GitHubTree represents the response from GitHub's tree API
type GitHubTree struct {
	Tree []GitHubTreeItem `json:"tree"`
}

// GitHubTreeItem represents a single item in the tree
type GitHubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"` // "blob" for files, "tree" for directories
	Size int    `json:"size"`
}

// fetchRepoTree fetches the repository tree from GitHub API
func (a *GitHubAnalyzer) fetchRepoTree() (*GitHubTree, error) {
	// Use GitHub API to get repository tree
	// Note: This uses the public API without authentication
	// Rate limit: 60 requests per hour for unauthenticated requests
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/HEAD?recursive=1", a.owner, a.repo)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "chassis-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s/%s", a.owner, a.repo)
	}
	if resp.StatusCode == 403 {
		// Check if it's rate limiting
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Please try again later")
		}
		return nil, fmt.Errorf("access forbidden: repository might be private")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tree GitHubTree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	return &tree, nil
}

// shouldSkipItem checks if an item should be filtered
func (a *GitHubAnalyzer) shouldSkipItem(item GitHubTreeItem, filter *analyze.Filter) bool {
	// Check depth limit
	depth := strings.Count(item.Path, "/")
	if depth >= a.maxDepth {
		return true
	}

	// Get the last component of the path
	parts := strings.Split(item.Path, "/")
	name := parts[len(parts)-1]

	// Check filter
	isDir := item.Type == "tree"
	return filter.ShouldFilter(name, isDir)
}

// addItemToTree adds a GitHub tree item to our node tree
func (a *GitHubAnalyzer) addItemToTree(root *parse.Node, item GitHubTreeItem) {
	parts := strings.Split(item.Path, "/")
	current := root

	// Navigate/create the path
	for i, part := range parts {
		isLastPart := (i == len(parts)-1)

		// Look for existing child
		var found *parse.Node
		for _, child := range current.Children {
			if child.Name == part {
				found = child
				break
			}
		}

		if found != nil {
			current = found
		} else {
			// Create new node
			node := &parse.Node{
				Name:  part,
				IsDir: !isLastPart || item.Type == "tree",
				Path:  item.Path,
			}
			current.Children = append(current.Children, node)
			current = node
		}
	}
}

// parseGitHubURL extracts owner and repo from various GitHub URL formats
func parseGitHubURL(url string) (owner, repo string) {
	// Remove protocol if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")
	url = strings.TrimPrefix(url, "github.com:")
	url = strings.TrimPrefix(url, "github.com/")

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Split by /
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}

	return "", ""
}
