# NEXT ACTION

Updated: 2026-09-19

## Published candidate

**v1.3.2-test.4** is published.

Automated verification:
- PR State Guard `35422413386` — PASS
- PR Windows Portable `35422413376` — PASS
- Publish Owner Test Candidate `35422493978` — PASS
- unit tests / vet / Windows x64 cross-build / public binary scan / packaging — PASS

## What test.4 changes

1. Removes the forced `SW_MAXIMIZE` call from `WM_SYSCOMMAND`.
2. Pins the application to monitor `rcWork`: full working area **with Windows taskbar visible**.
3. Blocks smaller resize/move/explicit maximize without recursive window-state changes.
4. Keeps minimize-to-taskbar and restores to full work-area size.
5. Moves native **XUẤT LOG...** Save As onto its own OS thread.
6. Keeps Log aggregation/file I/O off the main UI thread.
7. Removes synchronous Log flush from Log-page rendering.
8. Limits on-screen Log tail while exported Log remains comprehensive.
9. Adds a proper Windows background brush to remove grey/white broken-strip rendering.
10. Keeps bottom Excel-like tabs and shows the selected tab state.
11. Adds `UI_WATCHDOG_STALL` with goroutine stack capture if the main UI stops processing messages for >=6 seconds.

## Owner test now

- update from test.3 to test.4 or download test.4;
- verify taskbar remains visible at all times;
- minimize and restore;
- try resize/move/restore-down;
- rapidly switch all bottom tabs for 1–2 minutes;
- open LOG and use **XUẤT LOG...**;
- leave app idle for at least one minute.

If Windows still shows **Not Responding**, export/send the new Log. The watchdog stack should identify the exact blocked call rather than only showing the last page event.
