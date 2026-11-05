package main

import (
	"context"
	"fmt"
	"time"
)

func Preflight(ctx context.Context, runner Runner, app ComposeCmd, observer AgentObserver, settle time.Duration) error {
	if _, err := app.Run(ctx, runner, "ps", "--status", "running"); err != nil {
		return fmt.Errorf("workshop services are not running: %w", err)
	}
	if observer != nil {
		observation, err := observer.Current(ctx)
		if err != nil {
			return fmt.Errorf("observe agent: %w", err)
		}
		if observation.Open {
			return fmt.Errorf("agent already has open incident %s", observation.IncidentID)
		}
	}
	if settle > 0 {
		timer := time.NewTimer(settle)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}
