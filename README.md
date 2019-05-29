# gormcsv

Ed Crewe - June 2019

gormcsv provides loading of CSV data files / fixtures via the golang ORM, GORM

By default the CSV file name is used to match the Model name and the header line to match the Model fields.
Or each of these can be supplied as an argument.

CSV files can be inspected for data types and used to generate a GORM models.go file using the inspectcsv command, prior to building and loading data to it with the importcsv command.

The purpose of gormcsv is to cater for data population / migration or data test fixtures usage.
Either for GORM based applications, or simply to (build a schema and) populate a database based on CSV files.

Data is loadable to any of the GORM supported databases: mysql, postgres, sqlite, mssql