package optimiser

import (
	"context"
	"strings"
	"sync"
	"time"
)

type ChAnalyticalTunerAdapter struct {
	mu      sync.RWMutex
	metrics []LatencyMetric
}

func NewChAnalyticalTunerAdapter() *ChAnalyticalTunerAdapter {
	return &ChAnalyticalTunerAdapter{
		metrics: []LatencyMetric{},
	}
}

func (a *ChAnalyticalTunerAdapter) SeedLatencyMetric(ctx context.Context, m LatencyMetric) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics = append(a.metrics, m)
}

func (a *ChAnalyticalTunerAdapter) SweepLatencyMetrics(ctx context.Context, database string, minDuration time.Duration) ([]LatencyMetric, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	slow := []LatencyMetric{}
	for _, m := range a.metrics {
		if m.Duration >= minDuration {
			slow = append(slow, m)
		}
	}
	return slow, nil
}

func (a *ChAnalyticalTunerAdapter) AnalyzeProjectionSweeps(ctx context.Context, database string) ([]ProjectionSuggestion, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	suggestions := []ProjectionSuggestion{}
	for _, m := range a.metrics {
		qUpper := strings.ToUpper(m.QueryText)
		if !strings.Contains(qUpper, "GROUP BY") {
			continue
		}

		tableName := extractTableName(m.QueryText)
		if tableName == "" {
			continue
		}

		cols := extractGroupByColumns(m.QueryText)
		if len(cols) == 0 {
			continue
		}

		projName := "proj_" + tableName + "_" + strings.Join(cols, "_")
		suggestions = append(suggestions, ProjectionSuggestion{
			TableName:      tableName,
			ProjectionName: projName,
			Columns:        cols,
			QueryPattern:   m.QueryText,
		})
	}

	return suggestions, nil
}

func extractTableName(q string) string {
	parts := strings.Fields(q)
	for i, part := range parts {
		if strings.ToUpper(part) == "FROM" && i+1 < len(parts) {
			tbl := parts[i+1]
			tbl = strings.Trim(tbl, ";`\"' ")
			return tbl
		}
	}
	return ""
}

func extractGroupByColumns(q string) []string {
	qUpper := strings.ToUpper(q)
	idx := strings.Index(qUpper, "GROUP BY")
	if idx == -1 {
		return []string{}
	}

	sub := q[idx+len("GROUP BY"):]
	// Trim other potential clauses
	if stopIdx := strings.Index(strings.ToUpper(sub), "ORDER BY"); stopIdx != -1 {
		sub = sub[:stopIdx]
	}
	if stopIdx := strings.Index(strings.ToUpper(sub), "LIMIT"); stopIdx != -1 {
		sub = sub[:stopIdx]
	}
	if stopIdx := strings.Index(strings.ToUpper(sub), "HAVING"); stopIdx != -1 {
		sub = sub[:stopIdx]
	}

	sub = strings.Trim(sub, "; ")
	rawCols := strings.Split(sub, ",")
	cols := []string{}
	for _, rc := range rawCols {
		c := strings.TrimSpace(rc)
		c = strings.Trim(c, "`\"' ")
		if c != "" {
			cols = append(cols, c)
		}
	}
	return cols
}
