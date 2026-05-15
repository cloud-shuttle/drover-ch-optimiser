package graph

import (
	"strings"
	"testing"
)

func TestParseExcludes(t *testing.T) {
	e, err := ParseExcludes([]string{"otel", "default.events", "peerdb*", "audit_test.tmp_*"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		key  TableKey
		want bool
	}{
		{key: TableKey{Database: "otel", Name: "t"}, want: true},
		{key: TableKey{Database: "default", Name: "events"}, want: true},
		{key: TableKey{Database: "peerdb", Name: "x"}, want: true},
		{key: TableKey{Database: "audit_test", Name: "tmp_1"}, want: true},
		{key: TableKey{Database: "audit_test", Name: "events"}, want: false},
		{key: TableKey{Database: "analytics", Name: "events"}, want: false},
	}
	for _, tt := range tests {
		if got := e.Match(tt.key); got != tt.want {
			t.Errorf("Match(%+v) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestExcludeSQLFilter(t *testing.T) {
	e, _ := ParseExcludes([]string{"otel", "default.events*"})
	clause := e.SQLFilter()
	if clause == "" {
		t.Fatal("expected SQL clause")
	}
	for _, sub := range []string{"database NOT IN", "otel", "default", "events"} {
		if !strings.Contains(clause, sub) {
			t.Errorf("clause missing %q: %s", sub, clause)
		}
	}
}

func TestPruneExcluded(t *testing.T) {
	e, _ := ParseExcludes([]string{"noise"})
	g := &Graph{
		Nodes: map[TableKey]*Table{
			{Database: "noise", Name: "t"}: {Key: TableKey{Database: "noise", Name: "t"}},
			{Database: "keep", Name: "t"}:    {Key: TableKey{Database: "keep", Name: "t"}},
		},
		Edges: []Edge{
			{From: TableKey{Database: "noise", Name: "t"}, To: TableKey{Database: "keep", Name: "t"}},
		},
	}
	g.pruneExcluded(e)
	if len(g.Nodes) != 1 {
		t.Fatalf("nodes = %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Fatalf("edges = %d", len(g.Edges))
	}
}
