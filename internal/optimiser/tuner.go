package optimiser

import (
	"context"
	"time"
)

type LatencyMetric struct {
	QueryHash string        `json:"query_hash"`
	QueryText string        `json:"query_text"`
	Duration  time.Duration `json:"duration"`
	Freq      int           `json:"freq"`
}

type ProjectionSuggestion struct {
	TableName      string   `json:"table_name"`
	ProjectionName string   `json:"projection_name"`
	Columns        []string `json:"columns"`
	QueryPattern   string   `json:"query_pattern"`
}

type AnalyticalTuner interface {
	SweepLatencyMetrics(ctx context.Context, database string, minDuration time.Duration) ([]LatencyMetric, error)
	AnalyzeProjectionSweeps(ctx context.Context, database string) ([]ProjectionSuggestion, error)
}
