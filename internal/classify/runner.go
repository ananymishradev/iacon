package classify

import (
	"context"
	"time"

	"iacon/internal/detect"
	"iacon/internal/evidence"
)

type BundleFunc func(context.Context) evidence.Evidence
type PersistFunc func(detect.Incident, evidence.Evidence) error

type Runner struct {
	Collectors []evidence.Collector
	Config     Cfg
	Persist    PersistFunc
	Now        func() time.Time
}

func (r Runner) Run(ctx context.Context, incident *detect.Incident) error {
	if incident == nil {
		return nil
	}
	now := r.Now
	if now == nil {
		now = time.Now
	}
	ev := evidence.Run(ctx, r.Collectors)
	incident.Verdict = Classify(ev, r.Config)
	incident.ProvisionalAt = now()
	incident.Pending = []string{"confirm"}
	if r.Persist != nil {
		if err := r.Persist(*incident, ev); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) Confirm(ctx context.Context, incident *detect.Incident) error {
	if incident == nil {
		return nil
	}
	ev := evidence.Run(ctx, r.Collectors)
	incident.Verdict = Classify(ev, r.Config)
	incident.ConfirmedAt = time.Now()
	incident.Pending = nil
	if r.Persist != nil {
		return r.Persist(*incident, ev)
	}
	return nil
}
