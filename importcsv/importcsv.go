// Package importcsv imports CSV records into GORM models.
package importcsv

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/edcrewe/gormcsv/meta"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultBatchSize = 1000

// Config controls importer batching and concurrency.
type Config struct {
	BatchSize int
	Workers   int
}

// FileResult describes the outcome for one CSV file.
type FileResult struct {
	File       string
	Model      string
	Inserted   int64
	Duplicates int64
	Rejected   int64
}

// Result describes an import. A non-nil error can accompany partial results.
type Result struct {
	Files      []FileResult
	Inserted   int64
	Duplicates int64
	Rejected   int64
}

// Importer imports CSV files using an existing GORM connection.
type Importer struct {
	db        *gorm.DB
	factory   ModelFactory
	batchSize int
	workers   int
}

// New creates an importer. Zero values select a batch size of 1000 and one
// worker. SQLite is always restricted to one writer.
func New(db *gorm.DB, factory ModelFactory, config Config) (*Importer, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	if len(factory.models) == 0 {
		return nil, fmt.Errorf("at least one model is required")
	}
	if config.BatchSize < 0 {
		return nil, fmt.Errorf("batch size must not be negative")
	}
	if config.Workers < 0 {
		return nil, fmt.Errorf("worker count must not be negative")
	}
	if config.BatchSize == 0 {
		config.BatchSize = defaultBatchSize
	}
	if config.Workers == 0 {
		config.Workers = 1
	}
	return &Importer{
		db:        db,
		factory:   factory,
		batchSize: config.BatchSize,
		workers:   resolveWorkerCount(db, config.Workers),
	}, nil
}

func resolveWorkerCount(db *gorm.DB, configured int) int {
	if db.Name() == "sqlite" {
		return 1
	}
	sqlDB, err := db.DB()
	if err == nil {
		if maximum := sqlDB.Stats().MaxOpenConnections; maximum > 0 && configured > maximum {
			return maximum
		}
	}
	return configured
}

// OpenSQLite opens a SQLite database with portable GORM error translation.
// The caller owns the returned connection and must close its underlying sql.DB.
func OpenSQLite(path string) (*gorm.DB, error) {
	return OpenDatabase("sqlite", path)
}

// OpenDatabase opens a supported GORM database with portable error translation.
func OpenDatabase(driver, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch strings.ToLower(driver) {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", driver, err)
	}
	return db, nil
}

// Import imports one CSV file or every CSV file in a directory. Valid records
// are retained when other records fail, and the returned error describes every
// rejected record.
func (importer *Importer) Import(ctx context.Context, path string) (Result, error) {
	inputs, err := meta.CSVFiles(path)
	if err != nil {
		return Result{}, err
	}

	result := Result{Files: make([]FileResult, 0, len(inputs))}
	var importErrors []error
	for _, input := range inputs {
		fileResult, err := importer.importFile(ctx, input)
		result.Files = append(result.Files, fileResult)
		result.Inserted += fileResult.Inserted
		result.Duplicates += fileResult.Duplicates
		result.Rejected += fileResult.Rejected
		if err != nil {
			importErrors = append(importErrors, err)
		}
	}
	return result, errors.Join(importErrors...)
}

func (importer *Importer) importFile(ctx context.Context, input meta.CSVFile) (FileResult, error) {
	result := FileResult{File: input.Path, Model: input.Model}
	model := importer.factory.New(input.Model)
	if model == nil {
		return result, fmt.Errorf("%s: no model registered for %q", input.Path, input.Model)
	}
	if err := importer.db.WithContext(ctx).AutoMigrate(model); err != nil {
		return result, fmt.Errorf("%s: migrate model %q: %w", input.Path, input.Model, err)
	}

	file, err := os.Open(input.Path) // #nosec G304 -- path is explicitly supplied by the user
	if err != nil {
		return result, fmt.Errorf("open %q: %w", input.Path, err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return result, fmt.Errorf("%s: read header: %w", input.Path, err)
	}
	fieldMeta, err := meta.NewFieldMeta(model, header)
	if err != nil {
		return result, fmt.Errorf("%s: %w", input.Path, err)
	}

	modelType := reflect.TypeOf(model)
	models := reflect.MakeSlice(reflect.SliceOf(modelType), 0, importer.batchSize)
	row := 1
	var importErrors []error
	workerCtx, cancelWorkers := context.WithCancelCause(ctx)
	defer cancelWorkers(nil)
	jobs := make(chan batchJob, importer.workers)
	workerResults := runWorkers(workerCtx, importer.workers, jobs, func(workerCtx context.Context, job batchJob) batchResult {
		if context.Cause(workerCtx) != nil {
			return batchResult{rejected: job.rows}
		}
		dbResult := importer.db.WithContext(workerCtx).
			Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(job.models, importer.batchSize)
		if dbResult.Error != nil {
			err := fmt.Errorf("%s: insert batch ending at row %d: %w", input.Path, job.endRow, dbResult.Error)
			cancelWorkers(err)
			return batchResult{rejected: job.rows, err: err}
		}
		return batchResult{
			inserted:   dbResult.RowsAffected,
			duplicates: job.rows - dbResult.RowsAffected,
		}
	})

	type aggregate struct {
		inserted   int64
		duplicates int64
		rejected   int64
		errors     []error
	}
	aggregated := make(chan aggregate, 1)
	go func() {
		var total aggregate
		for batch := range workerResults {
			total.inserted += batch.inserted
			total.duplicates += batch.duplicates
			total.rejected += batch.rejected
			if batch.err != nil {
				total.errors = append(total.errors, batch.err)
			}
		}
		aggregated <- total
	}()

	queue := func() bool {
		if models.Len() == 0 {
			return true
		}
		job := batchJob{models: models.Interface(), rows: int64(models.Len()), endRow: row}
		select {
		case jobs <- job:
			models = reflect.MakeSlice(models.Type(), 0, importer.batchSize)
			return true
		case <-workerCtx.Done():
			result.Rejected += job.rows
			return false
		}
	}

	producing := true
	for {
		if context.Cause(workerCtx) != nil {
			result.Rejected += int64(models.Len())
			producing = false
			break
		}
		row++
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			result.Rejected++
			importErrors = append(importErrors, fmt.Errorf("%s row %d: %w", input.Path, row, readErr))
			continue
		}
		converted, convertErr := fieldMeta.RecordToModel(importer.factory.New(input.Model), record)
		if convertErr != nil {
			result.Rejected++
			importErrors = append(importErrors, fmt.Errorf("%s row %d: %w", input.Path, row, convertErr))
			continue
		}
		models = reflect.Append(models, reflect.ValueOf(converted))
		if models.Len() == importer.batchSize {
			if !queue() {
				producing = false
				break
			}
		}
	}
	if producing {
		_ = queue()
	}
	close(jobs)
	workerTotal := <-aggregated
	result.Inserted += workerTotal.inserted
	result.Duplicates += workerTotal.duplicates
	result.Rejected += workerTotal.rejected
	importErrors = append(importErrors, workerTotal.errors...)
	if cause := context.Cause(workerCtx); cause != nil && len(workerTotal.errors) == 0 {
		importErrors = append(importErrors, fmt.Errorf("%s: %w", input.Path, cause))
	}
	return result, errors.Join(importErrors...)
}
