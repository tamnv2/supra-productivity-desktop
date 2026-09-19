# NEXT ACTION

Updated: 2026-09-19

## Immediate next action

**Bind the locally encrypted runtime profile to full live operational synchronization, then verify the GitHub-built EXE.**

Order:

1. Implement Active-Picking and Payroll/Productivity live request execution from the DPAPI-protected runtime profile.
2. Parse/normalize live responses and feed the existing Pick/Pack/Shift/User-PDA business engines without embedding private endpoints in source.
3. Complete the remaining updater gap: manual/offline package selection and stronger startup-health rollback.
4. Trigger `Build Windows Portable` through a normal external push or manual Actions dispatch and verify the artifact.
5. Owner tests the GitHub-built EXE on PDA and restricted Office network.
6. Analyze diagnostics privately; commit only sanitized findings.
7. Update canonical state after every verified result.

## Already completed

- sanitized public source imported;
- public-repo/state guards;
- Windows x64 build workflow;
- versioned Release workflow;
- local DPAPI runtime-profile provisioning;
- staged GitHub updater download + SHA256 verification + backup + replace/restart;
- local core tests, Go vet, Windows cross-build, and binary public scan passed.

## Acceptance focus

- no production credential/private endpoint/raw operational data in public GitHub;
- no UI freeze during sync/update;
- core operation independent of GitHub availability;
- full live parity restored only through local private profile;
- PICK event-storm/flicker does not regress;
- owner explicitly accepts before stable release.
