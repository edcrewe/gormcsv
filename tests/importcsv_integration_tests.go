// +build integration

package tests

import (
	"database/sql"
	"github.com/edcrewe/gormcsv/importcsv"
	"testing"
)

var meta importcsv.FieldMeta
/*
Setup first then run the tests
*/
func TestMain(m *testing.M) {
	meta = importcsv.FieldMeta{}
	factory := importcsv.MakeModels()
	model := factory.New("country")
	meta.Setmeta(model, "name,code,latitude,longtitude,alias")
	m.Run()
}

func checkCount(rows *sql.Rows) (count int) {
	for rows.Next() {
		err:= rows.Scan(&count)
		if err != nil { return 0}
	}
	return count
}
/*
Integration test for importcsv
Run the import of test models.go to sqlite and check data is in the db
*/
func TestImportCSV(t *testing.T) {
	mcsv := importcsv.ModelCSV{}
	// Run import
	mcsv.ImportCSV("tests/fixtures/Country.csv")
	// Test database is populated
	db := mcsv.connectDB()
	rows, err := db.Query("SELECT COUNT(*) as count FROM countries")
	count := checkCount(rows)
	if count < 245 {
		t.Errorf("Total count: %d rows imported. Expected 245", count)
	}
}

