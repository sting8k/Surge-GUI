# macOS Visibility

Surge Desktop uses close-to-hide behavior on macOS.

## Modes

### Dock Mode

- The app appears in the Dock and app switcher.
- Closing the window hides it instead of quitting the app.
- Launching the app again should reveal, unminimize, focus, and activate the
  existing window.
- Switching back to dock mode should remove the menu bar item and reveal the
  main window.

### Menu Bar Mode

- The app uses a menu bar status item and does not show a Dock icon.
- Closing the window hides it instead of quitting the app.
- The menu bar item's **Open Surge** action should reveal, unminimize, focus, and
  activate the existing window.
- Launching the app again while it is already running should reveal the same
  existing window.
- Switching to menu bar mode should keep the current window visible and focused;
  it should not unexpectedly hide the app.

## Implementation Rule

All macOS reveal paths should use the same native helper so behavior stays
consistent across:

- Menu bar **Open Surge**.
- Second-instance launch.
- Dock/menu bar mode switching.

The desktop app should not create a second window just to recover from hidden or
minimized state.
