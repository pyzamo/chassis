package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Flags
	verbose    bool
	indentSize int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chassis",
	Short: "A lightweight CLI tool to scaffold project directory structures",
	Long: `Chassis is a lightweight Go command-line utility that scaffolds project 
directory structures from layout definition files.

It supports multiple input formats:
  - Plain-text indented tree (.txt, .tree)
  - YAML (.yaml, .yml)
  - JSON (.json)

The tool automatically detects the format from the file extension and creates
the specified directory structure, safely merging with existing directories.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print every path created/skipped")
}
