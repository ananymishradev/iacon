package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ListenAddr               string
	AppURL                   string
	PromURL                  string
	ClickHouseURL            string
	CHUser                   string
	CHPassword               string
	PGDSN                    string
	PGDSNDirect              string
	PGBDSN                   string
	PGPoolSize               int
	RedisAddr                string
	OTELExporterOTLPEndpoint string
	OTELServiceName          string
	DataDir                  string
	RetentionIncidents       int
	CanaryMode               string
	IaconMode                string
	NarrateWhen              string
	NarratorProvider         string
	NarratorModel            string
	NarratorAPIKey           string
	NarratorBaseURL          string
}

func Load() (Config, error) {
	get := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	cfg := Config{
		ListenAddr:               get("LISTEN_ADDR", ":8090"),
		AppURL:                   get("APP_URL", ""),
		PromURL:                  get("PROM_URL", ""),
		ClickHouseURL:            get("CLICKHOUSE_URL", ""),
		CHUser:                   get("CH_USER", ""),
		CHPassword:               get("CH_PASSWORD", ""),
		PGDSN:                    get("PG_DSN", ""),
		PGDSNDirect:              get("PG_DSN_DIRECT", ""),
		PGBDSN:                   get("PGB_DSN", ""),
		RedisAddr:                get("REDIS_ADDR", ""),
		OTELExporterOTLPEndpoint: get("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTELServiceName:          get("OTEL_SERVICE_NAME", "iacon"),
		DataDir:                  get("DATA_DIR", "/data"),
		CanaryMode:               get("CANARY_MODE", "plain"),
		IaconMode:                get("IACON_MODE", "rules"),
		NarrateWhen:              get("NARRATE_WHEN", "unknown"),
		NarratorProvider:         get("NARRATOR_PROVIDER", "fake"),
		NarratorModel:            get("NARRATOR_MODEL", ""),
		NarratorAPIKey:           get("NARRATOR_API_KEY", ""),
		NarratorBaseURL:          get("NARRATOR_BASE_URL", ""),
	}
	var err error
	cfg.PGPoolSize, err = parseInt("PG_POOL_SIZE", get("PG_POOL_SIZE", "5"))
	if err != nil {
		return Config{}, err
	}
	cfg.RetentionIncidents, err = parseInt("RETENTION_INCIDENTS", get("RETENTION_INCIDENTS", "50"))
	if err != nil {
		return Config{}, err
	}
	for _, key := range []struct {
		name  string
		value string
	}{
		{"APP_URL", cfg.AppURL}, {"PROM_URL", cfg.PromURL}, {"CLICKHOUSE_URL", cfg.ClickHouseURL},
		{"PG_DSN", cfg.PGDSN}, {"PG_DSN_DIRECT", cfg.PGDSNDirect}, {"PGB_DSN", cfg.PGBDSN},
		{"REDIS_ADDR", cfg.RedisAddr},
	} {
		if key.value == "" {
			return Config{}, fmt.Errorf("%s is required", key.name)
		}
	}
	if !strings.Contains(cfg.PGDSN, "default_query_exec_mode=simple_protocol") {
		return Config{}, fmt.Errorf("PG_DSN must contain default_query_exec_mode=simple_protocol")
	}
	if !strings.Contains(cfg.PGBDSN, "default_query_exec_mode=simple_protocol") {
		return Config{}, fmt.Errorf("PGB_DSN must contain default_query_exec_mode=simple_protocol")
	}
	if cfg.PGPoolSize <= 0 {
		return Config{}, fmt.Errorf("PG_POOL_SIZE must be positive")
	}
	if cfg.RetentionIncidents <= 0 {
		return Config{}, fmt.Errorf("RETENTION_INCIDENTS must be positive")
	}
	if !oneOf(cfg.CanaryMode, "plain", "padded_truncate") {
		return Config{}, fmt.Errorf("CANARY_MODE must be plain or padded_truncate")
	}
	if !oneOf(cfg.IaconMode, "rules", "model", "full") {
		return Config{}, fmt.Errorf("IACON_MODE must be rules, model, or full")
	}
	if !oneOf(cfg.NarrateWhen, "unknown", "always") {
		return Config{}, fmt.Errorf("NARRATE_WHEN must be unknown or always")
	}
	return cfg, nil
}

func parseInt(key, value string) (int, error) {
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return number, nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}
