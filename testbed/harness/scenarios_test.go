package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScenarioCommands(t *testing.T) {
	runner := new(recordingRunner)
	scenario, err := ScenarioByID(ComposeCmd{Project: "workshop-v4"}, "S1")
	if err != nil {
		t.Fatal(err)
	}
	if err := scenario.Inject(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "compose", "-p", "workshop-v4", "stop", "redis"}
	if !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("inject argv = %#v, want %#v", runner.args, want)
	}
	if err := scenario.Restore(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	want = []string{"docker", "compose", "-p", "workshop-v4", "start", "redis"}
	if !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("restore argv = %#v, want %#v", runner.args, want)
	}
}

func TestUnknownScenarioListsValidIDs(t *testing.T) {
	_, err := ScenarioByID(ComposeCmd{}, "bad")
	if err == nil {
		t.Fatal("expected unknown scenario error")
	}
}

func TestS4RequiresOverrideFile(t *testing.T) {
	t.Setenv("IACON_S4_OVERRIDE", filepath.Join(t.TempDir(), "missing.yml"))
	scenario, err := ScenarioByID(ComposeCmd{Project: "workshop-v4"}, "S4")
	if err != nil {
		t.Fatal(err)
	}
	if err := scenario.Inject(context.Background(), new(recordingRunner)); err == nil {
		t.Fatal("expected missing override error")
	}
}

func TestS4ComposeSequence(t *testing.T) {
	override := filepath.Join(t.TempDir(), "postgres-tmpfs.yml")
	if err := os.WriteFile(override, []byte("services: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IACON_S4_OVERRIDE", override)
	app := ComposeCmd{Project: "workshop-v4", Files: []string{"docker-compose-app.yml"}}
	runner := new(recordingRunner)
	scenario, err := ScenarioByID(app, "S4")
	if err != nil {
		t.Fatal(err)
	}
	if err := scenario.Inject(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "compose", "-p", "workshop-v4", "-f", "docker-compose-app.yml", "-f", override,
		"up", "-d", "--wait", "postgres"}
	if len(runner.calls) == 0 || !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("first inject argv = %#v, want %#v", runner.calls[0], want)
	}
}

func TestS5UsesPgBouncerForSaturatedSessions(t *testing.T) {
	t.Setenv("PG_POOL_SIZE", "2")
	runner := new(recordingRunner)
	scenario, err := ScenarioByID(ComposeCmd{Project: "workshop-v4"}, "S5")
	if err != nil {
		t.Fatal(err)
	}
	if err := scenario.Inject(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(runner.calls))
	}
	for _, call := range runner.calls {
		if !contains(call, "--host=pgbouncer") {
			t.Fatalf("saturation call = %#v, want pgbouncer host", call)
		}
	}
}

func TestS7StartsAndStopsEphemeralLoadgen(t *testing.T) {
	runner := &scriptedRunner{outputs: []string{"container-id"}}
	scenario, err := ScenarioByID(ComposeCmd{}, "S7")
	if err != nil {
		t.Fatal(err)
	}
	if err := scenario.Inject(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if !contains(runner.calls[0], "MODE=app-errors") {
		t.Fatalf("start call = %#v", runner.calls[0])
	}
	if err := scenario.Restore(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.calls[1], []string{"docker", "kill", "container-id"}) {
		t.Fatalf("restore call = %#v", runner.calls[1])
	}
}

type scriptedRunner struct {
	calls   [][]string
	outputs []string
}

func (r *scriptedRunner) Exec(_ context.Context, args ...string) (string, error) {
	r.calls = append(r.calls, append([]string(nil), args...))
	if len(r.outputs) == 0 {
		return "", nil
	}
	out := r.outputs[0]
	r.outputs = r.outputs[1:]
	return out, nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
