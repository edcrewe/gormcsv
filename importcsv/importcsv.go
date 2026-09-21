package importcsv

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/edcrewe/gormcsv/meta"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

// ImportCSV main command method for importcsv
func (mcsv *ModelCSV) ImportCSV(filePath string) {
	errorlist := []error{}
	db := mcsv.ConnectDB()
	factory := MakeModels()
	mcsv.CreateSchema(db, factory)
	filesMap, err := mcsv.FilesFetch(filePath)
	if err != nil {
		fmt.Printf("Failed to load CSV file(s) from %s, Due to %s\n", filePath, err)
		return
	}
	var count int = 0
	var duplicates int = 0
	csvmeta := meta.CSVMeta{}
	err = csvmeta.PopulateMeta(filePath)
	if err != nil {
		fmt.Printf("Failed to determine the fields, cannot import due to error: %s\n", err)
		return
	}
	fmt.Printf("Importing data from %s\n", filePath)
	for fileName, csvFile := range filesMap {
		reader := csv.NewReader(bufio.NewReader(csvFile))
		meta := meta.FieldMeta{}
		name, error := mcsv.getModel(fileName)
		fieldList := []string{}
		for _, field := range csvmeta.Fields[name] {
			if field.Name != "Model" {
				fieldList = append(fieldList, field.Name)
			}
		}
		mcsv.fields = strings.Join(fieldList, ",")
		if error != nil {
			fmt.Println(error)
			return
		}
		model := factory.New(name)
		meta.SetMeta(model, mcsv.fields)
		// Process with worker pool for performance
		var wg sync.WaitGroup
		jobs := make(chan [][]string, 100)
		results := make(chan batchResult, 100)
		workerCount := 4 // Use 4 concurrent workers

		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				worker(jobs, results, &meta, db, factory, name)
			}()
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		batchSize := 1000
		var currentBatch [][]string

		for {
			record, error := reader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				errorlist = append(errorlist, error)
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
		close(jobs)

		for res := range results {
			count += res.count
			duplicates += res.duplicates
			if len(res.errors) > 0 {
				errorlist = append(errorlist, res.errors...)
			}
		}
		fmt.Printf("Imported %d rows to %s\n", count, name)
		if duplicates > 0 {
			fmt.Printf("Skipped %d duplicate rows\n", duplicates)
		}
		if errorlist != nil {
			fmt.Printf("Failed import for %d rows due to errors:\n", len(errorlist))
			for _, error := range errorlist {
				fmt.Println(error)
			}
		}
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
