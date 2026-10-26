package main

import (
	"context"
	"fmt"
	"strings"
)

func startEphemeralLoadgen(ctx context.Context, runner Runner, mode string) (string, error) {
	out, err := runner.Exec(ctx,
		"docker", "run", "--rm", "-d", "--network", "workshop-net",
		"-e", "APP_URL=http://app:8080", "-e", "RPS=6", "-e", "MODE="+mode,
		"iacon-loadgen",
	)
	if err != nil {
		return "", fmt.Errorf("start ephemeral loadgen: %w", err)
	}
	id := strings.TrimSpace(out)
	if id == "" {
		return "", fmt.Errorf("start ephemeral loadgen: empty container id")
	}
	return id, nil
}

func stopEphemeral(ctx context.Context, runner Runner, id string) error {
	if id == "" {
		return nil
	}
	if _, err := runner.Exec(ctx, "docker", "kill", id); err != nil {
		return fmt.Errorf("kill ephemeral container %q: %w", id, err)
	}
	return nil
}
