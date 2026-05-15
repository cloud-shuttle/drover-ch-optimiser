package analyzer

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
	"gopkg.in/yaml.v3"
)

//go:embed rules.yaml
var rulesYAML []byte

type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
)

type Rule struct {
	ID          string    `yaml:"id"`
	Category    string    `yaml:"category"`
	Name        string    `yaml:"name"`
	Severity    Severity  `yaml:"severity"`
	Impact      int       `yaml:"impact"`
	Description string    `yaml:"description"`
	Condition   Condition `yaml:"condition"`
	Suggestion  string    `yaml:"suggestion"`
}

type Condition struct {
	Type            string `yaml:"type"`
	MessageTemplate string `yaml:"message_template"`
}

type OptimizationIssue struct {
	RuleID     string   `json:"rule_id"`
	RuleName   string   `json:"rule_name"`
	Severity   Severity `json:"severity"`
	Impact     int      `json:"impact"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
	Table      string   `json:"table"`
}

type AuditSummary struct {
	HealthScore float64        `json:"health_score"`
	TotalIssues int            `json:"total_issues"`
	Critical    int            `json:"critical"`
	High        int            `json:"high"`
	ByCategory  map[string]int `json:"by_category"`
}

type rulesFile struct {
	Rules []Rule `yaml:"rules"`
}

var (
	rulesOnce sync.Once
	rules     []Rule
	loadErr   error
)

func loadRules() error {
	rulesOnce.Do(func() {
		var rf rulesFile
		loadErr = yaml.Unmarshal(rulesYAML, &rf)
		rules = rf.Rules
	})
	return loadErr
}

func ListRules() {
	if err := loadRules(); err != nil {
		fmt.Println("error loading rules:", err)
		return
	}
	fmt.Printf("%-40s %-10s %s\n", "ID", "SEVERITY", "NAME")
	fmt.Println(strings.Repeat("-", 90))
	for _, r := range rules {
		fmt.Printf("%-40s %-10s %s\n", r.ID, r.Severity, r.Name)
	}
}

func Run(g *graph.Graph) (map[Severity][]OptimizationIssue, AuditSummary) {
	_ = loadRules()

	issuesBySeverity := make(map[Severity][]OptimizationIssue)
	summary := AuditSummary{ByCategory: make(map[string]int)}

	var wg sync.WaitGroup
	var mu sync.Mutex

	var healthSum float64
	var healthN int

	for key, tbl := range g.Nodes {
		if !graph.ShouldAnalyze(key, tbl.Engine) {
			continue
		}
		wg.Add(1)
		go func(key graph.TableKey, tbl *graph.Table) {
			defer wg.Done()
			tableIssues := evaluateTable(tbl, g)
			mu.Lock()
			defer mu.Unlock()
			for _, issue := range tableIssues {
				issue.Table = key.DisplayName()
				issuesBySeverity[issue.Severity] = append(issuesBySeverity[issue.Severity], issue)
				summary.TotalIssues++
				summary.ByCategory[issue.RuleID]++
			}
			applyTableHealth(tbl, tableIssues)
			healthSum += tbl.HealthScore
			healthN++
		}(key, tbl)
	}
	wg.Wait()

	summary.Critical = len(issuesBySeverity[Critical])
	summary.High = len(issuesBySeverity[High])
	if healthN > 0 {
		summary.HealthScore = healthSum / float64(healthN)
	} else {
		summary.HealthScore = 100
	}

	return issuesBySeverity, summary
}

func applyTableHealth(tbl *graph.Table, issues []OptimizationIssue) {
	penalty := 0
	worst := Low
	for _, iss := range issues {
		switch iss.Severity {
		case Critical:
			penalty += 25
			worst = Critical
		case High:
			penalty += 12
			if worst != Critical {
				worst = High
			}
		case Medium:
			penalty += 5
		}
	}
	tbl.HealthScore = 100 - float64(penalty)
	if tbl.HealthScore < 0 {
		tbl.HealthScore = 0
	}
	tbl.Color = severityColor(worst)
}

func severityColor(s Severity) string {
	switch s {
	case Critical:
		return "#ef4444"
	case High:
		return "#f97316"
	case Medium:
		return "#eab308"
	default:
		return "#22c55e"
	}
}

func evaluateTable(tbl *graph.Table, g *graph.Graph) []OptimizationIssue {
	var issues []OptimizationIssue
	for _, rule := range rules {
		if issue := evaluateRule(rule, tbl, g); issue != nil {
			issues = append(issues, *issue)
		}
	}
	return issues
}

func evaluateRule(rule Rule, tbl *graph.Table, g *graph.Graph) *OptimizationIssue {
	var msg string

	switch rule.Condition.Type {
	case "sorting_key_cardinality":
		if bad := detectHighCardFirst(tbl, tbl.SortingKey, tbl.Columns); bad != "" {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{
				"BadColumn": bad,
			})
		}
	case "missing_filter_in_pk":
		if col := missingFilterInPK(tbl); col != "" {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{"Column": col})
		}
	case "string_abuse":
		if cols := detectStringAbuse(tbl.Engine, tbl.Columns); len(cols) > 0 {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{
				"Count":  fmt.Sprintf("%d", len(cols)),
				"Column": cols[0],
			})
		}
	case "missing_lowcardinality":
		if col := missingLowCardinality(tbl.Columns); col != "" {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{"Column": col})
		}
	case "excessive_nullable":
		if n := countNullable(tbl.Columns); n >= 3 {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{
				"Count": fmt.Sprintf("%d", n),
			})
		}
	case "partition_cardinality":
		if hint := partitionKeyHint(tbl.PartitionKey); hint != "" {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{"Column": hint})
		}
	case "missing_skipping_index":
		if col := missingSkippingIndex(tbl); col != "" {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{"Column": col})
		}
	case "mutation_heavy":
		if n := tbl.IncompleteMutations; n > 0 {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{
				"Count": fmt.Sprintf("%d", n),
			})
		}
	case "too_many_parts":
		if tbl.ActiveParts > 150 {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{
				"Count":     fmt.Sprintf("%d", tbl.ActiveParts),
				"Threshold": "150",
			})
		}
	case "too_many_columns":
		if n := len(tbl.Columns); n > 80 {
			msg = replaceTemplate(rule.Condition.MessageTemplate, map[string]string{
				"Count":     fmt.Sprintf("%d", n),
				"Threshold": "80",
			})
		}
	default:
		return nil
	}

	if msg == "" {
		return nil
	}

	return &OptimizationIssue{
		RuleID:     rule.ID,
		RuleName:   rule.Name,
		Severity:   rule.Severity,
		Impact:     rule.Impact,
		Message:    msg,
		Suggestion: rule.Suggestion,
	}
}

func replaceTemplate(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{."+k+"}}", v)
	}
	return out
}
