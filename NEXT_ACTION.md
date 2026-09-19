# NEXT ACTION

Updated: 2026-09-19

## Current blocker

The public GitHub continuity/security bootstrap is complete.

The current **V1.3 TEST package is binary-only** in the active handoff artifacts:
- EXE
- test guide
- changelog
- SHA256

It does **not** contain the source tree used to build V1.3.

Therefore a reproducible GitHub Actions build/release is **not yet complete** and must not be represented as complete.

## Immediate next action

**Restore/create the canonical source tree that reproduces the current V1.3 behavior, then commit it after a public-repo security scrub.**

Required order:

1. Restore/recreate V1.3-equivalent source.
2. Run public-repo secret/sensitive-data validation.
3. Commit canonical source to `main`.
4. Add Windows x64 GitHub Actions build.
5. Verify GitHub-built EXE against local V1.3 behavior.
6. Add tagged GitHub Release: EXE + ZIP + SHA256 + release notes.
7. Add GitHub updater with non-fatal failure and manual/offline fallback for restricted Office network.
8. Continue owner testing on PDA + Office and record only sanitized diagnostic findings.

## Do not do

- Do not upload or reverse-publish real token/cookie/signature values.
- Do not commit raw diagnostic ZIPs or operational spreadsheets.
- Do not call GitHub build/release "done" before canonical source builds successfully in Actions.
