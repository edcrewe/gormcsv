package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

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
			command := newImportCommand(new(string))
			if err := command.ParseFlags(test.args); err != nil {
				t.Fatal(err)
			}
			gotDriver, gotDSN := connectionSettings(command, func(name string) string { return test.env[name] })
			if gotDriver != test.wantDriver || gotDSN != test.wantDSN {
				t.Errorf("connectionSettings() = %q, %q; want %q, %q", gotDriver, gotDSN, test.wantDriver, test.wantDSN)
			}
		})
	}
}

func TestImportUsesDSNFromEnvironment(t *testing.T) {
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
