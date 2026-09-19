# NEXT ACTION

Updated: 2026-09-19

## Current state

Owner **rejected `v1.3.1-test.1`** after real company-laptop testing:
- Windows displayed `Not Responding`;
- UI/layout was not accepted.

The rejected build must not be promoted stable.

## Fix committed for next candidate

Target: **`v1.3.1-test.2`**

Changes:
1. runtime log writes no longer execute filesystem work synchronously on the UI path;
2. runtime logs are kept in local AppData, independent from a stale/slow configured data folder;
3. diagnostic export runs in a worker goroutine;
4. log viewer reads only a bounded tail;
5. navigation moved from the left rail to one bottom row;
6. operational content now uses almost the full window width;
7. app opens at the full Windows work area and is not resizable; user can minimize it to the taskbar;
8. shell styling uses one restrained neutral palette with a dark header and status-aware text.

## Immediate next action

1. Verify GitHub Actions build and owner-test publish for `v1.3.1-test.2`.
2. If both PASS, owner downloads and tests `v1.3.1-test.2`.
3. Retest:
   - startup and repeated navigation: no `Not Responding`;
   - full-size/minimize-only behavior;
   - bottom navigation and content area;
   - Pick/Pack/Phân ca controls;
   - PDA sync;
   - restricted Office sync;
   - long-running CPU/RAM/log stability.

Do not publish stable until explicit owner acceptance.
