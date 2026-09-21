// Package cmd implements the gormcsv command-line interface.
package cmd

import "github.com/spf13/cobra"

var files string

var rootCmd = &cobra.Command{
	Use:          "gormcsv",
	Short:        "Generate GORM models from CSV files and import their data",
	SilenceUsage: true,
}

// Execute runs the command-line interface.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&files, "files", "f", "", "CSV file or directory")
}
