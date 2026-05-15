package analyzer

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

func TestMutationHeavy_IncompleteMutations(t *testing.T) {
	rule := findRuleByType(t, "mutation_heavy")
	tbl := &graph.Table{
		Key:                 graph.TableKey{Database: "db", Name: "t"},
		Engine:              "MergeTree",
		IncompleteMutations: 3,
	}
	issue := evaluateRule(rule, tbl, &graph.Graph{})
	if issue == nil {
		t.Fatal("expected mutation_heavy issue")
	}
	if issue.Message != "Table has 3 incomplete mutations" {
		t.Errorf("message = %q", issue.Message)
	}
}

func findRuleByType(t *testing.T, condType string) Rule {
	t.Helper()
	if err := loadRules(); err != nil {
		t.Fatal(err)
	}
	for _, r := range rules {
		if r.Condition.Type == condType {
			return r
		}
	}
	t.Fatalf("no rule with type %q", condType)
	return Rule{}
}
