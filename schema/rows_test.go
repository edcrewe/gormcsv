package schema

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
)

type Reading struct {
	Sensor string
	Value  int
}

func seed(t *testing.T) (*RowsQuery, func(RowsQuery) (Page, error)) {
	t.Helper()
	db := openDatabase(t)
	countries := make([]Country, 0, 5)
	for i, name := range []string{"Chad", "Albania", "Egypt", "Benin", "Denmark"} {
		countries = append(countries, Country{Code: fmt.Sprintf("C%d", i), Name: name})
	}
	if err := db.Create(&countries).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Migrator().DropTable(&Reading{}) })
	if err := db.Migrator().DropTable(&Reading{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Reading{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&[]Reading{{"b", 2}, {"a", 1}}).Error; err != nil {
		t.Fatal(err)
	}
	query := &RowsQuery{Table: "countries", Limit: 2}
	return query, func(q RowsQuery) (Page, error) { return Rows(context.Background(), db, q) }
}

func column(page Page, name string) []any {
	index := -1
	for i, c := range page.Columns {
		if c == name {
			index = i
		}
	}
	values := make([]any, len(page.Rows))
	for i, row := range page.Rows {
		values[i] = row[index]
	}
	return values
}

func TestRowsPagesInPrimaryKeyOrderByDefault(t *testing.T) {
	query, rows := seed(t)

	first, err := rows(*query)
	if err != nil {
		t.Fatal(err)
	}
	query.Offset = 4
	last, err := rows(*query)
	if err != nil {
		t.Fatal(err)
	}

	if first.Total != 5 || first.Order != "id" || first.Desc {
		t.Errorf("first page = total %d, order %q, desc %v", first.Total, first.Order, first.Desc)
	}
	if got := column(first, "name"); !reflect.DeepEqual(got, []any{"Chad", "Albania"}) {
		t.Errorf("first page names = %v", got)
	}
	if got := column(last, "name"); !reflect.DeepEqual(got, []any{"Denmark"}) {
		t.Errorf("last page names = %v", got)
	}
	if first.Columns[0] != "id" || len(first.Rows[0]) != len(first.Columns) {
		t.Errorf("rows should list values in column order: %v %v", first.Columns, first.Rows[0])
	}
}

func TestRowsSortsByAColumnEitherWay(t *testing.T) {
	query, rows := seed(t)
	query.Order, query.Limit = "name", 5

	ascending, err := rows(*query)
	if err != nil {
		t.Fatal(err)
	}
	query.Desc = true
	descending, err := rows(*query)
	if err != nil {
		t.Fatal(err)
	}

	want := []any{"Albania", "Benin", "Chad", "Denmark", "Egypt"}
	if got := column(ascending, "name"); !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v", got)
	}
	slices.Reverse(want)
	if got := column(descending, "name"); !reflect.DeepEqual(got, want) {
		t.Errorf("descending = %v", got)
	}
}

func TestRowsOrdersATableWithoutPrimaryKeyByItsFirstColumn(t *testing.T) {
	query, rows := seed(t)
	query.Table = "readings"

	page, err := rows(*query)
	if err != nil {
		t.Fatal(err)
	}

	if page.Order != "sensor" || !reflect.DeepEqual(column(page, "sensor"), []any{"a", "b"}) {
		t.Errorf("page = order %q, sensors %v", page.Order, column(page, "sensor"))
	}
}

func TestRowsPastTheEndIsEmptyNotNull(t *testing.T) {
	query, rows := seed(t)
	query.Offset = 50

	page, err := rows(*query)
	if err != nil {
		t.Fatal(err)
	}
	if page.Rows == nil || len(page.Rows) != 0 || page.Total != 5 {
		t.Errorf("page = %+v", page)
	}
}

func TestRowsRejectsUnknownNamesAndBadPaging(t *testing.T) {
	query, rows := seed(t)
	tests := map[string]struct {
		query    RowsQuery
		notFound bool
	}{
		"unknown table":          {RowsQuery{Table: "missing", Limit: 1}, true},
		"SQL in the table name":  {RowsQuery{Table: "countries; DROP TABLE countries", Limit: 1}, true},
		"unknown column":         {RowsQuery{Table: query.Table, Limit: 1, Order: "nope"}, true},
		"SQL in the column name": {RowsQuery{Table: query.Table, Limit: 1, Order: "name; DROP TABLE x"}, true},
		"zero limit":             {RowsQuery{Table: query.Table, Limit: 0}, false},
		"limit over the maximum": {RowsQuery{Table: query.Table, Limit: MaxLimit + 1}, false},
		"negative offset":        {RowsQuery{Table: query.Table, Limit: 1, Offset: -1}, false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := rows(test.query)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := errors.Is(err, ErrNotFound); got != test.notFound {
				t.Errorf("errors.Is(ErrNotFound) = %v for %v", got, err)
			}
		})
	}
	// The table survived.
	if page, err := rows(*query); err != nil || page.Total != 5 {
		t.Errorf("countries after the bad queries: %+v, %v", page, err)
	}
}

func TestJSONValueTurnsBytesIntoText(t *testing.T) {
	if got := jsonValue([]byte("abc")); got != "abc" {
		t.Errorf("jsonValue = %#v", got)
	}
	if got := jsonValue(int64(3)); got != int64(3) {
		t.Errorf("jsonValue = %#v", got)
	}
}
