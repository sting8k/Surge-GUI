# Product Docs

This directory holds the current product contract for Surge Desktop.

Current files:

- `overview.md`: current product position, documented capabilities, direction,
  and known gaps.
- `roadmap.md`: accepted product direction for backend freshness, core download
  manager completeness, queue control, desktop UX, and validation readiness.
- `download-flow.md`: backend-backed download distribution and management flow.
- `mac-visibility.md`: macOS dock/menu bar visibility and close-to-hide rules.

Create more domain files only when a selected story needs a durable product
contract.

## Update Rule

When behavior changes:

1. Update the affected product doc.
2. Update or create the story packet.
3. Update `docs/TEST_MATRIX.md`.
4. Record a decision if the change affects architecture, scope, risk, or a
   previously settled product rule.
