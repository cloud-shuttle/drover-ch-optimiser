package graph

import (
	"fmt"
	"strings"
)

// ImpactAnalysis summarizes upstream/downstream blast radius for a focus table.
type ImpactAnalysis struct {
	Focus       TableKey
	Upstream    []TableKey // tables this focus reads from (MV sources, etc.)
	Downstream  []TableKey // tables that depend on focus
	HotQueries  int
	FilterHints []string
}

// AnalyzeImpact walks dependency edges from a focus table (v1.1 what-if foundation).
func AnalyzeImpact(g *Graph, focus string) ImpactAnalysis {
	key := parseTableRef(focus)
	out := ImpactAnalysis{Focus: key}
	node, ok := g.Nodes[key]
	if ok {
		out.HotQueries = node.HotQueries
		out.FilterHints = append([]string(nil), node.FilterColumns...)
	}
	seenUp := map[TableKey]bool{key: true}
	seenDown := map[TableKey]bool{key: true}
	for _, e := range g.Edges {
		switch e.Type {
		case EdgeDependsOn, EdgeDistributed:
			if e.To == key && !seenUp[e.From] {
				seenUp[e.From] = true
				out.Upstream = append(out.Upstream, e.From)
			}
			if e.From == key && !seenDown[e.To] {
				seenDown[e.To] = true
				out.Downstream = append(out.Downstream, e.To)
			}
		}
	}
	return out
}

// FormatReport returns human-readable impact lines for CLI suggest output.
func (ia ImpactAnalysis) FormatReport() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Impact analysis for %s\n", ia.Focus.DisplayName()))
	if ia.HotQueries > 0 {
		b.WriteString(fmt.Sprintf("  Query-log hits (30d window): %d\n", ia.HotQueries))
	}
	if len(ia.FilterHints) > 0 {
		b.WriteString(fmt.Sprintf("  Frequent filter columns: %s\n", strings.Join(ia.FilterHints, ", ")))
	}
	if len(ia.Upstream) > 0 {
		b.WriteString("  Upstream (reads from): ")
		for i, k := range ia.Upstream {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(k.DisplayName())
		}
		b.WriteByte('\n')
	}
	if len(ia.Downstream) > 0 {
		b.WriteString("  Downstream (dependent): ")
		for i, k := range ia.Downstream {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(k.DisplayName())
		}
		b.WriteByte('\n')
	}
	if len(ia.Upstream) == 0 && len(ia.Downstream) == 0 {
		b.WriteString("  No MV/distributed edges found; review query_log filters before ORDER BY changes.\n")
	}
	b.WriteString("  Review ALTER ORDER BY with stakeholders; re-run with --focus after changes.\n")
	return b.String()
}
