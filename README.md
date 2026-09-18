# SUPRA PRODUCTIVITY DESKTOP

Windows portable desktop application for Pick / Pack / shift productivity operations.

> **PUBLIC REPOSITORY:** Never commit real tokens, cookies, signatures, passwords, internal endpoint URLs, raw diagnostic archives, operational spreadsheets, or unsanitized logs.

## Project authority

This repository is the canonical source for:
- source code and release artifacts;
- current project state and next action;
- owner decisions and stable invariants;
- sanitized test/debug findings;
- build/release/update documentation.

AI memory and old chat history are secondary. If they conflict with this repository, use the repository state and the owner's newest explicit instruction.

## AI / session continuity

A new AI session should **not read the whole repository first**.

Read in this order:
1. `AGENTS.md`
2. `ops/project-state.json`
3. `NEXT_ACTION.md`
4. Only the task-specific documents referenced by `AGENTS.md`.

See `docs/AI_BOOTSTRAP.md` for the full protocol.

## Current product constraints

- Windows x64 portable executable.
- Runs as a standard Windows user; no Administrator requirement.
- Must not depend on Excel at runtime.
- Core operations must work on both PDA network and restricted Office network when the internal services remain reachable.
- Local-first UI: network failure must not blank or freeze the application.
- Diagnostics are comprehensive but must redact sensitive credentials.
- Business behavior is based on the approved Excel workflow, except functions explicitly removed by the owner.

## Public-repo security

Read `docs/SECURITY_PUBLIC_REPO.md` before adding code, fixtures, logs, screenshots, configs, or diagnostic files.

## Status

Machine-readable status: `ops/project-state.json`

Immediate work: `NEXT_ACTION.md`
