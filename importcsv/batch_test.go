package importcsv

import (
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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
