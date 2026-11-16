package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/edcrewe/gormcsv/schema"
	"github.com/spf13/cobra"
)

var rowsQuery schema.RowsQuery

var rowsCommand = &cobra.Command{
	Use:   "rows",
	Short: "Print one page of a table's rows",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if rowsQuery.Table == "" {
			return fmt.Errorf("--table is required")
		}
		driver, dsn := connectionSettings(cmd, os.Getenv)
		db, err := importcsv.OpenDatabase(driver, dsn)
		if err != nil {
			return err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("access database connection: %w", err)
		}
		defer func() { _ = sqlDB.Close() }()

		page, err := schema.Rows(cmd.Context(), db, rowsQuery)
		if errors.Is(err, schema.ErrNotFound) {
			return fmt.Errorf("no %w", err)
		}
		if err != nil {
			return err
		}
		if jsonOutput {
			return writeJSON(cmd.OutOrStdout(), page)
		}
		return writeRows(cmd, page)
	},
}

func writeRows(cmd *cobra.Command, page schema.Page) error {
	lines := []string{strings.Join(page.Columns, "\t")}
	for _, row := range page.Rows {
		values := make([]string, len(row))
		for i, value := range row {
			values[i] = fmt.Sprint(value)
		}
		lines = append(lines, strings.Join(values, "\t"))
	}
	lines = append(lines, fmt.Sprintf("rows %d-%d of %d", page.Offset+1, page.Offset+len(page.Rows), page.Total))
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), strings.Join(lines, "\n")); err != nil {
		return fmt.Errorf("write rows: %w", err)
	}
	return nil
}

func init() {
	addConnectionFlags(rowsCommand)
	flags := rowsCommand.Flags()
	flags.StringVar(&rowsQuery.Table, "table", "", "table to read")
	flags.IntVar(&rowsQuery.Limit, "limit", 50, fmt.Sprintf("rows per page, at most %d", schema.MaxLimit))
	flags.IntVar(&rowsQuery.Offset, "offset", 0, "rows to skip")
	flags.StringVar(&rowsQuery.Order, "order", "", "column to sort by (default: the primary key)")
	flags.BoolVar(&rowsQuery.Desc, "desc", false, "sort in descending order")
	flags.BoolVar(&jsonOutput, "json", false, "print the page as JSON on stdout")
	rootCmd.AddCommand(rowsCommand)
}
