package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// mimicCmd represents the mimic command
var mimicCmd = &cobra.Command{
	Use:   "mimic <source> <target-dir>",
	Short: "Analyze and replicate directory structure from existing projects",
	Long: `Analyze an existing project directory or GitHub repository and create
a new project with the same directory structure.

Examples:
  # Mimic a local directory
  chassis mimic ./existing-project ./new-project
  
  # Mimic a GitHub repository
  chassis mimic https://github.com/facebook/react ./my-react-app
  chassis mimic github.com/vuejs/vue ./my-vue-app`,
	Args: cobra.ExactArgs(2),
	RunE: runMimic,
}

func init() {
	rootCmd.AddCommand(mimicCmd)
}

func runMimic(cmd *cobra.Command, args []string) error {
	source := args[0]
	targetDir := args[1]

	// Determine source type (GitHub URL or local path)
	sourceType := detectSourceType(source)

	fmt.Printf("🔍 Analyzing %s...\n", source)
	fmt.Printf("📁 Target directory: %s\n", targetDir)

	// TODO: Implement analysis based on source type
	switch sourceType {
	case "github":
		fmt.Printf("Detected GitHub repository: %s\n", source)
		// TODO: Implement GitHub analysis
		return fmt.Errorf("GitHub repository analysis not yet implemented")

	case "local":
		fmt.Printf("Detected local directory: %s\n", source)
		// TODO: Implement local directory analysis
		return fmt.Errorf("local directory analysis not yet implemented")

	default:
		return fmt.Errorf("unable to determine source type for: %s", source)
	}
}

// detectSourceType determines if the source is a GitHub URL or local path
func detectSourceType(source string) string {
	// Check for GitHub URL patterns
	if strings.HasPrefix(source, "https://github.com/") ||
		strings.HasPrefix(source, "http://github.com/") ||
		strings.HasPrefix(source, "github.com/") ||
		strings.HasPrefix(source, "git@github.com:") ||
		strings.HasPrefix(source, "www.github.com/") {
		return "github"
	}

	// Everything else is treated as a local path
	return "local"
}

// extractGitHubInfo parses GitHub repository information from various URL formats
func extractGitHubInfo(url string) (owner, repo string, err error) {
	// Remove common prefixes
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")
	url = strings.TrimPrefix(url, "www.")
	url = strings.TrimPrefix(url, "github.com/")
	url = strings.TrimPrefix(url, "github.com:")

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Split by / to get owner and repo
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format")
	}

	owner = parts[0]
	repo = parts[1]

	// Handle additional path segments (like /tree/main)
	// We only need owner/repo

	return owner, repo, nil
}
