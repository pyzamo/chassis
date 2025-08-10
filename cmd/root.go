package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	indentSize int
)

var rootCmd = &cobra.Command{
	Use:     "chassis",
	Short:   "A lightweight CLI tool to scaffold project directory structures",
	Version: "0.1.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print every path created/skipped")
}
