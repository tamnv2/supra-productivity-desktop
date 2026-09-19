# NEXT ACTION

Updated: 2026-09-19

## Published candidate

**v1.3.2-test.3** is published.

Automated verification before merge:
- Project State Guard `35420791580` — PASS
- Windows Portable `35420791549` — PASS
- Unit tests — PASS
- Go vet — PASS
- Windows x64 cross-build — PASS
- Public binary scan — PASS
- Packaging/artifact — PASS

## Included in test.3

1. Excel-like bottom business navigation.
2. Wider operational content/table area.
3. One comprehensive sanitized **LOG** instead of a separate diagnostic subsystem.
4. Native **XUẤT LOG...** Save As dialog.
5. Log coverage for UI page/command timing, table-fill timing, sync/network/update, periodic RAM/runtime state, row counts, panic capture and dropped-log detection.
6. Re-entrant page rendering blocked.
7. Same-tab clicks no longer rebuild the page.
8. Checkbox/combo refresh is deferred through the UI message queue.
9. ListView redraw is disabled during bulk fill to reduce freezes.
10. App always opens maximized; restore-down/resize/move is blocked; minimize-to-taskbar remains allowed.
11. Original Excel workbook has been analyzed and a sanitized keep/convert/remove proposal is stored in `docs/EXCEL_PARITY_PROPOSAL.md`.

## Owner test now

- rapidly switch all bottom tabs;
- operate Pick / Pack / Phân ca controls;
- minimize and restore — it must return maximized;
- try restore-down/resize — it must stay maximized;
- open LOG;
- use **XUẤT LOG...** and choose a save location;
- if any freeze/hang remains, send that exported Log.

Do not implement additional Excel-parity business functions until Owner confirms the pending keep/remove items.
