package main

import (
	"context"
	"time"
)

type Observation struct {
	IncidentID    string
	Open          bool
	OpenedAt      time.Time
	ProvisionalAt *time.Time
	ConfirmedAt   *time.Time
	NarratedAt    *time.Time
	ResolvedAt    *time.Time
	ReopenCount   int
	Cause         string
	Confidence    string
	ReportJSON    string
}

type AgentObserver interface {
	Current(context.Context) (Observation, error)
	Report(context.Context, string) (string, error)
}
