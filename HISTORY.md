# Changelog

## Unreleased

* `inspectcsv` skips rows with the wrong number of fields instead of failing the file, and
  reports their line numbers on stderr, matching how `importcsv` rejects them.
* Add `--json` to `inspectcsv` and `importcsv` for programs that drive the CLI. `inspectcsv` reports
  the generated models, skipped rows and warnings; `importcsv` reports per-file counts and each
  error, and still exits non-zero when any row is rejected. Each field lists the CSV `column` it came from, which
  is what `--key` takes.
* Add `inspectcsv --dedupe model` for data with no natural key. The model gets a `RowHash` field
  with a unique index, `importcsv` fills it with a SHA-256 of the row's values, and identical rows
  load once. It hashes the CSV text, so `1` and `1.0` are different rows.
* Add `inspectcsv --key model=column[+column]` to give a model a unique key. The key fields get a
  unique index, so `importcsv` skips rows whose key is already in the table and counts them as
  duplicates, instead of adding every row again on each import.
* Regenerate the committed `importcsv/models.go` from `static/fixtures`. It predated the current
  type inference (`Datecol string`, `Numbercol float32`), and a test now fails if it goes stale.
* Columns named after `gorm.Model` fields (`model`, `id`, `created_at`, `updated_at`,
  `deleted_at`) generate `ModelField`, `IdField` and so on, with a warning. Before, a `model`
  column was dropped silently, and an `id` column overwrote the primary key on import.
* `importcsv` now matches CSV headings to the Go names `inspectcsv` generates, so headings like
  `unit_price` or `1st col` import into `UnitPrice` and `Field1stCol`. Before, a generated model
  could not import its own CSV unless every heading was a single word.
* Print each command error once on stderr, prefixed with `Error:`. Cobra and `main` both printed
  it before.
* Add `gormcsv version` (and `--version`). It reports the build tag, or the latest tag plus
  commits since and short hash, stamped by `make build`, `make docker`, and release builds.

## 0.3.0 - Batch Loading and tidy up - 21 Sept 2026

* Add configurable GORM batch insertion with a bounded concurrent writer pool.
* Map records directly from validated CSV headers without relying on map order.
* Accept LF, CRLF, and classic Mac CR line endings in quoted or unquoted CSV.
* Return structured counts and errors for partial imports.
* Skip duplicate records portably without retrying failed batches row by row.
* Generate formatted models using deterministic, constant-memory type inference.
* Isolate database state in tests and verify imported values as well as counts.
* Add opt-in embedded PostgreSQL integration tests for concurrent batch writes.
* Modernise the CLI, container build, lint configuration, and CI checks.

## 0.2.0 - Upgrade to GORM v2 and Go 1.25 - 19 Sept 2026

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
