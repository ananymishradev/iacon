package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
		"S4": diskFullScenario(app),
		"S5": poolSaturationScenario(app),
		"S6": makeScenario("S6", "dev", "app_down", "stop", "app"),
		"S7": appErrorsScenario(),
		"H1": makeScenario("H1", "holdout", "redis_unavailable", "pause", "redis"),
		"H2": makeScenario("H2", "holdout", "postgres_down", "pause", "postgres"),
		"H3": makeScenario("H3", "holdout", "monitoring_degraded", "stop", "prometheus"),
	}
}

func poolSaturationScenario(app ComposeCmd) Scenario {
	sessionCount := 5
	if value, err := strconv.Atoi(os.Getenv("PG_POOL_SIZE")); err == nil && value > 0 {
		sessionCount = value
	}
	return Scenario{
		ID: "S5", Kind: "dev", TrueCause: "pool_saturated",
		Inject: func(ctx context.Context, runner Runner) error {
			for i := 0; i < sessionCount; i++ {
				if _, err := app.ExecDetached(ctx, runner, "postgres", "psql",
					"--host=pgbouncer", "--username=workshop", "--dbname=workshop",
					"--command=SELECT pg_sleep(3599)"); err != nil {
					return fmt.Errorf("start saturation session %d: %w", i+1, err)
				}
			}
			return nil
		},
		Restore: func(ctx context.Context, runner Runner) error {
			_, err := app.Exec(ctx, runner, "postgres", "psql",
				"--host=postgres", "--username=workshop", "--dbname=workshop",
				"--command=SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE query LIKE '%pg_sleep(3599)%'")
			return err
		},
	}
}

func appErrorsScenario() Scenario {
	var containerID string
	return Scenario{
		ID: "S7", Kind: "dev", TrueCause: "app_internal_errors",
		Inject: func(ctx context.Context, runner Runner) error {
			id, err := startEphemeralLoadgen(ctx, runner, "app-errors")
			if err == nil {
				containerID = id
			}
			return err
		},
		Restore: func(ctx context.Context, runner Runner) error {
			err := stopEphemeral(ctx, runner, containerID)
			containerID = ""
			return err
		},
	}
}

func diskFullScenario(app ComposeCmd) Scenario {
	const fillFile = "/tmp/iacon-disk-fill"
	return Scenario{
		ID: "S4", Kind: "dev", TrueCause: "database_disk_full",
		Inject: func(ctx context.Context, runner Runner) error {
			override := os.Getenv("IACON_S4_OVERRIDE")
			if override == "" {
				override = filepath.Join("testbed", "overrides", "postgres-tmpfs.yml")
			}
			if _, err := os.Stat(override); err != nil {
				return fmt.Errorf("S4 override file %q: %w", override, err)
			}
			withOverride := app.WithFiles(append(append([]string(nil), app.Files...), override)...)
			if _, err := withOverride.Run(ctx, runner, "up", "-d", "--wait", "postgres"); err != nil {
				return err
			}
			if _, err := app.Restart(ctx, runner, "app"); err != nil {
				return err
			}
			if _, err := app.Exec(ctx, runner, "app", "wget", "-q", "-O-", "http://localhost:8080/ready"); err != nil {
				return fmt.Errorf("wait for app readiness: %w", err)
			}
			_, err := app.Exec(ctx, runner, "postgres", "sh", "-c",
				"while dd if=/dev/zero of=/tmp/iacon-disk-fill bs=1M oflag=append conv=notrunc 2>/dev/null; do :; done")
			return err
		},
		Restore: func(ctx context.Context, runner Runner) error {
			if _, err := app.Exec(ctx, runner, "postgres", "rm", "-f", fillFile); err != nil {
				return err
			}
			if _, err := app.Run(ctx, runner, "up", "-d", "--wait", "postgres"); err != nil {
				return err
			}
			_, err := app.Exec(ctx, runner, "postgres", "pg_isready", "-U", "workshop", "-d", "workshop")
			return err
		},
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
