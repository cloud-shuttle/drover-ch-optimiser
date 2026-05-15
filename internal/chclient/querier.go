package chclient

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Querier is the read-only surface used by graph discovery (live or mock).
type Querier interface {
	Query(ctx context.Context, query string, args ...any) (driver.Rows, error)
	Close() error
}

// Conn implements Querier against a real ClickHouse server.
var _ Querier = (*Conn)(nil)
