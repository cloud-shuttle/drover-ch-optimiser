//go:build integration

package report_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/report"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/testutil"
)

func TestFullAuditPipeline_Integration(t *testing.T) {
	native := testutil.OpenConn(t)
	defer native.Close()

	conn := chclient.NewFromNative(native)
	g, err := graph.BuildGraph(conn, graph.Options{Database: testutil.TestDatabase, DaysBack: 7})
	if err != nil {
		t.Fatal(err)
	}
	issues, summary := analyzer.Run(g)
	rpt := report.New(g, issues, summary, 7)

	dir := t.TempDir()
	md := filepath.Join(dir, "audit.md")
	html := filepath.Join(dir, "audit.html")
	jsonPath := filepath.Join(dir, "audit.json")

	if err := rpt.WriteMarkdown(md); err != nil {
		t.Fatal(err)
	}
	if err := rpt.WriteHTML(html); err != nil {
		t.Fatal(err)
	}
	if err := rpt.WriteJSON(jsonPath); err != nil {
		t.Fatal(err)
	}

	htmlBytes, _ := os.ReadFile(html)
	if !strings.Contains(string(htmlBytes), "flowchart TB") {
		t.Error("HTML report missing mermaid diagram")
	}
	if summary.TotalIssues == 0 {
		t.Error("expected issues in fixture warehouse")
	}
}
