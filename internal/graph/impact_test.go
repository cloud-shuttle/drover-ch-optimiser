package graph

import "testing"

func TestAnalyzeImpact_MVUpstream(t *testing.T) {
	src := TableKey{Database: "audit_test", Name: "events_bad"}
	mv := TableKey{Database: "audit_test", Name: "events_bad_mv"}
	g := &Graph{
		Nodes: map[TableKey]*Table{
			src: {Key: src, HotQueries: 5, FilterColumns: []string{"day", "status"}},
			mv:  {Key: mv},
		},
		Edges: []Edge{{From: src, To: mv, Type: EdgeDependsOn}},
	}
	ia := AnalyzeImpact(g, "audit_test.events_bad_mv")
	if len(ia.Upstream) != 1 || ia.Upstream[0] != src {
		t.Fatalf("upstream = %+v", ia.Upstream)
	}
	if ia.HotQueries != 0 {
		t.Fatalf("focus node hot queries on mv, got %d", ia.HotQueries)
	}
}
