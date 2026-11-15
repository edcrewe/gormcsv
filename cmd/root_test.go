package cmd

import (
	"bytes"
	"testing"
)

// main prints errors, so the command tree must not print them too.
func TestExecuteReturnsErrorsWithoutPrintingThem(t *testing.T) {
	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs([]string{"inspectcsv"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	err := rootCmd.Execute()

	if err == nil || err.Error() != "--files is required" {
		t.Fatalf("Execute() error = %v, want --files is required", err)
	}
	if errOut.Len() != 0 || out.Len() != 0 {
		t.Errorf("unexpected output: stdout %q, stderr %q", out.String(), errOut.String())
	}
}
