package meta

import (
	"bufio"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var mapLength = map[int]string{2: "8", 4: "16", 9: "32", 20: "64"}
var keysLength = []int{2, 4, 9, 20}
var typeStrings = []string{"bool", "date", "string"}
var reNumber = regexp.MustCompile(`^[-+]?\d*\.?\d*$`)

// PopulateMeta read a sample set of data from each CSV file and determine the fields and field types
// Load that metadata to csvmeta struct Models and Fields
func (csvmeta *CSVMeta) PopulateMeta(path string) error {
	csvmeta.Now = time.Now()
	filesMap, err := csvmeta.FilesFetch(path)
	csvmeta.Models = map[string]string{}
	csvmeta.Fields = map[string][]field{}
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
						field := strings.Title(strings.ToLower(keys[index]))
						if field != "" && field != "Model" {
							names[field] = []string{}
						}
					} else {
						if len(keys) > index {
							field := strings.Title(strings.ToLower(keys[index]))
							if field != "" && field != "Model" {
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
	return nil
}

// GetField take a name and list of values from CSV then test the values to work out the type
func (csvmeta *CSVMeta) GetField(name string, valueStrings []string) field {
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
		if typeStr == "float8" || typeStr == "float16" {
			typeStr = "float32"
		}
	}
	f.Type = typeStr
	return f
}
