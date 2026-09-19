# KNOWN ISSUES

Updated: 2026-09-19

## Open / verification required

### GitHub-built candidate requires target-laptop verification
- Canonical sanitized source is present in the public repository.
- Pull-request CI successfully completed state/security guards, tests, vet, Windows x64 build, binary scan, package and artifact upload.
- Verified build run: `35412096189`; artifact id: `10574174267`.
- **Status:** build verified; owner runtime acceptance pending.

### V1.3 / V1.3.1 PICK flicker/event storm
- Earlier diagnostics identified repeated skip-20-minute UI events causing recalculation/render loops.
- The interaction was changed to avoid the recursive checkbox mechanism.
- **Status:** code changed; target-laptop verification required.

### Functional parity
- Live payroll/productivity and active-picking execution is wired through the locally encrypted runtime profile.
- Remaining parity items still require real environment verification, especially detailed CPU telemetry, all target/quota edit controls, and User/PDA behavior.
- **Status:** ongoing owner verification.

### Updater hardening
- GitHub stable-release lookup, download, SHA256 verification, backup, staged replacement, restart and replacement-failure restore are implemented.
- GitHub being blocked is non-fatal.
- Manual/offline package selection and startup-health rollback remain incomplete.
- **Status:** usable foundation; hardening remains.

### Stable Release
- Release workflow is committed and ready.
- A stable release is intentionally not created until the owner tests the GitHub-built candidate.
- **Status:** pending owner acceptance, not a technical source/build blocker.

## Resolved / superseded

- Public source import blocker → resolved.
- Missing canonical source claim → resolved/corrected.
- GitHub Actions reproducible build verification → resolved; PR CI passed.
- Public-repo guard false positives for source variable references/XML namespace URLs → fixed and verified.
- V1.1/V1.2 prototypes → superseded.

### v1.3.1-test.1 owner rejection — responsiveness and shell layout
- Owner runtime test showed Windows `Not Responding`.
- Owner rejected the visual shell/layout as insufficiently professional.
- Verified source risk: log writes and directory preparation were synchronous on UI-driven paths and could inherit a slow/stale configured data path.
- Fix in test.2: runtime log path is local AppData, log writes use a bounded non-blocking background queue, log viewing reads a bounded tail, and diagnostic export runs off the UI thread.
- UI correction in test.2: fixed full-work-area window, no resize flow, bottom navigation, neutral professional palette/status-aware header, larger operational content area.
- Test.2 GitHub build/publish checks passed; prerelease is available.
- **Status:** target-laptop verification of test.2 pending.

### Recovered V1.3 rebaseline verification
- The owner supplied a source reconstruction extracted from the stable V1.3 executable after the GitHub test builds diverged from stable behavior.
- Placeholder defaults/calculations in the prior GitHub reconstruction were confirmed to differ from V1.3 and are being replaced.
- Private endpoint/proxy literals found in the recovered binary are deliberately excluded from the public repository; runtime binding remains local/DPAPI-protected.
- `v1.3.1-test.2` is superseded as the business-behavior baseline. Its non-blocking logging fix is retained.
- PR #3 CI/build passed (State Guard `35419657684`, Windows Portable `35419657703`).
- Both updater verification releases `v1.3.2-test.1` and `v1.3.2-test.2` are published.
- **Status:** owner target-laptop runtime/updater verification pending.

### Runtime hang after short interaction
- Owner reported that the V1.3.2 test build can become unresponsive after brief UI interaction.
- Available log shows successful startup/page changes and successful GitHub update, but no exception or timing detail identifying the stall.
- Corrective candidate adds page/command/table timing, panic capture, periodic runtime telemetry, non-reentrant page rendering, deferred control refresh and batched ListView redraw suppression.
- **Status:** code changed; target-laptop verification required.


### v1.3.2-test.3 Win32 work-area / message-loop hang
- Owner reproduced **Not Responding** with zero business rows loaded; memory was low (~1.7 MB heap) and only four goroutines were active, so this is not a capacity/RAM failure.
- The test.3 shell forced `SW_MAXIMIZE` from inside `WM_SYSCOMMAND` while also suppressing restore/size/move. This is replaced by monitor `rcWork` pinning with no recursive maximize call.
- The test.3 log export Save As dialog blocked the main UI command for 2531 ms. The common dialog is moved to its own locked OS thread; log file aggregation remains background work.
- Parent window had no background brush, producing broken grey/white strip rendering. The shell now uses the standard Windows window background.
- Added UI watchdog logging: a main-thread stall >=6 seconds records all Go goroutine stacks as `UI_WATCHDOG_STALL`.
- **Status:** fixed in test.4 candidate; target-laptop verification required.

