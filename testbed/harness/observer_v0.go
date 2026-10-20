package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type V0Observer struct {
	Cmd     ComposeCmd
	Runner  Runner
	Service string
}

func (o V0Observer) read(ctx context.Context, path string) ([]byte, error) {
	service := o.Service
	if service == "" {
		service = "agent"
	}
	out, err := o.Cmd.Exec(ctx, o.Runner, service, "cat", path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return []byte(out), nil
}

func (o V0Observer) Current(ctx context.Context) (Observation, error) {
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
		ID       string    `json:"id"`
		OpenedAt time.Time `json:"openedAt"`
	}
	if err := json.Unmarshal(envelope.Incident, &incident); err != nil {
		return Observation{}, fmt.Errorf("decode incident: %w", err)
	}
	return Observation{
		IncidentID: incident.ID,
		Open:       true,
		OpenedAt:   incident.OpenedAt,
	}, nil
}

func (o V0Observer) Report(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("incident id is empty")
	}
	data, err := o.read(ctx, "/data/incidents/"+id+".report.json")
	return string(data), err
}
