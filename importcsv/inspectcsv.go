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
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
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
	Meta
	Files
	Models map[string]string
	Fields map[string][]field
}


func (csvmeta *CSVMeta) PopulateMeta(path string) error {
	filesMap, err := csvmeta.FilesFetch(path)
	if err != nil {
		return fmt.Errorf("Failed to find CSV file(s) from %s, Due to %s", path, err)
	}
	fmt.Print(filesMap)
	names := map[string][]string{}
	for model, csvFile := range filesMap {
		reader := csv.NewReader(bufio.NewReader(csvFile))
		modelLower := strings.ToLower(model)
		csvmeta.Models[modelLower] = model
		var keys []string
		if reader != nil {
			for i := 1; 1 <= 5; i++ {
				record, error := reader.Read()
				for index, _ := range record {
					if i == 1 {
						names[record[index]] = []string{}
						keys = record
					} else {
						names[keys[index]] = append(names[keys[index]], record[index])
					}
				}
				if error == io.EOF {
					break
				} else if error != nil {
					return error
				}
			}
		}
		for key, values := range names {
			field := csvmeta.GetField(key, values)
			csvmeta.Fields[modelLower] = append(csvmeta.Fields[modelLower], field)
		}
	}
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