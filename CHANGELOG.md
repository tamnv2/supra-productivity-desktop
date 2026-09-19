# CHANGELOG

All entries must be sanitized for a public repository.

## Unreleased

### Release automation
- Added file-driven owner-test prerelease publishing through `release/test-version.txt`.
- Added file-driven accepted-stable publishing through `release/stable-version.txt`.
- Published GitHub prerelease `v1.3.1-test.1` from canonical source; candidate publish and main portable build both passed.


### GitHub build verification
- Repaired two source-reconstruction defects found by real PR CI: DPAPI credential load closure/unmarshal and latest-release URL construction.
- Refined public-source guard to avoid false positives on credential variable references while still rejecting quoted credential literals.
- Allowed only standard XML namespace hosts required by XLSX parsing in the binary URL allowlist.
- GitHub-built candidate verified by PR CI run 35412096189: guards, tests, vet, Windows x64 cross-build, binary scan, packaging and artifact upload all passed.
- Live payroll/productivity and active-picking synchronization is wired through the locally encrypted runtime profile with proxy/direct fallback.
- Stable release remains pending owner runtime acceptance.


### Source / build
- Imported a sanitized public Go/Win32 source reconstruction for the Windows portable application.
- Added Excel-derived core business model and regression tests.
- Added Windows x64 GitHub Actions build artifact workflow.
- Added tagged/manual GitHub Release workflow.
- Stable semantic-version tags create normal releases; hyphenated test versions create prereleases.
- Added public binary scan to reject embedded credential-like data and non-approved hard-coded URL hosts.

### Security / runtime
- Production credential values, internal endpoint bindings, company datasets, and raw diagnostics remain excluded from the public repository.
- Added local runtime-profile provisioning: plaintext profile beside the EXE is validated, DPAPI-encrypted for the current Windows user, and removed after successful import where possible.
- cURL session input remains local and DPAPI-protected.

### Update
- Added non-fatal GitHub latest-release check.
- Added staged stable-release download, SHA256 verification, current-EXE backup, replace-after-exit, restart, and replacement-failure restore.
- Manual/offline package selection and startup-health rollback remain pending.

### Repository / continuity
- GitHub is canonical for project state, owner decisions, next action, invariants, sanitized issues, and session continuity.

## v1.3-test — 2026-09-18

- Improved responsiveness and background processing.
- Changed PICK skip-20-minute interaction to avoid repeated checkbox event loops.
- Added manual shift selection.
- Added double-click operational detail behavior.
- Added richer diagnostics.
- Moved business controls from generic Settings to relevant screens.
- Limited Settings primarily to Dashboard credential/cURL handling.

Status: private test build; owner acceptance pending.

## 2026-09-19 — v1.3.1-test.2 runtime/UI correction
- Marked test.1 owner-rejected.
- Removed UI-thread dependency on runtime log filesystem writes.
- Runtime logs now use a bounded background queue and local AppData path.
- Diagnostic export moved off the UI thread; log viewer reads a bounded tail.
- Replaced left navigation rail with a bottom navigation row.
- Expanded operational content to near full-window width.
- Window now opens full work-area and is non-resizable/minimize-only in normal use.
- Refined shell styling to a restrained neutral palette with status-aware header text.

## 2026-09-19 — Recovered V1.3 rebaseline

- Replaced placeholder business defaults/calculations with values and formulas recovered from the stable V1.3 executable.
- Restored V1.3 shift classification, per-DO SKU deduction, minute-based productivity calculation, site filtering, relative completion text and 1C1L calculation.
- Restored the V1.3 payroll duration semantics and the 31-column active-picking table shape.
- Rebased the desktop shell to the stable-style left-navigation implementation while retaining non-blocking diagnostics.
- Kept private endpoint bindings outside the public repository through the local DPAPI runtime profile.
- Extended the GitHub updater so test/prerelease builds can detect newer test releases; SHA256 verification, backup, staged replacement and restart remain mandatory.
- Planned an actual updater verification pair: `v1.3.2-test.1` → `v1.3.2-test.2`.

### V1.3 updater verification candidates published
- `v1.3.2-test.1` published from the recovered V1.3 rebaseline.
- `v1.3.2-test.2` published from identical source as the GitHub self-update target.
- PR #3 validation passed State Guard `35419657684` and Windows Portable `35419657703`.
- Owner runtime update verification remains required before stable promotion.

## 2026-09-19 — Bottom navigation / unified log / stability

- Moved business navigation to a bottom Excel-like tab row.
- Expanded operational tables to use the freed horizontal area.
- Replaced the separate diagnostic export concept with one comprehensive sanitized Log.
- Added user-selected Log export through a native Save As dialog.
- Added periodic runtime/performance snapshots, UI render timing, command timing, table-fill timing and panic capture to the log.
- Added buffered log flush/drop tracking so diagnostics cannot block the UI.
- Prevented re-entrant page rendering and same-tab redundant renders.
- Deferred control-triggered refresh via the Windows message queue instead of destroying controls inside their own event handler.
- Disabled ListView redraw while bulk-filling table rows to reduce UI stalls.
- Window now starts maximized, has no resize/restore-down flow, and still allows minimizing to the taskbar.
- Added sanitized Excel parity proposal for Owner review.

## 2026-09-19 — test.4 Win32 hang/work-area correction

- Removed forced `SW_MAXIMIZE` calls from `WM_SYSCOMMAND`.
- App is pinned to the current monitor work area (`rcWork`) so the taskbar remains visible.
- Restore-down, resize, move and explicit maximize are blocked without recursive window-state changes; minimize remains allowed.
- Native Save As for Log export now runs on a dedicated locked OS thread instead of blocking the main UI thread.
- Log viewer no longer performs synchronous log flush before rendering and only shows a bounded recent tail.
- Added standard Windows background brush to eliminate the broken grey/white strip rendering.
- Bottom navigation uses persistent checked/push-like tab state.
- Added UI watchdog stack capture for stalls of six seconds or longer.

## 2026-09-19 — test.5 UI-thread and professional-shell rebuild

### Runtime stability
- Lock the full Win32 GUI lifetime/message loop to one OS thread using `runtime.LockOSThread()`.
- Replace idle-time watchdog with explicit message-loop PING/PONG checks.
- Move periodic runtime/RAM snapshots off the UI thread.
- Move Log-tail file reading off the UI thread.
- Ignore anonymous edit-control `WM_COMMAND id=0` notifications to avoid unnecessary logging/event work.
- Keep Log Save As and export work outside the main UI thread.

### UI/layout rebuild
- Rebuilt the application shell around a consistent header, page-title area, content groups and Excel-like bottom navigation.
- Added structured summary/status cards on Tổng quan.
- Added consistent filter/rule toolbars for Đang lấy hàng, Pick, Pack and Phân ca.
- Reworked User/PDA and Log spacing.
- Rebuilt Thiết lập into separate Dashboard-session and session-status panels.
- Removed internal/repository wording from the operator-facing settings screen.
- Standardized spacing, typography and button sizing.

## 2026-09-19 — test.6 in-app live source setup

- Removed the separate runtime-profile file from the normal Owner workflow.
- Added **Sản lượng** and **Đang lấy hàng** source selection in Thiết lập.
- Added one-time cURL capture per source.
- Source request shape and Dashboard session are stored locally with Windows DPAPI.
- Added source-specific **KIỂM TRA NGUỒN** and **XOÁ NGUỒN ĐANG CHỌN** actions.
- Refreshing one source cURL updates session values without deleting the other source.
- Converted captured yesterday/today/tomorrow dates into runtime request templates.
- Đồng bộ can run with one or both configured sources.
- Partial sync keeps previous valid snapshots and reports missing/failed sources by business name.
- Removed operator-facing “runtime profile/binding” wording from the normal sync flow.
- Prevented source changes while a synchronization task is running.

## 2026-09-19 — test.7 single Dashboard Excel pipeline

- Replaced test.6's two-source configuration with one Dashboard cURL/session.
- Derive current-picking JSON and payroll/productivity XLSX requests internally from the captured Dashboard origin.
- Auto-migrate saved test.6 credentials/bindings on startup.
- Validate both request + parser paths in **KIỂM TRA ĐỒNG BỘ**, rather than only checking HTTP 200.
- Restored Excel-style production processing: payroll rows → employee enrichment → Phân ca → Pick/Pack.
- User/PDA now merges production users with current-picking employee/device fields.
- Preserve valid snapshots when either request fails.
- Added detailed sanitized sync counts for payroll, User/PDA, Pick, Pack, Phân ca and Đang lấy hàng.
- Fixed the reconstructed site filter from 1921 to **1291**.
- Added regression tests for Site1291 filtering and active/payroll employee enrichment.

