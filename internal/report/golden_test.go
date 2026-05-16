package report

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/golden files")

func TestHTMLGolden(t *testing.T) {
	g := goldenGraph()
	issues, summary := analyzer.Run(g)
	fixed := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	rpt := NewAt(g, issues, summary, 30, fixed)

	got, err := rpt.HTMLBytes()
	if err != nil {
		t.Fatal(err)
	}
	gotNorm := normalizeHTML(string(got))

	goldenPath := filepath.Join("..", "..", "testdata", "golden", "audit.html")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(gotNorm), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("updated golden %s", goldenPath)
		return
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update): %v", err)
	}
	wantNorm := normalizeHTML(string(wantBytes))
	if gotNorm != wantNorm {
		t.Fatalf("HTML differs from golden %s (run: go test ./internal/report -update)", goldenPath)
	}
}

func goldenGraph() *graph.Graph {
	return &graph.Graph{
		Nodes: map[graph.TableKey]*graph.Table{
			{Database: "audit_test", Name: "events_bad"}: {
				Key:              graph.TableKey{Database: "audit_test", Name: "events_bad"},
				Engine:           "MergeTree",
				SortingKey:       "user_uuid, day",
				PartitionKey:     "toYYYYMM(day)",
				CreateTableQuery: "CREATE TABLE audit_test.events_bad (`day` Date, `user_uuid` UUID, `created_at` String, `segment` String) ENGINE = MergeTree PARTITION BY toYYYYMM(day) ORDER BY (user_uuid, day) TTL toDate(day) + toIntervalDay(400)",
				HealthScore:      85,
				NodeType:         graph.NodeTable,
				Columns: []graph.Column{
					{Name: "created_at", Type: "String"},
					{Name: "segment", Type: "String"},
				},
				ActiveParts: 200,
			},
		},
	}
}

var wsCollapse = regexp.MustCompile(`\s+`)

func normalizeHTML(s string) string {
	s = strings.TrimSpace(s)
	s = wsCollapse.ReplaceAllString(s, " ")
	return s
}
