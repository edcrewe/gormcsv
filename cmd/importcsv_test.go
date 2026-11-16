package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// resetConnectionFlags puts --driver and --dsn back to unset, since cobra keeps flag state
// between Execute calls in one process.
func resetConnectionFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		for _, name := range []string{"driver", "dsn"} {
			flag := importCSVCommand.Flags().Lookup(name)
			_ = flag.Value.Set(flag.DefValue)
			flag.Changed = false
		}
	})
}

func TestConnectionSettings(t *testing.T) {
	env := map[string]string{"GORMCSV_DB_DRIVER": "postgres", "GORMCSV_DSN": "host=db password=secret"}
	tests := []struct {
		name       string
		args       []string
		env        map[string]string
		wantDriver string
		wantDSN    string
	}{
		{name: "defaults", wantDriver: "sqlite", wantDSN: "gormcsv.db"},
		{name: "environment", env: env, wantDriver: "postgres", wantDSN: "host=db password=secret"},
		{name: "flags win", args: []string{"--driver", "mysql", "--dsn", "x.db"}, env: env, wantDriver: "mysql", wantDSN: "x.db"},
		{name: "empty environment ignored", env: map[string]string{"GORMCSV_DSN": ""}, wantDriver: "sqlite", wantDSN: "gormcsv.db"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetConnectionFlags(t)
			if err := importCSVCommand.ParseFlags(test.args); err != nil {
				t.Fatal(err)
			}
			gotDriver, gotDSN := connectionSettings(importCSVCommand, func(name string) string { return test.env[name] })
			if gotDriver != test.wantDriver || gotDSN != test.wantDSN {
				t.Errorf("connectionSettings() = %q, %q; want %q, %q", gotDriver, gotDSN, test.wantDriver, test.wantDSN)
			}
		})
	}
}

func TestImportUsesDSNFromEnvironment(t *testing.T) {
	resetConnectionFlags(t)
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "TestTypes.csv")
	if err := os.WriteFile(csvPath, []byte("wordcol\nok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(directory, "from-env.db")
	t.Setenv("GORMCSV_DSN", dbPath)

	if _, err := run(t, "importcsv", "--files", csvPath); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database was not created at the GORMCSV_DSN path: %v", err)
	}
}
