# CHANGELOG

All entries must be sanitized for a public repository.

## Unreleased

### Source / build
- Imported a sanitized public Go/Win32 source reconstruction for the Windows portable application.
- Added Excel-derived core business model and regression tests.
- Added Windows x64 GitHub Actions build artifact workflow.
- Added tagged/manual GitHub Release workflow.
- Stable semantic-version tags create normal releases; hyphenated test versions create prereleases.
- Added public binary scan to reject embedded credential-like data and non-approved hard-coded URL hosts.

### Security / runtime
- Production credential values, internal endpoint bindings, company datasets, and raw diagnostics remain excluded from the public repository.
- Added documented private runtime-profile direction for local DPAPI-protected endpoint provisioning.
- cURL session input remains local and DPAPI-protected.

### Update
- Added non-fatal GitHub latest-release check foundation.
- Automatic staged install/SHA256/rollback/manual-offline update remains pending.

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
