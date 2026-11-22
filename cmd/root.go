// Package cmd implements the gormcsv command-line interface.
package cmd

import "github.com/spf13/cobra"

// newRootCommand builds the whole command tree. Each command keeps its flag values in its own
// variables, created here, so one Execute cannot leave settings behind for the next.
func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:          "gormcsv",
		Short:        "Generate GORM models from CSV files and import their data",
		SilenceUsage: true,
		// main prints the returned error once; without this cobra prints it as well.
		SilenceErrors: true,
		Version:       currentVersion(),
	}
	root.SetVersionTemplate("gormcsv {{.Version}}\n")
	files := root.PersistentFlags().StringP("files", "f", "", "CSV file or directory")
	root.AddCommand(
		newInspectCommand(files),
		newImportCommand(files),
		newTablesCommand(),
		newRowsCommand(),
		newVersionCommand(),
	)
	return root
}

// Execute runs the command-line interface.
func Execute() error {
	return newRootCommand().Execute()
}
