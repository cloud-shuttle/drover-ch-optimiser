//go:build integration

package graph_test

import (
	"context"
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/testutil"
)

func TestBuildGraph_Integration(t *testing.T) {
	native := testutil.OpenConn(t)
	defer native.Close()

	ctx := context.Background()
	// Seed query_log with filter patterns for events_bad.
	for _, q := range []string{
		"SELECT count() FROM audit_test.events_bad WHERE day >= '2024-01-01' AND status = 'ok'",
		"SELECT count() FROM audit_test.events_bad PREWHERE status = 'ok' WHERE user_uuid != '00000000-0000-0000-0000-000000000000'",
	} {
		if err := native.Exec(ctx, q); err != nil {
			t.Fatalf("warm query_log: %v", err)
		}
	}

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

	bad := g.Nodes[graph.TableKey{Database: testutil.TestDatabase, Name: "events_bad"}]
	if bad != nil && len(bad.FilterColumns) == 0 {
		t.Log("query_log filter columns empty (query_log may be disabled); skipping filter assertion")
	} else if bad != nil {
		hasDay := false
		for _, c := range bad.FilterColumns {
			if c == "day" || c == "status" {
				hasDay = true
			}
		}
		if !hasDay {
			t.Errorf("FilterColumns = %v, want day/status from warmed queries", bad.FilterColumns)
		}
	}

	if graph.ShouldAnalyze(graph.TableKey{Database: "audit_test", Name: ".inner_id.x"}, "MergeTree") {
		t.Error("inner_id should not be analyzed")
	}
}
