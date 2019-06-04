// Field interface uses reflect to get a map of field names and types for matching to CSV column names
package importcsv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
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
Type converter for CSV string fields to correct basic type or time
 */
func (meta *FieldMeta) convert(value string, to string) (interface{}, error) {
	if to == "string" {
		return value, nil
	}
	bits := 0
	var err error
	prefixes := []string{"float", "uint", "int"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(to, prefix){
			bitstr := to[len(prefix):]
			if bitstr != "" {
				bits, err = strconv.Atoi(bitstr)
			}
			if err != nil {
				return nil, err
			}
			to = prefix
		}
	}
	switch to {
	case "float":
		return strconv.ParseFloat(value, bits)
	case "uint":
		return strconv.ParseUint(value, 10, bits)
	case "int":
		return strconv.ParseInt(value, 10, bits)
	case "bool":
		return strconv.ParseBool(value)
	case "date":
		layouts := []string{"2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05", "28/02/2003"}
		var date time.Time
		for _, layout := range layouts {
			date, err = time.Parse(layout, value)
			if err != nil {
				return nil, fmt.Errorf("Could not convert time %s to %s, unknown date fmt", value, to)
			}
			return date, nil
		}
	}
	return nil, fmt.Errorf("Could not convert %s to %s", value, to)
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
