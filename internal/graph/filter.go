package graph

import "strings"

// ShouldAnalyze returns false for ClickHouse-internal or replication staging tables.
func ShouldAnalyze(key TableKey, engine string) bool {
	name := key.Name
	db := key.Database

	if strings.HasPrefix(name, ".inner_id.") || strings.HasPrefix(name, ".tmp.") {
		return false
	}
	if strings.HasPrefix(name, ".") && strings.Contains(name, "inner_id") {
		return false
	}
	if db == "peerdb" && strings.HasPrefix(name, "_peerdb_") {
		return false
	}
	// Ephemeral query pseudo-node
	if db == "_queries" {
		return false
	}
	_ = engine
	return true
}

// DisplayName formats db.table for humans (avoids confusing default..inner_id).
func (k TableKey) DisplayName() string {
	if strings.HasPrefix(k.Name, ".") {
		return k.Database + "." + "`" + k.Name + "`"
	}
	return k.String()
}
