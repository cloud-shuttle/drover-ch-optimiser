package chclient

import (
	"context"
	"testing"
)

func TestMockQuerier_QueryAndClose(t *testing.T) {
	m := &MockQuerier{
		Queries: []MockQuery{
			{
				Contains: "system.tables",
				Columns:  []string{"database", "name"},
				Rows:     [][]any{{"db", "events"}},
			},
		},
	}
	rows, err := m.Query(context.Background(), "SELECT 1 FROM system.tables")
	if err != nil {
		t.Fatal(err)
	}
	if !rows.Next() {
		t.Fatal("expected row")
	}
	var db, name string
	if err := rows.Scan(&db, &name); err != nil {
		t.Fatal(err)
	}
	if db != "db" || name != "events" {
		t.Fatalf("got %q.%q", db, name)
	}
	if err := m.Close(); err != nil || !m.Closed {
		t.Fatal("close not recorded")
	}
}
