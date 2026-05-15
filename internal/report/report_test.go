package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

func TestWriteMarkdown(t *testing.T) {
	g := sampleGraph()
	issues, summary := analyzer.Run(g)
	rpt := New(g, issues, summary, 30)

	dir := t.TempDir()
	path := filepath.Join(dir, "audit.md")
	if err := rpt.WriteMarkdown(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.Contains(content, "# ClickHouse Audit Report") {
		t.Error("missing title")
	}
	if !strings.Contains(content, "flowchart TB") {
		t.Error("missing mermaid block")
	}
}

func TestWriteHTML(t *testing.T) {
	g := sampleGraph()
	issues, summary := analyzer.Run(g)
	rpt := New(g, issues, summary, 30)

	dir := t.TempDir()
	path := filepath.Join(dir, "audit.html")
	if err := rpt.WriteHTML(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	for _, want := range []string{"drover-ch-optimiser", "mermaid", "flowchart TB"} {
		if !strings.Contains(content, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	g := sampleGraph()
	issues, summary := analyzer.Run(g)
	rpt := New(g, issues, summary, 30)

	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")
	if err := rpt.WriteJSON(path); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(path); err != nil || st.Size() == 0 {
		t.Fatal("json output empty or missing")
	}
}

func sampleGraph() *graph.Graph {
	return &graph.Graph{
		Nodes: map[graph.TableKey]*graph.Table{
			{Database: "db", Name: "events"}: {
				Key:         graph.TableKey{Database: "db", Name: "events"},
				Engine:      "MergeTree",
				SortingKey:  "day, id",
				HealthScore: 100,
				NodeType:    graph.NodeTable,
			},
		},
	}
}
