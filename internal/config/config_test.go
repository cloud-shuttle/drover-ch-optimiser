package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func resetViper(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"CLICKHOUSE_HOST", "CLICKHOUSE_PORT", "CLICKHOUSE_DATABASE",
		"CLICKHOUSE_USER", "CLICKHOUSE_PASSWORD", "CLICKHOUSE_SECURE",
	} {
		t.Setenv(k, "")
	}
	viper.Reset()
	viper.SetDefault("clickhouse.host", "localhost")
	viper.SetDefault("clickhouse.port", 9000)
	viper.SetDefault("clickhouse.database", "default")
	viper.SetDefault("clickhouse.username", "default")
	viper.SetDefault("clickhouse.password", "")
	viper.SetDefault("clickhouse.secure", false)
}

func TestLoad_FromFile(t *testing.T) {
	resetViper(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	err := os.WriteFile(path, []byte(`
clickhouse:
  host: ch.example.com
  port: 9440
  database: analytics
  username: reader
  password: secret
  secure: true
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := Load(path); err != nil {
		t.Fatal(err)
	}
	if C.ClickHouse.Host != "ch.example.com" || C.ClickHouse.Port != 9440 {
		t.Fatalf("unexpected config: %+v", C.ClickHouse)
	}
	if !C.ClickHouse.Secure {
		t.Error("expected secure true")
	}
}

func TestLoad_MissingFileUsesDefaults(t *testing.T) {
	resetViper(t)
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")
	if err := Load(path); err != nil {
		t.Fatal(err)
	}
	if C.ClickHouse.Host != "localhost" || C.ClickHouse.Port != 9000 {
		t.Fatalf("defaults not applied: %+v", C.ClickHouse)
	}
}

func TestApplyOverrides(t *testing.T) {
	resetViper(t)
	_ = Load(filepath.Join(t.TempDir(), "missing.yaml"))
	secure := true
	ApplyOverrides(Overrides{
		Host:     "override.host",
		Port:     1234,
		Username: "u",
		Password: "p",
		Secure:   &secure,
	})
	if C.ClickHouse.Host != "override.host" || C.ClickHouse.Port != 1234 || !C.ClickHouse.Secure {
		t.Fatalf("overrides not applied: %+v", C.ClickHouse)
	}
}
