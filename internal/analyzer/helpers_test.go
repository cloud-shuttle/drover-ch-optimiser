package analyzer

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

func TestDetectHighCardFirst_ReplacingMergeTreeIncidentID(t *testing.T) {
	tbl := &graph.Table{Engine: "SharedReplacingMergeTree"}
	got := detectHighCardFirst(tbl, "incident_id, event_time", nil)
	if got != "" {
		t.Errorf("incident_id on ReplacingMergeTree should not flag, got %q", got)
	}
}

func TestDetectHighCardFirst_OtelTraceID(t *testing.T) {
	tbl := &graph.Table{
		Key:    graph.TableKey{Database: "default", Name: "otel_traces_trace_id_ts"},
		Engine: "MergeTree",
	}
	got := detectHighCardFirst(tbl, "TraceId, Timestamp", nil)
	if got != "" {
		t.Errorf("TraceId on trace_id_ts table should be allowed, got %q", got)
	}
}

func TestDetectHighCardFirst_DynamoDBID(t *testing.T) {
	tbl := &graph.Table{
		Key:    graph.TableKey{Database: "default", Name: "dynamodb_traffic_incidents"},
		Engine: "MergeTree",
	}
	got := detectHighCardFirst(tbl, "id", nil)
	if got != "" {
		t.Errorf("id on dynamodb table should be allowed, got %q", got)
	}
}

func TestDetectHighCardFirst_UUIDFirst(t *testing.T) {
	tbl := &graph.Table{Engine: "MergeTree"}
	got := detectHighCardFirst(tbl, "user_uuid, day", nil)
	if got != "user_uuid" {
		t.Errorf("got %q, want user_uuid", got)
	}
}

func TestDetectStringAbuse_OtelAllowlist(t *testing.T) {
	cols := []graph.Column{
		{Name: "MetricDescription", Type: "String"},
		{Name: "created_at", Type: "String"},
	}
	got := detectStringAbuse("MergeTree", cols)
	if len(got) != 1 || got[0] != "created_at" {
		t.Fatalf("want [created_at], got %v", got)
	}
}

func TestDetectStringAbuse_StatusNotCritical(t *testing.T) {
	cols := []graph.Column{{Name: "status", Type: "String"}}
	got := detectStringAbuse("MergeTree", cols)
	if len(got) != 0 {
		t.Fatalf("status alone should not be string_abuse, got %v", got)
	}
}

func TestCountNullable(t *testing.T) {
	cols := []graph.Column{
		{Name: "a", Type: "Nullable(String)"},
		{Name: "b", Type: "UInt64"},
		{Name: "c", Type: "Nullable(Int32)"},
	}
	if n := countNullable(cols); n != 2 {
		t.Fatalf("countNullable = %d, want 2", n)
	}
}

func TestParseKeyColumns(t *testing.T) {
	got := parseKeyColumns("`foo`, bar")
	if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
		t.Fatalf("parseKeyColumns = %v", got)
	}
}
