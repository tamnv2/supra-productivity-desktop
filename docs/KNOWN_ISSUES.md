# KNOWN ISSUES

Updated: 2026-09-19

## Open / verification required

### V1.3 PICK flicker/event storm
- Earlier diagnostic showed a very large number of repeated skip-20-minute checkbox events, causing repeated recalculation/rendering.
- V1.3 changed this interaction to remove the checkbox event-loop mechanism.
- **Status:** implemented, owner verification still required on the target company laptop.

### Excel parity
- Live Pick/Pack/shift behavior is being ported from the Excel reference.
- **Status:** partial/ongoing verification; do not claim 100% parity until owner tests approve it.

### GitHub build/release
- Repository has continuity/security scaffolding but current application source has not yet been imported.
- **Status:** blocked until sanitized V1.3 source is committed.

### GitHub updater
- Required direction: GitHub Releases when reachable, but Office network may block GitHub.
- Must have manual/offline update fallback.
- **Status:** planned, not implemented.

## Resolved/superseded

- V1.1 sync button did not fully update operational live tables → superseded by later implementation.
- V1.2 PICK rendering/event behavior → superseded by V1.3 change; verification pending.
