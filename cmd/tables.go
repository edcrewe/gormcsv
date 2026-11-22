package cmd

import (
	"fmt"
	"os"

	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/edcrewe/gormcsv/schema"
	"github.com/spf13/cobra"
)

// TablesReport is the --json output of the tables command.
type TablesReport struct {
	Tables []schema.Table `json:"tables"`
}

func newTablesCommand() *cobra.Command {
	var jsonOutput bool
	command := &cobra.Command{
		Use:   "tables",
		Short: "List the database's tables with their row counts, columns and indexes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			tables, err := schema.Describe(cmd.Context(), db)
			if err != nil {
				return err
			}
			if jsonOutput {
				return writeJSON(cmd.OutOrStdout(), TablesReport{Tables: tables})
			}
			for _, table := range tables {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %d rows, %d columns\n",
					table.Name, table.Rows, len(table.Columns)); err != nil {
					return fmt.Errorf("write tables: %w", err)
				}
			}
			return nil
		},
	}
	addConnectionFlags(command)
	command.Flags().BoolVar(&jsonOutput, "json", false, "print the tables as JSON on stdout")
	return command
}
