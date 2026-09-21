package importcsv

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/edcrewe/gormcsv/meta"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// fileSizeWorkerThreshold is the CSV file size below which a single worker is used
// regardless of the database type.
const fileSizeWorkerThreshold = 5 * 1024 * 1024 // 5 MB

type ModelCSV struct {
	meta.Files
	fields string
}

// ConnectDB connect to the Database
func (mcsv *ModelCSV) ConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}
	return db
}

// CreateSchema create the schema in the db
func (mcsv *ModelCSV) CreateSchema(db *gorm.DB, factory ModelFactory) {
	for _, name := range factory.models {
		model := factory.New(name)
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("failed to automigrate schema for %s: %v", name, err)
		}
	}
}

func (mcsv *ModelCSV) getModel(name string) (string, error) {
	found := MakeModels().New(name)
	if found != nil {
		return name, nil
	} else {
		return "", errors.New("Model not found for " + name)
	}
}

// calcWorkerCount returns 1 for SQLite (which serialises writers) or for files
// under 5 MB where the concurrency overhead outweighs the gain. For larger
// files on other databases it uses runtime.NumCPU.
func calcWorkerCount(db *gorm.DB, csvFile *os.File) int {
	if db.Dialector.Name() == "sqlite" {
		return 1
	}
	if info, err := csvFile.Stat(); err == nil && info.Size() < fileSizeWorkerThreshold {
		return 1
	}
	if n := runtime.NumCPU(); n > 1 {
		return n
	}
	return 1
}

// ImportCSV main command method for importcsv
func (mcsv *ModelCSV) ImportCSV(filePath string) {
	db := mcsv.ConnectDB()
	factory := MakeModels()
	mcsv.CreateSchema(db, factory)
	filesMap, err := mcsv.FilesFetch(filePath)
	if err != nil {
		fmt.Printf("Failed to load CSV file(s) from %s, Due to %s\n", filePath, err)
		return
	}
	csvmeta := meta.CSVMeta{}
	err = csvmeta.PopulateMeta(filePath)
	if err != nil {
		fmt.Printf("Failed to determine the fields, cannot import due to error: %s\n", err)
		return
	}
	fmt.Printf("Importing data from %s\n", filePath)
	for fileName, csvFile := range filesMap {
		var count int
		var duplicates int
		var errorlist []error

		reader := csv.NewReader(bufio.NewReader(csvFile))
		fieldMeta := meta.FieldMeta{}
		name, modelErr := mcsv.getModel(fileName)
		fieldList := []string{}
		for _, field := range csvmeta.Fields[name] {
			if field.Name != "Model" {
				fieldList = append(fieldList, field.Name)
			}
		}
		mcsv.fields = strings.Join(fieldList, ",")
		if modelErr != nil {
			fmt.Println(modelErr)
			return
		}
		model := factory.New(name)
		fieldMeta.SetMeta(model, mcsv.fields)

		workerCount := calcWorkerCount(db, csvFile)
		var wg sync.WaitGroup
		jobs := make(chan [][]string, 100)
		results := make(chan batchResult, 100)

		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				worker(jobs, results, &fieldMeta, db, factory, name)
			}()
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		// Feed jobs from a dedicated goroutine so the main goroutine can drain
		// results concurrently. Without this, results and jobs buffers can both
		// fill simultaneously causing a deadlock on large files.
		// Parse errors are sent as batchResults so they flow through the same
		// channel and are collected with all other errors below.
		go func() {
			defer close(jobs)
			// Discard the header row; PopulateMeta already consumed it for
			// field-name and type inference via a separate file handle.
			if _, err := reader.Read(); err != nil {
				return
			}
			const batchSize = 1000
			var currentBatch [][]string
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					results <- batchResult{errors: []error{err}}
					continue
				}
				currentBatch = append(currentBatch, record)
				if len(currentBatch) >= batchSize {
					jobs <- currentBatch
					currentBatch = nil
				}
			}
			if len(currentBatch) > 0 {
				jobs <- currentBatch
			}
		}()

		for res := range results {
			count += res.count
			duplicates += res.duplicates
			errorlist = append(errorlist, res.errors...)
		}

		fmt.Printf("Imported %d rows to %s\n", count, name)
		if duplicates > 0 {
			fmt.Printf("Skipped %d duplicate rows\n", duplicates)
		}
		if len(errorlist) > 0 {
			fmt.Printf("Failed import for %d rows due to errors:\n", len(errorlist))
			for _, err := range errorlist {
				fmt.Println(err)
			}
		}
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
