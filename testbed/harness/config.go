package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"iacon/internal/causes"
)

type ScenarioConfig struct {
	ID                  string   `yaml:"id" json:"id"`
	Kind                string   `yaml:"kind" json:"kind"`
	TrueCause           string   `yaml:"true_cause" json:"true_cause"`
	Acceptable          []string `yaml:"acceptable,omitempty" json:"acceptable,omitempty"`
	SoakMinutes         int      `yaml:"soak_minutes,omitempty" json:"soak_minutes,omitempty"`
	ExpectNoAppIncident bool     `yaml:"expect_no_app_incident,omitempty" json:"expect_no_app_incident,omitempty"`
	ExpectNoIncident    bool     `yaml:"expect_no_incident,omitempty" json:"expect_no_incident,omitempty"`
}

type Config struct {
	WorkshopDir           string           `yaml:"workshop_dir"`
	AppProject            string           `yaml:"app_project"`
	AgentProject          string           `yaml:"agent_project"`
	IaconProject          string           `yaml:"iacon_project"`
	AppComposeFile        string           `yaml:"app_compose_file"`
	AgentComposeFile      string           `yaml:"agent_compose_file"`
	SteadySeconds         int              `yaml:"steady_seconds"`
	InjectJitterSeconds   int              `yaml:"inject_jitter_seconds"`
	ObserveTimeoutSeconds int              `yaml:"observe_timeout_seconds"`
	SettleSeconds         int              `yaml:"settle_seconds"`
	Scenarios             []ScenarioConfig `yaml:"scenarios"`
	HoldoutHash           string           `yaml:"-"`
}

func LoadConfig(repoDir string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(repoDir, "testbed", "harness", "config.yaml"))
	if err != nil {
		return Config{}, fmt.Errorf("read config.yaml: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config.yaml: %w", err)
	}
	localPath := filepath.Join(repoDir, "testbed", "harness", "config.local.yaml")
	if local, readErr := os.ReadFile(localPath); readErr == nil {
		if err := yaml.Unmarshal(local, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config.local.yaml: %w", err)
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read config.local.yaml: %w", readErr)
	}
	if env := os.Getenv("WORKSHOP_DIR"); env != "" {
		cfg.WorkshopDir = env
	}
	if cfg.WorkshopDir == "" {
		return Config{}, errors.New("WORKSHOP_DIR is required")
	}
	for _, scenario := range cfg.Scenarios {
		if scenario.TrueCause != "" && !causes.Valid(scenario.TrueCause) {
			return Config{}, fmt.Errorf("scenario %s has unknown cause %q", scenario.ID, scenario.TrueCause)
		}
		for _, acceptable := range scenario.Acceptable {
			if !causes.Valid(acceptable) {
				return Config{}, fmt.Errorf("scenario %s has unknown acceptable cause %q", scenario.ID, acceptable)
			}
		}
	}
	hashInput, _ := json.Marshal(cfg.Scenarios)
	sum := sha256.Sum256(hashInput)
	cfg.HoldoutHash = hex.EncodeToString(sum[:])
	return cfg, nil
}
