package graph

import "testing"

func TestShouldAnalyze(t *testing.T) {
	tests := []struct {
		db, name, engine string
		want             bool
	}{
		{"default", "events", "MergeTree", true},
		{"default", ".inner_id.e40f45be-3c4d-4b27-94d2-4f0bd3441b0c", "SharedReplacingMergeTree", false},
		{"default", ".tmp.foo", "MergeTree", false},
		{"peerdb", "_peerdb_raw_pg", "MergeTree", false},
		{"peerdb", "public_customers", "MergeTree", true},
		{"_queries", "workload", "", false},
	}
	for _, tt := range tests {
		got := ShouldAnalyze(TableKey{Database: tt.db, Name: tt.name}, tt.engine)
		if got != tt.want {
			t.Errorf("ShouldAnalyze(%s.%s) = %v, want %v", tt.db, tt.name, got, tt.want)
		}
	}
}

func TestTableKeyDisplayName(t *testing.T) {
	k := TableKey{Database: "default", Name: ".inner_id.abc"}
	want := "default.`" + ".inner_id.abc`"
	if got := k.DisplayName(); got != want {
		t.Errorf("DisplayName() = %q, want %q", got, want)
	}
	got2 := TableKey{Database: "default", Name: "events"}.DisplayName()
	if got2 != "default.events" {
		t.Errorf("DisplayName() = %q", got2)
	}
}
