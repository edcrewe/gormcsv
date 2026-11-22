package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/edcrewe/gormcsv/meta"
	"github.com/edcrewe/gormcsv/schema"
)

func TestNewInspectReportSortsModelsAndNeverHasNullFields(t *testing.T) {
	csvMeta := meta.CSVMeta{
		Models: map[string]string{"zebra": "Zebra", "apple": "Apple"},
		Fields: map[string][]meta.Field{"Zebra": {{Name: "Stripes", Type: "int8"}}},
	}

	report := newInspectReport(csvMeta, "models.go")

	if report.Models[0].Key != "apple" || report.Models[1].Key != "zebra" {
		t.Errorf("models not sorted by key: %+v", report.Models)
	}
	if report.Models[0].Fields == nil {
		t.Error("a model without fields should report [] not null")
	}
}

func TestNewImportReportListsEachJoinedErrorSeparately(t *testing.T) {
	rowErrors := errors.Join(errors.New("a.csv row 3: bad"), errors.New("a.csv row 4: bad"))
	importErr := errors.Join(rowErrors, errors.New("b.csv: no model registered"))
	result := importcsv.Result{
		Files:    []importcsv.FileResult{{File: "a.csv", Model: "A", Inserted: 2, Rejected: 2}},
		Inserted: 2,
		Rejected: 2,
	}

	report := newImportReport(result, importErr)

	want := []string{"a.csv row 3: bad", "a.csv row 4: bad", "b.csv: no model registered"}
	if !reflect.DeepEqual(report.Errors, want) {
		t.Errorf("errors = %q, want %q", report.Errors, want)
	}
	if report.Files[0] != (ImportFile{File: "a.csv", Model: "A", Inserted: 2, Rejected: 2}) {
		t.Errorf("files = %+v", report.Files)
	}
	if clean := newImportReport(importcsv.Result{}, nil); clean.Errors == nil || clean.Files == nil {
		t.Error("an empty report should have [] not null for errors and files")
	}
}

// run executes the CLI with args and returns stdout.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, _, err := execute(args...)
	return out, err
}

func TestInspectJSON(t *testing.T) {
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "thing.csv")
	if err := os.WriteFile(csvPath, []byte("model,count\nX,12\nshort\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	modelsPath := filepath.Join(directory, "models.go")

	stdout, err := run(t, "inspectcsv", "--files", csvPath, "--output", modelsPath, "--json")
	if err != nil {
		t.Fatal(err)
	}

	var report InspectReport
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	want := InspectReport{
		Output: modelsPath,
		Models: []InspectModel{{Key: "thing", Name: "Thing", Fields: []meta.Field{
			{Name: "ModelField", Type: "string", Column: "model"},
			{Name: "Count", Type: "int8", Column: "count"},
		}}},
		SkippedRows: map[string][]int{csvPath: {3}},
		Warnings: map[string][]string{csvPath: {
			`column "model" generates field ModelField, because Model is reserved by gorm.Model`,
		}},
	}
	if !reflect.DeepEqual(report, want) {
		t.Errorf("report = %+v\nwant %+v", report, want)
	}
}

func TestImportJSONReportsPartialResultsAndStillFails(t *testing.T) {
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "TestTypes.csv")
	contents := "wordcol,intcol\nok,1\nbad,notanumber\n"
	if err := os.WriteFile(csvPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, err := run(t, "importcsv", "--files", csvPath, "--dsn", filepath.Join(directory, "t.db"), "--json")

	if err == nil {
		t.Error("a rejected row should still make the command fail")
	}
	var report ImportReport
	if jsonErr := json.Unmarshal([]byte(stdout), &report); jsonErr != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", jsonErr, stdout)
	}
	if report.Inserted != 1 || report.Rejected != 1 || len(report.Errors) != 1 {
		t.Errorf("report = %+v", report)
	}
}

func TestTablesJSON(t *testing.T) {
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "TestTypes.csv")
	if err := os.WriteFile(csvPath, []byte("wordcol\nok\nalso\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(directory, "t.db")
	if _, err := run(t, "importcsv", "--files", csvPath, "--dsn", dbPath); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GORMCSV_DSN", dbPath)

	stdout, err := run(t, "tables", "--json")
	if err != nil {
		t.Fatal(err)
	}

	var report TablesReport
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if len(report.Tables) != 1 || report.Tables[0].Name != "test_types" || report.Tables[0].Rows != 2 {
		t.Errorf("report = %+v", report)
	}
}

func TestRowsJSONAndNotFound(t *testing.T) {
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "TestTypes.csv")
	if err := os.WriteFile(csvPath, []byte("wordcol\nb\na\nc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(directory, "t.db")
	if _, err := run(t, "importcsv", "--files", csvPath, "--dsn", dbPath); err != nil {
		t.Fatal(err)
	}
	stdout, err := run(t, "rows", "--dsn", dbPath, "--table", "test_types", "--limit", "2",
		"--offset", "1", "--order", "wordcol", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var page schema.Page
	if err := json.Unmarshal([]byte(stdout), &page); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	words := []any{}
	for _, row := range page.Rows {
		words = append(words, row[slices.Index(page.Columns, "wordcol")])
	}
	if page.Total != 3 || page.Order != "wordcol" || !reflect.DeepEqual(words, []any{"b", "c"}) {
		t.Errorf("page = %+v", page)
	}

	_, err = run(t, "rows", "--dsn", dbPath, "--table", "missing")
	if err == nil || err.Error() != `no table "missing": not found` {
		t.Errorf("err = %v", err)
	}
}

func TestPageSummary(t *testing.T) {
	rows := [][]any{{1}, {2}}
	tests := map[string]struct {
		page schema.Page
		want string
	}{
		"a page":        {schema.Page{Rows: rows, Offset: 40, Total: 245}, "rows 41-42 of 245"},
		"empty table":   {schema.Page{Rows: [][]any{}, Total: 0}, "no rows"},
		"past the end":  {schema.Page{Rows: [][]any{}, Offset: 999, Total: 6}, "no rows at offset 999; the table has 6"},
		"the last page": {schema.Page{Rows: rows[:1], Offset: 5, Total: 6}, "rows 6-6 of 6"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := pageSummary(test.page); got != test.want {
				t.Errorf("pageSummary() = %q, want %q", got, test.want)
			}
		})
	}
}
