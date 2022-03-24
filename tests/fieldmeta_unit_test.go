//go:build unit || u
// +build unit u

// Unit tests for fieldmeta - importcsv reflection subpackage
package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/edcrewe/gormcsv/common"
)

// TestConvert test the FieldMeta.Convert function
func TestConvert(t *testing.T) {
	meta := common.FieldMeta{}

	type TableTest struct {
		input   string
		convert string
		pass    bool
	}

	var tableTests = []TableTest{
		{"123", "int", true},
		{"214748364234123456789", "int", false},
		{"-123", "int", true},
		{"-123", "int8", true},
		{"32770", "int16", false},
		{"2147483646", "int32", true},
		{"214748364234", "int32", false},
		{"13213.3427734375", "float32", true},
		{"false", "bool", true},
	}

	for _, test := range tableTests {
		output, error := meta.Convert(test.input, test.convert)
		//fmt.Println(fmt.Sprint(output))
		if (fmt.Sprint(output) == test.input && !test.pass) || (fmt.Sprint(output) != test.input && test.pass) {
			t.Errorf("meta.Convert %s (%s) == %v is not %v %s", test.convert, test.input,
				output, test.pass, error)
		}
	}
	// DateTime Sprint not directly comparable so do separate test
	output, error := meta.Convert("2006-01-02T15:04:05", "date")
	if !strings.HasPrefix(fmt.Sprint(output), "2006-01-02") {
		t.Errorf("meta.Convert date failed. %s", error)
	}
}
