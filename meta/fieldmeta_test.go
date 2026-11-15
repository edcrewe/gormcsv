package meta

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type mappedModel struct {
	Name      string
	Count     int16
	Enabled   bool
	Price     float64
	CreatedAt time.Time
}

func TestRecordToModel(t *testing.T) {
	fieldMeta, err := NewFieldMeta(&mappedModel{}, []string{"enabled", "name", "price", "count", "createdat"})
	if err != nil {
		t.Fatalf("NewFieldMeta: %v", err)
	}
	converted, err := fieldMeta.RecordToModel(&mappedModel{}, []string{"true", "example", "12.5", "42", "2026-09-21"})
	if err != nil {
		t.Fatalf("RecordToModel: %v", err)
	}
	model := converted.(*mappedModel)
	if model.Name != "example" || model.Count != 42 || !model.Enabled || model.Price != 12.5 {
		t.Errorf("unexpected model: %+v", model)
	}
	if want := time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC); !model.CreatedAt.Equal(want) {
		t.Errorf("CreatedAt = %v, want %v", model.CreatedAt, want)
	}
}

func TestRecordToModelEmptyValuesUseZeroValues(t *testing.T) {
	fieldMeta, err := NewFieldMeta(&mappedModel{}, []string{"name", "count", "enabled", "price"})
	if err != nil {
		t.Fatal(err)
	}
	converted, err := fieldMeta.RecordToModel(&mappedModel{}, []string{"", "", "", ""})
	if err != nil {
		t.Fatal(err)
	}
	if got := converted.(*mappedModel); *got != (mappedModel{}) {
		t.Errorf("got %+v, want zero value", got)
	}
}

func TestNewFieldMetaRejectsInvalidHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header []string
	}{
		{name: "empty", header: nil},
		{name: "blank", header: []string{""}},
		{name: "duplicate", header: []string{"name", "NAME"}},
		{name: "unknown", header: []string{"missing"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewFieldMeta(&mappedModel{}, test.header); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestRecordToModelRejectsConversionAndShortRows(t *testing.T) {
	fieldMeta, err := NewFieldMeta(&mappedModel{}, []string{"name", "count"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fieldMeta.RecordToModel(&mappedModel{}, []string{"name", "invalid"}); err == nil {
		t.Fatal("expected conversion error")
	}
	if _, err := fieldMeta.RecordToModel(&mappedModel{}, []string{"name"}); err == nil {
		t.Fatal("expected short-row error")
	}
}

func TestNewFieldMetaMatchesGeneratedFieldNames(t *testing.T) {
	type OrderItem struct {
		UnitPrice   float64
		Field1stCol int8
		TotalCount  int16
	}
	headers := []string{"unit_price", "1st col", "total-count"}
	fieldMeta, err := NewFieldMeta(&OrderItem{}, headers)
	if err != nil {
		t.Fatalf("NewFieldMeta: %v", err)
	}
	model, err := fieldMeta.RecordToModel(&OrderItem{}, []string{"2.5", "7", "-3"})
	if err != nil {
		t.Fatalf("RecordToModel: %v", err)
	}
	got := *model.(*OrderItem)
	if want := (OrderItem{UnitPrice: 2.5, Field1stCol: 7, TotalCount: -3}); got != want {
		t.Errorf("RecordToModel() = %+v, want %+v", got, want)
	}
}

// Every name the generator produces must be importable, or a generated model cannot load its own CSV.
func TestGeneratedNamesRoundTripThroughFieldMeta(t *testing.T) {
	path := filepath.Join(t.TempDir(), "order-item.csv")
	contents := "item name,unit_price,2nd-qty\nwidget,1.5,3\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	var csvMeta CSVMeta
	if err := csvMeta.PopulateMeta(path); err != nil {
		t.Fatalf("PopulateMeta: %v", err)
	}
	fields := csvMeta.Fields["OrderItem"]
	structFields := make([]reflect.StructField, len(fields))
	for i, field := range fields {
		structFields[i] = reflect.StructField{Name: field.Name, Type: reflect.TypeOf("")}
	}
	model := reflect.New(reflect.StructOf(structFields)).Interface()
	if _, err := NewFieldMeta(model, []string{"item name", "unit_price", "2nd-qty"}); err != nil {
		t.Fatalf("generated fields %v cannot import their own headers: %v", fields, err)
	}
}

type gormLike struct {
	ID        uint
	CreatedAt time.Time
}

type generatedThing struct {
	gormLike
	IdField    int8
	ModelField string
	Name       string
}

func TestNewFieldMetaMapsReservedHeadingsToRenamedFields(t *testing.T) {
	fieldMeta, err := NewFieldMeta(&generatedThing{}, []string{"id", "model", "name"})
	if err != nil {
		t.Fatalf("NewFieldMeta: %v", err)
	}
	model, err := fieldMeta.RecordToModel(&generatedThing{}, []string{"7", "X1", "alpha"})
	if err != nil {
		t.Fatalf("RecordToModel: %v", err)
	}
	got := model.(*generatedThing)
	if got.ID != 0 || got.IdField != 7 || got.ModelField != "X1" || got.Name != "alpha" {
		t.Errorf("RecordToModel() = %+v; the id column must not set the primary key", got)
	}
}

type olderGenerated struct {
	gormLike
	Name string
}

func TestNewFieldMetaProtectsEmbeddedFieldsForOlderModels(t *testing.T) {
	// A model generated before renaming has no IdField: an id column must fail, not set ID.
	if _, err := NewFieldMeta(&olderGenerated{}, []string{"id", "name"}); err == nil {
		t.Error("expected an id heading to be rejected rather than written to the embedded ID")
	}
	// Its "model" column was always ignored, and still is.
	if _, err := NewFieldMeta(&olderGenerated{}, []string{"model", "name"}); err != nil {
		t.Errorf("model heading should be ignored for older models: %v", err)
	}
}

type hashedModel struct {
	Name    string
	Code    string
	RowHash string
}

func hashOf(t *testing.T, fieldMeta *FieldMeta, record []string) string {
	t.Helper()
	model, err := fieldMeta.RecordToModel(&hashedModel{}, record)
	if err != nil {
		t.Fatal(err)
	}
	return model.(*hashedModel).RowHash
}

func TestRecordToModelFillsRowHash(t *testing.T) {
	fieldMeta, err := NewFieldMeta(&hashedModel{}, []string{"name", "code", "model"})
	if err != nil {
		t.Fatal(err)
	}
	first := hashOf(t, fieldMeta, []string{"Afghanistan", "AF", "ignored"})

	if len(first) != 64 {
		t.Fatalf("RowHash = %q, want 64 hex characters", first)
	}
	if again := hashOf(t, fieldMeta, []string{"Afghanistan", "AF", "different"}); again != first {
		t.Error("an ignored column should not change the hash")
	}
	if other := hashOf(t, fieldMeta, []string{"Afghanistan", "AL", "ignored"}); other == first {
		t.Error("different values should give a different hash")
	}
	// Length prefixes keep ("ab","c") and ("a","bc") apart.
	if hashOf(t, fieldMeta, []string{"ab", "c", ""}) == hashOf(t, fieldMeta, []string{"a", "bc", ""}) {
		t.Error("values must not run together in the hash")
	}
}

func TestRowHashIsNotOverwrittenWhenTheCSVFillsIt(t *testing.T) {
	fieldMeta, err := NewFieldMeta(&hashedModel{}, []string{"name", "rowhash"})
	if err != nil {
		t.Fatal(err)
	}
	if got := hashOf(t, fieldMeta, []string{"x", "from-csv"}); got != "from-csv" {
		t.Errorf("RowHash = %q, want the CSV value", got)
	}
}
