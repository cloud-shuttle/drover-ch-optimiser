package graph

import (
	"strings"
	"testing"
)

func TestSanitizeID(t *testing.T) {
	id := sanitizeID(TableKey{Database: "default", Name: "my-table"})
	if strings.Contains(id, "-") || strings.Contains(id, ".") {
		t.Fatalf("sanitizeID should not contain . or -: %q", id)
	}
	if id[0] >= '0' && id[0] <= '9' {
		t.Fatalf("sanitizeID should not start with digit: %q", id)
	}
}

func TestEscapeMermaidText(t *testing.T) {
	in := `foo "bar" (baz) | pipe`
	out := escapeMermaidText(in)
	if strings.Contains(out, `"`) {
		t.Fatalf("quotes should be escaped: %q", out)
	}
	if !strings.Contains(out, "#40;") {
		t.Fatalf("parens should be escaped: %q", out)
	}
}

func TestToMermaidNoSyntaxBreakers(t *testing.T) {
	g := &Graph{
		Nodes: map[TableKey]*Table{
			{Database: "default", Name: "events"}: {
				Key:         TableKey{Database: "default", Name: "events"},
				Engine:      "MergeTree",
				HealthScore: 50,
				Color:       "#ef4444",
				NodeType:    NodeTable,
			},
			{Database: "default", Name: "mv_events"}: {
				Key:      TableKey{Database: "default", Name: "mv_events"},
				Engine:   "MaterializedView",
				NodeType: NodeMV,
			},
		},
		Edges: []Edge{
			{From: TableKey{"default", "events"}, To: TableKey{"default", "mv_events"}, Type: EdgeDependsOn, Label: "MV reads"},
		},
	}
	out := g.ToMermaid(MermaidOpts{MaxNodes: 10})
	for _, bad := range []string{"-->", "flowchart TB"} {
		if !strings.Contains(out, bad) {
			t.Errorf("expected mermaid to contain %q", bad)
		}
	}
	if strings.Contains(out, "->>") {
		t.Error("self-loop or invalid arrow syntax")
	}
}

func TestToMermaidSkipsSelfLoop(t *testing.T) {
	key := TableKey{Database: "default", Name: "t"}
	g := &Graph{
		Nodes: map[TableKey]*Table{key: {Key: key, Engine: "MergeTree", NodeType: NodeTable}},
		Edges: []Edge{{From: key, To: key, Type: EdgePopulates, Label: "populates"}},
	}
	out := g.ToMermaid(MermaidOpts{MaxNodes: 5})
	// self-loop edge should not appear as From->To with same id twice in one line
	lines := strings.Split(out, "\n")
	selfEdges := 0
	id := sanitizeID(key)
	for _, line := range lines {
		if strings.Contains(line, id) && strings.Count(line, id) >= 2 && strings.Contains(line, "-->") {
			selfEdges++
		}
	}
	if selfEdges > 0 {
		t.Errorf("self-loop should not be rendered: %s", out)
	}
}
