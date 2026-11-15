# gormcsv

`gormcsv` inspects CSV files, generates typed GORM models, and imports CSV data
into PostgreSQL, MySQL, or SQLite in batches.

The project demonstrates a small, testable Go import pipeline:

- CSV headers are matched case-insensitively to exported model fields while
  preserving the source column order.
- Quoted, unquoted, and mixed CSV records support LF, CRLF, and classic Mac CR
  line endings.
- Type inference scans all records in constant memory and supports booleans,
  signed integers, floating-point values, common date formats, and strings.
- Imports use configurable GORM batches and a bounded database-writer pool.
- Unique-constraint conflicts are skipped portably with `ON CONFLICT DO
  NOTHING`; all other conversion, parsing, and database failures are returned.
- Core packages return structured results and errors. Only the CLI writes user
  output or chooses process exit status.

## Requirements

- Go 1.25 or newer
- A C compiler for the SQLite driver

## Build and test

```sh
go build ./...
go test -race ./...
go vet ./...
```

With golangci-lint v2.13.2 installed:

```sh
golangci-lint run --config=golangci-lint.yml
```

The same commands are available through `make build`, `make test`, and
`make lint`.

### PostgreSQL integration tests

Importer tests use isolated SQLite databases by default. To run the same tests
with four concurrent writers against a real embedded PostgreSQL server:

```sh
GORMCSV_TEST_POSTGRES=1 go test -race -count=1 ./importcsv
# or
make test-postgres
```

This uses
[`fergusstrange/embedded-postgres`](https://github.com/fergusstrange/embedded-postgres)
v1.34.0 and PostgreSQL 18.3. The first run downloads a platform-specific
PostgreSQL archive from Maven Central and caches it under
`~/.embedded-postgres-go`. Each package run starts one server on an available
local port, gives every test an isolated schema, and removes its temporary data
and process on completion. Leave `GORMCSV_TEST_POSTGRES` unset for fast,
offline SQLite tests.

## Generate models

The `inspectcsv` command reads one CSV file or every `.csv` file in a directory
and writes formatted Go model definitions:

```sh
go run . inspectcsv --files static/fixtures --output importcsv/models.go
```

The filename determines the model name and each header determines a field name.
For example, `order-item.csv` generates `OrderItem`, and `unit_price` generates
`UnitPrice`.

Columns named after a field that `gorm.Model` already provides (`model`, `id`,
`created_at`, `updated_at`, `deleted_at`) get a `Field` suffix, such as
`ModelField`, and the command warns about each one.

Generated models have no unique key, so importing the same CSV twice stores every
row twice. To skip rows that are already loaded, give the model a key from one or
more columns:

```sh
go run . inspectcsv --files static/fixtures --output importcsv/models.go --key country=code
```

`--key` names the model by its CSV filename and can be repeated. Join columns
with `+` for a composite key, such as `--key order-line=order_id+line`. The key
fields get a unique index (a composite one for several columns), and
`importcsv` counts rows whose key already exists as skipped duplicates. Generated
fields are never pointers, so an empty key value counts as a value, and all rows
with an empty key are duplicates of each other.

If the data has no natural key, `--dedupe model` skips rows identical to one
already loaded instead. The model gets a `RowHash` field with a unique index, and
`importcsv` fills it with a SHA-256 of the row's values. It compares the CSV text,
so `1` and `1.0` are different rows. Use either `--key` or `--dedupe` for a model,
not both.

Both cost an index update per row. In a single run on SQLite, loading 200,000 rows
took about 1.0s with no key, 1.1s with `--key` on a sequential code column, and
2.2s with `--dedupe`, whose random hashes spread across the index. Re-importing
the same file was faster than the first load in both cases, because every row was
skipped.

Rows whose field count differs from the header are left out of type inference,
matching `importcsv`, which rejects them. The command lists their line numbers
on stderr and still exits successfully. Other parse errors, such as unbalanced
quotes, fail the inspection.

Review generated models before using them for a long-lived schema. In
particular, decide whether empty values should use pointers or nullable database
types and add application-specific indexes or constraints.

## Import CSV data

```sh
go run . importcsv \
  --files static/fixtures/TestTypes.csv \
  --driver sqlite \
  --dsn gormcsv.db \
  --batch-size 1000 \
  --workers 4
```

The model must be registered by `importcsv.MakeModels`, normally by generating
`importcsv/models.go` before building the command. Tables are migrated before
their files are imported.

An import is best effort. Valid records are committed even when another record
is malformed. The command prints inserted, duplicate, and rejected counts and
returns a non-zero exit status if any record fails.

For PostgreSQL or MySQL, pass the corresponding driver and its GORM
connection string. For example:

```sh
go run . importcsv --files ./fixtures --driver postgres \
  --dsn 'host=localhost user=gormcsv password=secret dbname=gormcsv sslmode=disable' \
  --batch-size 1000 --workers 8
```

Each worker can hold one batch insert in flight. GORM uses Go's
`database/sql` connection pool underneath, and the CLI sets the pool's maximum
open connections to the requested worker count. SQLite is always restricted to
one writer. Applications using the `importcsv` package can inject an existing
`*gorm.DB`; worker count is also capped when that database has a lower
`SetMaxOpenConns` limit.

Concurrent databases use first-commit-wins semantics when duplicate keys occur
in different batches. Use one worker when strict source-order duplicate
resolution is required.

## JSON output

For programs that drive the CLI, `--json` prints a machine-readable report on stdout instead of
the text summaries. `inspectcsv --json` lists each generated model with its factory key and
fields, plus skipped rows and warnings per file. `importcsv --json` prints per-file counts, the
totals and every error as a separate string. It still exits non-zero when any row is rejected,
so check `errors` rather than the exit status to tell a partial import from a failure.

## Container

```sh
make docker
docker run --rm --volume "$PWD:/data" gormcsv:latest \
  importcsv --files /data/static/fixtures/TestTypes.csv \
  --driver sqlite --dsn /data/gormcsv.db
```

The runtime image uses an unprivileged user and `/data` as its working
directory.

## History

See [HISTORY.md](HISTORY.md).

```sh
gormcsv version    # or gormcsv --version
```
