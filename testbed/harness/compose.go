package main

import (
	"context"
	"os/exec"
	"strconv"
)

type Runner interface {
	Exec(ctx context.Context, args ...string) (string, error)
}

type ComposeCmd struct {
	Project string
	Files   []string
	Workdir string
}

type ExecRunner struct{}

func (ExecRunner) Exec(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (c ComposeCmd) argv(args ...string) []string {
	argv := []string{"docker", "compose", "-p", c.Project}
	for _, file := range c.Files {
		argv = append(argv, "-f", file)
	}
	return append(argv, args...)
}

func (c ComposeCmd) Run(ctx context.Context, runner Runner, args ...string) (string, error) {
	return runner.Exec(ctx, c.argv(args...)...)
}

func (c ComposeCmd) Up(ctx context.Context, runner Runner) (string, error) {
	return c.Run(ctx, runner, "up", "-d")
}
func (c ComposeCmd) Down(ctx context.Context, runner Runner) (string, error) {
	return c.Run(ctx, runner, "down")
}
func (c ComposeCmd) Stop(ctx context.Context, runner Runner, services ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"stop"}, services...)...)
}
func (c ComposeCmd) Start(ctx context.Context, runner Runner, services ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"start"}, services...)...)
}
func (c ComposeCmd) Pause(ctx context.Context, runner Runner, services ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"pause"}, services...)...)
}
func (c ComposeCmd) Unpause(ctx context.Context, runner Runner, services ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"unpause"}, services...)...)
}
func (c ComposeCmd) Restart(ctx context.Context, runner Runner, services ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"restart"}, services...)...)
}
func (c ComposeCmd) Kill(ctx context.Context, runner Runner, services ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"kill"}, services...)...)
}
func (c ComposeCmd) Exec(ctx context.Context, runner Runner, service string, command ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"exec", "-T", service}, command...)...)
}
func (c ComposeCmd) ExecDetached(ctx context.Context, runner Runner, service string, command ...string) (string, error) {
	return c.Run(ctx, runner, append([]string{"exec", "-d", service}, command...)...)
}
func (c ComposeCmd) LogsTail(ctx context.Context, runner Runner, service string, lines int) (string, error) {
	return c.Run(ctx, runner, "logs", "--tail", strconv.Itoa(lines), service)
}
