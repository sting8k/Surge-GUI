# Download Flow

Surge Desktop uses a backend-backed download flow.

## Ownership

Surge backend owns:

- Download scheduling.
- Active worker/concurrency decisions.
- Per-download connections.
- Pause, resume, delete, and URL update execution.
- Persistence of download state and history.

Surge Desktop owns:

- URL intake and validation before sending requests to Surge.
- Optional user-provided filename, destination path, and mirrors.
- Normalizing backend statuses for UI consistency.
- Queue, active, completed, and error visibility.
- Bulk control commands that call Surge APIs for each matching item.
- Recovery UX for failed or expired URLs.
- Validation evidence for supported flows.

## Flow

```text
User input
  -> desktop validation
  -> Surge API request
  -> Surge queue/scheduler
  -> Surge status/events
  -> desktop normalization
  -> UI tabs, details, bulk controls, recovery actions
```

## Distribution Inputs

The desktop app should expose only distribution inputs the backend already
supports. Today that means mirrors on add:

```json
{
  "url": "https://example.com/file.zip",
  "mirrors": [
    "https://mirror1.example.com/file.zip",
    "https://mirror2.example.com/file.zip"
  ]
}
```

The desktop app must not fake priority, worker limits, or queue reordering unless
Surge exposes stable API support for those behaviors.

## Management States

The desktop UI treats backend statuses as these canonical states:

- `queued`
- `downloading`
- `paused`
- `completed`
- `error`

Unknown statuses may be displayed, but they should not be used to claim support
for a new management flow until documented and validated.

## Current Non-Goals

- Do not implement a second scheduler in the desktop app.
- Do not persist an app-side queue that can diverge from Surge.
- Do not claim active download limit or priority support until the backend API is
  confirmed.
