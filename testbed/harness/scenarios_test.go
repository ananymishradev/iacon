package main

import (
	"context"
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
