# Download Manager Foundation Validation

Date: 2026-05-18

## Scope

Validated the first foundation slice toward a more professional Surge Desktop
client:

- Backend status and CLI version visibility.
- Canonical download status normalization.
- URL validation for add/update flows.
- Failed-download URL update and resume flow.
- Bulk pause, resume, and clear-completed controls.
- Backend-backed mirror input in the advanced add flow.
- Offline engine UI state.
- Consistent macOS reveal behavior for dock mode, menu bar mode, second launch,
  and menu bar Open action.

## Commands Run

```text
go test ./...
cd frontend && npm run check
cd frontend && npm run build
wails build
markdownlint README.md AGENTS.md build/README.md frontend/README.md 'docs/**/*.md'
```

## Results

| Check | Result | Notes |
| --- | --- | --- |
| Go test | pass | No Go test files yet; package compiles. |
| Frontend typecheck | pass | `svelte-check` passed cleanly. |
| Frontend build | pass | Vite production build completed. |
| Wails build | pass | macOS arm64 app packaged successfully. |
| Markdown lint | pass | Project documentation linted cleanly. |

## Evidence

- Local Surge CLI detected as `Surge v0.8.0` via `surge --version`.
- Advanced add flow now sends optional `mirrors` to Surge instead of managing a
  separate desktop scheduler.
- Wails build output: `build/bin/Surge.app/Contents/MacOS/Surge`.
- macOS visibility behavior is documented in `docs/product/mac-visibility.md`.

## Gaps

- No automated integration test spins up Surge and exercises the HTTP API yet.
- Backend latest upstream compatibility is surfaced but not fully enforced by a
  minimum/supported version policy.
- Queue ordering and active download limit controls still depend on Surge backend
  behavior and are not yet exposed as first-class settings in the desktop UI.
- macOS show/hide behavior still needs manual smoke testing on a signed packaged
  app across Dock launch, Finder launch, menu bar mode, close-to-hide, and
  second-instance launch.
