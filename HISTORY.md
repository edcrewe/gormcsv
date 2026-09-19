# Changelog

## 0.2.0 - Upgrade to GORM v2 and Go 1.24 - 19 Sept 2026

* Migrate to modern GORM v2 (`gorm.io/gorm`) and updated driver plugins.
  - github.com/jinzhu/gorm (up to v1.9.x) = GORM v1
  - gorm.io/gorm (from v1.20.0 up to today's v1.31.2) = GORM v2
* Update Go version requirement to 1.24.
* Update Dockerfile to use `golang:1.24-alpine`.
* Update Makefile `test` and `build` commands to use `--no-cache` and modern `go test ./...` syntax.
* Fix GORM v2 deprecations, including database connection handling, logmode removal, and `Count` pointer typing.

## 0.1.0 - Initial Release - 24 March 2022

* Inspect CSV files to dynamically generate GORM models (`inspectcsv`).
* Import CSV data directly into GORM-supported databases (`importcsv`).
* Automatic type inference for CSV fields (integers, floats, dates, booleans, etc).
* Automatic table creation via GORM AutoMigrate.
* Handle duplicate row skipping on import.
* Database support for MySQL, Postgres, SQLite, and MSSQL.
