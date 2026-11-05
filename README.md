# Iacon

Iacon is a verified incident responder for the workshop stack. It provides the
v0-compatible `/live`, `/ready`, and `/health` contract on `127.0.0.1:8090`,
collects tri-state evidence, classifies incidents with topology-specific rules,
and optionally produces a validated narration.

```text
workshop services -> 1 Hz /ready + 5 s read/canary/Prometheus windows
                  -> detector -> parallel evidence bundle -> rule verdict
                  -> optional narrator -> durable incident and report files
```

The chaos harness runs the same scenarios against v0 and Iacon without shell
interpolation. It stores run data under `eval/runs/` and produces separate dev
and hold-out reports.

## Status

The repository contains the Ship 1 implementation and focused unit tests.
Live workshop spikes, fault runs, the two-hour soak, and the demo recording
require the workshop stack and remain operational validation steps.

## Limitations

Faults are synthetic and results are measured on one laptop. Evaluation counts
are reported as `k/n` (N is 10, or 5 for some hold-outs), not population rates.
Rules are specific to this topology; v0 sees HTTP statuses while Iacon uses
additional evidence. Zero false incidents in a two-hour soak supports only
"under about 1.5 per hour" at 95% confidence. S4 and S5 availability depend on
the results of Spikes C and B.
