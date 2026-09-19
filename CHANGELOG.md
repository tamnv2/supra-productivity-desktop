# CHANGELOG

All entries must be sanitized for a public repository.

## Unreleased

### Release automation
- Added file-driven owner-test prerelease publishing through `release/test-version.txt`.
- Added file-driven accepted-stable publishing through `release/stable-version.txt`.
- Published GitHub prerelease `v1.3.1-test.1` from canonical source; candidate publish and main portable build both passed.


### GitHub build verification
- Repaired two source-reconstruction defects found by real PR CI: DPAPI credential load closure/unmarshal and latest-release URL construction.
- Refined public-source guard to avoid false positives on credential variable references while still rejecting quoted credential literals.
- Allowed only standard XML namespace hosts required by XLSX parsing in the binary URL allowlist.
- GitHub-built candidate verified by PR CI run 35412096189: guards, tests, vet, Windows x64 cross-build, binary scan, packaging and artifact upload all passed.
- Live payroll/productivity and active-picking synchronization is wired through the locally encrypted runtime profile with proxy/direct fallback.
- Stable release remains pending owner runtime acceptance.


### Source / build
- Imported a sanitized public Go/Win32 source reconstruction for the Windows portable application.
- Added Excel-derived core business model and regression tests.
- Added Windows x64 GitHub Actions build artifact workflow.
- Added tagged/manual GitHub Release workflow.
- Stable semantic-version tags create normal releases; hyphenated test versions create prereleases.
- Added public binary scan to reject embedded credential-like data and non-approved hard-coded URL hosts.

### Security / runtime
- Production credential values, internal endpoint bindings, company datasets, and raw diagnostics remain excluded from the public repository.
- Added local runtime-profile provisioning: plaintext profile beside the EXE is validated, DPAPI-encrypted for the current Windows user, and removed after successful import where possible.
- cURL session input remains local and DPAPI-protected.

### Update
- Added non-fatal GitHub latest-release check.
- Added staged stable-release download, SHA256 verification, current-EXE backup, replace-after-exit, restart, and replacement-failure restore.
- Manual/offline package selection and startup-health rollback remain pending.

### Repository / continuity
- GitHub is canonical for project state, owner decisions, next action, invariants, sanitized issues, and session continuity.

## v1.3-test — 2026-09-18

- Improved responsiveness and background processing.
- Changed PICK skip-20-minute interaction to avoid repeated checkbox event loops.
- Added manual shift selection.
- Added double-click operational detail behavior.
- Added richer diagnostics.
- Moved business controls from generic Settings to relevant screens.
- Limited Settings primarily to Dashboard credential/cURL handling.

Status: private test build; owner acceptance pending.

## 2026-09-19 — v1.3.1-test.2 runtime/UI correction
- Marked test.1 owner-rejected.
- Removed UI-thread dependency on runtime log filesystem writes.
- Runtime logs now use a bounded background queue and local AppData path.
- Diagnostic export moved off the UI thread; log viewer reads a bounded tail.
- Replaced left navigation rail with a bottom navigation row.
- Expanded operational content to near full-window width.
- Window now opens full work-area and is non-resizable/minimize-only in normal use.
- Refined shell styling to a restrained neutral palette with status-aware header text.

