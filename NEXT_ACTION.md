# NEXT ACTION

Updated: 2026-09-19

## Verified state

`v1.3.1-test.1` is **owner-rejected** and must not be promoted.

Replacement candidate **`v1.3.1-test.2`** is now published and all required GitHub checks passed:

- Project State Guard: run `35416300901` — PASS
- Build Windows Portable: run `35416300869` — PASS
- Publish Owner Test Candidate: run `35416300861` — PASS
- release target commit: `59ebff41eafc90acf74c954315b49bf6b3f36521`

### test.2 changes

- runtime logging removed from synchronous UI filesystem paths;
- runtime logs use a bounded background queue in local AppData;
- diagnostic export runs off the UI thread;
- log viewer reads a bounded tail only;
- business navigation moved to one bottom row;
- operational content uses near-full window width;
- app opens full Windows work-area and is not resizable; minimize-to-taskbar remains available;
- shell styling is restrained and consistent, with a dark header and status-aware text.

## Immediate next action

**Owner tests `v1.3.1-test.2` on the target company laptop.**

Priority verification:
1. open app and switch screens repeatedly — Windows must not show `Not Responding`;
2. confirm full-size / minimize-only window behavior;
3. confirm bottom navigation and usable content area;
4. test Pick / Pack / Phân ca controls;
5. sync on PDA network;
6. sync on restricted Office network;
7. leave running long enough to inspect CPU/RAM/log stability.

If any item fails, send the private diagnostic and screenshot; record only sanitized findings in the public repository.

Do not publish stable until explicit owner acceptance.
