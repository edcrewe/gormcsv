/* Estimate the field types from a CSV file and create a model for it
Add method to try convert and narrow down to type then write meta
Need csv read method from ModelCSV.ImportCSV too so maybe refactor OOP for that
*/
package importcsv

type field struct {
	Name string
	Type string
	Tag string
}

type CSVMeta struct {
	Meta
	Models map[string]string
	Fields map[string][]field
}

