package chclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/config"
)

type Conn struct {
	conn clickhouse.Conn
}

func NewConnection() (*Conn, error) {
	cfg := config.C.ClickHouse
	opts := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 300,
		},
		DialTimeout: 10 * time.Second,
	}
	if cfg.Secure {
		opts.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		_ = conn.Close()
		return nil, authHintError(err)
	}
	return &Conn{conn: conn}, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

// NewFromNative wraps an existing connection (for integration tests).
func NewFromNative(conn clickhouse.Conn) *Conn {
	return &Conn{conn: conn}
}

func (c *Conn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	return c.conn.Query(ctx, query, args...)
}

func (c *Conn) Native() clickhouse.Conn {
	return c.conn
}

func authHintError(err error) error {
	cfg := config.C.ClickHouse
	msg := err.Error()
	if containsAny(msg, "516", "Authentication failed", "password is incorrect") {
		return fmt.Errorf(
			"clickhouse authentication failed for user %q at %s:%d: %w\n\n"+
				"Fix credentials via:\n"+
				"  • config file: --config ~/.drover/ch-optimiser.yaml (see .drover/ch-optimiser.example.yaml)\n"+
				"  • env: CLICKHOUSE_USER, CLICKHOUSE_PASSWORD, CLICKHOUSE_HOST, CLICKHOUSE_PORT\n"+
				"  • flags: --user, --password, --host, --port, --secure\n\n"+
				"ClickHouse Cloud: set secure: true, port 9440, and the service password from the console.\n"+
				"Local installs often require a password (see users.d/default-password.xml on the server).",
			cfg.Username,
			cfg.Host,
			cfg.Port,
			err,
		)
	}
	return fmt.Errorf("ping clickhouse at %s:%d: %w", cfg.Host, cfg.Port, err)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if sub != "" && strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
