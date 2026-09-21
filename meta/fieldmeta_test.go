package meta

import (
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
