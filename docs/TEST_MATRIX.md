# Test Matrix

This file maps product behavior to proof.

Product direction is now defined in `docs/product/overview.md` and
`docs/product/roadmap.md`. Do not mark a row implemented until tests or
validation evidence exist.

## Status Values

| Status | Meaning |
| --- | --- |
| planned | Accepted as intended behavior, not implemented |
| in_progress | Actively being built |
| implemented | Implemented and proof exists |
| changed | Contract changed after earlier implementation |
| retired | No longer part of the product contract |

## Matrix

- E01 Backend Freshness
  - Contract: latest supported Surge backend compatibility.
  - Proof: Go compile, frontend typecheck/build, Wails build; integration needed.
  - Status: in_progress.
  - Evidence: `docs/validation/2026-05-18-download-manager-foundation.md`.
- E02 Core Download Manager
  - Contract: core download states and controls.
  - Proof: Go compile, frontend typecheck/build, Wails build; E2E needed.
  - Status: in_progress.
  - Evidence: `docs/validation/2026-05-18-download-manager-foundation.md`.
- E03 Queue And Control
  - Contract: backend-backed distribution and management flow.
  - Proof: Go compile, frontend typecheck/build, Wails build; integration needed.
  - Status: in_progress.
  - Evidence: `docs/validation/2026-05-18-download-manager-foundation.md`.
- E04 Professional Desktop UX
  - Contract: professional desktop UX states and polish.
  - Proof: frontend typecheck/build and Wails build; platform smoke needed.
  - Status: in_progress.
  - Evidence: `docs/validation/2026-05-18-download-manager-foundation.md`.
- E05 Validation Readiness
  - Contract: validation and release readiness.
  - Proof: Go compile, frontend typecheck/build, Wails build.
  - Status: in_progress.
  - Evidence: `docs/validation/2026-05-18-download-manager-foundation.md`.

## Evidence Rules

- Unit proof covers pure domain and application rules.
- Integration proof covers backend enforcement, data integrity, provider
  behavior, jobs, or service contracts.
- E2E proof covers user-visible flows across the relevant UI surfaces.
- Platform proof covers only shell, deployment, mobile, desktop, or runtime
  behavior that cannot be proven in lower layers.
- A story can be implemented without every proof column if the story packet
  explains why.
