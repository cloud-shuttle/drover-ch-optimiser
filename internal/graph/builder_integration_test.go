//go:build integration

package graph_test

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/testutil"
)

func TestBuildGraph_Integration(t *testing.T) {
	native := testutil.OpenConn(t)
	defer native.Close()

	conn := chclient.NewFromNative(native)
	g, err := graph.BuildGraph(conn, graph.Options{Database: testutil.TestDatabase, DaysBack: 7})
	if err != nil {
		t.Fatal(err)
	}

	if len(g.Nodes) < 2 {
		t.Fatalf("expected fixture tables, got %d nodes", len(g.Nodes))
	}

	var foundBad bool
	for key := range g.Nodes {
		if key.Name == "events_bad" {
			foundBad = true
			break
		}
	}
	if !foundBad {
		t.Fatal("expected audit_test.events_bad in graph")
	}

	if graph.ShouldAnalyze(graph.TableKey{Database: "audit_test", Name: ".inner_id.x"}, "MergeTree") {
		t.Error("inner_id should not be analyzed")
	}
}
