# Surge Desktop Frontend

This directory contains the Svelte + TypeScript frontend used by the Wails
desktop shell.

## Role

The frontend is responsible for user-facing download-manager workflows:

- Adding download URLs.
- Showing download progress, speed, ETA, and status.
- Exposing pause, resume, delete, and future queue controls.
- Showing desktop app settings such as launch-at-login and dock/menu bar mode.

## Current Status

The UI is early-stage. It should not be treated as a complete professional
download-manager experience yet.

Known product gaps are tracked in:

- `../docs/product/overview.md`
- `../docs/product/roadmap.md`
- `../docs/stories/backlog.md`
- `../docs/TEST_MATRIX.md`

## Development

Run the app from the repository root with Wails:

```bash
wails dev
```

Use frontend package commands only when they exist and are relevant to the
selected story. Do not claim validation has passed unless the command was run and
evidence was recorded.
