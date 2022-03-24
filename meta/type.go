package meta

import "time"

// Files embedded stuct to hang the FilesFetch method on
type Files struct {
}

// Meta embedded stuct to hang the convert method on
type Meta struct {
}

// FieldMeta struct for metadata about a Model
// fieldcols = position in CSV parsed array mapped to field name
// fieldtype = type of field used for converting from CSV string
type FieldMeta struct {
	Meta
	fieldcols map[string]int
	// fieldtypes map[string]reflect.Type
}

type field struct {
	Name string
	Type string
	Tag  string
}

// CSVMeta - main metadata struct used by inspectcsv and importcsv
type CSVMeta struct {
	Now time.Time
	Meta
	Files
	Models map[string]string
	Fields map[string][]field
}
