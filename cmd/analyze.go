package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pyzamo/chassis/internal/ai"
	"github.com/pyzamo/chassis/internal/analyze"
	"github.com/pyzamo/chassis/internal/github"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	maxDepth     int
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze <source>",
	Short: "Analyze a project with AI to extract a reusable scaffolding template",
	Long: `Analyze an existing project directory or GitHub repository using AI to generate
a generalized, reusable scaffolding template that can be used with the 'chassis build' command.

This command uses Google Gemini AI to:
- Identify architectural patterns in your project
- Extract a generalized skeleton (e.g., "controllers/" instead of "UserController.js")
- Create reusable project templates, not project-specific copies
- Remove implementation details while preserving structure

Requires GEMINI_API_KEY environment variable. Get your free API key at:
https://aistudio.google.com/app/apikey

Examples:
  # Analyze local directory and save template
  export GEMINI_API_KEY='your-key-here'
  chassis analyze ./my-project > template.txt
  chassis build template.txt ./new-project
  
  # Analyze GitHub repository
  chassis analyze https://github.com/user/repo > structure.txt
  
  # Limit analysis depth
  chassis analyze ./deep-project --max-depth 3 > layout.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Local flags for analyze command
	analyzeCmd.Flags().StringVar(&outputFormat, "format", "tree", "Output format: tree, yaml, or json")
	analyzeCmd.Flags().IntVar(&maxDepth, "max-depth", 5, "Maximum depth to analyze")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Validate output format
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "tree" && outputFormat != "yaml" && outputFormat != "json" {
		return fmt.Errorf("invalid format: %s (must be tree, yaml, or json)", outputFormat)
	}

	// Validate max depth
	if maxDepth < 1 {
		return fmt.Errorf("max-depth must be at least 1")
	}

	// Print progress to stderr so it doesn't mix with output
	fmt.Fprintf(os.Stderr, "Analyzing '%s'...\n", source)

	// Determine if source is GitHub URL or local path
	var analyzer analyze.Analyzer
	if isGitHubURL(source) {
		fmt.Fprintf(os.Stderr, "Detected GitHub repository\n")
		analyzer = github.NewAnalyzer(source)
	} else {
		// Local directory
		analyzer = analyze.NewLocalAnalyzer(source, maxDepth)
	}

	// Perform analysis
	result, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Print statistics to stderr
	fmt.Fprintf(os.Stderr, "Found: %d directories, %d files\n", result.DirCount, result.FileCount)
	if result.FilteredCount > 0 {
		fmt.Fprintf(os.Stderr, "Filtered: %d items (build artifacts, dependencies, etc.)\n", result.FilteredCount)
	}

	// First, export the raw structure to send to AI
	exporter := analyze.NewExporter(result.Nodes)
	rawStructure, err := exporter.ToTreeSimple()
	if err != nil {
		return fmt.Errorf("failed to export structure: %w", err)
	}

	// Initialize Gemini client
	fmt.Fprintf(os.Stderr, "Analyzing patterns with AI...\n")
	geminiClient, err := ai.NewGeminiClient()
	if err != nil {
		// If AI is not available, provide helpful message
		fmt.Fprintf(os.Stderr, "\n⚠️  AI analysis unavailable: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nTo enable AI-powered skeleton extraction:\n")
		fmt.Fprintf(os.Stderr, "1. Get a free API key from: https://aistudio.google.com/app/apikey\n")
		fmt.Fprintf(os.Stderr, "2. Set the environment variable: export GEMINI_API_KEY='your-key-here'\n")
		fmt.Fprintf(os.Stderr, "\nFalling back to raw structure output...\n\n")

		// Fall back to raw structure
		fmt.Print(rawStructure)
		return nil
	}

	// Detect project type for better AI analysis
	projectType := ai.DetectProjectType(rawStructure)
	fmt.Fprintf(os.Stderr, "Detected project type: %s\n", projectType)

	// Get AI-generated skeleton
	skeleton, err := geminiClient.ExtractSkeleton(rawStructure, projectType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  AI analysis failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Falling back to raw structure output...\n\n")
		fmt.Print(rawStructure)
		return nil
	}

	// Convert skeleton to requested format if needed
	var output string
	if outputFormat == "tree" {
		output = skeleton // AI already returns in tree format
	} else {
		// Parse the skeleton and convert to other formats
		// For now, we'll need to parse the AI output back into nodes
		fmt.Fprintf(os.Stderr, "Note: YAML/JSON output currently shows raw structure. AI-generated skeleton is available in tree format only.\n")

		// Fall back to raw structure for non-tree formats
		switch outputFormat {
		case "yaml":
			output, err = exporter.ToYAML()
		case "json":
			output, err = exporter.ToJSON()
		}
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
	}

	// Print the layout to stdout
	fmt.Print(output)
	if !strings.HasSuffix(output, "\n") {
		fmt.Println()
	}

	// Success message to stderr
	fmt.Fprintf(os.Stderr, "\n✓ AI-powered analysis complete\n")

	return nil
}

// isGitHubURL checks if the source is a GitHub URL
func isGitHubURL(source string) bool {
	lower := strings.ToLower(source)
	return strings.HasPrefix(lower, "https://github.com/") ||
		strings.HasPrefix(lower, "http://github.com/") ||
		strings.HasPrefix(lower, "github.com/") ||
		strings.HasPrefix(lower, "git@github.com:")
}
