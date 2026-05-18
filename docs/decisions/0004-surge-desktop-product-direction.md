# 0004 Surge Desktop Product Direction

Date: 2026-05-18

## Status

Accepted

## Context

The repository contains an early-stage Wails, Svelte, and Go desktop client for
Surge. The previous harness wording still treated the project as if no product
implementation existed.

The human clarified two product concerns:

1. Surge backend compatibility is not updated or verified against the latest
   supported upstream version.
2. The app is not yet professional or feature-complete enough to be considered a
   full download manager.

## Decision

Treat Surge Desktop as an early-stage product that must mature through the
harness workflow.

The current product direction is:

- Verify and maintain compatibility with the supported Surge backend version.
- Build toward a professional download manager with complete core workflows,
  queue control, reliable state handling, polished desktop UX, and validation
  evidence.
- Keep product truth in `docs/product/overview.md`, `docs/product/roadmap.md`,
  `docs/stories/backlog.md`, and `docs/TEST_MATRIX.md`.

## Alternatives Considered

1. Keep the harness as a generic pre-implementation shell. Rejected because the
   repository already contains a working app and the human provided product
   direction.
2. Start coding immediately. Rejected because the requested work was
   documentation normalization and the product gaps need explicit contracts
   first.

## Consequences

Positive:

- README and docs now describe the project honestly as early-stage.
- Future agents have a concrete roadmap instead of generic placeholders.
- Product gaps are tracked without claiming implementation proof.

Tradeoffs:

- The existing app may still be ahead of or behind the docs until stories verify
  behavior with tests.
- Surge backend version support remains unverified until a dedicated story runs
  compatibility checks.

## Follow-Up

- Select the first story, likely backend compatibility verification.
- Replace remaining generated template docs when they block product clarity.
- Add validation commands before marking roadmap items implemented.
