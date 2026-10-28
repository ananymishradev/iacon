package main

import (
	"strings"
	"testing"
)

func TestBuildReport(t *testing.T) {
	open8, open10 := int64(8000), int64(10000)
	report := BuildReport([]RunResult{
		{Scenario: "S1", Kind: "dev", RunIndex: 1, TrueCause: "redis_unavailable", TOpenMS: &open8, Cause: "redis_unavailable", Confidence: "high"},
		{Scenario: "S1", Kind: "dev", RunIndex: 2, TrueCause: "redis_unavailable", TOpenMS: &open10, Cause: "redis_unavailable", Confidence: "high"},
		{Scenario: "S1", Kind: "dev", RunIndex: 3, TrueCause: "redis_unavailable"},
	})
	if len(report.Dev) != 1 || report.Dev[0].Detected != "2/3" {
		t.Fatalf("report = %#v", report)
	}
	if !strings.Contains(report.Markdown, "9000 / 10000 ms") {
		t.Fatalf("markdown = %s", report.Markdown)
	}
}

func TestHoldoutHashesMatch(t *testing.T) {
	runs := []RunResult{
		{Kind: "holdout", HoldoutHash: "a"},
		{Kind: "holdout", HoldoutHash: "b"},
	}
	if err := HoldoutHashesMatch(runs); err == nil {
		t.Fatal("expected mismatch")
	}
}
