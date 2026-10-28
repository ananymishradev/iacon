package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Scenario struct {
	ID              string
	Kind            string
	TrueCause       string
	Acceptable      []string
	Inject          func(context.Context, Runner) error
	Restore         func(context.Context, Runner) error
	RequiresTraffic bool
}

type RunConfig struct {
	RepoDir        string
	Tag            string
	Agent          string
	RunIndex       int
	Seed           int64
	HoldoutHash    string
	Steady         time.Duration
	InjectJitter   time.Duration
	ObserveTimeout time.Duration
	Settle         time.Duration
}

type RunResult struct {
	Scenario       string    `json:"scenario"`
	Kind           string    `json:"kind"`
	Agent          string    `json:"agent"`
	Tag            string    `json:"tag"`
	RunIndex       int       `json:"run_index"`
	Seed           int64     `json:"seed"`
	HoldoutHash    string    `json:"holdout_hash,omitempty"`
	Status         string    `json:"status"`
	TInject        time.Time `json:"t_inject"`
	TRestore       time.Time `json:"t_restore"`
	InjectOffsetMS int64     `json:"inject_offset_ms"`
	TrueCause      string    `json:"true_cause"`
	IncidentID     string    `json:"incident_id,omitempty"`
	Cause          string    `json:"cause,omitempty"`
	Confidence     string    `json:"confidence,omitempty"`
	TOpenMS        *int64    `json:"t_open_ms,omitempty"`
	TReportMS      *int64    `json:"t_report_ms,omitempty"`
	TResolvedMS    *int64    `json:"t_resolved_ms,omitempty"`
	ReopenCount    int       `json:"reopen_count"`
}

type groundTruth struct {
	Scenario  string    `json:"scenario"`
	TInject   time.Time `json:"t_inject"`
	TRestore  time.Time `json:"t_restore"`
	TrueCause string    `json:"true_cause"`
	RunDir    string    `json:"run_dir"`
}

func RunScenario(ctx context.Context, scenario Scenario, runner Runner, observer AgentObserver, cfg RunConfig) (RunResult, error) {
	if cfg.Tag == "" {
		cfg.Tag = "manual"
	}
	if cfg.Agent == "" {
		cfg.Agent = "v0"
	}
	dir := filepath.Join(cfg.RepoDir, "eval", "runs", cfg.Tag, scenario.ID, fmt.Sprint(cfg.RunIndex))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return RunResult{}, err
	}
	result := RunResult{
		Scenario: scenario.ID, Kind: scenario.Kind, Agent: cfg.Agent, Tag: cfg.Tag,
		RunIndex: cfg.RunIndex, Seed: cfg.Seed, TrueCause: scenario.TrueCause,
		HoldoutHash: cfg.HoldoutHash,
	}
	before, err := observer.Current(ctx)
	if err != nil {
		return result, err
	}
	if before.Open {
		result.Status = "pre_open"
		return persistRunResult(dir, result)
	}
	jitter := time.Duration(0)
	if cfg.InjectJitter > 0 {
		jitter = time.Duration(rand.New(rand.NewSource(cfg.Seed)).Int63n(int64(cfg.InjectJitter) + 1))
	}
	result.InjectOffsetMS = jitter.Milliseconds()
	timer := time.NewTimer(cfg.Steady + jitter)
	select {
	case <-ctx.Done():
		return result, ctx.Err()
	case <-timer.C:
	}
	result.TInject = time.Now().UTC()
	if err := scenario.Inject(ctx, runner); err != nil {
		return result, err
	}
	deadline := time.Now().Add(cfg.ObserveTimeout)
	for time.Now().Before(deadline) {
		current, currentErr := observer.Current(ctx)
		if currentErr == nil && current.Open {
			result.IncidentID = current.IncidentID
			result.Cause = current.Cause
			result.Confidence = current.Confidence
			openMS := time.Since(result.TInject).Milliseconds()
			result.TOpenMS = &openMS
			break
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	if result.IncidentID == "" {
		result.Status = "observe_timeout"
	} else {
		result.Status = "ok"
	}
	result.TRestore = time.Now().UTC()
	if err := scenario.Restore(ctx, runner); err != nil {
		result.Status = "restore_failed"
		_, _ = persistRunResult(dir, result)
		return result, err
	}
	persisted, err := persistRunResult(dir, result)
	if err != nil {
		return result, err
	}
	if err := appendGroundTruth(cfg.RepoDir, groundTruth{
		Scenario: scenario.ID, TInject: result.TInject, TRestore: result.TRestore,
		TrueCause: scenario.TrueCause, RunDir: dir,
	}); err != nil {
		return result, err
	}
	return persisted, nil
}

func persistRunResult(dir string, result RunResult) (RunResult, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return result, err
	}
	tmp, err := os.CreateTemp(dir, "run-*.json")
	if err != nil {
		return result, err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return result, err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return result, err
	}
	if err := tmp.Close(); err != nil {
		return result, err
	}
	if err := os.Rename(tmpName, filepath.Join(dir, "run.json")); err != nil {
		return result, err
	}
	return result, nil
}

func appendGroundTruth(repoDir string, record groundTruth) error {
	path := filepath.Join(repoDir, "eval", "ground_truth.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(data, '\n'))
	return err
}
