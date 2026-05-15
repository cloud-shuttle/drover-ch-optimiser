//go:build integration

package analyzer_test

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/testutil"
)

func TestRun_Integration_FixtureWarehouse(t *testing.T) {
	native := testutil.OpenConn(t)
	defer native.Close()

	conn := chclient.NewFromNative(native)
	g, err := graph.BuildGraph(conn, graph.Options{Database: testutil.TestDatabase, DaysBack: 7})
	if err != nil {
		t.Fatal(err)
	}

	issues, summary := analyzer.Run(g)
	if summary.TotalIssues == 0 {
		t.Fatal("expected rule hits on audit_test fixtures")
	}
	if summary.Critical == 0 && summary.High == 0 {
		t.Fatalf("expected critical or high issues: %+v", summary)
	}

	// events_bad should trigger string + PK rules
	var hits int
	for _, list := range issues {
		for _, iss := range list {
			if iss.Table == "audit_test.events_bad" {
				hits++
			}
		}
	}
	if hits < 2 {
		t.Errorf("expected multiple issues on events_bad, got %d", hits)
	}
}
