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
	var country Country

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
	}
	fmt.Println("test")
	// Read
	db.First(&country, 1)                // find country with id 1
	db.First(&country, "code = ?", "AZ") // find country with code l1212

	// Update - update country's alias
	db.Model(&country).Update("Alias", "New Azer")

	// Delete - delete country
	db.Delete(&country)
	db.Close()
}
