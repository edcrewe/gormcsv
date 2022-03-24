/* Estimate the field types from a CSV file and create a model for it
Add method to try convert and narrow down to type then write meta
Need csv read method from ModelCSV.ImportCSV too so maybe refactor OOP for that

uint8       the set of all unsigned  8-bit integers (0 to 255)
uint16      the set of all unsigned 16-bit integers (0 to 65535)
uint32      the set of all unsigned 32-bit integers (0 to 4294967295)
uint64      the set of all unsigned 64-bit integers (0 to 18446744073709551615)

int8        the set of all signed  8-bit integers (-128 to 127)
int16       the set of all signed 16-bit integers (-32768 to 32767)
int32       the set of all signed 32-bit integers (-2147483648 to 2147483647)
int64       the set of all signed 64-bit integers (-9223372036854775808 to 9223372036854775807)

float32     the set of all IEEE-754 32-bit floating-point numbers
float64     the set of all IEEE-754 64-bit floating-point numbers

*/
package inspectcsv

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/edcrewe/gormcsv/meta"
)

//go:embed models.tmpl
var modelsTemplate string

// Generate use csvmeta to generate each model for models.go
func Generate(csvMeta meta.CSVMeta) error {
	t := template.Must(template.New("models").Parse(modelsTemplate))
	f, err := os.Create("importcsv/models.go")
	if err != nil {
		return err
	}
	err = t.Execute(f, csvMeta)
	if err != nil {
		return err
	} else {
		fmt.Println("Generated importcsv/models.go")
	}
	return nil
}
