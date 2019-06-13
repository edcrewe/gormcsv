package tests


import (
	"github.com/edcrewe/gormcsv/importcsv"
	"strings"
	"testing"
)

var csvmeta importcsv.CSVMeta
/*
Setup first then run the tests
*/
func TestMain(m *testing.M) {
	csvmeta = importcsv.CSVMeta{}
	m.Run()
}

/*
Test the CSVMeta.GetField function
*/
func TestGetField(t *testing.T) {

	type TableTest struct {
		input   string
		typeStr string
		pass    bool
	}

	var tableTests = []TableTest {
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
		if (output.Type == test.typeStr && !test.pass || (output.Type != test.typeStr && test.pass)) {
			t.Errorf("csvmeta.GetField (%s) %s == %v is not %v", test.input, test.typeStr,
				output.Type, test.pass)
		}
	}
}

