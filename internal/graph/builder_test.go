package graph

import "testing"

func TestParseTableRef(t *testing.T) {
	tests := []struct {
		in       string
		wantDB   string
		wantName string
	}{
		{"default.events", "default", "events"},
		{"default..inner_id.uuid", "default", ".inner_id.uuid"},
		{"events", "default", "events"},
		{"  db.tbl  ", "db", "tbl"},
	}
	for _, tt := range tests {
		got := parseTableRef(tt.in)
		if got.Database != tt.wantDB || got.Name != tt.wantName {
			t.Errorf("parseTableRef(%q) = %+v, want %s.%s", tt.in, got, tt.wantDB, tt.wantName)
		}
	}
}
