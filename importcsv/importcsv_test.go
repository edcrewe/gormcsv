package importcsv

import (
	"fmt"
	"os"
	"testing"
)

// TestMain setup for all the tests in package tests, remove the db, then run the tests
func TestMain(m *testing.M) {
	code := m.Run()
	if _, err := os.Stat("test.db"); !os.IsNotExist(err) {
		err := os.Remove("test.db")
		if err != nil {
			fmt.Printf("Failed to wipe database %s", err)
			return
		}
	}
	os.Exit(code)
}

// TestImportCountry integration test for importcsv
// Run the import of test models.go to sqlite and check data is in the db
func TestImportCountry(t *testing.T) {
	mcsv := ModelCSV{}
	db := mcsv.ConnectDB()
	// Run import
	mcsv.ImportCSV("../static/fixtures/Country.csv")
	// Test database is populated
	var count int
	db.Table("countries").Count(&count)
	//count := checkCount(rows)
	if count < 245 {
		t.Errorf("Total count: %d rows imported. Expected 245", count)
	}
}

// TestImportTestTypes integration test for importcsv
// Run the import of test models.go to sqlite and check data is in the db
func TestImportTestTypes(t *testing.T) {
	mcsv := ModelCSV{}
	db := mcsv.ConnectDB()
	// Run import
	mcsv.ImportCSV("../static/fixtures/TestTypes.csv")
	// Test database is populated
	var count int
	db.Table("test_types").Count(&count)
	//count := checkCount(rows)
	if count < 6 {
		t.Errorf("Total count: %d rows imported. Expected 7", count)
	}
	//model := db.Model(&importcsv.TestTypes{}).Where("codecol = ?", "RF024")

	//fmt.Println(model.Codecol)
}
