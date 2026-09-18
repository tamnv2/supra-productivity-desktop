# PRIVATE RUNTIME PROFILE

The repository is public, so private operational endpoint bindings are deliberately excluded from source control.

## Goal

A GitHub-built public executable must be able to receive the site-specific operational bindings **locally**, without publishing them in:

- source files;
- Git history;
- GitHub Actions logs;
- Release notes;
- public release metadata.

## Model

The application has two separate local secret/config layers:

1. **Session credentials**
   - supplied by the user by pasting Copy-as-cURL (bash);
   - token/cookie/signature values are parsed locally;
   - persisted with Windows DPAPI for the current Windows user.

2. **Runtime operational profile**
   - contains private data-source endpoint/request templates needed for live Active-Picking and Payroll/Pick/Pack synchronization;
   - provisioned locally;
   - stored DPAPI-encrypted under the current Windows user;
   - never committed to this repository.

## Public schema only

The public code may know binding names such as:

- `active-picking`
- `payroll-productivity`

but must not contain their production hosts, URLs, cookies, tokens, or signatures.

## Provisioning direction

The supported provisioning path will be:

1. place a private provisioning file beside the EXE or select it locally;
2. application validates schema and required binding names;
3. application encrypts the profile with Windows DPAPI;
4. plaintext provisioning file is deleted after successful import where possible;
5. logs record only profile version/binding names, never URLs or credentials.

## Office / PDA

The runtime profile is independent of GitHub. Once provisioned, core operations must continue when:

- GitHub is blocked;
- public Internet is blocked;
- internal operational services remain reachable.

## Security rule

Never attach a real runtime profile to a public GitHub issue, commit, release, Action artifact, or public diagnostic package.
