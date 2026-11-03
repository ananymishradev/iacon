package telemetry

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Metrics struct {
	mu       sync.Mutex
	counters map[string]float64
	gauges   map[string]float64
}

func NewMetrics() *Metrics {
	return &Metrics{counters: map[string]float64{}, gauges: map[string]float64{}}
}

func (m *Metrics) Inc(name string, value float64) {
	m.mu.Lock()
	m.counters[name] += value
	m.mu.Unlock()
}

func (m *Metrics) Set(name string, value float64) {
	m.mu.Lock()
	m.gauges[name] = value
	m.mu.Unlock()
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		m.mu.Lock()
		defer m.mu.Unlock()
		for name, value := range m.counters {
			fmt.Fprintf(w, "%s %g\n", name, value)
		}
		for name, value := range m.gauges {
			fmt.Fprintf(w, "%s %g\n", name, value)
		}
	})
}

func (m *Metrics) String() string {
	var b strings.Builder
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, value := range m.counters {
		fmt.Fprintf(&b, "%s %g\n", name, value)
	}
	for name, value := range m.gauges {
		fmt.Fprintf(&b, "%s %g\n", name, value)
	}
	return b.String()
}
