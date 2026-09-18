# PUBLIC REPOSITORY SECURITY POLICY

This repository is public.

## Never commit

- real token / Authorization values;
- real cookies;
- APISID / USID values;
- x-signature / nonce values;
- passwords;
- private keys or certificates containing private keys;
- internal/private service URLs unless explicitly approved for publication;
- raw diagnostic ZIPs/log files;
- raw company operational spreadsheets/data;
- screenshots containing credentials or sensitive operational details;
- local SQLite/cache databases.

## Allowed

- placeholder names such as `<TOKEN>`, `<COOKIE>`;
- sanitized logs with values redacted;
- synthetic test fixtures;
- public GitHub repository/release URLs;
- structural header names without values;
- sanitized summaries of bugs/performance findings.

## Before every release/commit

Run:
- `python scripts/validate_project_state.py`
- `python scripts/validate_public_repo.py`

The GitHub Actions state guard runs the same checks.

## Diagnostic handling

Raw diagnostic files should be provided privately to the active debugging session only.

After analysis, store only a short sanitized finding in:
- `docs/KNOWN_ISSUES.md`;
- `ops/session-ledger.jsonl`;
- `CHANGELOG.md` if behavior changed.
