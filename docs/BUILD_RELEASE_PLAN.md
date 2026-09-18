# BUILD / RELEASE / UPDATE PLAN

## Repository

`tamnv2/supra-productivity-desktop` — public.

## Build goal

GitHub Actions must produce a reproducible Windows x64 portable build from canonical source.

Planned release assets:
- `SupraProductivity.exe`
- `SupraProductivity_<version>.zip`
- `SHA256SUMS.txt`
- release notes derived from `CHANGELOG.md`.

## Release trigger

Preferred:
- CI build/test on push and pull request.
- Release build on version tag `vX.Y.Z`.

## Updater behavior

When GitHub is reachable:
1. query latest compatible GitHub Release;
2. compare semantic version;
3. download release ZIP/EXE;
4. validate SHA256;
5. stage update;
6. close current app;
7. atomically replace executable using a helper/updater;
8. restart;
9. rollback if replacement/startup fails.

## Restricted Office network

GitHub may be blocked. Therefore:
- updater failure must never block normal application startup or operation;
- show only a non-fatal "cannot check updates on this network" state;
- support manual/offline update package selection;
- optionally add an owner-approved reachable mirror later.

## Public-repo security

Builds must never require production credentials from repository files.
Runtime credentials are user-supplied locally.

## Current status

Build/release workflow is intentionally not finalized until the current V1.3 source is committed and sanitized.
