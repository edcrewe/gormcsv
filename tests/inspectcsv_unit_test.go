//go:build unit || u
// +build unit u

package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/edcrewe/gormcsv/inspectcsv"
)

// TestPopulateMeta test CSVMeta.PopulateMeta
func TestPopulateMeta(t *testing.T) {
	csvmeta := inspectcsv.CSVMeta{}
	path := "fixtures"
	err := csvmeta.PopulateMeta(path)
	if err != nil {
		fmt.Printf("Failed to populate meta for path %s due to:\n %s\n", path, err)
	}
	type TableTest struct {
		model   string
		field   string
		typeStr string
		pass    bool
	}

	var tableTests = []TableTest{
		{"Country", "Name", "string", true},
		{"Country", "Code", "string", true},
		{"Country", "Latitude", "float32", true},
		{"Country", "Alias", "int16", false},
		//		{"item","DESCRIPTION", "string", true},
		//		{"item","QUANTITY", "int16", true},
		{"TestTypes", "Wordcol", "string", true},
		{"TestTypes", "Codecol", "int8", false},
		{"TestTypes", "Textcol", "string", true},
		{"TestTypes", "Numbercol", "float32", true},
		{"TestTypes", "Intcol", "int16", true},
		{"TestTypes", "Boolcol", "bool", true},
		//		{"testtypes","datecol", "date", true},
	}

	for _, test := range tableTests {
		fields, ok := csvmeta.Fields[test.model]
		if !ok {
			t.Errorf("csvmeta.Fields is missing the model %s", test.model)
			continue
		}
		// fmt.Print(csvmeta.Fields[test.model])
		found := false
		for _, field := range fields {
			if field.Name == test.field {
				found = true
				if field.Type == test.typeStr && !test.pass || (field.Type != test.typeStr && test.pass) {
					t.Errorf("csvmeta.Fields[%s]-%s %s == %v is not %v", test.model, test.field, test.typeStr,
						field.Type, test.pass)
				}
				break
			}
		}
		if !found {
			t.Errorf("csvmeta.Fields[%s] is missing the field %s", test.model, test.field)
		}
	}
}

// TestGetField test the CSVMeta.GetField function
func TestGetField(t *testing.T) {
	csvmeta := inspectcsv.CSVMeta{}

	type TableTest struct {
		input   string
		typeStr string
		pass    bool
	}

	var tableTests = []TableTest{
		{"23,24,55", "int8", true},
		{"21474836423412345,234567,23423423", "int64", true},
		{"-123,24,+12", "uint16", true},
		{"-123,24,201", "int16", false},
		{"32770,234234", "int16", false},
		{"214748366,1234562", "int32", true},
		{"214748364234,234234234233", "int32", false},
		{"13213.33,1234.23,-123123", "float32", true},
		{"2006-01-02T15:04:05", "date", true},
		{"false", "bool", true},
	}

	for _, test := range tableTests {
		input := strings.Split(test.input, ",")
		output := csvmeta.GetField("test", input)
		//fmt.Println(fmt.Sprint(output))
		if output.Type == test.typeStr && !test.pass || (output.Type != test.typeStr && test.pass) {
			t.Errorf("csvmeta.GetField (%s) %s == %v is not %v", test.input, test.typeStr,
				output.Type, test.pass)
		}
	}
}
