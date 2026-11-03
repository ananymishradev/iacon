package telemetry

import (
	"context"
	"strings"
	"testing"
)

func TestMetricsAndSpans(t *testing.T) {
	m := NewMetrics()
	m.Inc("iacon_rule_matches_total", 1)
	if !strings.Contains(m.String(), "iacon_rule_matches_total") {
		t.Fatal("metric missing")
	}
	tracer := NewTracer()
	span, end := tracer.Start(context.Background(), "agent.detect", nil)
	span.Attributes["source"] = "app_ready"
	end()
	if len(tracer.Spans) != 1 || tracer.Spans[0].Name != "agent.detect" {
		t.Fatalf("spans = %#v", tracer.Spans)
	}
}
