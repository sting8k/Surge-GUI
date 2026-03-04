# Notes

## 2026-03-01: Initial Build

### Architecture Decision: HTTP API Client (not Embedded Engine)
- Go `internal/` packages cannot be imported from outside the module
- Wails app spawns `surge server` as a subprocess, communicates via HTTP API on port 1700
- SSE event stream from `/events` endpoint is forwarded via Wails EventsEmit to Svelte frontend
- If surge server is already running, app connects to it instead of spawning a new one

### Key Files
- `app.go` — HTTP API client, SSE streaming, Wails bindings
- `main.go` — Wails app config, macOS titlebar, dark theme
- `frontend/src/App.svelte` — main dashboard UI
- `frontend/src/style.css` — design system (Electric Noir theme)

### Frontend Design
- Color palette: Electric Cyan (#00e5ff) accent on deep charcoal surfaces
- Typography: Outfit (display) + JetBrains Mono (data/mono)
- Shimmer animation on active download progress bars
- Status-colored left borders on download cards
- Transparent macOS titlebar with drag region

### Known Limitations
- Token discovery relies on `surge token` CLI command or common file paths
- No embedded engine — requires `surge` binary in PATH
- Browser extension compatibility preserved (HTTP server on port 1700 served by surge itself)
