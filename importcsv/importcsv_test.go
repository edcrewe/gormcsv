package importcsv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestImportTestTypesPreservesValues(t *testing.T) {
	importer, db := testImporter(t, MakeModels(), 2)
	result, err := importer.Import(context.Background(), "../static/fixtures/TestTypes.csv")
	if err != nil {
		t.Fatalf("Import returned an error: %v", err)
	}
	if result.Inserted != 6 || result.Duplicates != 0 || result.Rejected != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}

	var record TestTypes
	if err := db.Where("codecol = ? AND bigtextcol = ?", "RF024", "Tent Family 17.5m2").First(&record).Error; err != nil {
		t.Fatalf("query imported record: %v", err)
	}
	if record.Wordcol != "tent" || record.Textcol != "Save UK" || record.Bigtextcol != "Tent Family 17.5m2" {
		t.Errorf("string fields were not preserved: %+v", record)
	}
	if record.Numbercol != 32.45 || record.Intcol != 45 || !record.Boolcol || record.Datecol != "2014-02-21" {
		t.Errorf("typed fields were not preserved: %+v", record)
	}
}

func TestImportCountryReportsMalformedRowAndKeepsValidRows(t *testing.T) {
	importer, db := testImporter(t, MakeModels(), 100)
	result, err := importer.Import(context.Background(), "../static/fixtures/Country.csv")
	if err == nil || !strings.Contains(err.Error(), "wrong number of fields") {
		t.Fatalf("expected malformed-row error, got %v", err)
	}
	if result.Inserted != 245 || result.Rejected != 1 {
		t.Fatalf("unexpected result: %+v, error: %v", result, err)
	}

	var country Country
	if err := db.Where("code = ?", "AF").First(&country).Error; err != nil {
		t.Fatalf("query imported country: %v", err)
	}
	if country.Name != "AFGHANISTAN" || country.Latitude != 33 || country.Longitude != 65 || country.Alias != "Afghanistan" {
		t.Errorf("country fields were not preserved: %+v", country)
	}
}

func TestImportBatchesAndSkipsDuplicatesDeterministically(t *testing.T) {
	type BatchUser struct {
		ID    uint
		Name  string `gorm:"uniqueIndex"`
		Score int
	}
	factory := NewModelFactory(map[string]func() any{
		"batchuser": func() any { return &BatchUser{} },
	})
	importer, db := testImporterWithWorkers(t, factory, 2, 1)
	path := writeCSV(t, "BatchUser.csv", [][]string{
		{"name", "score"},
		{"Alice", "10"},
		{"Bob", "20"},
		{"Alice", "99"},
		{"Carol", "not-a-number"},
	})

	result, err := importer.Import(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "not-a-number") {
		t.Fatalf("expected conversion error, got %v", err)
	}
	if result.Inserted != 2 || result.Duplicates != 1 || result.Rejected != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	var alice BatchUser
	if err := db.Where("name = ?", "Alice").First(&alice).Error; err != nil {
		t.Fatalf("query Alice: %v", err)
	}
	if alice.Score != 10 {
		t.Errorf("first duplicate should win, got score %d", alice.Score)
	}
}

func TestImportMultipleBatches(t *testing.T) {
	type BatchUser struct {
		ID   uint
		Name string
	}
	factory := NewModelFactory(map[string]func() any{
		"batchuser": func() any { return &BatchUser{} },
	})
	importer, db := testImporter(t, factory, 1000)
	records := make([][]string, 1, 2101)
	records[0] = []string{"name"}
	for i := 0; i < 2100; i++ {
		records = append(records, []string{fmt.Sprintf("user-%04d", i)})
	}
	path := writeCSV(t, "BatchUser.csv", records)

	result, err := importer.Import(context.Background(), path)
	if err != nil {
		t.Fatalf("Import returned an error: %v", err)
	}
	if result.Inserted != 2100 {
		t.Fatalf("inserted %d rows, want 2100", result.Inserted)
	}
	var count int64
	if err := db.Model(&BatchUser{}).Count(&count).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 2100 {
		t.Fatalf("stored %d users, want 2100", count)
	}
	for _, name := range []string{"user-0000", "user-1099", "user-2099"} {
		var user BatchUser
		if err := db.Where("name = ?", name).First(&user).Error; err != nil {
			t.Errorf("query %s: %v", name, err)
		}
	}
}

func TestImportUnquotedCRSeparatedCSV(t *testing.T) {
	type Item struct {
		ID          uint
		Name        string
		Description string
	}
	factory := NewModelFactory(map[string]func() any{
		"item": func() any { return &Item{} },
	})
	importer, db := testImporter(t, factory, 2)
	path := filepath.Join(t.TempDir(), "Item.csv")
	data := "name,description\rtent,\"Family, 17.5m2\"\rbucket,20 litre\r"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := importer.Import(context.Background(), path)
	if err != nil {
		t.Fatalf("Import returned an error: %v", err)
	}
	if result.Inserted != 2 || result.Rejected != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	var item Item
	if err := db.Where("name = ?", "tent").First(&item).Error; err != nil {
		t.Fatalf("query imported item: %v", err)
	}
	if item.Description != "Family, 17.5m2" {
		t.Errorf("description = %q, want %q", item.Description, "Family, 17.5m2")
	}
}

func TestImportRejectsInvalidInput(t *testing.T) {
	importer, _ := testImporter(t, MakeModels(), 0)
	if _, err := importer.Import(context.Background(), "nonexistent.csv"); err == nil {
		t.Fatal("expected a missing-file error")
	}
	path := writeCSV(t, "Unknown.csv", [][]string{{"name"}, {"value"}})
	if _, err := importer.Import(context.Background(), path); err == nil || !strings.Contains(err.Error(), "no model registered") {
		t.Fatalf("expected an unknown-model error, got %v", err)
	}
}

func TestNewValidatesConfiguration(t *testing.T) {
	if _, err := New(nil, MakeModels(), Config{BatchSize: 1}); err == nil {
		t.Fatal("expected nil database error")
	}
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	closeDB(t, db)
	if _, err := New(db, ModelFactory{}, Config{BatchSize: 1}); err == nil {
		t.Fatal("expected empty factory error")
	}
	if _, err := New(db, MakeModels(), Config{BatchSize: -1}); err == nil {
		t.Fatal("expected negative batch size error")
	}
	if _, err := New(db, MakeModels(), Config{Workers: -1}); err == nil {
		t.Fatal("expected negative worker count error")
	}
}

func TestOpenDatabaseRejectsUnsupportedDriver(t *testing.T) {
	if _, err := OpenDatabase("unknown", "dsn"); err == nil {
		t.Fatal("expected unsupported-driver error")
	}
}

func testImporter(t *testing.T, factory ModelFactory, batchSize int) (*Importer, *gorm.DB) {
	t.Helper()
	workers := 1
	if usePostgresTests {
		workers = 4
	}
	return testImporterWithWorkers(t, factory, batchSize, workers)
}

func testImporterWithWorkers(t *testing.T, factory ModelFactory, batchSize, workers int) (*Importer, *gorm.DB) {
	t.Helper()
	var db *gorm.DB
	if usePostgresTests {
		db = openPostgresTestDatabase(t)
	} else {
		var err error
		db, err = OpenSQLite(filepath.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatalf("OpenSQLite: %v", err)
		}
	}
	closeDB(t, db)
	importer, err := New(db, factory, Config{BatchSize: batchSize, Workers: workers})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return importer, db
}

func closeDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
}

func writeCSV(t *testing.T, name string, records [][]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create CSV: %v", err)
	}
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		t.Fatalf("write CSV: %v", err)
	}
	if err := writer.Error(); err != nil {
		t.Fatalf("flush CSV: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close CSV: %v", err)
	}
	return path
}
