# NEXT ACTION

Updated: 2026-09-19

## Immediate next action

**Provision private live-data bindings locally without publishing them, then validate the GitHub-built executable.**

Order:

1. Implement the encrypted local runtime-profile import described in `docs/RUNTIME_PROFILE.md`.
2. Bind the sanitized public app to the locally provisioned Active-Picking and Payroll/Productivity request templates.
3. Complete GitHub updater installation: SHA256 validation, staged replacement, backup/rollback, and manual/offline package fallback.
4. Trigger the GitHub Actions Windows x64 build from an external/manual workflow event and verify the produced artifact.
5. Owner tests the GitHub-built EXE on PDA and restricted Office network.
6. Analyze diagnostics privately; commit only sanitized findings.
7. Update `ops/project-state.json`, `docs/KNOWN_ISSUES.md`, and `CHANGELOG.md` after each verified result.

## Current acceptance focus

- Public repository contains no production credentials/private endpoint URLs/raw operational data.
- GitHub can reproducibly build the portable Windows x64 executable.
- Core app remains usable if GitHub/public Internet is blocked.
- Local runtime profile restores full live-data capability without exposing internal infrastructure.
- PICK remains free of the prior recursive event/flicker behavior.
- Owner acceptance is required before any build is called stable.
