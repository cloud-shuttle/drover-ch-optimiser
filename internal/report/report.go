package report

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

//go:embed template.html
var templateFS embed.FS

type Report struct {
	data ReportData
}

type ReportData struct {
	Database       string
	GeneratedAt    string
	HealthScore    float64
	TotalTables    int
	QueryDays      int
	CriticalIssues []analyzer.OptimizationIssue
	HighIssues     []analyzer.OptimizationIssue
	MediumIssues   []analyzer.OptimizationIssue
	MermaidContent template.HTML
	Tables         []TableSummary
}

type TableSummary struct {
	Key         string
	Engine      string
	HealthScore float64
	IssueCount  int
	TotalRows   uint64
	TotalBytesGB float64
}

func New(g *graph.Graph, issues map[analyzer.Severity][]analyzer.OptimizationIssue, summary analyzer.AuditSummary, queryDays int) *Report {
	issueCountByTable := make(map[string]int)
	for sev, list := range issues {
		_ = sev
		for _, iss := range list {
			issueCountByTable[iss.Table]++
		}
	}

	var tables []TableSummary
	for key, node := range g.Nodes {
		if !graph.ShouldAnalyze(key, node.Engine) {
			continue
		}
		tables = append(tables, TableSummary{
			Key:          key.DisplayName(),
			Engine:       node.Engine,
			HealthScore:  node.HealthScore,
			IssueCount:   issueCountByTable[key.String()],
			TotalRows:    node.TotalRows,
			TotalBytesGB: float64(node.TotalBytes) / (1024 * 1024 * 1024),
		})
	}

	db := "all"
	if len(g.Nodes) > 0 {
		for k := range g.Nodes {
			db = k.Database
			break
		}
	}

	return &Report{
		data: ReportData{
			Database:       db,
			GeneratedAt:    time.Now().Format(time.RFC1123),
			HealthScore:    summary.HealthScore,
			TotalTables:    len(tables),
			QueryDays:      queryDays,
			CriticalIssues: issues[analyzer.Critical],
			HighIssues:     issues[analyzer.High],
			MediumIssues:   issues[analyzer.Medium],
			MermaidContent: template.HTML(g.ToMermaid(graph.MermaidOpts{Title: "Warehouse dependency graph"})),
			Tables:         tables,
		},
	}
}

func (r *Report) WriteHTML(path string) error {
	tmpl, err := template.ParseFS(templateFS, "template.html")
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, r.data)
}

func (r *Report) WriteJSON(path string) error {
	b, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func (r *Report) WriteMarkdown(path string) error {
	var b strings.Builder
	d := r.data
	b.WriteString("# ClickHouse Audit Report\n\n")
	b.WriteString(fmt.Sprintf("**Generated:** %s  \n", d.GeneratedAt))
	b.WriteString(fmt.Sprintf("**Health score:** %.1f / 100  \n", d.HealthScore))
	b.WriteString(fmt.Sprintf("**Tables analyzed:** %d  \n\n", d.TotalTables))

	b.WriteString("## Critical issues\n\n")
	if len(d.CriticalIssues) == 0 {
		b.WriteString("_None_\n\n")
	} else {
		for _, iss := range d.CriticalIssues {
			writeIssue(&b, iss)
		}
	}

	b.WriteString("## High issues\n\n")
	if len(d.HighIssues) == 0 {
		b.WriteString("_None_\n\n")
	} else {
		for _, iss := range d.HighIssues {
			writeIssue(&b, iss)
		}
	}

	b.WriteString("## Dependency graph\n\n```mermaid\n")
	b.WriteString(string(d.MermaidContent))
	b.WriteString("\n```\n")

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeIssue(b *strings.Builder, iss analyzer.OptimizationIssue) {
	b.WriteString(fmt.Sprintf("### `%s` — %s\n\n", iss.Table, iss.RuleID))
	b.WriteString(fmt.Sprintf("- **%s**\n", iss.Message))
	b.WriteString(fmt.Sprintf("- Suggestion: %s\n\n", iss.Suggestion))
}

func WriteGraphOnlyHTML(g *graph.Graph, path string) error {
	r := &Report{
		data: ReportData{
			GeneratedAt:    time.Now().Format(time.RFC1123),
			HealthScore:    100,
			MermaidContent: template.HTML(g.ToMermaid(graph.MermaidOpts{})),
		},
	}
	return r.WriteHTML(path)
}

// PrintSummary writes a short summary to stdout.
func PrintSummary(summary analyzer.AuditSummary, path string) {
	fmt.Printf("Report: %s\n", path)
	fmt.Printf("Health score: %.1f/100\n", summary.HealthScore)
	if summary.Critical > 0 {
		fmt.Printf("Critical issues: %d\n", summary.Critical)
	}
}

// For tests: render HTML to buffer.
func (r *Report) HTMLBytes() ([]byte, error) {
	tmpl, err := template.ParseFS(templateFS, "template.html")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
