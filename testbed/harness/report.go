package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type ReportRow struct {
	Scenario       string
	Detected       string
	OpenMedianP95  string
	Accuracy       string
	ConfidentWrong string
	Abstain        string
	Flaps          int
}

type ReportTables struct {
	Dev      []ReportRow
	Holdout  []ReportRow
	Soak     string
	Markdown string
}

func BuildReport(runs []RunResult) ReportTables {
	grouped := map[string][]RunResult{}
	for _, run := range runs {
		grouped[run.Scenario] = append(grouped[run.Scenario], run)
	}
	var report ReportTables
	for scenario, values := range grouped {
		sort.Slice(values, func(i, j int) bool { return values[i].RunIndex < values[j].RunIndex })
		row := summarizeRuns(scenario, values)
		if values[0].Kind == "holdout" {
			report.Holdout = append(report.Holdout, row)
		} else {
			report.Dev = append(report.Dev, row)
		}
	}
	sort.Slice(report.Dev, func(i, j int) bool { return report.Dev[i].Scenario < report.Dev[j].Scenario })
	sort.Slice(report.Holdout, func(i, j int) bool { return report.Holdout[i].Scenario < report.Holdout[j].Scenario })
	report.Markdown = renderReport(report)
	return report
}

func summarizeRuns(scenario string, runs []RunResult) ReportRow {
	detected := 0
	var opens []int64
	correct, acceptable, wrong, abstain, flaps := 0, 0, 0, 0, 0
	for _, run := range runs {
		if run.TOpenMS != nil {
			detected++
			opens = append(opens, *run.TOpenMS)
		}
		flaps += run.ReopenCount
		scenarioConfig := Scenario{ID: run.Scenario, Kind: run.Kind, TrueCause: run.TrueCause}
		grade := Grade(run, scenarioConfig)
		if grade.Correct {
			correct++
		}
		if grade.Acceptable {
			acceptable++
		}
		if grade.ConfidentWrong {
			wrong++
		}
		if grade.Abstained {
			abstain++
		}
	}
	accuracy := correct
	if runs[0].Kind == "holdout" {
		accuracy = acceptable
	}
	return ReportRow{
		Scenario: scenario, Detected: fmt.Sprintf("%d/%d", detected, len(runs)),
		OpenMedianP95: formatStats(opens), Accuracy: fmt.Sprintf("%d/%d", accuracy, len(runs)),
		ConfidentWrong: fmt.Sprintf("%d/%d", wrong, len(runs)),
		Abstain:        fmt.Sprintf("%d/%d", abstain, len(runs)), Flaps: flaps,
	}
}

func formatStats(values []int64) string {
	if len(values) == 0 {
		return "-"
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
	p95 := percentile(values, 0.95)
	return fmt.Sprintf("%d / %d ms", median, p95)
}

func percentile(values []int64, p float64) int64 {
	if len(values) == 1 {
		return values[0]
	}
	index := int(math.Ceil(p*float64(len(values))) - 1)
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func renderReport(report ReportTables) string {
	var b strings.Builder
	b.WriteString("| scenario | detected k/n | median / p95 t_open_ms | accuracy or acceptable k/n | confident_wrong k/n | abstain k/n | flaps |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|\n")
	writeRows := func(rows []ReportRow) {
		for _, row := range rows {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %d |\n",
				row.Scenario, row.Detected, row.OpenMedianP95, row.Accuracy,
				row.ConfidentWrong, row.Abstain, row.Flaps)
		}
	}
	b.WriteString("\n## Dev\n")
	writeRows(report.Dev)
	b.WriteString("\n## Hold-out\n")
	writeRows(report.Holdout)
	if report.Soak != "" {
		b.WriteString("\n## Soak\n")
		b.WriteString(report.Soak)
		b.WriteByte('\n')
	}
	return b.String()
}

func HoldoutHashesMatch(runs []RunResult) error {
	var hash string
	for _, run := range runs {
		if run.Kind != "holdout" {
			continue
		}
		if hash == "" {
			hash = runHoldoutHash(run)
			continue
		}
		if current := runHoldoutHash(run); current != hash {
			return fmt.Errorf("hold-out hash mismatch: %q != %q", current, hash)
		}
	}
	return nil
}

func runHoldoutHash(run RunResult) string {
	return run.HoldoutHash
}
