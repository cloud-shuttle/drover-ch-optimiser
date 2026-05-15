package graph

import "testing"

func TestExtractFilterColumns(t *testing.T) {
	cols := []string{"day", "status", "user_uuid", "created_at"}
	q := `SELECT count() FROM audit_test.events_bad PREWHERE status = 'ok' WHERE day >= '2024-01-01' AND user_uuid = 'x'`
	got := ExtractFilterColumns(q, cols)
	want := map[string]bool{"status": true, "day": true, "user_uuid": true}
	if len(got) < 2 {
		t.Fatalf("got %v", got)
	}
	for _, c := range got {
		if !want[c] {
			t.Fatalf("unexpected column %q in %v", c, got)
		}
	}
}

func TestExtractFilterColumns_IgnoresUnknown(t *testing.T) {
	got := ExtractFilterColumns("WHERE unknown_col = 1 AND day = 1", []string{"day"})
	if len(got) != 1 || got[0] != "day" {
		t.Fatalf("got %v", got)
	}
}
