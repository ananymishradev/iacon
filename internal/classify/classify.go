package classify

import (
	"fmt"
	"strings"

	"iacon/internal/causes"
	"iacon/internal/detect"
	"iacon/internal/evidence"
)

type rule struct {
	id    string
	cause string
	kind  string
	match func(map[string]any, Cfg) bool
	keys  []string
}

func Classify(ev evidence.Evidence, cfg Cfg) detect.Verdict {
	values := map[string]any{}
	ids := map[string]string{}
	missing := false
	for _, item := range ev.Items {
		for key, value := range item.Data {
			values[key] = value
			ids[key] = item.ID
		}
		if !item.OK {
			missing = true
		}
	}
	rules := rules()
	var matches []rule
	for _, candidate := range rules {
		if candidate.match(values, cfg) {
			matches = append(matches, candidate)
		}
	}
	if len(matches) == 0 {
		return detect.Verdict{Cause: causes.Unknown, Confidence: "low", Kind: "outage", MissingEvidence: missingKeys(ev)}
	}
	primary := matches[0]
	verdict := detect.Verdict{
		Cause: primary.cause, Confidence: "high", Kind: primary.kind,
		MatchedRules: []string{primary.id}, EvidenceIDs: evidenceIDs(ids, primary.keys),
		MissingEvidence: missingKeys(ev),
	}
	for _, match := range matches[1:] {
		verdict.Secondary = append(verdict.Secondary, match.id)
	}
	if missing || len(verdict.EvidenceIDs) == 0 {
		verdict.Confidence = "medium"
	}
	if len(matches) > 1 {
		verdict.Confidence = "low"
	}
	return verdict
}

func rules() []rule {
	return []rule{
		{"R2", causes.PgBouncerDown, "outage", func(v map[string]any, _ Cfg) bool {
			return boolValue(v, "pgb_reachable") == false && canaryBad(v)
		}, []string{"pgb_reachable", "canary_last"}},
		{"R3", causes.PostgresDown, "outage", func(v map[string]any, _ Cfg) bool {
			return boolValue(v, "pgb_reachable") && !boolValue(v, "direct_ok") && canaryBad(v) &&
				(postgresTransport(v) || strings.Contains(fmt.Sprint(v["err_code"]), "57P03"))
		}, []string{"pgb_reachable", "direct_ok", "canary_last"}},
		{"R4", causes.DatabaseDiskFull, "outage", func(v map[string]any, _ Cfg) bool {
			return canaryDiskFull(v) && (numberValue(v, "get_status") == 200 || boolValue(v, "direct_ok"))
		}, []string{"canary_last", "get_status"}},
		{"R1", causes.RedisUnavailable, "outage", func(v map[string]any, _ Cfg) bool {
			return !boolValue(v, "ping_ok") && !canaryBad(v) && boolValue(v, "pgb_reachable")
		}, []string{"ping_ok", "pgb_reachable"}},
		{"R5", causes.PoolSaturated, "degraded", func(v map[string]any, c Cfg) bool {
			return c.AdminEnabled && boolValue(v, "admin_ok") && numberValue(v, "cl_waiting") > 0 &&
				numberValue(v, "sv_active") >= float64(c.PGPoolSize) && boolValue(v, "direct_ok") &&
				(canaryTimeout(v) || strings.Contains(strings.ToLower(fmt.Sprint(v["get_err"])), "timeout"))
		}, []string{"admin_ok", "cl_waiting", "sv_active", "direct_ok"}},
		{"R6", causes.AppDown, "outage", func(v map[string]any, _ Cfg) bool {
			return appUnavailable(v) && boolValue(v, "ping_ok") && boolValue(v, "pgb_reachable") && boolValue(v, "direct_ok")
		}, []string{"ready_status", "get_status", "ping_ok", "direct_ok"}},
		{"R7", causes.AppInternalErrors, "degraded", func(v map[string]any, _ Cfg) bool {
			_, known := v["err_ratio_30s"]
			return numberValue(v, "ready_status") == 200 && boolValue(v, "ping_ok") &&
				boolValue(v, "pgb_reachable") && boolValue(v, "direct_ok") && known && numberValue(v, "err_ratio_30s") >= 0.02
		}, []string{"ready_status", "err_ratio_30s"}},
		{"R8", causes.MonitoringDegraded, "monitoring", func(v map[string]any, _ Cfg) bool {
			return (!boolValue(v, "prom_ok") || !boolValue(v, "ch_ok")) &&
				boolValue(v, "ping_ok") && boolValue(v, "pgb_reachable") && boolValue(v, "direct_ok") &&
				numberValue(v, "ready_status") == 200
		}, []string{"prom_ok", "ch_ok"}},
	}
}

func boolValue(v map[string]any, key string) bool {
	value, ok := v[key].(bool)
	return ok && value
}
func numberValue(v map[string]any, key string) float64 {
	switch n := v[key].(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}
func canaryBad(v map[string]any) bool {
	text := strings.ToLower(fmt.Sprint(v["canary_last"]))
	return strings.Contains(text, "bad") || strings.Contains(text, "53100") || strings.Contains(text, "no space")
}
func canaryDiskFull(v map[string]any) bool {
	text := strings.ToLower(fmt.Sprint(v["canary_last"]))
	return strings.Contains(text, "53100") || strings.Contains(text, "no space")
}
func canaryTimeout(v map[string]any) bool {
	return strings.Contains(strings.ToLower(fmt.Sprint(v["canary_last"])), "timeout")
}
func postgresTransport(v map[string]any) bool {
	text := strings.ToLower(fmt.Sprint(v["err"]))
	return strings.Contains(text, "timeout") || strings.Contains(text, "refused") || strings.Contains(text, "transport")
}
func appUnavailable(v map[string]any) bool {
	return numberValue(v, "ready_status") == 0 && numberValue(v, "get_status") == 0
}
func evidenceIDs(ids map[string]string, keys []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, key := range keys {
		if id := ids[key]; id != "" && !seen[id] {
			out, seen[id] = append(out, id), true
		}
	}
	return out
}
func missingKeys(ev evidence.Evidence) []string {
	var out []string
	for _, item := range ev.Items {
		if !item.OK {
			out = append(out, item.ID)
		}
	}
	return out
}
