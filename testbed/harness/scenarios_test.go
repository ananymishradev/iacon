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
