# NEXT ACTION

Updated: 2026-09-19

## Current canonical candidate

Recovered stable V1.3 behavior is now the rebuild baseline.

Published GitHub prereleases:
- `v1.3.2-test.1` — updater source build.
- `v1.3.2-test.2` — identical-source updater target build.

PR verification before merge:
- Project State Guard: run `35419657684` — PASS.
- Windows Portable build: run `35419657703` — PASS.
- This includes public-repo guard, unit tests, vet, Windows x64 cross-build and binary scan.

## Owner test now

1. Download/run **v1.3.2-test.1** on the target laptop.
2. Verify the V1.3 baseline opens and core Pick/Pack/Phân ca behavior is usable.
3. Click **CẬP NHẬT**.
4. It must detect **v1.3.2-test.2**.
5. It must download `SupraProductivity.exe` + `SHA256SUMS.txt`, verify SHA256, back up the current EXE, replace after exit and restart.
6. After restart, the window/version must show **v1.3.2-test.2**.
7. If this passes, only then resume the owner's next UI/feature changes.

Stable release remains blocked until explicit owner acceptance.
