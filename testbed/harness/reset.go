package main

import (
	"context"
	"fmt"
)

func Reset(ctx context.Context, runner Runner, app ComposeCmd) error {
	for _, args := range [][]string{
		{"unpause"},
		{"start"},
		{"rm", "-f", "-s", "-v"},
		{"up", "-d"},
	} {
		if _, err := app.Run(ctx, runner, args...); err != nil {
			return fmt.Errorf("reset compose %v: %w", args, err)
		}
	}
	return nil
}
