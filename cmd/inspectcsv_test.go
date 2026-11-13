package cmd

import (
	"bytes"
	"testing"
)

func TestWriteSkippedRows(t *testing.T) {
	var out bytes.Buffer
	skipped := map[string][]int{
		"b.csv": {3, 9},
		"a.csv": {148},
	}
	if err := writeSkippedRows(&out, skipped); err != nil {
		t.Fatal(err)
	}
	want := "warning: a.csv: skipped 1 row with the wrong number of fields (line 148)\n" +
		"warning: b.csv: skipped 2 rows with the wrong number of fields (lines 3, 9)\n"
	if got := out.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteSkippedRowsWritesNothingWhenClean(t *testing.T) {
	var out bytes.Buffer
	if err := writeSkippedRows(&out, map[string][]int{}); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Errorf("unexpected output %q", out.String())
	}
}
