# Product Overview

Surge Desktop is a Wails, Svelte, and Go desktop client for managing downloads
through the Surge backend.

## Current Position

The app is a working early-stage desktop client, not yet a professional or
feature-complete download manager.

Current documented capabilities:

- Add a download URL.
- Pause, resume, and delete downloads.
- Show real-time progress, speed, and ETA from backend events.
- Toggle launch-at-login on macOS.
- Switch between dock and menu bar mode.
- Enforce a single running app instance.

## Product Direction

The target is a reliable desktop download manager that feels complete enough for
daily use, with clear backend compatibility, predictable download control, useful
queue management, and polished desktop UX.

## Known Gaps

- The Surge backend dependency/integration may not match the latest upstream
  Surge version.
- Core download-manager workflows are still too thin for a professional app.
- Queue behavior, retry behavior, completed-download handling, and error recovery
  need explicit product contracts.
- The app needs stronger validation before claiming production readiness.

## Non-Goals For The Current Docs Pass

- Do not implement new download behavior yet.
- Do not claim compatibility with the latest Surge release until verified.
- Do not mark product stories implemented without validation evidence.
