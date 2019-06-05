package importcsv

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"log"
	"os"
	"strings"
)

type ModelCSV struct {
	 fields string
}

/*
Connect to the Database
 */
func (mcsv *ModelCSV) connectDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal("failed to connect database")
	}
	return db
}

/*
Create the schema in the db
 */
func (mcsv *ModelCSV) createSchema(db *gorm.DB, factory ModelFactory) {
	for _, name := range factory.models {
		model := factory.New(name)
		db.AutoMigrate(model)
	}
}

func (mcsv *ModelCSV) modelFromFile(filePath string) (string, error) {
	parts := strings.Split(filePath, "/")
	parts = strings.Split(parts[len(parts) - 1], ".")
	name := strings.ToLower(parts[0])
	found := MakeModels().New(name)
	if found != nil {
		return name, nil
	} else {
		return "", errors.New("Model not found for " + name)
	}
}

/*
Main command method for importcsv
 */
func (mcsv *ModelCSV) ImportCSV(file string) {
	errorlist := []error{}
	db := mcsv.connectDB()
	factory := MakeModels()
	mcsv.createSchema(db, factory)
	/*files := []string
	if strings.HasSuffix(file, ".csv") {
		files.append(file)
	} */
	csvFile, _ := os.Open(file)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	mcsv.fields = "name,code,latitude,longtitude,alias"
	meta := FieldMeta{}
	name, error := mcsv.modelFromFile(file)
	if error != nil {
		fmt.Println(error)
		return
	}
	model := factory.New(name)
	meta.Setmeta(model, mcsv.fields)
	var count int = 0
	var duplicates int = 0
	for {
		record, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			errorlist = append(errorlist, error)
			continue
		}
		model, error := meta.RecordToModel(factory.New(name) , record)
		if error != nil {
			errorlist = append(errorlist, error)
			continue
		}
		// Create
		result := db.Create(model)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
				duplicates += 1
			} else {
				errorlist = append(errorlist, error)
			}
		} else {
			count += 1
		}
	}
	fmt.Printf("Imported %d rows to Country\n", count)
	if duplicates > 0 {
		fmt.Printf("Skipped %d duplicate rows\n", duplicates)
	}
	if errorlist != nil {
		fmt.Printf("Failed import for %d rows due to errors:\n", len(errorlist))
		for _, error := range errorlist {
			fmt.Println(error)
		}
	}
	db.Close()
}
