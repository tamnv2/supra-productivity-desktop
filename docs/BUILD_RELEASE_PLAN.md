# BUILD / RELEASE / UPDATE PLAN

## Canonical repository

`tamnv2/supra-productivity-desktop` — public.

Production credentials, private endpoint bindings, raw operational data and raw diagnostics are not repository assets.

## Continuous build

`.github/workflows/build.yml`

On main/PR it validates project state and public-repo safety, runs tests/vet, cross-builds Windows x64, scans the produced binary, creates SHA256 and uploads the portable artifact.

Verified PR build:
- run `35412096189` — PASS.

Verified main build for current candidate request:
- run `35416300869` — PASS.

## Owner-test prerelease

`.github/workflows/publish-test.yml`

Canonical trigger:
- change `release/test-version.txt` to a new version matching `vX.Y.Z-test.N`.

The workflow rebuilds from canonical source, reruns guards/tests/vet/binary scan and publishes a GitHub **prerelease**.

Current candidate:
- `v1.3.2-test.1` — recovered V1.3 rebaseline candidate.
- After test.1 is published, `v1.3.2-test.2` will be published from identical source as the in-app updater verification target.

Previous `v1.3.1-test.1` and `v1.3.1-test.2` are superseded for business-behavior baseline purposes by the owner-provided recovered V1.3 source.

Prereleases are for owner testing and are not treated as accepted stable builds.

## Accepted stable release

`.github/workflows/publish-stable.yml`

After explicit owner acceptance, create/update:

`release/stable-version.txt`

with a stable semantic version such as `v1.3.1`.

That change automatically rebuilds, validates and publishes a normal GitHub Release. Existing stable release tags are not silently overwritten.

The older tag/manual `.github/workflows/release.yml` remains available as an alternate release path.

## In-app updater

Update path:
1. stable builds query the latest stable GitHub Release;
2. test builds query recent GitHub Releases and select a newer test/stable version only;
3. compare semantic/test version to prevent downgrade;
4. download `SupraProductivity.exe` and `SHA256SUMS.txt`;
5. verify SHA256;
6. save current executable as backup;
7. stage replacement after app exit;
8. restart;
9. restore backup if replacement itself fails.

The rebaseline is verified with a real `v1.3.2-test.1` → `v1.3.2-test.2` update pair.

GitHub/public Internet being unavailable is non-fatal; core internal operations continue.

Remaining hardening:
- manual/offline package picker for Office networks where GitHub is blocked;
- startup-health rollback after a replacement that launches but later fails.

## Acceptance rule

A successful GitHub build/prerelease is **not owner acceptance**.

Only publish/promote a stable version after explicit owner runtime approval.
