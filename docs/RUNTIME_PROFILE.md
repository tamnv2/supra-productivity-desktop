# PRIVATE RUNTIME PROFILE

The repository is public, so private operational endpoint bindings are deliberately excluded from source control.

## Purpose

A GitHub-built public executable receives site-specific live-data bindings locally without publishing them in source code, Git history, Actions logs, Release notes, or public artifacts.

## Implemented provisioning foundation

The application now supports a local provisioning file named:

`SupraProductivity.profile.json`

Place it beside `SupraProductivity.exe`.

On startup the app:

1. detects the provisioning file;
2. validates `schema_version = 1` and the request bindings;
3. encrypts the profile with Windows DPAPI for the current Windows user;
4. stores the encrypted profile under the local secure application directory;
5. removes the plaintext provisioning file after successful import where possible;
6. logs only the sanitized profile ID and binding count — never URLs or credentials.

A synthetic example is available at:

`samples/public/runtime-profile.example.json`

## Two local layers

### Session credentials
- user pastes Copy-as-cURL (bash);
- token/cookie/signature values are extracted locally;
- sensitive fields are masked by default;
- values are persisted with Windows DPAPI.

### Runtime operational profile
- contains private request templates/bindings for operational data sources;
- is provisioned locally;
- is DPAPI-protected;
- is never committed to the public repository.

## Public schema

The public code may know generic binding names such as:

- `active-picking`
- `payroll-productivity`

but production hosts, paths, payload templates, tokens, cookies, and signatures must remain private.

## Next implementation step

The profile import/storage layer is implemented. The remaining work is to bind those locally provisioned requests to the live parsers/engines for:

- Active Picking;
- Payroll/Productivity → Pick/Pack/Shift/User-PDA as applicable.

## Office / PDA

The runtime profile is independent of GitHub. Once provisioned, core operations must continue when GitHub/public Internet is blocked but the internal services remain reachable.

## Security rule

Never attach a real runtime profile to a public GitHub issue, commit, release, Action artifact, or public diagnostic package.
