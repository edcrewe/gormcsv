package cmd

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

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
		if err := writeSkippedRows(cmd.ErrOrStderr(), csvMeta.SkippedRows); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", output); err != nil {
			return fmt.Errorf("write generation result: %w", err)
		}
		return nil
	},
}

// writeSkippedRows warns about rows left out of type inference, in path order.
func writeSkippedRows(w io.Writer, skipped map[string][]int) error {
	paths := make([]string, 0, len(skipped))
	for path := range skipped {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		lines := make([]string, len(skipped[path]))
		for i, line := range skipped[path] {
			lines[i] = strconv.Itoa(line)
		}
		rows, lineWord := "rows", "lines"
		if len(lines) == 1 {
			rows, lineWord = "row", "line"
		}
		if _, err := fmt.Fprintf(w, "warning: %s: skipped %d %s with the wrong number of fields (%s %s)\n",
			path, len(lines), rows, lineWord, strings.Join(lines, ", ")); err != nil {
			return fmt.Errorf("write skipped rows: %w", err)
		}
	}
	return nil
}

func init() {
	inspectCSVCommand.Flags().StringVarP(&output, "output", "o", "importcsv/models.go", "generated Go file")
	rootCmd.AddCommand(inspectCSVCommand)
}
