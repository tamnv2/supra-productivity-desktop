# OWNER DECISIONS

Canonical record of explicit owner decisions. Newest explicit owner instruction overrides older entries.

## 2026-09-18 — Product scope

- Keep the Excel-derived Pick/Pack/shift operational logic unless explicitly removed.
- Remove: R01, R02, R03, R04, R07, R08.
- Portable Windows EXE, standard-user permission, no Excel runtime dependency.
- Prioritize speed, stability, low RAM/CPU, and professional compact UI.

## 2026-09-18 — Security and credentials

- Token/session information must remain user-editable.
- Preferred input: paste the same **Copy as cURL (bash)** text used with the Excel workflow, then the app extracts required fields.
- Sensitive values must be hidden by default.
- Credentials must not be written to logs.
- Runtime credential persistence should be protected for the current Windows user.

## 2026-09-18 — Diagnostics and storage

- Keep comprehensive logs for optimization and debugging, excluding sensitive information.
- User can choose the local data storage folder.
- Do not use an actively-open SQLite database directly on an unreliable network share.

## 2026-09-18 — Network behavior

- App must work on both PDA and restricted Office network when required internal services are reachable.
- Public Internet/GitHub being blocked must not be treated as total network failure.
- Core operational features must not depend on GitHub availability.

## 2026-09-18 — UI / workflow

- Content in operational tables should be left aligned while preserving correct data types.
- Table headers must sort correctly by type: text, numeric, percent, date/time, and tenure duration.
- Double-click should open operational detail comparable to the Excel workflow.
- Business controls belong on their relevant operational screen, not in a generic Settings page.
- Settings should be limited mainly to Dashboard credential/cURL handling.
- Select/checkbox operational filters should apply immediately where appropriate.
- Pack currently has no deduction control; keep only its required display/filter controls.
- Manual shift selection must work similarly to Excel.

## 2026-09-19 — GitHub continuity

- Repository: `tamnv2/supra-productivity-desktop`.
- Repository is public.
- Never publish sensitive information.
- GitHub becomes the canonical persistence layer for project state, decisions, next action, build/release history, and sanitized debugging findings.
- Any new AI session should continue from repository state instead of relying on AI memory.

## 2026-09-19 — Runtime/UI correction after test.1 rejection

- `v1.3.1-test.1` is rejected and must not be promoted stable.
- UI must be professional and restrained: consistent neutral palette, clear work/status color semantics, no visually mixed ad-hoc styling.
- Primary business navigation moves to a single bottom row so the operational content area gets maximum width/height.
- The desktop window opens at the full Windows work area and is not user-resizable; normal user flow is full-size or minimized to the taskbar.
- Runtime responsiveness has priority over diagnostics: filesystem/network logging must not block the UI thread.

## 2026-09-19 — Recovered V1.3 source becomes rebuild baseline

- The owner-provided `SUPRA_PRODUCTIVITY_V1.3_TEST_FULL_SOURCE.zip`, reconstructed from the stable V1.3 executable, is the authority for the next rebuild.
- Preserve the recovered V1.3 business logic and operating scenarios before making any further feature changes.
- The previous GitHub test.1/test.2 implementation must not override recovered V1.3 behavior where they differ.
- Further UI/feature changes are deferred until the V1.3 rebaseline and GitHub updater path pass owner testing.
- Public-repository security remains unchanged: recovered private endpoint/proxy values must not be committed; local encrypted runtime binding remains required.
