# EXCEL PARITY PROPOSAL

Updated: 2026-09-19

Source basis: owner-provided original macro workbook `SanLuongNgay_V29_Optimized_UpdateUser0509.xlsm`.

Security note: the workbook contains operational data and private runtime configuration. This document records only sanitized UI/business structure. No employee data, credentials, tokens, private endpoint URLs, API secrets or raw workbook content is committed.

## What the original Excel establishes

Visible operational areas:
- Dữ liệu User PDA
- Phân ca
- Pick
- Pack
- Site1291
- Site1399
- Config

Supporting/hidden areas include mapping, raw production import, active-picking data, recap, manual-shift storage and temporary calculation/cache sheets.

The Excel interaction pattern is:
- operational controls occupy a compact area above the data;
- the business table receives most of the screen;
- worksheet tabs are at the bottom;
- technical/helper storage is not part of normal operator interaction.

## Recommended desktop mapping

### Recommend KEEP / reproduce in native desktop UI

1. **Bottom business navigation**
   - Tổng quan
   - Đang lấy hàng
   - Pick
   - Pack
   - Phân ca
   - User / PDA
   - Log
   - Thiết lập

2. **Pick — retain Excel working model**
   - ca filter;
   - định mức khoán chẵn;
   - SKU deduction behavior;
   - target chẵn/lẻ by Inhouse/other provider;
   - target pick speed;
   - 1 chẵn 1 lẻ check;
   - incomplete-even / 1C1L-error filters;
   - all-site/support visibility where still operationally required;
   - typed sorting and detail view.

3. **Pack — retain compact Excel model**
   - ca filter;
   - table columns and productivity calculations;
   - top summary/target area where it is actually used.

4. **Phân ca**
   - automatic shift result;
   - manual override;
   - first/last order times;
   - visible note when manual shift wins.

5. **Đang lấy hàng**
   - Refresh/load data;
   - only-running vs all records;
   - optional technical-column visibility;
   - progress/status/warning columns;
   - double-click details.

6. **User / PDA**
   - operational user/PDA list;
   - manual user synchronization where required;
   - no hidden helper-sheet concept in the desktop UI.

7. **Synchronization controls**
   - manual production sync;
   - manual user sync where needed;
   - clear last-sync state and failure reason;
   - network failure must keep the previous valid snapshot.

### Recommend CONVERT, not copy literally

1. Excel hidden/very-hidden sheets → in-memory model/local cache.
2. Config cells containing session values → DPAPI-protected local settings.
3. Excel Form Controls → native desktop controls.
4. Excel temporary sheets used for clipboard/ranges/1C1L → internal application state.
5. Excel Save As/report buttons → native file dialogs and explicit export commands.
6. Excel formula/status helper cells → code with regression tests.
7. Technical columns → hidden by default, user can reveal when troubleshooting.

### Recommend REMOVE / keep out unless Owner confirms a need

1. Raw worksheet-like technical storage screens.
2. Any UI that exposes raw token/APISID/USID/signature values by default.
3. Separate “diagnostic” subsystem — one comprehensive sanitized Log is enough.
4. Duplicate controls that perform the same sync/refresh action.
5. Site1399-specific screens/data if this desktop application's actual scope is now only 1291.
6. Legacy temporary/clipboard helper concepts that existed only because Excel/VBA required them.

## Items that need Owner decision before implementation

1. **Pack SKU deduction control**
   - Original Excel shows this setting.
   - Current canonical desktop decision says Pack should not expose a deduction control.
   - Recommendation: keep it hidden/disabled unless Owner explicitly restores it.

2. **Site1399**
   - Original workbook includes Site1291 and Site1399.
   - Recommendation: remove Site1399 from the new desktop UI if the current application is 1291-only.

3. **N-1 production report**
   - Original workbook contains a “Tạo báo cáo sản lượng N-1” action.
   - Recommendation: keep only if still used operationally; otherwise remove from the core app.

4. **Automatic sync interval**
   - Excel exposes/manualizes sync modes and 15/30/60-minute choices.
   - Recommendation: core data refresh remains manual-first until runtime stability is accepted; automatic interval can be added after Owner confirms desired cadence.

5. **User/PDA source behavior**
   - Excel has dedicated sync/cache behavior.
   - Recommendation: retain the operational result, but do not reproduce the workbook's hidden-sheet plumbing.

## UI direction

Use Excel as the information architecture reference, not as a pixel-for-pixel visual copy:
- light neutral workspace;
- compact control strip above each table;
- clear Excel-like bottom navigation;
- table consumes the majority of screen;
- status colors only where they carry operational meaning;
- no decorative color mixing;
- always maximized desktop workspace, minimize-to-taskbar allowed.
