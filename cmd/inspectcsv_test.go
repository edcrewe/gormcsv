package cmd

import (
	"bytes"
	"reflect"
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

func TestWriteWarnings(t *testing.T) {
	var out bytes.Buffer
	warnings := map[string][]string{
		"b.csv": {"second"},
		"a.csv": {"first", "also first"},
	}
	if err := writeWarnings(&out, warnings); err != nil {
		t.Fatal(err)
	}
	want := "warning: a.csv: first\nwarning: a.csv: also first\nwarning: b.csv: second\n"
	if got := out.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestParseKeys(t *testing.T) {
	got, err := parseKeys([]string{"Country=code", "order-line = order_id + line"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"country": {"code"}, "order-line": {"order_id", "line"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseKeys() = %v, want %v", got, want)
	}
	for _, bad := range [][]string{
		{"country"},
		{"=code"},
		{"country="},
		{"country=code+"},
		{"x=a", "X=b"},
	} {
		if _, err := parseKeys(bad); err == nil {
			t.Errorf("parseKeys(%q) should fail", bad)
		}
	}
}
