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

func ImportCSV() {
	db := connectDB()
	createModels(db)
	csvFile, _ := os.Open("tests/fixtures/countries.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	errors := []error{}
	fields := map[string]int{
		"Name":       0,
		"Code":       1,
		"Latitude":   2,
		"Longtitude": 3,
		"Alias":      4,
	}
	var count int = 0
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			errors = append(errors, error)
			continue
		}
		// Create
		lat, _ := strconv.ParseFloat(line[fields["Latitude"]], 64)
		longt, _ := strconv.ParseFloat(line[fields["Longtitude"]], 64)
		db.Create(&Country{
			Code:       line[fields["Code"]],
			Name:       line[fields["Name"]],
			Latitude:   lat,
			Longtitude: longt,
			Alias:      line[fields["Alias"]],
		})
		count += 1
	}
	fmt.Printf("Imported %d rows to Country", count)
	db.Close()
}
