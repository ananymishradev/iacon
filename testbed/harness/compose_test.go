package main

import (
	"context"
	"reflect"
	"testing"
)

type recordingRunner struct {
	args  []string
	calls [][]string
}

func (r *recordingRunner) Exec(_ context.Context, args ...string) (string, error) {
	r.args = append([]string(nil), args...)
	r.calls = append(r.calls, append([]string(nil), args...))
	return "", nil
}

func TestComposeKeepsPathsAsSingleArguments(t *testing.T) {
	runner := new(recordingRunner)
	cmd := ComposeCmd{
		Project: "workshop-v4",
		Files:   []string{"docker-compose-app.yml"},
		Workdir: "/tmp/path with spaces",
	}
	if _, err := cmd.Stop(context.Background(), runner, "redis"); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "compose", "-p", "workshop-v4", "-f", "docker-compose-app.yml", "stop", "redis"}
	if !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("argv = %#v, want %#v", runner.args, want)
	}
}
