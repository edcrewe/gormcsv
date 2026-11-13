package meta

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGetField(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{name: "empty", values: []string{"", ""}, want: "string"},
		{name: "bool", values: []string{"true", "False"}, want: "bool"},
		{name: "date", values: []string{"2026-09-21", "2024-01-01"}, want: "time.Time"},
		{name: "signed", values: []string{"-123", "24", "201"}, want: "int16"},
		{name: "int64", values: []string{"214748364234", "234234234233"}, want: "int64"},
		{name: "uint64", values: []string{"18446744073709551615"}, want: "uint64"},
		{name: "float", values: []string{"13213.33", "-123.2"}, want: "float64"},
		{name: "string", values: []string{"RF024", "WA041"}, want: "string"},
	}
	var csvMeta CSVMeta
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := csvMeta.GetField("Value", test.values).Type; got != test.want {
				t.Errorf("GetField() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestExportedNameStartsWithLetter(t *testing.T) {
	if got := exportedName("2026_sales"); got != "Field2026Sales" {
		t.Errorf("exportedName() = %q, want Field2026Sales", got)
	}
}

func TestPopulateMetaPreservesHeaderOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "event-log.csv")
	contents := "event_name,event_date,total-count\nlaunch,2026-09-21,-12\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	var csvMeta CSVMeta
	if err := csvMeta.PopulateMeta(path); err != nil {
		t.Fatalf("PopulateMeta: %v", err)
	}
	if got := csvMeta.Models["event-log"]; got != "EventLog" {
		t.Errorf("model = %q, want EventLog", got)
	}
	want := []Field{
		{Name: "EventName", Type: "string"},
		{Name: "EventDate", Type: "time.Time"},
		{Name: "TotalCount", Type: "int8"},
	}
	if got := csvMeta.Fields["EventLog"]; !reflect.DeepEqual(got, want) {
		t.Errorf("fields = %#v, want %#v", got, want)
	}
	if !csvMeta.UsesTime {
		t.Error("UsesTime is false, want true")
	}
}

func TestCSVFilesFiltersAndSorts(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{"B.csv", "A.CSV", "ignored.txt"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("name\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := CSVFiles(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Model != "A" || files[1].Model != "B" {
		t.Errorf("unexpected files: %#v", files)
	}
}

func TestPopulateMetaSkipsRowsWithWrongFieldCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ragged.csv")
	contents := "name,count\nalpha,1\nshort\nbeta,2,extra\ngamma,300\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	var csvMeta CSVMeta
	if err := csvMeta.PopulateMeta(path); err != nil {
		t.Fatalf("PopulateMeta: %v", err)
	}
	if got, want := csvMeta.SkippedRows[path], []int{3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("skipped rows = %v, want %v", got, want)
	}
	// Only the well-formed rows drive inference: 1 and 300 need int16, and "2,extra" is ignored.
	want := []Field{{Name: "Name", Type: "string"}, {Name: "Count", Type: "int16"}}
	if got := csvMeta.Fields["Ragged"]; !reflect.DeepEqual(got, want) {
		t.Errorf("fields = %#v, want %#v", got, want)
	}
}

func TestPopulateMetaInspectsCountryFixtureDespiteShortRow(t *testing.T) {
	path := "../static/fixtures/Country.csv"
	var csvMeta CSVMeta
	if err := csvMeta.PopulateMeta(path); err != nil {
		t.Fatalf("PopulateMeta: %v", err)
	}
	if got, want := csvMeta.SkippedRows[path], []int{148}; !reflect.DeepEqual(got, want) {
		t.Errorf("skipped rows = %v, want %v", got, want)
	}
	if got := csvMeta.Models["country"]; got != "Country" {
		t.Errorf("model = %q, want Country", got)
	}
	if fields := csvMeta.Fields["Country"]; len(fields) == 0 || fields[0] != (Field{Name: "Name", Type: "string"}) {
		t.Errorf("unexpected Country fields: %#v", fields)
	}
}

func TestPopulateMetaStillFailsOnBadQuoting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes.csv")
	if err := os.WriteFile(path, []byte("a,b\n1,\"open\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var csvMeta CSVMeta
	err := csvMeta.PopulateMeta(path)
	if err == nil || !strings.Contains(err.Error(), "quoted-field") {
		t.Fatalf("expected quoting error, got %v", err)
	}
}

func TestPopulateMetaHasNoSkippedRowsForCleanFile(t *testing.T) {
	var csvMeta CSVMeta
	if err := csvMeta.PopulateMeta("../static/fixtures/TestTypes.csv"); err != nil {
		t.Fatalf("PopulateMeta: %v", err)
	}
	if len(csvMeta.SkippedRows) != 0 {
		t.Errorf("skipped rows = %v, want none", csvMeta.SkippedRows)
	}
}
