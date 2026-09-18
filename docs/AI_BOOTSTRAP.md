# AI Bootstrap / Session Continuity

Goal: any new chat or AI worker can continue the project correctly without rereading the full history.

## Minimal bootstrap

Read:
1. `ops/project-state.json`
2. `NEXT_ACTION.md`

That is sufficient to answer: what is this project, where is it now, and what should happen next.

## Conditional reads

### Implementing or changing business/UI behavior
Read:
- `docs/OWNER_DECISIONS.md`
- `docs/STABLE_INVARIANTS.md`

### Debugging
Read:
- `docs/KNOWN_ISSUES.md`
- `docs/REGRESSION_GUARD_POLICY.md`
- only the relevant source files/log summaries.

### Build, release, updater, GitHub Actions
Read:
- `docs/BUILD_RELEASE_PLAN.md`
- `ops/resource-registry.json`

### Broad historical question
Only then read:
- `docs/HANDOVER.md`
- `CHANGELOG.md`
- `ops/session-ledger.jsonl`

## Information authority

Never reconstruct project state from model memory when canonical files exist.

If information conflicts:
1. current owner instruction wins;
2. newest dated owner decision wins;
3. stable invariants apply unless the owner explicitly changed them;
4. project-state controls current status;
5. old handover/history is context only.

## Smart update rule

Do not dump chat transcripts into GitHub. Store compact durable state:
- decision → OWNER_DECISIONS;
- invariant → STABLE_INVARIANTS;
- current status → project-state;
- immediate action → NEXT_ACTION;
- bug → KNOWN_ISSUES;
- code behavior change → CHANGELOG;
- session summary → session-ledger.

This keeps context small and high-signal.

## Security

Public repository:
- redact or omit credentials;
- do not upload raw logs;
- do not upload raw company spreadsheets;
- do not publish private internal URLs;
- use synthetic/sanitized fixtures only.

When analyzing a diagnostic, commit only findings such as:
`PICK render loop caused by repeated control event; fixed in v1.3`
—not the raw diagnostic archive.
