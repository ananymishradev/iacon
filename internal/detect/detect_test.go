package detect

import (
	"testing"
	"time"

	"iacon/internal/signals"
)

func push(w *signals.Window, state signals.SampleState, n int, definitive bool) {
	for i := 0; i < n; i++ {
		w.Push(signals.Sample{At: time.Now().Add(time.Duration(i) * time.Second), State: state, Definitive: definitive})
	}
}

func TestDetectorThresholds(t *testing.T) {
	app, read, canary, prom := new(signals.Window), new(signals.Window), new(signals.Window), new(signals.Window)
	push(app, signals.StateBAD, 5, false)
	if got := (Detector{}).Evaluate(time.Now(), map[string]*signals.Window{"app_ready": app, "read": read, "canary": canary, "prom": prom}); got != EventOpen {
		t.Fatalf("app open = %v", got)
	}
	_ = read
	_ = canary
}
