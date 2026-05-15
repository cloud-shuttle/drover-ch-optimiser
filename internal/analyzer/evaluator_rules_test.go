package analyzer

import (
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

func TestReplaceTemplate(t *testing.T) {
	got := replaceTemplate("col {{.Column}} count {{.Count}}", map[string]string{
		"Column": "status",
		"Count":  "3",
	})
	if got != "col status count 3" {
		t.Fatalf("got %q", got)
	}
}

func TestEvaluateRule_TooManyParts(t *testing.T) {
	_ = loadRules()
	rule := Rule{
		ID: "schema-parts-too-many-active", Severity: High, Impact: 8,
		Condition: Condition{Type: "too_many_parts", MessageTemplate: "{{.Count}} parts"},
		Suggestion: "fix",
	}
	tbl := &graph.Table{ActiveParts: 200, Key: graph.TableKey{Database: "db", Name: "t"}, Engine: "MergeTree"}
	issue := evaluateRule(rule, tbl, &graph.Graph{})
	if issue == nil {
		t.Fatal("expected issue")
	}
}

func TestEvaluateRule_ExcessiveNullable(t *testing.T) {
	_ = loadRules()
	var rule Rule
	for _, r := range rules {
		if r.Condition.Type == "excessive_nullable" {
			rule = r
			break
		}
	}
	tbl := &graph.Table{
		Key:    graph.TableKey{Database: "db", Name: "weather"},
		Engine: "MergeTree",
		Columns: []graph.Column{
			{Name: "a", Type: "Nullable(String)"},
			{Name: "b", Type: "Nullable(Int32)"},
			{Name: "c", Type: "Nullable(Float64)"},
		},
	}
	issue := evaluateRule(rule, tbl, &graph.Graph{})
	if issue == nil {
		t.Fatal("expected nullable issue")
	}
}

func TestEvaluateRule_PartitionHint(t *testing.T) {
	_ = loadRules()
	var rule Rule
	for _, r := range rules {
		if r.Condition.Type == "partition_cardinality" {
			rule = r
			break
		}
	}
	tbl := &graph.Table{
		Key:          graph.TableKey{Database: "db", Name: "t"},
		Engine:       "MergeTree",
		PartitionKey: "user_uuid",
	}
	issue := evaluateRule(rule, tbl, &graph.Graph{})
	if issue == nil {
		t.Fatal("expected partition hint")
	}
}

func TestEveryRuleHasHandler(t *testing.T) {
	_ = loadRules()
	known := map[string]bool{
		"sorting_key_cardinality": true, "missing_filter_in_pk": true,
		"string_abuse": true, "missing_lowcardinality": true,
		"excessive_nullable": true, "partition_cardinality": true,
		"missing_skipping_index": true, "mutation_heavy": true,
		"too_many_parts": true,
	}
	for _, r := range rules {
		if !known[r.Condition.Type] {
			t.Errorf("rule %s has unhandled condition type %q", r.ID, r.Condition.Type)
		}
	}
}

func TestRun_FindsIssuesOnBadSchema(t *testing.T) {
	g := &graph.Graph{
		Nodes: map[graph.TableKey]*graph.Table{
			{Database: "audit_test", Name: "events_bad"}: {
				Key:          graph.TableKey{Database: "audit_test", Name: "events_bad"},
				Engine:       "MergeTree",
				SortingKey:   "user_uuid, day",
				PartitionKey: "toYYYYMM(day)",
				Columns: []graph.Column{
					{Name: "created_at", Type: "String"},
					{Name: "status", Type: "String"},
				},
			},
		},
	}
	issues, summary := Run(g)
	if summary.TotalIssues == 0 {
		t.Fatal("expected issues on bad schema fixture")
	}
	var foundString, foundPK bool
	for _, iss := range issues[Critical] {
		if iss.RuleID == "schema-types-native-types" {
			foundString = true
		}
		if iss.RuleID == "schema-pk-cardinality-order" {
			foundPK = true
		}
	}
	if !foundString {
		t.Error("expected string abuse on created_at")
	}
	if !foundPK {
		t.Error("expected PK cardinality on user_uuid")
	}
}
