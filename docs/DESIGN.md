# Iacon Design

## Locked decisions

1. Iacon is a dedicated repository at `~/projects/iacon` (this checkout uses
   `/home/anany/Main Projects/iacon`), with its own Compose project and history.
2. Ship 1 watches the workshop stack only. Collectors read targets from
   configuration and sit behind interfaces.
3. Portfolio value comes first: every phase produces a number, table, or demo
   clip for the README.
4. Narration uses a direct HTTP client behind a `Narrator` interface, with one
   real provider and a fake.
5. Ship 1 includes the spikes, harness, v0 baseline, agent core, validated
   narration, hardening, evaluation, and README. Ship 2 contains slow-burn,
   compound-fault, disagreement, injection-corpus, and model-only work.

## Global constraints

- The agent listens on `127.0.0.1:8090`, joins the external `workshop-net`, and
  uses project name `iacon`.
- `/live` is always 200 with `{"status":"alive","version":"iacon-ship1"}`.
  `/ready` and `/health` expose `version`, `ready`, `lastPoll`, and `incident`,
  returning 200 only after a recent monitor tick and while not stopping.
- State is stored under `/data`, using atomic writes and v0-compatible state
  and incident paths.
- The container is single-replica and hardened with dropped capabilities,
  `no-new-privileges`, 512 MiB memory, 0.5 CPU, and no Docker socket.
- Samples are tri-state: OK, BAD, or UNKNOWN. UNKNOWN is never healthy; it
  blocks primary-source resolution but advisory UNKNOWN is only annotated.
- Evidence collection is parallel, deadline-bound, partial-result tolerant,
  and uses stable `E1..En` IDs.
- Canary writes use only the agent-owned `iacon_canary` schema and every
  PgBouncer DSN uses `default_query_exec_mode=simple_protocol`.
- Detection thresholds are per-source constants and are tuned only with the
  harness; tuning is recorded in `docs/DECISIONS.md`.
- Cause slugs are defined once and reused by rules, narration, harness, and
  grading.
- Development and hold-out scenarios remain separate. Hold-out expectations
  are registered before evaluation and post-hoc changes are recorded.
- Model prompts contain only allowlisted, redacted, size-capped evidence.
  Invalid model output is retried once and never blocks incident reporting.
- Workshop credentials are used for Ship 1; least privilege is documented as
  future hardening, not claimed as implemented.
