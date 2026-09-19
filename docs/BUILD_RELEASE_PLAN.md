# BUILD / RELEASE / UPDATE PLAN

## Repository

`tamnv2/supra-productivity-desktop` — public.

## Current build pipeline

`.github/workflows/build.yml`

On main pushes, pull requests, or manual dispatch it will:

1. validate canonical project state;
2. scan public text files for likely sensitive values;
3. run core unit tests and Go vet;
4. cross-build Windows x64 portable EXE;
5. scan the built binary for embedded secrets/private URL hosts;
6. create SHA256;
7. package and upload the Windows artifact.

## Release pipeline

`.github/workflows/release.yml`

Version format: `vX.Y.Z` or a hyphenated test version such as `v1.4.0-test.1`.

Assets:
- `SupraProductivity.exe`
- `SHA256SUMS.txt`
- `SupraProductivity_<version>_windows_x64.zip`

Stable tags without a hyphen create normal releases. Hyphenated test versions are prereleases.

## Updater

Implemented foundation:
- non-fatal GitHub release check;
- no dependency on GitHub for core startup/operation;
- blocked GitHub on Office network is treated as an update-check limitation, not network failure.

Implemented:
1. download stable release EXE and SHA256 asset;
2. verify SHA256 before replacement;
3. stage new executable beside the current portable EXE;
4. back up the current executable;
5. replace after app exit with retry;
6. restart the application;
7. restore the backup if replacement itself fails.

Still required:
- manual/offline package selection for restricted Office network;
- stronger startup-health rollback if a newly replaced EXE launches but fails shortly afterward.

## Private operational configuration

Production endpoint bindings are not embedded in public source or GitHub Releases.

See `docs/RUNTIME_PROFILE.md`.

## Security

A release is invalid if any guard detects credentials, private endpoint URLs, raw company data, or unsanitized diagnostics.

## Current verification status

Local core tests, Windows cross-build, and binary public guard passed on the reconstructed sanitized source.

Remote GitHub Actions execution still needs an external/manual trigger because connector-created events did not expose a run during this session.
