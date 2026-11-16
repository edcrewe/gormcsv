package schema

import (
	"context"
	"errors"
	"fmt"
	"slices"

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

// Page is one page of rows. Each row lists its values in Columns order.
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
// own catalogue, and GORM quotes them, so neither reaches SQL as raw input.
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
	columnTypes, err := db.Migrator().ColumnTypes(query.Table)
	if err != nil {
		return Page{}, fmt.Errorf("read columns of %q: %w", query.Table, err)
	}
	page := Page{Table: query.Table, Rows: [][]any{}, Limit: query.Limit, Offset: query.Offset, Desc: query.Desc}
	for _, columnType := range columnTypes {
		page.Columns = append(page.Columns, columnType.Name())
		if query.Order == "" {
			if primary, _ := columnType.PrimaryKey(); primary {
				page.Order = columnType.Name()
			}
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

	if err := db.Table(query.Table).Count(&page.Total).Error; err != nil {
		return Page{}, fmt.Errorf("count rows in %q: %w", query.Table, err)
	}
	var records []map[string]any
	order := clause.OrderByColumn{Column: clause.Column{Name: page.Order}, Desc: query.Desc}
	if err := db.Table(query.Table).Order(order).Limit(query.Limit).Offset(query.Offset).Find(&records).Error; err != nil {
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

// jsonValue turns driver values that do not encode well, such as []byte text, into strings.
func jsonValue(value any) any {
	if bytes, ok := value.([]byte); ok {
		return string(bytes)
	}
	return value
}
