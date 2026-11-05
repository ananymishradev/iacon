# Decisions

## Spike results

Record workshop facts and the results of Spikes A-D here.

Automated implementation is complete through the agent, harness, telemetry and
Compose code paths. The live-stack values and measurements below still require
the workshop repository and are intentionally not guessed.

## Verification checklist

- [ ] A `-f` override preserves the workshop Compose project name.
- [ ] The PgBouncer admin console works for the workshop user.
- [ ] The SQLSTATE produced by a full disk is confirmed.
- [ ] Workshop service names, credentials, and ports match `.env.example`.

## Surprises

## Post-hoc changes

No post-hoc evaluation changes have been made.
