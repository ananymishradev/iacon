package detect

import (
	"fmt"
	"time"
)

type Verdict struct {
	Cause           string   `json:"cause"`
	Confidence      string   `json:"confidence"`
	Kind            string   `json:"kind"`
	MatchedRules    []string `json:"matched_rules,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
	Secondary       []string `json:"secondary,omitempty"`
	MissingEvidence []string `json:"missing_evidence,omitempty"`
}

type Incident struct {
	ID                  string    `json:"id"`
	OpenedAt            time.Time `json:"opened_at"`
	ProvisionalAt       time.Time `json:"provisional_at,omitempty"`
	ConfirmedAt         time.Time `json:"confirmed_at,omitempty"`
	NarratedAt          time.Time `json:"narrated_at,omitempty"`
	ResolvedAt          time.Time `json:"resolved_at,omitempty"`
	ReopenCount         int       `json:"reopen_count"`
	Verdict             Verdict   `json:"verdict"`
	ResolvedWithUnknown []string  `json:"resolved_with_unknown,omitempty"`
	Pending             []string  `json:"pending,omitempty"`
}

type Machine struct {
	Incident *Incident
	lastID   int64
}

func (m *Machine) Apply(now time.Time, event Event) (*Incident, error) {
	switch event {
	case EventOpen:
		if m.Incident != nil && m.Incident.ResolvedAt.IsZero() {
			return m.Incident, nil
		}
		if m.Incident != nil && !m.Incident.ResolvedAt.IsZero() &&
			now.Sub(m.Incident.ResolvedAt) <= MergeWindow {
			m.Incident.ReopenCount++
			m.Incident.ResolvedAt = time.Time{}
			return m.Incident, nil
		}
		m.lastID++
		m.Incident = &Incident{ID: fmt.Sprintf("incident-%d", m.lastID), OpenedAt: now}
	case EventResolve:
		if m.Incident != nil && m.Incident.ResolvedAt.IsZero() {
			m.Incident.ResolvedAt = now
		}
	}
	return m.Incident, nil
}
