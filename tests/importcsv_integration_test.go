// +build integration

package tests

import (
	"fmt"
	"github.com/edcrewe/gormcsv/importcsv"
	"os"
	"testing"
)

var meta importcsv.FieldMeta
/*
Setup first, remove the db, then run the tests
*/
func TestMain(m *testing.M) {
	if _, err := os.Stat("test.db"); !os.IsNotExist(err) {
		err := os.Remove("test.db")
		if err != nil {
			fmt.Printf("Failed to wipe database %s", err)
			return
		}
	}
	m.Run()
}

/*
Integration test for importcsv
Run the import of test models.go to sqlite and check data is in the db
*/
func TestImportCSV(t *testing.T) {
	meta = importcsv.FieldMeta{}
	factory := importcsv.MakeModels()
	model := factory.New("country")
	meta.Setmeta(model, "name,code,latitude,longtitude,alias")
	mcsv := importcsv.ModelCSV{}
	db := mcsv.ConnectDB()
	mcsv.CreateSchema(db, factory)

	// Run import
	mcsv.ImportCSV("fixtures/Country.csv")
	// Test database is populated
	var count int
	db.Table("countries").Count(&count)
	//count := checkCount(rows)
	if count < 245 {
		t.Errorf("Total count: %d rows imported. Expected 245", count)
	}
}

