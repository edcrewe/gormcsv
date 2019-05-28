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
)

// Connect to the Database
func connectDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
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
	// Create
	db.Create(&Country{Code: "AZ", Name: "AZERBAIJAN", Latitude: 40.5, Longtitude: 47.5, Alias: "Azerbaijan"})

	csvFile, _ := os.Open("tests/fixtures/countries.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		//		fmt.Println(line)
	}
	fmt.Println("test")
	// Read
	db.First(&country, 1)                // find country with id 1
	db.First(&country, "code = ?", "AZ") // find country with code l1212

	// Update - update country's alias
	db.Model(&country).Update("Alias", "New Azer")

	// Delete - delete country
	db.Delete(&country)
}
