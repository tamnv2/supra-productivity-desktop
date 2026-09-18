# NEXT ACTION

Updated: 2026-09-19

## Immediate next action

**Import the canonical V1.3 source into this repository after a public-repo security scrub.**

Then, in order:

1. Add a Windows x64 GitHub Actions build that reproduces the portable EXE.
2. Add tagged GitHub Release packaging: EXE + ZIP + SHA256 + changelog.
3. Add application update logic using GitHub Releases when reachable, with a manual/offline fallback for restricted Office network.
4. Continue owner testing of V1.3 on PDA and Office.
5. Convert each received diagnostic into a **sanitized issue summary**, never commit raw diagnostic ZIPs.
6. Fix only verified issues and update `ops/project-state.json`, `docs/KNOWN_ISSUES.md`, and `CHANGELOG.md`.

## Acceptance focus for V1.3

- PICK no longer flickers/blanks or creates an event storm.
- UI remains responsive during sync.
- Manual shift can be selected and persisted.
- Double-click details match the intended Excel workflow.
- Active-picking status filter applies immediately.
- Logs contain enough technical detail for diagnosis without credentials.
- CPU/RAM behavior remains stable during extended use.
