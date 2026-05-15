package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
)

func TestRulesCommand(t *testing.T) {
	out := captureStdout(t, func() {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"rules"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "insert-mutation") && !strings.Contains(out, "mutation") {
		t.Errorf("rules output missing mutation rule: %q", truncate(out, 300))
	}
}

func TestSuggestRequiresTableArg(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"suggest"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error without table arg")
	}
}

func TestAudit_MockConnection(t *testing.T) {
	old := newCHConnection
	defer func() { newCHConnection = old }()

	newCHConnection = func() (chclient.Querier, error) {
		return &chclient.MockQuerier{Queries: auditMockQueries()}, nil
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "audit.html")

	out := captureStdout(t, func() {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{
			"audit", "-f", "html", "-o", outPath,
			"--host", "test", "--user", "u", "--password", "p",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Audit completed") {
		t.Errorf("stdout = %q", out)
	}
	if st, err := os.Stat(outPath); err != nil || st.Size() == 0 {
		t.Fatalf("audit html: %v", err)
	}
}

func TestRootVersion(t *testing.T) {
	if NewRootCmd().Version == "" {
		t.Error("expected version set")
	}
}

func TestExcludeInvalidPattern(t *testing.T) {
	exclude = []string{"*"}
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"graph", "--exclude", "*", "--host", "x"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected exclude parse error")
	}
	exclude = nil
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func auditMockQueries() []chclient.MockQuery {
	return []chclient.MockQuery{
		{
			Contains: "system.tables",
			Rows: [][]any{
				{
					"audit_test", "events_bad", "MergeTree", "", "CREATE TABLE",
					"toYYYYMM(day)", "user_uuid, day", "user_uuid, day",
					[]string{}, []string{},
					uint64(10), uint64(100),
				},
			},
		},
		{
			Contains: "system.columns",
			Rows: [][]any{
				{"audit_test", "events_bad", "created_at", "String", "", ""},
				{"audit_test", "events_bad", "status", "String", "", ""},
			},
		},
		{Contains: "system.parts", Rows: [][]any{{"audit_test", "events_bad", uint64(200)}}},
		{Contains: "system.mutations", Rows: nil},
		{Contains: "system.query_log", Rows: nil},
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
