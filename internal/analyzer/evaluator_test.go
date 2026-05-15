package analyzer

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

func TestRun_HealthScoreAverage(t *testing.T) {
	g := &graph.Graph{
		Nodes: map[graph.TableKey]*graph.Table{
			{Database: "db", Name: "good"}: {
				Key:         graph.TableKey{Database: "db", Name: "good"},
				Engine:      "MergeTree",
				SortingKey:  "day, user_id",
				HealthScore: 100,
			},
			{Database: "db", Name: "bad"}: {
				Key:         graph.TableKey{Database: "db", Name: "bad"},
				Engine:      "MergeTree",
				SortingKey:  "user_uuid, day",
				Columns:     []graph.Column{{Name: "created_at", Type: "String"}},
				HealthScore: 100,
			},
		},
	}
	issues, summary := Run(g)
	if summary.HealthScore <= 0 || summary.HealthScore >= 100 {
		t.Fatalf("expected blended health between 0 and 100, got %.1f", summary.HealthScore)
	}
	if len(issues[Critical]) == 0 && len(issues[High]) == 0 {
		t.Fatal("expected at least one issue on bad table")
	}
}

func TestRun_SkipsInnerTables(t *testing.T) {
	g := &graph.Graph{
		Nodes: map[graph.TableKey]*graph.Table{
			{Database: "default", Name: ".inner_id.abc"}: {
				Key:        graph.TableKey{Database: "default", Name: ".inner_id.abc"},
				Engine:     "SharedReplacingMergeTree",
				SortingKey: "incident_id",
			},
		},
	}
	issues, summary := Run(g)
	if summary.TotalIssues != 0 {
		t.Fatalf("inner tables should be skipped, got %d issues", summary.TotalIssues)
	}
	if len(issues[Critical]) > 0 {
		t.Fatal("inner table should not produce critical issues")
	}
}

func TestLoadRules(t *testing.T) {
	if err := loadRules(); err != nil {
		t.Fatal(err)
	}
	if len(rules) < 9 {
		t.Fatalf("expected at least 9 rules, got %d", len(rules))
	}
}
