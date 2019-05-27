package importcsv

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func ImportCSV() {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Country{})

	// Create
	db.Create(&Country{Code: "AZ", Name: "AZERBAIJAN", Latitude: 40.5, Longtitude: 47.5, Alias: "Azerbaijan"})
	// Read
	var country Country
	db.First(&country, 1)                // find country with id 1
	db.First(&country, "code = ?", "AZ") // find country with code l1212

	// Update - update country's alias
	db.Model(&country).Update("Alias", "New Azer")

	// Delete - delete country
	db.Delete(&country)
}
