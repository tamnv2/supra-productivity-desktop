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
