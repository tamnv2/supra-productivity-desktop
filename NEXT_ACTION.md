# NEXT ACTION

Updated: 2026-09-19

## Verified state

GitHub is now the canonical source and build system.

Completed:
- sanitized canonical source in `main`;
- public-repo/state guards;
- Go tests and vet;
- Windows x64 GitHub build;
- binary public scan;
- Actions artifact packaging;
- live runtime-profile binding;
- GitHub update foundation;
- automatic owner-test prerelease workflow;
- automatic accepted-stable release workflow.

### Published owner-test candidate

`v1.3.1-test.1`

GitHub Actions:
- candidate publish run: `35412269300` — PASS
- main portable build run: `35412269285` — PASS

Release contains:
- `SupraProductivity.exe`
- `SHA256SUMS.txt`
- `SupraProductivity_v1.3.1-test.1_windows_x64.zip`

## Immediate next action

**Owner tests `v1.3.1-test.1` on the target company laptop.**

Test:
1. standard Windows user, no Admin;
2. import session with Copy-as-cURL bash;
3. sync on PDA;
4. sync on restricted Office network;
5. verify PICK has no flicker/blank event loop;
6. verify manual shift, double-click detail, typed sorting and screen-specific controls;
7. leave app running long enough to inspect CPU/RAM/log stability.

### If accepted

Create/update:

`release/stable-version.txt`

with the approved semantic version, for example:

`v1.3.1`

That single canonical change automatically runs `.github/workflows/publish-stable.yml`, rebuilds from source, validates security/tests, and publishes the stable GitHub Release. The in-app updater checks stable GitHub Releases and verifies SHA256 before replacement.

### If not accepted

Send the diagnostic privately. Record only sanitized findings in GitHub, fix the verified issue, increment `release/test-version.txt`, and GitHub automatically publishes the next prerelease candidate.

## Security

Never commit the runtime cURL/token/cookie/signature, raw company data, raw diagnostic ZIPs, or private endpoint bindings to this public repository.
