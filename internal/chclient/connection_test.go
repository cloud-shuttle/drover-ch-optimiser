package chclient

import (
	"errors"
	"strings"
	"testing"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/config"
)

func TestAuthHintError_AuthFailure(t *testing.T) {
	config.C.ClickHouse.Host = "localhost"
	config.C.ClickHouse.Port = 9000
	config.C.ClickHouse.Username = "default"

	err := authHintError(errors.New("code: 516, message: Authentication failed: password is incorrect"))
	if err == nil {
		t.Fatal("expected error")
	}
	s := err.Error()
	for _, want := range []string{"authentication failed", "CLICKHOUSE_PASSWORD", "--password"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in: %s", want, s)
		}
	}
}

func TestAuthHintError_Generic(t *testing.T) {
	config.C.ClickHouse.Host = "ch"
	config.C.ClickHouse.Port = 1
	err := authHintError(errors.New("connection refused"))
	if !strings.Contains(err.Error(), "ping clickhouse") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("code: 516", "516") {
		t.Error("expected match")
	}
	if containsAny("ok", "516") {
		t.Error("unexpected match")
	}
}
