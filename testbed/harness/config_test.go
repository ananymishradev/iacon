package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigRequiresWorkshopDir(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "testbed", "harness"), 0o755); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "testbed", "harness", "config.yaml"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WORKSHOP_DIR", "")
	_, err = LoadConfig(repo)
	if err == nil || !strings.Contains(err.Error(), "WORKSHOP_DIR") {
		t.Fatalf("expected WORKSHOP_DIR error, got %v", err)
	}
}

func TestLoadConfigValidatesCauseAndHash(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "testbed", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := []byte("workshop_dir: /tmp/workshop\nscenarios:\n  - id: H1\n    kind: holdout\n    true_cause: redis_unavailable\n    acceptable: [redis_unavailable, unknown]\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), yaml, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WORKSHOP_DIR", "")
	first, err := LoadConfig(repo)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadConfig(repo)
	if err != nil || first.HoldoutHash != second.HoldoutHash {
		t.Fatalf("unstable hash: %q, %q, %v", first.HoldoutHash, second.HoldoutHash, err)
	}
}
