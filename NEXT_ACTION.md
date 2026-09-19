# NEXT ACTION

Updated: 2026-09-19

## Owner finding on test.3

`v1.3.2-test.3` is rejected:
- UI rendered incorrectly with grey/white bands.
- App entered Windows **Not Responding**.
- Owner requires “full” to mean full **work area with taskbar visible**, not fullscreen/taskbar coverage.

The exported log shows the failure was not caused by heavy data:
- payroll/pick/pack/active rows were all zero;
- heap remained very small;
- only a few goroutines were active;
- Save As command 401 held the main UI command for 2531 ms.

## test.4 corrective candidate

1. Remove forced `SW_MAXIMIZE` from `WM_SYSCOMMAND`.
2. Pin outer window to monitor `rcWork`; Windows taskbar remains visible.
3. Block resize/move/explicit maximize without recursively changing window state.
4. Keep minimize-to-taskbar.
5. Move the native Log Save As dialog to its own locked OS thread.
6. Keep Log aggregation/background I/O off the UI thread.
7. Remove synchronous Log flush from Log-page rendering.
8. Limit on-screen Log tail while exported Log remains comprehensive.
9. Give the parent Win32 window a standard background brush so static controls no longer appear as broken strips.
10. Keep bottom Excel-like tabs with a visible selected state.
11. Add `UI_WATCHDOG_STALL` stack capture for any main-thread stall >=6 seconds.

## Immediate next action

Run PR CI for **v1.3.2-test.4**. Publish only after full CI passes.

Owner runtime verification after publication:
- taskbar always visible;
- app fills remaining monitor work area;
- minimize/restore works and restores full work-area size;
- no restore-down/resize smaller;
- rapidly switch all bottom tabs;
- open/export Log;
- leave app idle for at least one minute;
- if Windows still shows Not Responding, export/send the new Log; watchdog stack should identify the exact blocking call.
