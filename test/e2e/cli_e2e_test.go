//go:build e2e

package e2e_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/testutil"
)

func TestCLI_Rules(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "rules")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rules failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Fatal("empty output")
	}
}

func TestCLI_AuditJSON_Integration(t *testing.T) {
	testutil.ApplyConfig(t)
	if err := testutil.WaitForClickHouse(5 * time.Second); err != nil {
		t.Skip(err)
	}

	bin := buildBinary(t)
	dir := t.TempDir()
	outPath := filepath.Join(dir, "audit.json")

	cmd := exec.Command(bin, "audit",
		"--format", "json",
		"--output", outPath,
		"--database", testutil.TestDatabase,
		"--days", "7",
		"--host", testutil.CHConfig().ClickHouse.Host,
		"--port", portStr(testutil.CHConfig().ClickHouse.Port),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit failed: %v\n%s", err, output)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		HealthScore float64 `json:"HealthScore"`
		TotalTables int     `json:"TotalTables"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, data)
	}
	if payload.TotalTables < 2 {
		t.Errorf("expected tables in report, got %d", payload.TotalTables)
	}
}

func TestCLI_AuditHTML_Integration(t *testing.T) {
	testutil.ApplyConfig(t)
	if err := testutil.WaitForClickHouse(5 * time.Second); err != nil {
		t.Skip(err)
	}

	bin := buildBinary(t)
	dir := t.TempDir()
	outPath := filepath.Join(dir, "audit.html")

	cmd := exec.Command(bin, "audit",
		"-f", "html", "-o", outPath,
		"--database", testutil.TestDatabase,
		"--host", testutil.CHConfig().ClickHouse.Host,
		"--port", portStr(testutil.CHConfig().ClickHouse.Port),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("audit html failed: %v\n%s", err, out)
	}
	html, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(html)
	for _, want := range []string{"drover-ch-optimiser", "flowchart TB", "mermaid"} {
		if !strings.Contains(s, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "drover-ch")
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/drover-ch")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func portStr(p int) string {
	return strconv.Itoa(p)
}
