package graph

import "testing"

func TestCrossJoinUnboundedHeuristic(t *testing.T) {
	if !CrossJoinUnboundedHeuristic("SELECT * FROM a CROSS JOIN b LIMIT 1") {
		t.Fatal("expected cross join without filter")
	}
	if CrossJoinUnboundedHeuristic("SELECT * FROM a CROSS JOIN b WHERE a.id = b.id") {
		t.Fatal("where should clear flag")
	}
	if CrossJoinUnboundedHeuristic("SELECT * FROM a PREWHERE x > 0 CROSS JOIN b") {
		t.Fatal("prewhere should clear flag")
	}
	if CrossJoinUnboundedHeuristic("SELECT * FROM a INNER JOIN b ON a.id=b.id") {
		t.Fatal("inner join not cross join")
	}
}
