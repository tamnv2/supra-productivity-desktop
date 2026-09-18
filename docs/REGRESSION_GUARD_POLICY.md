# REGRESSION GUARD POLICY

## Before changing code

1. Read `ops/project-state.json` and `NEXT_ACTION.md`.
2. Read the specific owner decision/invariant for the area being changed.
3. Identify which accepted behavior could regress.
4. Do not reintroduce removed features R01/R02/R03/R04/R07/R08.

## Required regression checks

### UI responsiveness
- Startup shell renders without waiting for network.
- Sync does not make Windows report Not Responding.
- Repeated UI events do not trigger render/recalculate loops.

### PICK
- No flicker/blank loop.
- Typed sorting works.
- Filters and rules apply correctly.
- Skip-20-minute handling does not create recursive events.
- Double-click detail remains available.

### PACK
- No unsupported deduction control is reintroduced.
- Display-by-shift applies immediately.

### Shift
- Auto shift remains available.
- Manual override can be selected, saved, and takes precedence when expected.

### Network
- PDA path works.
- Restricted Office path works when internal services are reachable.
- GitHub/public Internet failure does not block core operation.

### Security
- Build contains no embedded production credentials.
- Logs contain no credentials.
- Repo guard passes.
- No raw diagnostic archive/company workbook is committed.

## Acceptance

Implementation complete ≠ owner accepted.

Record acceptance only when the owner explicitly reports the item as OK/accepted after testing.
