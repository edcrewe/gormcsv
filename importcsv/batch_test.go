package importcsv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BatchUser struct {
	ID   uint
	Name string
}

func TestDynamicBatch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&BatchUser{})

	dummy := &BatchUser{}
	modelType := reflect.TypeOf(dummy)
	sliceType := reflect.SliceOf(modelType)
	sliceVal := reflect.MakeSlice(sliceType, 0, 0)

	u1 := &BatchUser{Name: "Alice"}
	u2 := &BatchUser{Name: "Bob"}
	sliceVal = reflect.Append(sliceVal, reflect.ValueOf(u1))
	sliceVal = reflect.Append(sliceVal, reflect.ValueOf(u2))

	res := db.Create(sliceVal.Interface())
	if res.Error != nil {
		t.Fatalf("GORM create error: %v", res.Error)
	}
	if res.RowsAffected != 2 {
		t.Fatalf("Expected 2 rows affected, got %d", res.RowsAffected)
	}
}

// TestWorkerPipelineMultiBatch generates a CSV with 2100 rows (spanning three
// batches of 1000) and imports it via ImportCSV. It verifies the full pipeline
// including the concurrent job-feeder goroutine, worker pool, and results
// collection.
func TestWorkerPipelineMultiBatch(t *testing.T) {
	const rowCount = 2100

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "TestTypes.csv")
	f, err := os.Create(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	w := csv.NewWriter(f)
	if err := w.Write([]string{
		"wordcol", "codecol", "textcol", "bigtextcol",
		"numbercol", "intcol", "boolcol", "datecol",
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < rowCount; i++ {
		if err := w.Write([]string{
			fmt.Sprintf("word%d", i),
			fmt.Sprintf("CD%04d", i),
			"test text",
			"test big text",
			"1.23",
			"42",
			"true",
			"2024-01-01T00:00:00",
		}); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	mcsv := ModelCSV{}
	db := mcsv.ConnectDB()
	// Ensure the schema exists before counting so GORM does not log a warning.
	mcsv.CreateSchema(db, MakeModels())

	var before int64
	db.Table("test_types").Count(&before)

	mcsv.ImportCSV(tmpDir)

	var after int64
	db.Table("test_types").Count(&after)

	added := after - before
	if added < rowCount {
		t.Errorf("expected at least %d rows added, got %d", rowCount, added)
	}
}
