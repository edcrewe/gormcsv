package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/edcrewe/gormcsv/meta"
)

// jsonOutput is set by --json on inspectcsv and importcsv, for programs that drive the CLI.
var jsonOutput bool

// InspectReport is the --json output of inspectcsv.
type InspectReport struct {
	Output      string              `json:"output"`
	Models      []InspectModel      `json:"models"`
	SkippedRows map[string][]int    `json:"skippedRows"`
	Warnings    map[string][]string `json:"warnings"`
}

// InspectModel is one generated model.
type InspectModel struct {
	// Key is the factory key importcsv matches against a CSV filename.
	Key    string       `json:"key"`
	Name   string       `json:"name"`
	Fields []meta.Field `json:"fields"`
}

// ImportReport is the --json output of importcsv. Errors lists every rejected row or file;
// the command still exits non-zero when it is not empty.
type ImportReport struct {
	Files      []ImportFile `json:"files"`
	Inserted   int64        `json:"inserted"`
	Duplicates int64        `json:"duplicates"`
	Rejected   int64        `json:"rejected"`
	Errors     []string     `json:"errors"`
}

// ImportFile is the outcome for one CSV file.
type ImportFile struct {
	File       string `json:"file"`
	Model      string `json:"model"`
	Inserted   int64  `json:"inserted"`
	Duplicates int64  `json:"duplicates"`
	Rejected   int64  `json:"rejected"`
}

func newInspectReport(csvMeta meta.CSVMeta, output string) InspectReport {
	report := InspectReport{
		Output:      output,
		Models:      make([]InspectModel, 0, len(csvMeta.Models)),
		SkippedRows: csvMeta.SkippedRows,
		Warnings:    csvMeta.Warnings,
	}
	for key, name := range csvMeta.Models {
		fields := csvMeta.Fields[name]
		if fields == nil {
			fields = []meta.Field{}
		}
		report.Models = append(report.Models, InspectModel{Key: key, Name: name, Fields: fields})
	}
	sort.Slice(report.Models, func(i, j int) bool { return report.Models[i].Key < report.Models[j].Key })
	return report
}

func newImportReport(result importcsv.Result, importErr error) ImportReport {
	report := ImportReport{
		Files:      make([]ImportFile, 0, len(result.Files)),
		Inserted:   result.Inserted,
		Duplicates: result.Duplicates,
		Rejected:   result.Rejected,
		Errors:     []string{},
	}
	for _, file := range result.Files {
		report.Files = append(report.Files, ImportFile{
			File:       file.File,
			Model:      file.Model,
			Inserted:   file.Inserted,
			Duplicates: file.Duplicates,
			Rejected:   file.Rejected,
		})
	}
	for _, err := range flattenErrors(importErr) {
		report.Errors = append(report.Errors, err.Error())
	}
	return report
}

// flattenErrors unpacks errors.Join trees, so each rejected row is its own entry.
func flattenErrors(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var all []error
		for _, inner := range joined.Unwrap() {
			all = append(all, flattenErrors(inner)...)
		}
		return all
	}
	return []error{err}
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}
