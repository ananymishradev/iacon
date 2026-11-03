package narrate

import (
	"context"

	"iacon/internal/detect"
	"iacon/internal/evidence"
)

type Usage struct {
	InTokens  int64 `json:"in_tokens"`
	OutTokens int64 `json:"out_tokens"`
}

type Input struct {
	Bundle      evidence.Evidence
	RuleVerdict *detect.Verdict
	IncidentID  string
}

type Output struct {
	Summary       string   `json:"summary"`
	Cause         string   `json:"cause"`
	CitedEvidence []string `json:"citedEvidence"`
	Confidence    string   `json:"confidence"`
	NextChecks    []string `json:"nextChecks"`
	Usage         Usage    `json:"usage"`
}

type Narrator interface {
	Narrate(context.Context, Input) (Output, error)
}
