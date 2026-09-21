package cmd

import (
	"fmt"

	"github.com/edcrewe/gormcsv/inspectcsv"
	"github.com/edcrewe/gormcsv/meta"
	"github.com/spf13/cobra"
)

var output string

var inspectCSVCommand = &cobra.Command{
	Use:   "inspectcsv",
	Short: "Generate GORM models from CSV files",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if files == "" {
			return fmt.Errorf("--files is required")
		}
		var csvMeta meta.CSVMeta
		if err := csvMeta.PopulateMeta(files); err != nil {
			return err
		}
		if err := inspectcsv.GenerateFile(csvMeta, output); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", output); err != nil {
			return fmt.Errorf("write generation result: %w", err)
		}
		return nil
	},
}

func init() {
	inspectCSVCommand.Flags().StringVarP(&output, "output", "o", "importcsv/models.go", "generated Go file")
	rootCmd.AddCommand(inspectCSVCommand)
}
