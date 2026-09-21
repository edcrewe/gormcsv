package cmd

import (
	"fmt"

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
		for _, file := range result.Files {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: inserted %d, skipped %d duplicates, rejected %d\n",
				file.Model, file.Inserted, file.Duplicates, file.Rejected); err != nil {
				return fmt.Errorf("write import result: %w", err)
			}
		}
		return importErr
	},
}

func init() {
	importCSVCommand.Flags().StringVar(&driver, "driver", "sqlite", "database driver: sqlite, postgres, or mysql")
	importCSVCommand.Flags().StringVarP(&dsn, "dsn", "d", "gormcsv.db", "database connection string or SQLite path")
	importCSVCommand.Flags().IntVarP(&batchSize, "batch-size", "b", 1000, "records per database batch")
	importCSVCommand.Flags().IntVarP(&workers, "workers", "w", 4, "concurrent database writers (SQLite is always one)")
	rootCmd.AddCommand(importCSVCommand)
}
