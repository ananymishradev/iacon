package telemetry

import (
	"context"
	"sync"
	"time"
)

type Span struct {
	Name       string
	Attributes map[string]string
	Start      time.Time
	End        time.Time
	Parent     *Span
}

type Tracer struct {
	mu    sync.Mutex
	Spans []Span
}

func NewTracer() *Tracer { return &Tracer{} }

func (t *Tracer) Start(_ context.Context, name string, parent *Span) (*Span, func()) {
	span := &Span{Name: name, Parent: parent, Start: time.Now(), Attributes: map[string]string{
		"service.name": "iacon", "service.version": "iacon-ship1",
	}}
	return span, func() {
		span.End = time.Now()
		t.mu.Lock()
		t.Spans = append(t.Spans, *span)
		t.mu.Unlock()
	}
}
