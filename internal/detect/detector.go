package detect

import (
	"time"

	"iacon/internal/signals"
)

type Event int

const (
	EventNone Event = iota
	EventOpen
	EventResolve
)

type Detector struct{}

func (Detector) Evaluate(_ time.Time, windows map[string]*signals.Window) Event {
	app, read, canary, prom := windows["app_ready"], windows["read"], windows["canary"], windows["prom"]
	if app != nil {
		bad, total := app.BADOfLast(AppReadyOpenWindow)
		if (total < AppReadyOpenWindow && bad >= AppReadyOpenBad) || bad >= AppReadyOpenBad {
			return EventOpen
		}
	}
	if read != nil && read.ConsecutiveBAD() >= ReadOpenBad {
		return EventOpen
	}
	if canary != nil && canary.ConsecutiveBAD() >= CanaryOpenBad &&
		(canary.ConsecutiveBADDefinitive() || canary.ConsecutiveBAD() >= 3) {
		return EventOpen
	}
	if prom != nil && prom.LastState() == signals.StateBAD {
		return EventNone
	}
	if app == nil || read == nil || canary == nil {
		return EventNone
	}
	if !app.AllOKConsecutive(AppReadyResolveOK) || app.OKSpan() < AppReadyResolveAge ||
		!read.AllOKConsecutive(ReadResolveOK) || !canary.AllOKConsecutive(CanaryResolveOK) {
		return EventNone
	}
	return EventResolve
}
