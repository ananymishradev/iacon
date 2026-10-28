package main

import (
	"testing"

	"iacon/internal/causes"
)

func TestGrade(t *testing.T) {
	run := RunResult{Kind: "holdout", Cause: causes.RedisUnavailable, Confidence: "high"}
	scenario := Scenario{Kind: "holdout", TrueCause: causes.PostgresDown, Acceptable: []string{causes.PostgresDown, causes.Unknown}}
	grade := Grade(run, scenario)
	if !grade.ConfidentWrong || grade.Abstained {
		t.Fatalf("grade = %#v", grade)
	}
	run.Cause, run.Confidence = causes.Unknown, "low"
	if grade := Grade(run, scenario); !grade.Abstained {
		t.Fatalf("unknown should abstain: %#v", grade)
	}
}

func TestGradeV0(t *testing.T) {
	cause, confident := GradeV0("Redis connection refused while checking the cache")
	if cause != causes.RedisUnavailable || !confident {
		t.Fatalf("got %q, %v", cause, confident)
	}
	if cause, confident := GradeV0(""); cause != causes.Unknown || confident {
		t.Fatalf("empty report = %q, %v", cause, confident)
	}
}
