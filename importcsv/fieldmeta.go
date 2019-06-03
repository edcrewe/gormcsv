// Field interface uses reflect to get a map of field names and types for matching to CSV column names
package importcsv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/* Struct for metadata about a Model
   fieldcols = position in CSV parsed array mapped to field name
   fieldtype = type of field used for converting from CSV string
 */
type FieldMeta struct {
	fieldcols map[string]int
	// fieldtypes map[string]reflect.Type
}

/* Create index lookup for csv parsed line array fields
   Match lowercased name to field name and create index map
*/
func (meta *FieldMeta) setmeta(model interface{}, csvfields string) {
	meta.fieldcols = make(map[string]int)
	// meta.fieldtypes = make(map[string]reflect.Type)
	csv := strings.Split(csvfields, ",")
	structValue := reflect.ValueOf(model).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		typeField := structValue.Type().Field(i)
		field := strings.ToLower(typeField.Name)
		for j := range csv {
			if csv[j] == field {
				meta.fieldcols[typeField.Name] = j
				// Dont currently use field type or tag metadata
				// meta.fieldtypes[typeField.Name] = typeField.Type
				// tag := typeField.Tag
			}
		}
	}
	fmt.Println("Using field mapping = ", meta.fieldcols)
}

/*
Map record values
 */
func (meta *FieldMeta) getMap(record []string) map[string]string {
	mData := make(map[string]string)
	for name, index := range meta.fieldcols {
		mData[name] = record[index]
	}
	return mData
}

/*
Type converter for CSV string fields to correct Struct type
 */
func (meta *FieldMeta) convert(value string, to string) (interface{}, error) {
	switch to {
	case "float64":
		return strconv.ParseFloat(value, 64)
	case "int64":
		return strconv.Atoi(value)
	}
	return value, nil
}

/*
Populate Struct fields from maprecord
*/
func (meta *FieldMeta) RecordToModel(model interface{}, record []string) (interface{}, error) {
	mData := meta.getMap(record)
	structValue := reflect.ValueOf(model).Elem()
	for name, value := range mData {
		structFieldValue := structValue.FieldByName(name)
		fieldType := structFieldValue.Type()
		converted, error := meta.convert(value,  fieldType.Name())
		if error != nil {
			return nil, error
		}
		val := reflect.ValueOf(converted)
		structFieldValue.Set(val.Convert(fieldType))
	}
	return model, nil
}
