package main

import (
	"strings"

	"iacon/internal/causes"
)

type Grading struct {
	Correct        bool
	Acceptable     bool
	Confident      bool
	Abstained      bool
	ConfidentWrong bool
}

func Grade(run RunResult, scenario Scenario) Grading {
	confident := run.Confidence == "high" || run.Confidence == "medium"
	abstained := run.Cause == "" || run.Cause == causes.Unknown || run.Confidence == "low"
	correct := run.Cause != "" && run.Cause == scenario.TrueCause
	acceptable := correct
	if scenario.Kind == "holdout" {
		acceptable = false
		for _, cause := range scenario.Acceptable {
			if run.Cause == cause {
				acceptable = true
				break
			}
		}
	}
	return Grading{
		Correct: correct, Acceptable: acceptable, Confident: confident,
		Abstained:      abstained,
		ConfidentWrong: confident && !acceptable && run.Cause != "" && run.Cause != causes.Unknown,
	}
}

func GradeV0(reportText string) (string, bool) {
	text := strings.ToLower(reportText)
	keywords := []struct {
		cause string
		terms []string
	}{
		{causes.RedisUnavailable, []string{"redis", "connection refused"}},
		{causes.PgBouncerDown, []string{"pgbouncer"}},
		{causes.PostgresDown, []string{"postgres", "postgresql"}},
		{causes.DatabaseDiskFull, []string{"disk full", "no space", "53100"}},
		{causes.PoolSaturated, []string{"pool", "saturated", "waiting"}},
		{causes.AppDown, []string{"app down", "application unavailable", "connection refused"}},
		{causes.AppInternalErrors, []string{"internal error", "http 500", "status 500"}},
		{causes.MonitoringDegraded, []string{"prometheus", "monitoring"}},
	}
	for _, keyword := range keywords {
		for _, term := range keyword.terms {
			if strings.Contains(text, term) {
				return keyword.cause, true
			}
		}
	}
	return causes.Unknown, false
}
