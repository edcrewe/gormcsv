// Field interface uses reflect to get a map of field names and types for matching to CSV column names
package importcsv

import (
	"fmt"
	"reflect"
	"strings"
)

// Struct for metadata about a Model
type FieldMeta struct {
	fieldcols map[string]int
	// fieldtype() map[string] reflect.Value.Type
}

/* Create index lookup for csv parsed line array fields
   Match lowercased name to field name and create index map
*/
func (m *FieldMeta) getmeta(model interface{}, csvfields string) {
	m.fieldcols = make(map[string]int)
	// model := Country{}
	csv := strings.Split(csvfields, ",")
	val := reflect.ValueOf(model).Elem()
	for i := 0; i < val.NumField(); i++ {
		// valueField := val.Field(i)
		typeField := val.Type().Field(i)
		field := strings.ToLower(typeField.Name)
		for j := range csv {
			if csv[j] == field {
				// Found!
				m.fieldcols[typeField.Name] = j
			}
		}
		// tag := typeField.Tag
	}
	fmt.Println("Using field mapping = ", m.fieldcols)
}
