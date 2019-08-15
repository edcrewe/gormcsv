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
package importcsv

import (
	"time"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type field struct {
	Name string
	Type string
	Tag string
}

var mapLength = map[int]string{2:"8", 4:"16", 9:"32", 20:"64"}
var keysLength = []int {2, 4, 9, 20}
var typeStrings = []string {"bool", "date", "string"}
var reNumber = regexp.MustCompile(`^[-+]?\d*\.?\d*$`)


type CSVMeta struct {
	Now time.Time
	Meta
	Files
	Models map[string]string
	Fields map[string][]field
}

/*
Use csvmeta to generate each model for models.go
 */
func (csvmeta *CSVMeta) Generate() error {
	t := template.Must(template.New("models").Parse(modelsTemplate))
	f, err := os.Create("importcsv/models.go")
	if err != nil {
		return err
	} else {
		fmt.Println("Generated importcsv/models.go")
	}
	t.Execute(f, csvmeta)
	return nil
}

/*
Read a sample set of data from each CSV file and determine the fields and field types
Load that metadata to csvmeta struct Models and Fields
 */
func (csvmeta *CSVMeta) PopulateMeta(path string) error {
	csvmeta.Now = time.Now()
	filesMap, err := csvmeta.FilesFetch(path)
	csvmeta.Models =  map[string]string{}
	csvmeta.Fields =  map[string][]field{}
	if err != nil {
		return fmt.Errorf("Failed to find CSV file(s) from %s, Due to %s", path, err)
	}
	sample := 5
	for model, csvFile := range filesMap {
		names := map[string][]string{}
		reader := csv.NewReader(bufio.NewReader(csvFile))
		modelLower := strings.ToLower(model)
		csvmeta.Models[modelLower] = model
		var keys []string
		if reader != nil {
			for i := 1; i <= sample; i++ {
				record, error := reader.Read()
				//fmt.Println(i)

				for index, _ := range record {
					if i == 1 {
						keys = make([]string, len(record))
						copy(keys, record)
						names[keys[index]] = []string{}
					} else {
						if len(keys) > index {
							field := keys[index]
							if field != "" {
								names[field] = append(names[field], record[index])
							}
						}
					}
				}
				if error == io.EOF {
					i = sample
				} else if error != nil {
					return fmt.Errorf("Failed to inspect %s due to %s", model, error)
				}
			}
		}
		for key, values := range names {
			field := csvmeta.GetField(key, values)
			csvmeta.Fields[model] = append(csvmeta.Fields[model], field)
		}
	}
	//fmt.Println(csvmeta.Fields["country"])
	return nil
}

/*
Take a name and list of values from CSV then test the values to work out the type
 */
func (csvmeta *CSVMeta) GetField(name string, valueStrings []string) field{
	var f = field{Name: name}
	var typeStr = ""
	var vLength = 0
	for _, valueStr := range valueStrings {
		if typeStr != "string" && reNumber.MatchString(valueStr) {
			if strings.Index(valueStr, ".") > -1 {
				typeStr = "float"
			}
			if typeStr != "float" {
				if strings.HasPrefix(valueStr, "-") || strings.HasPrefix(valueStr, "+") {
					typeStr = "uint"
				} else if typeStr != "uint" {
					typeStr = "int"
				}
			}
			if len(valueStr) > vLength {
				vLength = len(valueStr)
			}
		}
		if typeStr == "" {
			for _, typeString := range typeStrings {
				_, err := csvmeta.Convert(valueStr, typeString)
				if err == nil {
					typeStr = typeString
					break
				}
			}
		}
	}
	if typeStr != "string" && typeStr != "date" && typeStr != "bool" {
		for _, length := range keysLength {
			if vLength <= length {
				typeStr = typeStr + mapLength[length]
				break
			}
		}
		if typeStr == "float8" || typeStr == "float16" { typeStr = "float32" }
	}
	f.Type = typeStr
	return f
}

var modelsTemplate = `// Models for loading CSV data  - generated {{ .Now }}
package importcsv

import (
	"github.com/jinzhu/gorm"
	"strings"
	// "time"
)

// GORM Model factory
type ModelFactory struct {
	models []string
}

/*
Make a single instance of the model factory for use in importcsv
 */
func MakeModels() ModelFactory {
	factory := ModelFactory{}{{range $k, $v := .Models}}
    factory.models = append(factory.models, "{{$k}}"){{end}}
	return factory
}

{{range $k, $v := .Fields}}
type {{ $k }} struct {
   gorm.Model
   {{ range $f := $v }}{{$f.Name}} {{$f.Type}} {{$f.Tag}}
   {{end}}
}
{{end}}

func (f ModelFactory) New(name string) interface{} {
	name = strings.ToLower(name)
	switch name {
	{{range $k, $v := .Models}}
	case "{{$k}}":
		return &{{$v}}{}
	{{end}}
	}
	return nil
}
`