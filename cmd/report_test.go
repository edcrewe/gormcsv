package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/edcrewe/gormcsv/meta"
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

// run executes the CLI with args and returns stdout, resetting the flags it sets.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		jsonOutput = false
		files = ""
	})
	err := rootCmd.Execute()
	return out.String(), err
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
	resetConnectionFlags(t)
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "TestTypes.csv")
	if err := os.WriteFile(csvPath, []byte("wordcol\nok\nalso\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(directory, "t.db")
	if _, err := run(t, "importcsv", "--files", csvPath, "--dsn", dbPath); err != nil {
		t.Fatal(err)
	}
	resetFlags := func() {
		flag := importCSVCommand.Flags().Lookup("dsn")
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	}
	resetFlags()
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
