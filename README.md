# gormcsv

Ed Crewe - March 2022

gormcsv provides loading of CSV data files / fixtures via the golang ORM, GORM

By default the CSV file name is used to match the Model name and the header line to match the Model fields.
Or each of these can be supplied as an argument.

CSV files can be inspected for data types and used to generate a GORM models.go file using the inspectcsv command, prior to building and loading data to it with the importcsv command.

The purpose of gormcsv is to cater for data population / migration or data test fixtures usage.
Either for GORM based applications, or simply to (build a schema and) populate a database based on CSV files.

Data is loadable to any of the GORM supported databases: mysql, postgres, sqlite, mssql

Model generation is a separate step that can be skipped for importing into existing Tables.

# Usage:
  gormcsv [command]

Available Commands:
  help        Help about any command
  importcsv   Populates one or more tables from CSV file(s)
  inspectcsv  Create models from CSV files

Flags:
  -f, --files string   CSV file or folder of files
  -h, --help           help for gormcsv

Use "gormcsv [command] --help" for more information about a command.

# Tests

Unit:
  cd tests & go test -v --tags=unit
  
Integration: 
  cd tests &  go test -v --tags=i
