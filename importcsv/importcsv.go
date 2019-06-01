package importcsv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"log"
	"os"
	"strconv"
)

// Connect to the Database
func connectDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal("failed to connect database")
	}
	return db
}

func createModels(db *gorm.DB) {
	Models := []interface{}{&Country{}, &UnitOfMeasure{}, &Organisation{}, &Item{}}
	// Migrate the schema
	for _, model := range Models {
		db.AutoMigrate(model)
	}
}

func ImportCSV(files string) {
	db := connectDB()
	createModels(db)
	csvFile, _ := os.Open(files)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	errors := []error{}
	meta := FieldMeta{}
	meta.getmeta(&Country{}, "name,code,latitude,longtitude,alias")
	fields := meta.fieldcols
	var count int = 0
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			errors = append(errors, error)
			continue
		}
		lat, error := strconv.ParseFloat(line[fields["Latitude"]], 64)
		if error != nil {
			errors = append(errors, error)
			continue
		}
		longt, error := strconv.ParseFloat(line[fields["Longtitude"]], 64)
		if error != nil {
			errors = append(errors, error)
			continue
		}
		// Create
		db.Create(&Country{
			Code:       line[fields["Code"]],
			Name:       line[fields["Name"]],
			Latitude:   lat,
			Longtitude: longt,
			Alias:      line[fields["Alias"]],
		})
		count += 1
	}
	fmt.Printf("Imported %d rows to Country\n", count)
	if errors != nil {
		fmt.Printf("Failed import for %d rows due to errors:\n", len(errors))
		for _, error := range errors {
			fmt.Println(error)
		}
	}
	db.Close()
}
