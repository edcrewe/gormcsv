package inspectcsv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edcrewe/gormcsv/meta"
)

func TestGenerate(t *testing.T) {
	// Create mock CSVMeta
	csvMeta := meta.CSVMeta{
		UsesTime: true,
		Models:   map[string]string{"country": "Country"},
		Fields: map[string][]meta.Field{
			"Country": {
				{Name: "Name", Type: "string", Tag: ""},
				{Name: "Code", Type: "string", Tag: ""},
				{Name: "CreatedAt", Type: "time.Time", Tag: ""},
			},
		},
	}

	var buf bytes.Buffer
	err := Generate(&buf, csvMeta)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := buf.String()

	// Check output contains expected strings
	expectedStrings := []string{
		"package importcsv",
		"NewModelFactory",
		"\"country\": func() any",
		"type Country struct",
		"Name      string",
		"Code      string",
		"CreatedAt time.Time",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}

	// Make sure it doesn't contain sqlite import
	if strings.Contains(output, "gorm.io/driver/sqlite") {
		t.Errorf("Generated output should not contain sqlite driver import")
	}
}

func TestGenerateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.go")
	csvMeta := meta.CSVMeta{
		Models: map[string]string{"country": "Country"},
		Fields: map[string][]meta.Field{"Country": {{Name: "Name", Type: "string"}}},
	}
	if err := GenerateFile(csvMeta, path); err != nil {
		t.Fatalf("GenerateFile: %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "type Country struct") {
		t.Errorf("generated file did not contain Country model:\n%s", contents)
	}
}
