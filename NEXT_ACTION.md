# NEXT ACTION

Updated: 2026-09-19

## Owner finding on test.4

`v1.3.2-test.4` is rejected:
- Windows still shows **Not Responding** after short runtime.
- UI/layout remains too raw, inconsistent and visually unsuitable.

The exported test.4 Log provides the key technical clue:
- business row counts were still zero, so the hang is not caused by table size;
- memory use remained low;
- the watchdog captured goroutine 1 in the Win32 message-pump syscall;
- the prior GUI lifetime was **not locked to one OS thread**;
- the prior watchdog itself was also flawed because idle `GetMessageW` time was treated as a stall.

## test.5 corrective candidate

### Runtime root fix
1. Call `runtime.LockOSThread()` before all Win32 initialization and keep the GUI/message loop on that OS thread for its complete lifetime.
2. Replace idle-time watchdog with `WM_APP_PING` / pong acknowledgement; only a missing pong is considered a real message-loop stall.
3. Move periodic RAM/Go runtime telemetry off the UI thread.
4. Load Log-tail text asynchronously instead of reading the file during page rendering.
5. Suppress anonymous edit-control command logging.
6. Keep native Log Save As/export off the UI thread.

### UI rebuild
1. Consistent fixed header with app/version, status, Update and Sync.
2. Consistent page title + short operational subtitle.
3. Excel-like bottom business navigation with selected state.
4. Tổng quan rebuilt into production/system/sync cards.
5. Đang lấy hàng/Pick/Pack/Phân ca use consistent grouped filter/rule bars.
6. User/PDA and Log use the same content alignment.
7. Thiết lập rebuilt into Dashboard-session and Session-status panels.
8. Remove public-repository/internal engineering text from operator UI.
9. Normalize spacing, button sizes and typography.
10. App remains full Windows work area with taskbar visible.

## Immediate next action

Run PR CI for **v1.3.2-test.5** and publish only if all automated checks pass.

Owner runtime test after publication:
- open and leave idle for at least two minutes;
- switch bottom tabs continuously;
- type/paste inside Thiết lập;
- open LOG;
- minimize/restore;
- verify taskbar remains visible;
- confirm whether Windows ever shows Not Responding;
- review professional layout separately from functional correctness.

If test.5 still hangs, the PING/PONG watchdog will now represent a real blocked message loop and its stack can be trusted.
