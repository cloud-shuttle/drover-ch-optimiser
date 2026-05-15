// Package testutil provides helpers for integration and e2e tests against ClickHouse.
package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/config"
)

const (
	DefaultTestHost = "localhost"
	DefaultTestPort = 19000
	TestDatabase    = "audit_test"
)

// CHConfig returns connection settings from env or docker-compose defaults.
func CHConfig() config.Config {
	host := envOr("CLICKHOUSE_HOST", DefaultTestHost)
	port := DefaultTestPort
	if p := os.Getenv("CLICKHOUSE_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	var c config.Config
	c.ClickHouse.Host = host
	c.ClickHouse.Port = port
	c.ClickHouse.Database = TestDatabase
	c.ClickHouse.Username = envOr("CLICKHOUSE_USER", "default")
	c.ClickHouse.Password = os.Getenv("CLICKHOUSE_PASSWORD")
	c.ClickHouse.Secure = os.Getenv("CLICKHOUSE_SECURE") == "true"
	return c
}

// ApplyConfig sets package-level config used by chclient.
func ApplyConfig(t *testing.T) config.Config {
	t.Helper()
	cfg := CHConfig()
	config.C = cfg
	return cfg
}

// EnsureFixtures runs testdata/clickhouse/init.sql if fixture tables are missing.
func EnsureFixtures(t *testing.T, conn clickhouse.Conn) {
	t.Helper()
	ctx := context.Background()
	var n uint8
	if err := conn.QueryRow(ctx, "SELECT 1 FROM audit_test.events_bad LIMIT 1").Scan(&n); err == nil {
		return
	}
	sqlPath := filepath.Join(repoRoot(), "testdata", "clickhouse", "init.sql")
	body, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read init.sql: %v", err)
	}
	for _, stmt := range splitSQL(string(body)) {
		if err := conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("fixture SQL failed: %v\nstmt: %s", err, stmt)
		}
	}
}

// OpenConn dials test ClickHouse or skips the test.
func OpenConn(t *testing.T) clickhouse.Conn {
	t.Helper()
	cfg := ApplyConfig(t)
	opts := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.Username,
			Password: cfg.ClickHouse.Password,
		},
		DialTimeout: 5 * time.Second,
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		t.Skipf("clickhouse not available: %v (run: docker compose -f docker-compose.test.yml up -d)", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		t.Skipf("clickhouse ping failed: %v", err)
	}
	EnsureFixtures(t, conn)
	return conn
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func splitSQL(sql string) []string {
	var out []string
	var b strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			out = append(out, strings.TrimSpace(b.String()))
			b.Reset()
		}
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		out = append(out, s)
	}
	return out
}

// WaitForClickHouse blocks until CH responds or timeout (for scripts).
func WaitForClickHouse(timeout time.Duration) error {
	cfg := CHConfig()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		opts := &clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
			Auth: clickhouse.Auth{Database: cfg.ClickHouse.Database, Username: cfg.ClickHouse.Username, Password: cfg.ClickHouse.Password},
			DialTimeout: 2 * time.Second,
		}
		conn, err := clickhouse.Open(opts)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = conn.Ping(ctx)
			cancel()
			_ = conn.Close()
			if err == nil {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("clickhouse not ready after %s", timeout)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
