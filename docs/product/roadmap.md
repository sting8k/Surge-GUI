# Product Roadmap

This roadmap captures accepted product direction. It is not implementation proof.
Stories should be created from these areas when work is selected.

## Implemented Foundation Slice

The 2026-05-18 foundation slice added:

Selected flow direction: Surge backend owns scheduling/concurrency; Surge
Desktop owns UX, state normalization, queue visibility, and supported API
controls. See `docs/product/download-flow.md`.

- Backend status reporting with base URL, CLI path, CLI version, token detection,
  and offline messages.
- Canonical status normalization for queued, downloading, paused, completed, and
  error states.
- URL validation before add and update operations.
- Failed-download URL update and resume flow using Surge `/update-url`.
- Bulk pause, resume, and clear-completed controls based on existing Surge API
  capabilities.
- Validation report at
  `docs/validation/2026-05-18-download-manager-foundation.md`.

Remaining roadmap items below still need deeper story work before the app should
be called a complete professional download manager.

## R1 Backend Freshness

Goal: keep the desktop client compatible with the current supported Surge
backend.

Candidate outcomes:

- Identify the latest supported Surge backend version.
- Document the required backend startup command and API assumptions.
- Detect unsupported backend versions and show a clear user-facing message.
- Add validation that proves the desktop client works against the supported
  backend version.

## R2 Core Download Manager Completeness

Goal: cover the workflows users expect from a daily download manager.

Candidate outcomes:

- Add and validate single URL downloads.
- Support pause, resume, cancel, delete, and retry with clear states.
- Define completed, failed, queued, active, paused, and cancelled states.
- Preserve enough download history for users to understand past activity.
- Handle backend disconnects and failed operations gracefully.

## R3 Queue And Control

Goal: make multi-download behavior predictable.

Candidate outcomes:

- Define queue ordering and active download limits.
- Let users prioritize, pause, resume, or remove queued items.
- Show aggregate speed and per-download progress.
- Keep UI state consistent with backend state after restart or reconnect.

## R4 Professional Desktop UX

Goal: make the app feel polished and trustworthy on desktop.

Candidate outcomes:

- Replace template docs and rough edges with product-specific documentation.
- Provide empty, loading, error, offline, and completed states.
- Improve accessibility, keyboard behavior, and visual feedback.
- Clarify platform support and package expectations.

## R5 Validation And Release Readiness

Goal: avoid claiming production quality without proof.

Candidate outcomes:

- Define quick validation commands for Go, frontend, and Wails packaging.
- Add integration proof against the supported Surge backend.
- Add platform smoke checks for macOS first.
- Track every implemented story in `docs/TEST_MATRIX.md` with evidence.
