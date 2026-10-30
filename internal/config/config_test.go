package config

import "testing"

func validEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_URL", "http://app:8080")
	t.Setenv("PROM_URL", "http://prometheus:9090")
	t.Setenv("CLICKHOUSE_URL", "http://clickhouse:8123")
	t.Setenv("PG_DSN", "postgres://x?default_query_exec_mode=simple_protocol")
	t.Setenv("PG_DSN_DIRECT", "postgres://x")
	t.Setenv("PGB_DSN", "postgres://x?default_query_exec_mode=simple_protocol")
	t.Setenv("REDIS_ADDR", "redis:6379")
}

func TestLoadValidation(t *testing.T) {
	validEnv(t)
	t.Setenv("LISTEN_ADDR", ":9911")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != ":9911" {
		t.Fatalf("listen addr = %q", cfg.ListenAddr)
	}
	t.Setenv("APP_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing APP_URL")
	}
	validEnv(t)
	t.Setenv("PG_DSN", "postgres://x")
	if _, err := Load(); err == nil {
		t.Fatal("expected simple protocol validation")
	}
	validEnv(t)
	t.Setenv("IACON_MODE", "bad")
	if _, err := Load(); err == nil {
		t.Fatal("expected mode validation")
	}
}
