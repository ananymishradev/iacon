package main

import (
	"context"
	"fmt"
	"sort"
)

func Scenarios(app ComposeCmd) map[string]Scenario {
	makeScenario := func(id, kind, cause string, verb string, service string) Scenario {
		return Scenario{
			ID: id, Kind: kind, TrueCause: cause,
			Inject: func(ctx context.Context, runner Runner) error {
				_, err := app.Run(ctx, runner, verb, service)
				return err
			},
			Restore: func(ctx context.Context, runner Runner) error {
				restoreVerb := "start"
				if verb == "pause" {
					restoreVerb = "unpause"
				}
				_, err := app.Run(ctx, runner, restoreVerb, service)
				return err
			},
		}
	}
	return map[string]Scenario{
		"S1": makeScenario("S1", "dev", "redis_unavailable", "stop", "redis"),
		"S2": makeScenario("S2", "dev", "pgbouncer_down", "stop", "pgbouncer"),
		"S3": makeScenario("S3", "dev", "postgres_down", "stop", "postgres"),
		"S6": makeScenario("S6", "dev", "app_down", "stop", "app"),
		"H1": makeScenario("H1", "holdout", "redis_unavailable", "pause", "redis"),
		"H2": makeScenario("H2", "holdout", "postgres_down", "pause", "postgres"),
		"H3": makeScenario("H3", "holdout", "monitoring_degraded", "stop", "prometheus"),
	}
}

func ScenarioByID(app ComposeCmd, id string) (Scenario, error) {
	scenarios := Scenarios(app)
	scenario, ok := scenarios[id]
	if ok {
		return scenario, nil
	}
	ids := make([]string, 0, len(scenarios))
	for known := range scenarios {
		ids = append(ids, known)
	}
	sort.Strings(ids)
	return Scenario{}, fmt.Errorf("unknown scenario %q; valid scenarios: %v", id, ids)
}
