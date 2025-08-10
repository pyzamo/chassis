package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/pyzamo/chassis/internal/generate"
	"github.com/pyzamo/chassis/internal/parse"
	"github.com/pyzamo/chassis/internal/validate"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build <layout-file|-> [target-dir]",
	Short: "Build directory structure from a layout definition file",
	Long: `Build directory structure from a layout definition file.
	
The layout file can be in plain-text tree format, YAML, or JSON.
Format is auto-detected from the file extension.

Use "-" as the layout file to read from stdin.
If target-dir is not specified, the current directory is used.

Examples:
  # Build structure in new directory
  chassis build layout.txt my-project
  
  # Build structure in current directory
  chassis build layout.yaml .
  
  # Read layout from stdin
  echo "src/\n  main.go" | chassis build - .
  
  # Build structure with verbose output
  chassis build -v layout.json my-app`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Local flags for build command
	buildCmd.Flags().IntVar(&indentSize, "indent", 2, "Expected space width for plain-text parser (auto-detects tabs)")
}

func runBuild(cmd *cobra.Command, args []string) error {
	layoutFile := args[0]
	targetDir := "."

	if len(args) > 1 {
		targetDir = args[1]
	}

	// Step 1: Open the input source
	var reader io.Reader
	var format parse.Format

	if layoutFile == "-" {
		// Read from stdin
		reader = os.Stdin
		format = parse.FormatPlainText // Default to plain-text for stdin
		if verbose {
			fmt.Println("Reading from stdin...")
		}
	} else {
		// Read from file
		file, err := os.Open(layoutFile)
		if err != nil {
			return fmt.Errorf("failed to open layout file: %w", err)
		}
		defer file.Close()
		reader = file

		// Detect format from file extension
		format = parse.DetectFormat(layoutFile)
		if format == parse.FormatUnknown {
			return fmt.Errorf("unknown file format for %s (supported: .txt, .tree, .yaml, .yml, .json)", layoutFile)
		}

		if verbose {
			fmt.Printf("Reading %s format from %s\n", format, layoutFile)
		}
	}

	// Step 2: Parse the input
	var nodes []*parse.Node
	var err error

	if format == parse.FormatPlainText {
		// Use the indent size flag for plain-text
		nodes, err = parse.ParseWithIndent(reader, format, indentSize)
	} else {
		nodes, err = parse.Parse(reader, format)
	}

	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if verbose {
		nodeCount := 0
		for _, node := range nodes {
			nodeCount += node.CountNodes()
		}
		fmt.Printf("Parsed %d nodes\n", nodeCount)
	}

	// Step 3: Validate the tree
	if err := validate.Validate(nodes); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if verbose {
		fmt.Println("Validation passed")
	}

	// Step 4: Generate the filesystem structure
	result, err := generate.Generate(nodes, targetDir, verbose)
	if err != nil {
		// Even with errors, show what was done
		if result != nil {
			result.PrintSummary()
		}
		return err
	}

	// Step 5: Show summary
	result.PrintSummary()

	// Show success message
	if result.Created > 0 || result.Skipped > 0 {
		fmt.Printf("\n✓ Structure built in %s\n", targetDir)
	} else {
		fmt.Println("\nNo changes made (all paths already exist)")
	}

	return nil
}
