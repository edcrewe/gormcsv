package meta

// FieldMeta struct for metadata about a Model
// fieldcols = position in CSV parsed array mapped to field name
// fieldtype = type of field used for converting from CSV string
type FieldMeta struct {
	fields []mappedField
}

type mappedField struct {
	column int
	name   string
}

type Field struct {
	Name string
	Type string
	Tag  string
}

// CSVMeta - main metadata struct used by inspectcsv and importcsv
type CSVMeta struct {
	Models   map[string]string
	Fields   map[string][]Field
	UsesTime bool
	// SkippedRows lists, per CSV path, the line numbers of records left out of type
	// inference because their field count did not match the header.
	SkippedRows map[string][]int
}
