# NEXT ACTION

Updated: 2026-09-19

## Current verified state

The canonical sanitized source is now in GitHub and the pull-request build has passed:

- project-state validation: PASS
- public-repo sensitive-value guard: PASS
- `go test ./...`: PASS
- `go vet ./...`: PASS
- Windows x64 cross-build: PASS
- public binary scan: PASS
- package + artifact upload: PASS

Verified CI:
- workflow run: `35412096189`
- artifact: `SupraProductivity-windows-x64`
- artifact id: `10574174267`
- artifact digest: `sha256:1103ef127b603408516979bdfbe8c87c3164b4562ef4ab1deee6a938f953bf8c`

## Immediate next action

**Owner-test the GitHub-built artifact on the target company laptop.**

Test in this order:

1. Start as standard Windows user.
2. Import Dashboard session using Copy-as-cURL bash.
3. Sync on PDA network.
4. Sync on restricted Office network.
5. Verify PICK no longer flickers/blanks.
6. Verify manual shift, double-click detail, typed sorting and operational controls.
7. Leave the app running long enough to review CPU/RAM/log stability.

If accepted:
- create the next semantic version tag;
- let `.github/workflows/release.yml` publish the Windows portable GitHub Release;
- future accepted clients update through GitHub Release with SHA256 verification.

If not accepted:
- export diagnostics privately;
- commit only sanitized findings;
- fix the verified regression and let GitHub Actions rebuild.

## Important

A **stable release is intentionally not published yet**. Publishing a stable tag before the GitHub-built candidate is tested on the company laptop would incorrectly mark an unverified build as accepted.
