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

var (
	output string
	keys   []string
	dedupe []string
)

var inspectCSVCommand = &cobra.Command{
	Use:   "inspectcsv",
	Short: "Generate GORM models from CSV files",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if files == "" {
			return fmt.Errorf("--files is required")
		}
		parsedKeys, err := parseKeys(keys)
		if err != nil {
			return err
		}
		csvMeta := meta.CSVMeta{Keys: parsedKeys, Dedupe: make(map[string]bool, len(dedupe))}
		for _, model := range dedupe {
			csvMeta.Dedupe[strings.ToLower(strings.TrimSpace(model))] = true
		}
		if err := csvMeta.PopulateMeta(files); err != nil {
			return err
		}
		if err := inspectcsv.GenerateFile(csvMeta, output); err != nil {
			return err
		}
		if err := writeSkippedRows(cmd.ErrOrStderr(), csvMeta.SkippedRows); err != nil {
			return err
		}
		if err := writeWarnings(cmd.ErrOrStderr(), csvMeta.Warnings); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", output); err != nil {
			return fmt.Errorf("write generation result: %w", err)
		}
		return nil
	},
}

// parseKeys reads --key values of the form model=heading or model=heading+heading.
func parseKeys(values []string) (map[string][]string, error) {
	parsed := make(map[string][]string, len(values))
	for _, value := range values {
		model, columns, ok := strings.Cut(value, "=")
		model = strings.ToLower(strings.TrimSpace(model))
		if !ok || model == "" || strings.TrimSpace(columns) == "" {
			return nil, fmt.Errorf("--key %q must look like model=column or model=column+column", value)
		}
		if _, exists := parsed[model]; exists {
			return nil, fmt.Errorf("--key given twice for %q", model)
		}
		for _, column := range strings.Split(columns, "+") {
			if column = strings.TrimSpace(column); column == "" {
				return nil, fmt.Errorf("--key %q has an empty column", value)
			}
			parsed[model] = append(parsed[model], column)
		}
	}
	return parsed, nil
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

// writeWarnings prints other inspection warnings, such as renamed columns, in path order.
func writeWarnings(w io.Writer, warnings map[string][]string) error {
	paths := make([]string, 0, len(warnings))
	for path := range warnings {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		for _, warning := range warnings[path] {
			if _, err := fmt.Fprintf(w, "warning: %s: %s\n", path, warning); err != nil {
				return fmt.Errorf("write warnings: %w", err)
			}
		}
	}
	return nil
}

func init() {
	inspectCSVCommand.Flags().StringVarP(&output, "output", "o", "importcsv/models.go", "generated Go file")
	inspectCSVCommand.Flags().StringArrayVar(&keys, "key", nil,
		"unique key for a model, e.g. country=code or order-item=order+line; repeatable")
	inspectCSVCommand.Flags().StringArrayVar(&dedupe, "dedupe", nil,
		"skip rows identical to one already imported for this model, e.g. country; repeatable")
	rootCmd.AddCommand(inspectCSVCommand)
}
