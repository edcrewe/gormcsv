package inspectcsv

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/edcrewe/gormcsv/meta"
)

func TestGenerate(t *testing.T) {
	// Create mock CSVMeta
	csvMeta := meta.CSVMeta{
		Now:    time.Now(),
		Models: map[string]string{"country": "Country"},
		Fields: map[string][]meta.Field{
			"Country": {
				{Name: "Name", Type: "string", Tag: ""},
				{Name: "Code", Type: "string", Tag: ""},
				{Name: "Population", Type: "int64", Tag: ""},
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
		"type ModelFactory struct",
		"factory.models = append(factory.models, \"country\")",
		"type Country struct",
		"Name string",
		"Code string",
		"Population int64",
		"case \"country\":",
		"return &Country{}",
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
