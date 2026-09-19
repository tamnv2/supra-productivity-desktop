# NEXT ACTION

Updated: 2026-09-19

## Current baseline

The owner supplied `SUPRA_PRODUCTIVITY_V1.3_TEST_FULL_SOURCE.zip`, reconstructed from the stable V1.3 executable. That recovered behavior is now the authority for this rebuild.

The previous GitHub candidates `v1.3.1-test.1` and `v1.3.1-test.2` are not the baseline for business behavior.

## Rebuild included

- recovered V1.3 business defaults and formulas;
- V1.3 shift classification and manual-shift priority;
- per-DO SKU deduction using `ceil(SKU/10)`;
- payroll duration kept in minutes and productivity speed calculated from minutes;
- recovered active-picking 31-column shape;
- stable-style left-navigation desktop shell;
- asynchronous/non-blocking diagnostics retained;
- private endpoint bindings remain local via encrypted runtime profile;
- GitHub updater supports both stable and test/prerelease channels;
- updater still verifies SHA256, creates a backup, replaces after exit and restores the backup if replacement itself fails.

## Immediate next action

1. Run GitHub PR CI for candidate source `v1.3.2-test.1`.
2. Require PASS for project/public guards, Go tests, vet, Windows x64 build and binary scan.
3. Merge only after CI passes.
4. Publish `v1.3.2-test.1`.
5. Publish identical-source `v1.3.2-test.2` as the updater target.
6. Owner runs test.1 on the target company laptop and clicks **CẬP NHẬT**. It must detect test.2, verify SHA256, replace/restart successfully and show the new version.
7. Only after that updater/runtime pass should further feature/UI changes resume.

Stable release remains blocked until explicit owner acceptance.
