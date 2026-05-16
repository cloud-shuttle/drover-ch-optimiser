package graph

import "strings"

// CrossJoinUnboundedHeuristic is true when the query uses CROSS JOIN and has no
// WHERE or PREWHERE clause (naïve; good enough for query_log triage).
func CrossJoinUnboundedHeuristic(query string) bool {
	lower := strings.ToLower(query)
	if !strings.Contains(lower, " cross join ") {
		return false
	}
	return !strings.Contains(lower, " where ") && !strings.Contains(lower, " prewhere ")
}
