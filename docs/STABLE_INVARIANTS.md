# STABLE INVARIANTS

Do not break these without an explicit new owner decision.

## Runtime

- Windows x64 portable application.
- Runs with standard user permissions.
- No Excel installation/runtime dependency for normal application operation.
- Application must remain usable when public Internet is unavailable but internal business services are reachable.
- Network work must not block the UI thread.

## Data safety

- Failed sync must not destroy the last valid local snapshot.
- New data replaces active data only after successful download/parse/validation.
- Credentials are never logged.
- Raw diagnostics and company data are never committed to the public repository.

## UI

- Operational content area has priority over navigation chrome.
- UI must remain compact, readable, and professional on weaker company laptops.
- Correct data types must be preserved for rendering and sorting.
- Business controls live with the relevant business screen.
- Settings is not a dumping ground for operational controls.

## Diagnostics

- Logging must be detailed enough to investigate performance, network, parsing, business logic, UI events, and exceptions.
- Sensitive values are redacted/omitted.
- Long-running builds must include periodic CPU/RAM/process telemetry sufficient to identify leaks or event storms.

## Project continuity

- GitHub canonical state overrides model memory.
- Every substantive task updates project state and next action.
- Owner acceptance is explicit; an untested implementation is not recorded as accepted.

## Report and synchronization

- Production/report synchronization is user-triggered; background telemetry timers must not trigger business-data downloads.
- Selected date range defaults to today; multi-day report ranges must not cause Pick/Pack/Phân ca targets to aggregate across days.
- Historical production cache is date-addressed and reusable. A sync must fetch only missing historical ranges, while the current day is refreshable on every manual sync.
- Native Recap calculations do not depend on Google Sheets availability.
