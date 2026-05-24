package optimiser

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticalTuner_SweepLatencyMetrics(t *testing.T) {
	ctx := context.Background()
	
	// Create mock tuner seeded with sample execution patterns
	tuner := NewChAnalyticalTunerAdapter()

	// Seed latency items
	tuner.SeedLatencyMetric(ctx, LatencyMetric{
		QueryHash: "qh-1",
		QueryText: "SELECT * FROM sales GROUP BY store_id",
		Duration:  120 * time.Millisecond,
		Freq:      10,
	})
	tuner.SeedLatencyMetric(ctx, LatencyMetric{
		QueryHash: "qh-2",
		QueryText: "SELECT id FROM logs WHERE severity = 'error'",
		Duration:  10 * time.Millisecond,
		Freq:      5,
	})

	// Test sweeping with a min duration filter
	metrics, err := tuner.SweepLatencyMetrics(ctx, "default", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("expected 1 slow query, got: %d", len(metrics))
	}

	if metrics[0].QueryHash != "qh-1" {
		t.Errorf("expected slow query qh-1, got: %s", metrics[0].QueryHash)
	}
}

func TestAnalyticalTuner_AnalyzeProjectionSweeps(t *testing.T) {
	ctx := context.Background()
	tuner := NewChAnalyticalTunerAdapter()

	// Seed query logs that represent heavy aggregate/filtering patterns suitable for projections
	tuner.SeedLatencyMetric(ctx, LatencyMetric{
		QueryHash: "qh-3",
		QueryText: "SELECT count() FROM pageviews GROUP BY country, device",
		Duration:  300 * time.Millisecond,
		Freq:      45,
	})

	suggestions, err := tuner.AnalyzeProjectionSweeps(ctx, "default")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("expected 1 projection suggestion, got: %d", len(suggestions))
	}

	s := suggestions[0]
	if s.TableName != "pageviews" {
		t.Errorf("expected table name pageviews, got: %s", s.TableName)
	}
	if s.ProjectionName != "proj_pageviews_country_device" {
		t.Errorf("expected projection name proj_pageviews_country_device, got: %s", s.ProjectionName)
	}
	if len(s.Columns) != 2 || s.Columns[0] != "country" || s.Columns[1] != "device" {
		t.Errorf("expected columns [country device], got: %v", s.Columns)
	}
}
