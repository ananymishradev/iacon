package evidence

import (
	"context"
	"fmt"
	"strings"
)

type Queryer interface {
	Query(context.Context, string) ([]map[string]any, error)
}

type PGBouncerCollector struct {
	DSN     string
	Queryer Queryer
}

func (c PGBouncerCollector) Name() string { return "pgbouncer" }

func (c PGBouncerCollector) Collect(ctx context.Context) (map[string]any, error) {
	if c.Queryer == nil {
		return map[string]any{"pgb_reachable": false, "admin_ok": false, "err": "queryer unavailable"}, nil
	}
	data := map[string]any{"pgb_reachable": true, "admin_ok": false}
	rows, err := c.Queryer.Query(ctx, "SHOW POOLS")
	if err != nil {
		data["err"] = err.Error()
		if isTransport(err) {
			data["pgb_reachable"] = false
		}
		return data, nil
	}
	data["admin_ok"] = len(rows) > 0
	if len(rows) > 0 {
		data["cl_waiting"] = number(rows[0]["cl_waiting"])
		data["sv_active"] = number(rows[0]["sv_active"])
		data["maxwait_s"] = number(rows[0]["maxwait"])
	}
	return data, nil
}

type PostgresCollector struct {
	DirectDSN string
	Queryer   Queryer
}

func (c PostgresCollector) Name() string { return "postgres" }

func (c PostgresCollector) Collect(ctx context.Context) (map[string]any, error) {
	if c.Queryer == nil {
		return map[string]any{"direct_ok": false, "err": "queryer unavailable"}, nil
	}
	_, err := c.Queryer.Query(ctx, "SELECT 1")
	if err != nil {
		data := map[string]any{"direct_ok": false, "err": err.Error()}
		if code := sqlState(err); code != "" {
			data["err_code"] = code
		}
		return data, nil
	}
	return map[string]any{"direct_ok": true}, nil
}

func number(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

func isTransport(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "refused") || strings.Contains(text, "timeout") ||
		strings.Contains(text, "connection")
}

func sqlState(err error) string {
	text := err.Error()
	for _, marker := range []string{"SQLSTATE ", "sqlstate ", "code "} {
		if index := strings.Index(text, marker); index >= 0 && len(text) >= index+len(marker)+5 {
			return text[index+len(marker) : index+len(marker)+5]
		}
	}
	return ""
}

type StaticQueryer struct {
	Rows  []map[string]any
	Error error
}

func (q StaticQueryer) Query(context.Context, string) ([]map[string]any, error) {
	if q.Error != nil {
		return nil, q.Error
	}
	if q.Rows == nil {
		return nil, fmt.Errorf("no rows")
	}
	return q.Rows, nil
}
