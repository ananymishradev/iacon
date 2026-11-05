package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "help" || os.Args[1] == "--help" {
		fmt.Fprintln(os.Stderr, "usage: harness {run|report|preflight|reset|help}")
		return
	}
	switch os.Args[1] {
	case "run":
		if err := runCommand(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "report":
		if err := reportCommand(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "preflight", "reset":
		if err := lifecycleCommand(os.Args[1], os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", os.Args[1])
		os.Exit(2)
	}
}

func lifecycleCommand(command string, args []string) error {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	repo := "."
	fs.StringVar(&repo, "repo", repo, "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := LoadConfig(repo)
	if err != nil {
		return err
	}
	app := ComposeCmd{Project: cfg.AppProject, Files: []string{filepath.Join(cfg.WorkshopDir, cfg.AppComposeFile)}}
	observer := V0Observer{Cmd: ComposeCmd{Project: cfg.AgentProject, Files: []string{filepath.Join(cfg.WorkshopDir, cfg.AgentComposeFile)}}, Runner: ExecRunner{}, Service: "agent"}
	if command == "preflight" {
		return Preflight(context.Background(), ExecRunner{}, app, observer, time.Duration(cfg.SettleSeconds)*time.Second)
	}
	return Reset(context.Background(), ExecRunner{}, app)
}

func runCommand(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	repo, tag, observerName := ".", "manual", "v0"
	repeat := 1
	seed := int64(1)
	fs.StringVar(&repo, "repo", repo, "repository root")
	fs.StringVar(&tag, "tag", tag, "evaluation tag")
	fs.StringVar(&observerName, "observer", observerName, "v0 or iacon")
	fs.IntVar(&repeat, "repeat", repeat, "runs per scenario")
	fs.Int64Var(&seed, "seed", seed, "scenario seed")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := LoadConfig(repo)
	if err != nil {
		return err
	}
	app := ComposeCmd{Project: cfg.AppProject, Files: []string{filepath.Join(cfg.WorkshopDir, cfg.AppComposeFile)}}
	var observer AgentObserver
	switch observerName {
	case "v0":
		observer = V0Observer{Cmd: ComposeCmd{Project: cfg.AgentProject, Files: []string{filepath.Join(cfg.WorkshopDir, cfg.AgentComposeFile)}}, Runner: ExecRunner{}, Service: "agent"}
	case "iacon":
		observer = IaconObserver{Cmd: ComposeCmd{Project: cfg.IaconProject, Files: []string{"docker-compose.yml"}}, Runner: ExecRunner{}, Service: "iacon"}
	default:
		return fmt.Errorf("unknown observer %q", observerName)
	}
	scenarios := Scenarios(app)
	for id, scenario := range scenarios {
		for i := 0; i < repeat; i++ {
			runCfg := RunConfig{
				RepoDir: repo, Tag: tag, Agent: observerName, RunIndex: i, Seed: seed + int64(i),
				HoldoutHash: cfg.HoldoutHash, Steady: time.Duration(cfg.SteadySeconds) * time.Second,
				InjectJitter:   time.Duration(cfg.InjectJitterSeconds) * time.Second,
				ObserveTimeout: time.Duration(cfg.ObserveTimeoutSeconds) * time.Second,
				Settle:         time.Duration(cfg.SettleSeconds) * time.Second,
			}
			if _, err := RunScenario(context.Background(), scenario, ExecRunner{}, observer, runCfg); err != nil {
				return fmt.Errorf("%s run %d: %w", id, i, err)
			}
		}
	}
	return nil
}

func reportCommand(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	repo, tag := ".", "manual"
	fs.StringVar(&repo, "repo", repo, "repository root")
	fs.StringVar(&tag, "tag", tag, "evaluation tag")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := filepath.Join(repo, "eval", "runs", tag)
	var runs []RunResult
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil || info.IsDir() || filepath.Base(path) != "run.json" {
			return walkErr
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var run RunResult
		if err := json.Unmarshal(data, &run); err != nil {
			return err
		}
		runs = append(runs, run)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := HoldoutHashesMatch(runs); err != nil {
		return err
	}
	report := BuildReport(runs)
	fmt.Print(strings.TrimSpace(report.Markdown) + "\n")
	return os.WriteFile(filepath.Join(root, "report.md"), []byte(report.Markdown), 0o644)
}
