// Package schema describes the tables in a database, for tools that browse what importcsv built.
package schema

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Column is one column as the database reports it.
type Column struct {
	Name string `json:"name"`
	// Type is the database's own type, such as VARCHAR(64) or bigint.
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primaryKey"`
}

// Index is an index on a table, such as the unique index --key or --dedupe adds.
type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// Table is one table with its row count, columns in database order, and indexes by name.
type Table struct {
	Name    string   `json:"name"`
	Rows    int64    `json:"rows"`
	Columns []Column `json:"columns"`
	Indexes []Index  `json:"indexes"`
}

// Describe lists every table in db's current database or schema, sorted by name.
func Describe(ctx context.Context, db *gorm.DB) ([]Table, error) {
	db = db.WithContext(ctx)
	migrator := db.Migrator()
	names, err := migrator.GetTables()
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	sort.Strings(names)
	tables := make([]Table, 0, len(names))
	for _, name := range names {
		// SQLite's own bookkeeping, such as sqlite_sequence, is not part of the data.
		if db.Name() == "sqlite" && strings.HasPrefix(name, "sqlite_") {
			continue
		}
		table, err := describeTable(db, name)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func describeTable(db *gorm.DB, name string) (Table, error) {
	table := Table{Name: name, Columns: []Column{}, Indexes: []Index{}}
	// The name comes from the database's own catalogue, and GORM quotes it.
	if err := db.Table(name).Count(&table.Rows).Error; err != nil {
		return table, fmt.Errorf("count rows in %q: %w", name, err)
	}
	columnTypes, err := db.Migrator().ColumnTypes(name)
	if err != nil {
		return table, fmt.Errorf("read columns of %q: %w", name, err)
	}
	for _, columnType := range columnTypes {
		column := Column{Name: columnType.Name(), Type: columnType.DatabaseTypeName()}
		if full, ok := columnType.ColumnType(); ok && full != "" {
			column.Type = full
		}
		column.Nullable, _ = columnType.Nullable()
		column.PrimaryKey, _ = columnType.PrimaryKey()
		table.Columns = append(table.Columns, column)
	}
	// The SQLite migrator logs its index queries in debug mode; a discarding logger stays silent.
	quiet := db.Session(&gorm.Session{Logger: logger.Discard})
	indexes, err := quiet.Migrator().GetIndexes(name)
	if err != nil {
		return table, fmt.Errorf("read indexes of %q: %w", name, err)
	}
	for _, index := range indexes {
		unique, _ := index.Unique()
		table.Indexes = append(table.Indexes, Index{Name: index.Name(), Columns: index.Columns(), Unique: unique})
	}
	sort.Slice(table.Indexes, func(i, j int) bool { return table.Indexes[i].Name < table.Indexes[j].Name })
	return table, nil
}
