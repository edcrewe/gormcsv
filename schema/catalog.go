package schema

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GORM's Table and its migrators' ColumnTypes and GetIndexes treat a name containing a space or
// backtick as raw SQL, and split one containing a dot into schema and table. GetIndexes also
// leaves out indexes that back UNIQUE constraints on SQLite and PostgreSQL. So this file reads
// columns and indexes from each database's catalogue with the table name as a bound parameter,
// and quotes names itself where they must appear in SQL.

// catalogQuery reads one table's columns as (name, type, nullable, primary key) in table order,
// and its indexes as (index name, unique, column) rows, ordered by index then key position. A
// column is NULL for an expression in an index. Primary key indexes are left out: the primary
// key is shown on its columns.
type catalogQuery struct {
	columns string
	indexes string
}

var catalogQueries = map[string]catalogQuery{
	"sqlite": {
		columns: `SELECT name, type, "notnull" = 0, pk > 0 FROM pragma_table_info(?) ORDER BY cid`,
		indexes: `SELECT il.name, il."unique", ii.name
			FROM pragma_index_list(?) AS il, pragma_index_info(il.name) AS ii
			WHERE il.origin <> 'pk'
			ORDER BY il.name, ii.seqno`,
	},
	"postgres": {
		columns: `SELECT a.attname, format_type(a.atttypid, a.atttypmod), NOT a.attnotnull,
				EXISTS (SELECT 1 FROM pg_index i
					WHERE i.indrelid = c.oid AND i.indisprimary AND a.attnum = ANY (i.indkey))
			FROM pg_attribute a
			JOIN pg_class c ON c.oid = a.attrelid
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = current_schema() AND c.relname = ? AND a.attnum > 0 AND NOT a.attisdropped
			ORDER BY a.attnum`,
		indexes: `SELECT ic.relname, i.indisunique, a.attname
			FROM pg_index i
			JOIN pg_class t ON t.oid = i.indrelid
			JOIN pg_namespace n ON n.oid = t.relnamespace
			JOIN pg_class ic ON ic.oid = i.indexrelid
			CROSS JOIN LATERAL unnest(i.indkey::int2[]) WITH ORDINALITY AS k(attnum, position)
			LEFT JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
			WHERE n.nspname = current_schema() AND t.relname = ? AND NOT i.indisprimary
				AND k.position <= i.indnkeyatts
			ORDER BY ic.relname, k.position`,
	},
	"mysql": {
		columns: `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE = 'YES', COLUMN_KEY = 'PRI'
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION`,
		indexes: `SELECT INDEX_NAME, NON_UNIQUE = 0, COLUMN_NAME
			FROM information_schema.STATISTICS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME <> 'PRIMARY'
			ORDER BY INDEX_NAME, SEQ_IN_INDEX`,
	},
}

// expressionColumn stands in for an index part that is an expression rather than a column.
const expressionColumn = "(expression)"

func catalogFor(db *gorm.DB) (catalogQuery, error) {
	query, ok := catalogQueries[db.Name()]
	if !ok {
		return catalogQuery{}, fmt.Errorf("describing %s databases is not supported", db.Name())
	}
	return query, nil
}

func readColumns(db *gorm.DB, table string) ([]Column, error) {
	query, err := catalogFor(db)
	if err != nil {
		return nil, err
	}
	rows, err := db.Raw(query.columns, table).Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	columns := []Column{}
	for rows.Next() {
		var column Column
		if err := rows.Scan(&column.Name, &column.Type, &column.Nullable, &column.PrimaryKey); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func readIndexes(db *gorm.DB, table string) ([]Index, error) {
	query, err := catalogFor(db)
	if err != nil {
		return nil, err
	}
	rows, err := db.Raw(query.indexes, table).Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	indexes := []Index{}
	for rows.Next() {
		var (
			name   string
			unique bool
			column sql.NullString
		)
		if err := rows.Scan(&name, &unique, &column); err != nil {
			return nil, err
		}
		part := column.String
		if !column.Valid {
			part = expressionColumn
		}
		// Rows arrive grouped by index, so a new name starts a new index.
		if last := len(indexes) - 1; last >= 0 && indexes[last].Name == name {
			indexes[last].Columns = append(indexes[last].Columns, part)
			continue
		}
		indexes = append(indexes, Index{Name: name, Columns: []string{part}, Unique: unique})
	}
	return indexes, rows.Err()
}

// quoteIdentifier quotes name as one SQL identifier, whatever it contains.
func quoteIdentifier(db *gorm.DB, name string) string {
	quote := `"`
	if db.Name() == "mysql" {
		quote = "`"
	}
	return quote + strings.ReplaceAll(name, quote, quote+quote) + quote
}

// fromTable is db.Table(name) with name always quoted as a single identifier.
func fromTable(db *gorm.DB, name string) *gorm.DB {
	return db.Table("?", clause.Table{Name: quoteIdentifier(db, name), Raw: true})
}
