package schema

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edcrewe/gormcsv/importcsv"
	"gorm.io/gorm"
)

type Country struct {
	gorm.Model
	Code string `gorm:"uniqueIndex"`
	Name string
}

type OrderLine struct {
	ID      uint
	OrderID string `gorm:"uniqueIndex:idx_orderline_key"`
	Line    int8   `gorm:"uniqueIndex:idx_orderline_key"`
}

func openDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	driver, dsn := "sqlite", filepath.Join(t.TempDir(), "schema.db")
	if pg := os.Getenv("GORMCSV_TEST_POSTGRES_DSN"); pg != "" {
		driver, dsn = "postgres", pg
	}
	db, err := importcsv.OpenDatabase(driver, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Migrator().DropTable(&Country{}, &OrderLine{})
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.Migrator().DropTable(&Country{}, &OrderLine{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Country{}, &OrderLine{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestDescribeListsTablesWithRowsColumnsAndIndexes(t *testing.T) {
	db := openDatabase(t)
	if err := db.Create(&[]Country{{Code: "AF", Name: "Afghanistan"}, {Code: "AL", Name: "Albania"}}).Error; err != nil {
		t.Fatal(err)
	}

	tables, err := Describe(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	names := make([]string, len(tables))
	for i, table := range tables {
		names[i] = table.Name
	}
	if !reflect.DeepEqual(names, []string{"countries", "order_lines"}) {
		t.Fatalf("tables = %v, want countries and order_lines only", names)
	}
	countries, orderLines := tables[0], tables[1]
	if countries.Rows != 2 || orderLines.Rows != 0 {
		t.Errorf("rows = %d and %d, want 2 and 0", countries.Rows, orderLines.Rows)
	}

	columnNames := make([]string, len(countries.Columns))
	for i, column := range countries.Columns {
		columnNames[i] = column.Name
	}
	want := []string{"id", "created_at", "updated_at", "deleted_at", "code", "name"}
	if !reflect.DeepEqual(columnNames, want) {
		t.Errorf("columns = %v, want %v in table order", columnNames, want)
	}
	if !countries.Columns[0].PrimaryKey || countries.Columns[4].PrimaryKey {
		t.Errorf("only id should be the primary key: %+v", countries.Columns)
	}
	if countries.Columns[0].Type == "" {
		t.Error("columns should report a database type")
	}

	if !hasIndex(countries.Indexes, Index{Name: "idx_countries_code", Columns: []string{"code"}, Unique: true}) {
		t.Errorf("missing unique code index: %+v", countries.Indexes)
	}
	composite := Index{Name: "idx_orderline_key", Columns: []string{"order_id", "line"}, Unique: true}
	if !hasIndex(orderLines.Indexes, composite) {
		t.Errorf("missing composite key index: %+v", orderLines.Indexes)
	}
}

func hasIndex(indexes []Index, want Index) bool {
	for _, index := range indexes {
		if reflect.DeepEqual(index, want) {
			return true
		}
	}
	return false
}

func TestDescribeEmptyDatabase(t *testing.T) {
	db, err := importcsv.OpenSQLite(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	tables, err := Describe(context.Background(), db)
	if err != nil || tables == nil || len(tables) != 0 {
		t.Errorf("Describe() = %v, %v; want an empty, non-nil list", tables, err)
	}
}
