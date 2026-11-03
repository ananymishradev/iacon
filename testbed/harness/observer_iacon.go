package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type IaconObserver struct {
	Cmd     ComposeCmd
	Runner  Runner
	Service string
}

func (o IaconObserver) read(ctx context.Context, path string) ([]byte, error) {
	service := o.Service
	if service == "" {
		service = "iacon"
	}
	out, err := o.Cmd.Exec(ctx, o.Runner, service, "cat", path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return []byte(out), nil
}

func parseTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return &t
}

func (o IaconObserver) Current(ctx context.Context) (Observation, error) {
	state, err := o.read(ctx, "/data/state.json")
	if err != nil {
		return Observation{}, err
	}
	var envelope struct {
		Incident json.RawMessage `json:"incident"`
	}
	if err := json.Unmarshal(state, &envelope); err != nil {
		return Observation{}, fmt.Errorf("decode state: %w", err)
	}
	if len(envelope.Incident) == 0 || string(envelope.Incident) == "null" {
		return Observation{}, nil
	}
	var incident struct {
		ID            string         `json:"id"`
		OpenedAt      time.Time      `json:"opened_at"`
		OpenedAtCamel time.Time      `json:"openedAt"`
		ProvisionalAt string         `json:"provisional_at"`
		ConfirmedAt   string         `json:"confirmed_at"`
		NarratedAt    string         `json:"narrated_at"`
		ResolvedAt    string         `json:"resolved_at"`
		ReopenCount   int            `json:"reopen_count"`
		Verdict       detectSnapshot `json:"verdict"`
	}
	if err := json.Unmarshal(envelope.Incident, &incident); err != nil {
		return Observation{}, fmt.Errorf("decode incident: %w", err)
	}
	opened := incident.OpenedAt
	if opened.IsZero() {
		opened = incident.OpenedAtCamel
	}
	return Observation{
		IncidentID: incident.ID, Open: true, OpenedAt: opened,
		ProvisionalAt: parseTime(incident.ProvisionalAt), ConfirmedAt: parseTime(incident.ConfirmedAt),
		NarratedAt: parseTime(incident.NarratedAt), ResolvedAt: parseTime(incident.ResolvedAt),
		ReopenCount: incident.ReopenCount, Cause: incident.Verdict.Cause,
		Confidence: incident.Verdict.Confidence,
	}, nil
}

type detectSnapshot struct {
	Cause      string `json:"cause"`
	Confidence string `json:"confidence"`
}

func (o IaconObserver) Report(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("incident id is empty")
	}
	data, err := o.read(ctx, "/data/incidents/"+id+".report.json")
	return string(data), err
}
