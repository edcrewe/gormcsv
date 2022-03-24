// Field interface uses reflect to get a map of field names and types for matching to CSV column names
package meta

import (
	"fmt"
	"reflect"
	"strings"
)

// SetMeta create index lookup for csv parsed line array fields
// Match lowercased name to field name and create index map
func (meta *FieldMeta) SetMeta(model interface{}, csvfields string) {
	meta.fieldcols = make(map[string]int)
	if csvfields == "" {
		fmt.Println("Mapping failed because no csvfields were supplied")
		return
	}
	// meta.fieldtypes = make(map[string]reflect.Type)
	csv := strings.Split(csvfields, ",")
	structValue := reflect.ValueOf(model).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		typeField := structValue.Type().Field(i)
		// Need uppercase first letter for public attribute
		field := strings.Title(strings.ToLower(typeField.Name))
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

// getMap map record values
func (meta *FieldMeta) getMap(record []string) map[string]string {
	mData := make(map[string]string)
	for name, index := range meta.fieldcols {
		mData[name] = record[index]
	}
	return mData
}

// RecordToModel populate Struct fields from maprecord
func (meta *FieldMeta) RecordToModel(model interface{}, record []string) (interface{}, error) {
	mData := meta.getMap(record)
	structValue := reflect.ValueOf(model).Elem()
	for name, value := range mData {
		structFieldValue := structValue.FieldByName(strings.ToTitle(strings.ToLower(name)))
		if structFieldValue.IsValid() {
			fieldType := structFieldValue.Type()
			converted, error := meta.Convert(value, fieldType.Name())
			if error != nil {
				return nil, error
			}
			val := reflect.ValueOf(converted)
			structFieldValue.Set(val.Convert(fieldType))
		}
	}
	return model, nil
}
