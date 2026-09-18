# AGENTS.md — Canonical AI working contract

This repository is the source of truth for **SUPRA PRODUCTIVITY DESKTOP**.

## Mandatory bootstrap — minimal reads

At the beginning of a new chat/session/task, read only:
1. `ops/project-state.json`
2. `NEXT_ACTION.md`

Then read task-specific context:
- Business/UI behavior change → `docs/OWNER_DECISIONS.md` + `docs/STABLE_INVARIANTS.md`
- Bug/regression → `docs/KNOWN_ISSUES.md` + `docs/REGRESSION_GUARD_POLICY.md`
- Build/release/update → `docs/BUILD_RELEASE_PLAN.md` + `ops/resource-registry.json`
- Broad handover/history only when needed → `docs/HANDOVER.md` + `CHANGELOG.md`

Do **not** read the whole repository by default.

## Precedence

1. Owner's newest explicit instruction in the current conversation.
2. `docs/OWNER_DECISIONS.md`
3. `docs/STABLE_INVARIANTS.md`
4. `ops/project-state.json` and `NEXT_ACTION.md`
5. `docs/HANDOVER.md` / `CHANGELOG.md`
6. Model memory / old chat summaries.

Never override a newer owner decision using stale AI memory.

## Required end-of-task update

After any substantive implementation, debug session, accepted decision, release, or test result:
- update `ops/project-state.json`;
- update `NEXT_ACTION.md`;
- append a concise sanitized entry to `ops/session-ledger.jsonl`;
- update `CHANGELOG.md` when code/behavior changed;
- update `docs/OWNER_DECISIONS.md` only for actual owner decisions;
- update `docs/KNOWN_ISSUES.md` when an issue is discovered/resolved.

## Public repository rule

Never commit:
- real authorization/token/cookie values;
- APISID/USID or signature values;
- passwords;
- private keys;
- internal/private endpoint URLs unless the owner explicitly approves publication;
- raw operational logs/diagnostic ZIPs;
- raw company data or unsanitized Excel workbooks;
- personally identifying operational datasets.

Store only sanitized summaries and placeholders.

## Engineering rule

Do not create duplicate parallel implementations when a canonical implementation exists. Before adding a replacement workflow/module/state file, inspect the current canonical files and update them instead.
