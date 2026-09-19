# NEXT ACTION

Updated: 2026-09-19

## Published candidate

**v1.3.2-test.5** is published.

Automated verification:
- PR State Guard `35422949366` — PASS
- PR Windows Portable `35422949403` — PASS
- Publish Owner Test Candidate `35422993953` — PASS
- unit tests / vet / Windows x64 cross-build / public binary scan / packaging — PASS

## Runtime root fix

1. Win32 GUI initialization, child controls and message loop are locked to one OS thread with `runtime.LockOSThread()`.
2. The old idle-time watchdog is removed.
3. New watchdog posts `WM_APP_PING` and requires a pong from the real UI message loop before reporting a stall.
4. Periodic RAM/Go telemetry is background work.
5. Log-tail file loading is background work.
6. Anonymous edit-control notifications are ignored instead of being logged continuously.
7. Log Save As/export remains off the UI thread.

## UI rebuild

- consistent fixed header;
- app/version + status + Update/Sync aligned;
- Excel-like bottom navigation;
- Tổng quan split into production/system/sync cards;
- Đang lấy hàng, Pick, Pack and Phân ca use consistent grouped toolbars;
- User/PDA and Log aligned to the same content grid;
- Thiết lập rebuilt into Dashboard-session and Session-status panels;
- internal repository/engineering wording removed from operator screens;
- spacing, typography and control sizes normalized;
- full Windows work area while taskbar remains visible.

## Owner verification now

1. Run test.5 and leave idle for at least **2 minutes**.
2. Switch all bottom tabs repeatedly.
3. Enter/paste text in Thiết lập.
4. Open LOG and export a Log.
5. Minimize/restore.
6. Confirm taskbar remains visible.
7. Review the new UI/layout separately from functional correctness.

If Windows still shows **Not Responding**, export/send the new test.5 Log. Any `UI_WATCHDOG_STALL` from test.5 now means a real failed PING/PONG and is actionable.
