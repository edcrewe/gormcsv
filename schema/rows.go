package schema

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MaxLimit caps how many rows one page can ask for.
const MaxLimit = 1000

// ErrNotFound is returned for a table or column the database does not have.
var ErrNotFound = errors.New("not found")

// RowsQuery selects one page of a table.
type RowsQuery struct {
	Table  string
	Limit  int
	Offset int
	// Order is a column to sort by; empty sorts by the primary key, or the first column.
	Order string
	Desc  bool
}

// Page is one page of rows. Each row lists its values in Columns order. Integers too large for a
// JavaScript number to hold exactly are strings of digits.
type Page struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
	// Total counts every row in the table, for paging.
	Total  int64  `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Order  string `json:"order"`
	Desc   bool   `json:"desc"`
}

// Rows reads one page of a table. The table and order column are checked against the database's
// own catalogue and quoted as identifiers, so neither reaches SQL as raw input.
func Rows(ctx context.Context, db *gorm.DB, query RowsQuery) (Page, error) {
	if query.Limit < 1 || query.Limit > MaxLimit {
		return Page{}, fmt.Errorf("limit must be from 1 to %d", MaxLimit)
	}
	if query.Offset < 0 {
		return Page{}, fmt.Errorf("offset must not be negative")
	}
	db = db.WithContext(ctx)
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return Page{}, fmt.Errorf("list tables: %w", err)
	}
	if !slices.Contains(tables, query.Table) {
		return Page{}, fmt.Errorf("table %q: %w", query.Table, ErrNotFound)
	}
	columns, err := readColumns(db, query.Table)
	if err != nil {
		return Page{}, fmt.Errorf("read columns of %q: %w", query.Table, err)
	}
	page := Page{
		Table: query.Table, Columns: []string{}, Rows: [][]any{},
		Limit: query.Limit, Offset: query.Offset, Desc: query.Desc,
	}
	for _, column := range columns {
		page.Columns = append(page.Columns, column.Name)
		if query.Order == "" && column.PrimaryKey && page.Order == "" {
			page.Order = column.Name
		}
	}
	switch {
	case query.Order != "" && !slices.Contains(page.Columns, query.Order):
		return Page{}, fmt.Errorf("column %q in %q: %w", query.Order, query.Table, ErrNotFound)
	case query.Order != "":
		page.Order = query.Order
	case page.Order == "" && len(page.Columns) > 0:
		// No primary key: still sort, so pages do not overlap.
		page.Order = page.Columns[0]
	}

	if err := fromTable(db, query.Table).Count(&page.Total).Error; err != nil {
		return Page{}, fmt.Errorf("count rows in %q: %w", query.Table, err)
	}
	var records []map[string]any
	order := clause.OrderByColumn{
		Column: clause.Column{Name: quoteIdentifier(db, page.Order), Raw: true},
		Desc:   query.Desc,
	}
	err = fromTable(db, query.Table).Order(order).Limit(query.Limit).Offset(query.Offset).Find(&records).Error
	if err != nil {
		return Page{}, fmt.Errorf("read rows of %q: %w", query.Table, err)
	}
	for _, record := range records {
		row := make([]any, len(page.Columns))
		for i, column := range page.Columns {
			row[i] = jsonValue(record[column])
		}
		page.Rows = append(page.Rows, row)
	}
	return page, nil
}

// maxSafeInteger is the largest integer a JSON reader using float64 numbers, such as JavaScript,
// holds exactly (2^53 - 1).
const maxSafeInteger = 1<<53 - 1

// jsonValue turns driver values that do not encode well into strings: []byte text, and integers
// beyond ±maxSafeInteger, which a JavaScript reader would silently round.
func jsonValue(value any) any {
	switch v := value.(type) {
	case []byte:
		return string(v)
	case int64:
		if v > maxSafeInteger || v < -maxSafeInteger {
			return strconv.FormatInt(v, 10)
		}
	case int:
		return jsonValue(int64(v))
	case uint64:
		if v > maxSafeInteger {
			return strconv.FormatUint(v, 10)
		}
	case uint:
		return jsonValue(uint64(v))
	}
	return value
}
