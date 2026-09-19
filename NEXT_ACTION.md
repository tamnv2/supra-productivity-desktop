# NEXT ACTION

Updated: 2026-09-19

## Verified from Owner logs

The real GitHub self-update path has passed:
- `v1.3.2-test.1` detected `v1.3.2-test.2`;
- updater staged the release after SHA256 verification;
- application exited;
- replacement completed;
- application restarted as `v1.3.2-test.2`.

## Current candidate

Target: **`v1.3.2-test.3`**

Changes:
1. Excel-like business navigation moved to the bottom.
2. Operational content/table area expands into the space previously used by the left navigation.
3. One comprehensive sanitized **LOG** replaces the separate diagnostic concept.
4. Log captures UI/page/command/table timing, sync/network/update events, periodic RAM/runtime state, row counts and recovered panic details without credentials.
5. **XUẤT LOG...** opens a native Save As dialog so the Owner chooses where to save the combined log file.
6. Page rendering is non-reentrant; same-tab clicks do not recreate the screen.
7. Checkbox/combo changes post refresh back to the UI queue instead of destroying controls inside their own event handler.
8. ListView redraw is suspended during bulk row insertion to reduce freezes on weak laptops.
9. Window starts maximized and cannot restore/resize to a smaller desktop window; minimize-to-taskbar remains allowed.
10. Sanitized Excel parity proposal is in `docs/EXCEL_PARITY_PROPOSAL.md`; detailed feature parity waits for Owner decisions.

## Immediate next action

Run PR CI for test.3. Merge/publish only if state guard, public-repo guard, unit tests, vet, Windows x64 build and binary scan pass.

After publish, Owner tests:
- rapid switching across all bottom tabs;
- Pick/Pack/Phân ca controls;
- minimize then restore (must return maximized);
- attempt resize/restore-down (must remain maximized);
- LOG display;
- XUẤT LOG... Save As;
- if any hang remains, send the exported Log.

Stable release remains blocked until explicit Owner acceptance.
