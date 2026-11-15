package meta

// FieldMeta struct for metadata about a Model
// fieldcols = position in CSV parsed array mapped to field name
// fieldtype = type of field used for converting from CSV string
type FieldMeta struct {
	fields []mappedField
	// hashField is set when the model has a RowHash field the CSV does not fill.
	hashField string
}

type mappedField struct {
	column int
	name   string
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
}

// CSVMeta - main metadata struct used by inspectcsv and importcsv
type CSVMeta struct {
	// Keys is input, set before PopulateMeta: per model key (the lower-cased filename without
	// .csv), the CSV headings that together must be unique. Their fields get a unique index, so
	// importcsv skips rows whose key is already in the table.
	Keys map[string][]string
	// Dedupe is input, set before PopulateMeta: model keys whose identical rows should be
	// skipped. Those models get a RowHash field with a unique index, filled by importcsv.
	Dedupe map[string]bool

	Models   map[string]string
	Fields   map[string][]Field
	UsesTime bool
	// SkippedRows lists, per CSV path, the line numbers of records left out of type
	// inference because their field count did not match the header.
	SkippedRows map[string][]int
	// Warnings lists, per CSV path, other problems worth reporting, such as renamed columns.
	Warnings map[string][]string
}
