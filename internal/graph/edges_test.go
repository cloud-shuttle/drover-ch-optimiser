package graph

import "testing"

func TestBuildDependencyEdges_MV(t *testing.T) {
	mvKey := TableKey{Database: "audit_test", Name: "events_bad_mv"}
	srcKey := TableKey{Database: "audit_test", Name: "events_bad"}
	g := &Graph{
		Nodes: map[TableKey]*Table{
			srcKey: {Key: srcKey, Engine: "MergeTree"},
			mvKey: {
				Key:             mvKey,
				Engine:          "MaterializedView",
				DependenciesDB:  []string{"audit_test"},
				DependenciesTbl: []string{"events_bad"},
				CreateTableQuery: "CREATE MATERIALIZED VIEW events_bad_mv TO events_daily AS SELECT 1",
			},
		},
	}
	g.buildDependencyEdges()
	if len(g.Edges) == 0 {
		t.Fatal("expected MV dependency edge")
	}
	found := false
	for _, e := range g.Edges {
		if e.From == srcKey && e.To == mvKey && e.Type == EdgeDependsOn {
			found = true
		}
	}
	if !found {
		t.Fatalf("edges: %+v", g.Edges)
	}
}

func TestSelectMermaidNodes_PrioritizesUnhealthy(t *testing.T) {
	g := &Graph{
		Nodes: map[TableKey]*Table{
			{Database: "db", Name: "healthy"}: {Key: TableKey{"db", "healthy"}, Engine: "MergeTree", HealthScore: 100},
			{Database: "db", Name: "sick"}:   {Key: TableKey{"db", "sick"}, Engine: "MergeTree", HealthScore: 40, HotQueries: 2},
		},
	}
	include := g.selectMermaidNodes(10)
	if !include[TableKey{"db", "sick"}] {
		t.Error("sick table should be included")
	}
}
