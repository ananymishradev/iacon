package classify

import (
	"testing"

	"iacon/internal/causes"
	"iacon/internal/evidence"
)

func TestClassifyRedisUsesDirectEvidence(t *testing.T) {
	v := Classify(evidence.Evidence{Items: []evidence.Item{
		{ID: "E1", Source: "app", OK: true, Data: map[string]any{"ready_status": 503}},
		{ID: "E2", Source: "redis", OK: true, Data: map[string]any{"ping_ok": false}},
		{ID: "E3", Source: "pgbouncer", OK: true, Data: map[string]any{"pgb_reachable": true}},
		{ID: "E4", Source: "postgres", OK: true, Data: map[string]any{"direct_ok": true}},
	}}, Cfg{})
	if v.Cause != causes.RedisUnavailable {
		t.Fatalf("cause = %#v", v)
	}
}
