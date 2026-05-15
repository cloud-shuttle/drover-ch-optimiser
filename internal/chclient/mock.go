package chclient

import (
	"context"
	"reflect"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// MockQuery routes a query substring to tabular results.
type MockQuery struct {
	Contains string
	Columns  []string
	Rows     [][]any
	Err      error
}

// MockQuerier implements Querier for unit tests.
type MockQuerier struct {
	Queries []MockQuery
	Closed  bool
}

func (m *MockQuerier) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	for _, q := range m.Queries {
		if strings.Contains(query, q.Contains) {
			if q.Err != nil {
				return nil, q.Err
			}
			return newSliceRows(q.Columns, q.Rows), nil
		}
	}
	return newSliceRows(nil, nil), nil
}

func (m *MockQuerier) Close() error {
	m.Closed = true
	return nil
}

// sliceRows is a minimal driver.Rows over in-memory data.
type sliceRows struct {
	columns []string
	rows    [][]any
	idx     int
}

func newSliceRows(columns []string, rows [][]any) *sliceRows {
	if columns == nil {
		columns = []string{}
	}
	return &sliceRows{columns: columns, rows: rows}
}

func (r *sliceRows) Next() bool {
	if r.idx >= len(r.rows) {
		return false
	}
	r.idx++
	return true
}

func (r *sliceRows) Scan(dest ...any) error {
	row := r.rows[r.idx-1]
	for i, d := range dest {
		if i >= len(row) {
			continue
		}
		assignScan(d, row[i])
	}
	return nil
}

func (r *sliceRows) Close() error { return nil }

func (r *sliceRows) Err() error { return nil }

func (r *sliceRows) ScanStruct(dest any) error { return nil }

func (r *sliceRows) Totals(dest ...any) error { return nil }

func (r *sliceRows) Columns() []string { return r.columns }

func (r *sliceRows) ColumnTypes() []driver.ColumnType { return nil }

func assignScan(dest, src any) {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return
	}
	ev := dv.Elem()
	sv := reflect.ValueOf(src)
	if !sv.IsValid() {
		return
	}
	if sv.Type().AssignableTo(ev.Type()) {
		ev.Set(sv)
		return
	}
	if sv.Type().ConvertibleTo(ev.Type()) {
		ev.Set(sv.Convert(ev.Type()))
	}
}
