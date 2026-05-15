package graph

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type MermaidOpts struct {
	Title    string
	MaxNodes int
}

const defaultMermaidMaxNodes = 40

func sanitizeID(key TableKey) string {
	s := key.Database + "__" + key.Name
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	id := re.ReplaceAllString(s, "_")
	if id == "" {
		id = "node"
	}
	if id[0] >= '0' && id[0] <= '9' {
		id = "n_" + id
	}
	return id
}

func escapeMermaidText(s string) string {
	repl := strings.NewReplacer(
		`"`, "#quot;",
		"[", "#91;",
		"]", "#93;",
		"(", "#40;",
		")", "#41;",
		"|", "#124;",
		"#", "#35;",
		"<", "#lt;",
		">", "#gt;",
		"&", "#amp;",
		"\n", "<br>",
		"\r", "",
	)
	return repl.Replace(s)
}

func (g *Graph) ToMermaid(opts MermaidOpts) string {
	maxNodes := opts.MaxNodes
	if maxNodes <= 0 {
		maxNodes = defaultMermaidMaxNodes
	}

	include := g.selectMermaidNodes(maxNodes)
	if len(include) == 0 {
		return "flowchart TB\n    empty[\"No tables to display\"]\n"
	}

	var sb strings.Builder
	sb.WriteString("flowchart TB\n")
	if opts.Title != "" {
		sb.WriteString("    %% ")
		sb.WriteString(opts.Title)
		sb.WriteByte('\n')
	}

	for key, node := range g.Nodes {
		if !include[key] {
			continue
		}
		id := sanitizeID(key)
		label := escapeMermaidText(fmt.Sprintf("%s<br>%s", key.DisplayName(), node.Engine))
		if node.HealthScore > 0 && node.HealthScore < 100 {
			label += escapeMermaidText(fmt.Sprintf("<br>health: %.0f", node.HealthScore))
		}
		sb.WriteString(mermaidNode(id, label, node.NodeType))
		sb.WriteByte('\n')
		if node.Color != "" {
			sb.WriteString(fmt.Sprintf("    style %s stroke:%s,stroke-width:3px\n", id, node.Color))
		}
	}

	edgeCount := 0
	const maxEdges = 80
	for _, e := range g.Edges {
		if edgeCount >= maxEdges {
			break
		}
		if e.From == e.To {
			continue
		}
		if e.From.Database == "_queries" {
			continue // skip workload hub (clutters + breaks layout)
		}
		if !include[e.From] || !include[e.To] {
			continue
		}
		from := sanitizeID(e.From)
		to := sanitizeID(e.To)
		arrow := mermaidArrow(e.Type)
		lbl := e.Label
		if lbl == "" {
			lbl = string(e.Type)
		}
		if e.Weight > 0 {
			lbl = fmt.Sprintf("%s %.0f", lbl, e.Weight)
		}
		lbl = escapeMermaidText(lbl)
		sb.WriteString(fmt.Sprintf("    %s %s|%s| %s\n", from, arrow, lbl, to))
		edgeCount++
	}

	return sb.String()
}

func (g *Graph) selectMermaidNodes(max int) map[TableKey]bool {
	type scored struct {
		key   TableKey
		score float64
	}
	var list []scored
	for key, node := range g.Nodes {
		if !ShouldAnalyze(key, node.Engine) {
			continue
		}
		sc := node.HealthScore
		if sc > 100 {
			sc = 100
		}
		// Prefer unhealthy + hot tables
		sc = 100 - sc
		sc += float64(node.HotQueries) * 5
		if node.HealthScore < 100 {
			sc += 20
		}
		list = append(list, scored{key: key, score: sc})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].score > list[j].score
	})
	if len(list) > max {
		list = list[:max]
	}
	out := make(map[TableKey]bool, len(list))
	for _, s := range list {
		out[s.key] = true
	}
	// Always include MV dependency endpoints for selected nodes
	for _, e := range g.Edges {
		if e.Type == EdgeDependsOn || e.Type == EdgePopulates {
			if out[e.From] {
				out[e.To] = true
			}
			if out[e.To] {
				out[e.From] = true
			}
		}
	}
	return out
}

func mermaidNode(id, label string, t NodeType) string {
	switch t {
	case NodeMV:
		return fmt.Sprintf(`    %s{{"%s"}}`, id, label)
	case NodeView:
		return fmt.Sprintf(`    %s("%s")`, id, label)
	case NodeDictionary:
		return fmt.Sprintf(`    %s[["%s"]]`, id, label)
	default:
		return fmt.Sprintf(`    %s["%s"]`, id, label)
	}
}

func mermaidArrow(t EdgeType) string {
	switch t {
	case EdgeDependsOn, EdgeQueryUses:
		return "-->"
	default:
		return "-.->"
	}
}
