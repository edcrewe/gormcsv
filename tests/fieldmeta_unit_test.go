// Unit tests for fieldmeta - importcsv reflection subpackage
package tests

import (
	"fmt"
	"github.com/edcrewe/gormcsv/importcsv"
	"testing"
)

var meta importcsv.FieldMeta
/*
Setup first then run the tests
 */
func TestMain(m *testing.M) {
	meta = importcsv.FieldMeta{}
	factory := importcsv.MakeModels()
	model := factory.New("country")
	meta.Setmeta(model, "name,code,latitude,longtitude,alias")
	m.Run()
}

/*
Test the FieldMeta.Convert function
 */
func TestConvert(t *testing.T) {

	type TableTest struct {
		input   string
		convert string
		pass    bool
	}

	var tableTests = []TableTest {
		{"123", "int", true},
		{"214748364234", "int", false},
		{"214748364234", "int32", true},
		{"13213.3427734375", "float32", true},
	}



	for _, test := range tableTests {
			output, _:= meta.Convert(test.input, test.convert)
			//fmt.Println(fmt.Sprint(output))
			if ((fmt.Sprint(output) == test.input && !test.pass) || (fmt.Sprint(output) != test.input && test.pass)) {
				t.Errorf("meta.Convert(%s) == %v is not %v", test.input, output, test.pass)
			}
	}

}
