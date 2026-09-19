# KNOWN ISSUES

Updated: 2026-09-19

## Open / verification required

### Public source is a sanitized reconstruction
- The earlier private V1.3 test source is not treated as a public canonical source.
- A sanitized Go/Win32 source tree has now been reconstructed and committed.
- It intentionally excludes internal endpoint bindings, real company snapshots, credentials, and raw diagnostics.
- **Status:** local core tests/cross-build passed; owner validation of the GitHub-built artifact is still required.

### Private live-data bindings
- Public source now imports and DPAPI-protects a local runtime profile without publishing production endpoint templates.
- Full live request execution/response parsing is not yet wired to the profile.
- **Status:** provisioning complete; live binding open.

### GitHub Actions verification
- Windows build/release workflows are committed.
- Connector-created commits/PR events did not expose a workflow run for verification in this session.
- **Status:** trigger through a normal external push or manual Actions dispatch and verify artifact/logs.

### GitHub updater
- Public source now downloads the stable release EXE/checksum, verifies SHA256, backs up the current EXE, stages replacement after exit, and restarts.
- Replacement failure attempts to restore the backup.
- Manual/offline package selection and startup-health rollback remain incomplete.
- **Status:** partially implemented.

### UI / functional parity
- Full process/system CPU telemetry still needs parity work.
- All PICK numeric target/quota fields are not yet exposed as editable controls in the reconstructed public UI.
- User/PDA public build contains no bundled company snapshot by design; live/local binding is pending.
- **Status:** open.

### V1.3 PICK flicker/event storm
- Earlier diagnostics showed a large repeated skip-20-minute checkbox event storm.
- V1.3 changed the interaction to avoid the recursive render mechanism.
- **Status:** implementation exists; target-laptop owner verification required.

## Resolved / superseded

- V1.1 incomplete live-sync behavior → superseded.
- V1.2 PICK recursive checkbox/render mechanism → superseded by the V1.3 interaction change, pending verification.
- GitHub continuity/source-state bootstrap → complete.
- Public source import blocker → resolved; source is now in the repository.
