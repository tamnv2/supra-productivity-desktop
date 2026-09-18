# HANDOVER — SUPRA PRODUCTIVITY DESKTOP

This is durable background context. For a new task, start with `ops/project-state.json` and `NEXT_ACTION.md`; read this file only when broader context is needed.

## Purpose

Replace an increasingly complex Excel/VBA productivity workbook with a lightweight portable Windows desktop application while retaining approved operational behavior.

## Main operational areas

- Overview
- Active Picking / current picking sessions
- Pick
- Pack
- Shift / manual shift override
- User / PDA
- Diagnostics / logs
- Dashboard credential input

## Scope removed by owner

R01, R02, R03, R04, R07, R08.

## Core design

- Portable Windows x64.
- Standard user privileges.
- Local-first cache/state.
- Adaptive networking for PDA and restricted Office environments.
- Credential entry from Copy-as-cURL bash.
- Sensitive values masked and omitted from logs.
- Detailed diagnostics for CPU/RAM/network/business/UI errors.
- Operational settings placed on the related business screen.

## Iteration history

### V1 / V1.1
Initial native portable prototype and UI/network test. UI blocking and live-sync limitations were identified.

### V1.2
Expanded live data and business controls. Owner testing found PICK rendering/flicker problems and further UI/workflow parity gaps.

### V1.3
Focused on responsiveness, event-loop prevention, manual shift, screen-specific controls, double-click detail, richer diagnostics, and settings cleanup.

Current owner verification is still required.

## Current repository direction

GitHub public repository is now the canonical persistence layer. Raw secrets and operational data are forbidden. Source import and GitHub Actions build/release are the next engineering steps.
