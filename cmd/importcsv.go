package cmd

import (
	"fmt"
	"os"

	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/spf13/cobra"
)

var (
	driver    string
	dsn       string
	batchSize int
	workers   int
)

var importCSVCommand = &cobra.Command{
	Use:   "importcsv",
	Short: "Populate database tables from CSV files",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if files == "" {
			return fmt.Errorf("--files is required")
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
		if workers > 0 {
			sqlDB.SetMaxOpenConns(workers)
		}

		importer, err := importcsv.New(db, importcsv.MakeModels(), importcsv.Config{
			BatchSize: batchSize,
			Workers:   workers,
		})
		if err != nil {
			return err
		}
		result, importErr := importer.Import(cmd.Context(), files)
		if jsonOutput {
			if err := writeJSON(cmd.OutOrStdout(), newImportReport(result, importErr)); err != nil {
				return err
			}
			return importErr
		}
		for _, file := range result.Files {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: inserted %d, skipped %d duplicates, rejected %d\n",
				file.Model, file.Inserted, file.Duplicates, file.Rejected); err != nil {
				return fmt.Errorf("write import result: %w", err)
			}
		}
		return importErr
	},
}

// connectionSettings returns the driver and DSN: from --driver and --dsn when given, otherwise
// from GORMCSV_DB_DRIVER and GORMCSV_DSN, otherwise the flag defaults. A DSN on the command line
// is visible to every local user in the process list, so callers holding a password should use
// the environment.
func connectionSettings(cmd *cobra.Command, getenv func(string) string) (string, string) {
	pick := func(flag, flagValue, envName string) string {
		if !cmd.Flags().Changed(flag) {
			if value := getenv(envName); value != "" {
				return value
			}
		}
		return flagValue
	}
	return pick("driver", driver, "GORMCSV_DB_DRIVER"), pick("dsn", dsn, "GORMCSV_DSN")
}

func init() {
	importCSVCommand.Flags().StringVar(&driver, "driver", "sqlite", "database driver: sqlite, postgres, or mysql; GORMCSV_DB_DRIVER applies when not given")
	importCSVCommand.Flags().StringVarP(&dsn, "dsn", "d", "gormcsv.db", "database connection string or SQLite path; GORMCSV_DSN applies when not given, keeping a password out of the process list")
	importCSVCommand.Flags().IntVarP(&batchSize, "batch-size", "b", 1000, "records per database batch")
	importCSVCommand.Flags().IntVarP(&workers, "workers", "w", 4, "concurrent database writers (SQLite is always one)")
	importCSVCommand.Flags().BoolVar(&jsonOutput, "json", false,
		"print per-file counts and every error as JSON on stdout; still exits non-zero on errors")
	rootCmd.AddCommand(importCSVCommand)
}
