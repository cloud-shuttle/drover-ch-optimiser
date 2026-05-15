package graph

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
)

func TestBuildGraph_MockQuerier(t *testing.T) {
	mock := &chclient.MockQuerier{
		Queries: []chclient.MockQuery{
			{
				Contains: "system.tables",
				Rows: [][]any{
					{
						"audit_test", "events", "MergeTree", "", "CREATE TABLE",
						"toYYYYMM(day)", "day, id", "day, id",
						[]string{}, []string{},
						uint64(100), uint64(1024),
					},
				},
			},
			{
				Contains: "system.columns",
				Rows: [][]any{
					{"audit_test", "events", "id", "UInt64", "", ""},
				},
			},
			{
				Contains: "system.parts",
				Rows:     [][]any{{"audit_test", "events", uint64(5)}},
			},
			{
				Contains: "system.mutations",
				Rows:     [][]any{{"audit_test", "events", uint64(2)}},
			},
		},
	}

	g, err := BuildGraph(mock, Options{DaysBack: 7})
	if err != nil {
		t.Fatal(err)
	}
	key := TableKey{Database: "audit_test", Name: "events"}
	node, ok := g.Nodes[key]
	if !ok {
		t.Fatal("missing events table")
	}
	if node.ActiveParts != 5 {
		t.Errorf("ActiveParts = %d, want 5", node.ActiveParts)
	}
	if node.IncompleteMutations != 2 {
		t.Errorf("IncompleteMutations = %d, want 2", node.IncompleteMutations)
	}
	if len(node.Columns) != 1 || node.Columns[0].Name != "id" {
		t.Errorf("columns = %+v", node.Columns)
	}
}
