package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"iacon/internal/evidence"
)

type Report struct {
	IncidentID      string            `json:"incident_id"`
	Summary         string            `json:"summary"`
	Verdict         Verdict           `json:"verdict"`
	ProvisionalAt   string            `json:"provisional_at,omitempty"`
	ConfirmedAt     string            `json:"confirmed_at,omitempty"`
	Evidence        evidence.Evidence `json:"evidence"`
	MissingEvidence []string          `json:"missing_evidence,omitempty"`
	NextChecks      []string          `json:"next_checks,omitempty"`
}

func WriteReport(dir string, incident Incident, ev evidence.Evidence) error {
	report := Report{
		IncidentID: incident.ID, Summary: "Workshop service health incident",
		Verdict: incident.Verdict, Evidence: ev,
		MissingEvidence: incident.Verdict.MissingEvidence,
		ProvisionalAt:   formatTime(incident.ProvisionalAt),
		ConfirmedAt:     formatTime(incident.ConfirmedAt),
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, incident.ID+".report.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func formatTime(value interface {
	IsZero() bool
	String() string
}) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.String()
}
