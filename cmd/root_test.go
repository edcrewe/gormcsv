package cmd

import (
	"bytes"
	"testing"
)

// execute runs args through a fresh command tree, as main does, and returns its stdout, stderr
// and error.
func execute(args ...string) (string, string, error) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), errOut.String(), err
}

// main prints errors, so the command tree must not print them too.
func TestExecuteReturnsErrorsWithoutPrintingThem(t *testing.T) {
	out, errOut, err := execute("inspectcsv")

	if err == nil || err.Error() != "--files is required" {
		t.Fatalf("Execute() error = %v, want --files is required", err)
	}
	if errOut != "" || out != "" {
		t.Errorf("unexpected output: stdout %q, stderr %q", out, errOut)
	}
}
