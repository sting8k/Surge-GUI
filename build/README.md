# Build Assets

This directory contains Wails build assets and platform-specific packaging files
for Surge Desktop.

## Structure

- `appicon.png`: source app icon used by Wails builds.
- `darwin/`: macOS plist files for development and production builds.
- `windows/`: Windows manifest, icon, installer, and metadata files.

## macOS

`build/darwin/` contains:

- `Info.plist`: production macOS plist used by `wails build`.
- `Info.dev.plist`: development macOS plist used by `wails dev`.

macOS is the currently documented platform surface. Platform behavior should be
validated before release claims are made.

## Windows

`build/windows/` contains packaging files generated or used by Wails:

- `icon.ico`: Windows application icon.
- `info.json`: application metadata for Windows builds and installer details.
- `wails.exe.manifest`: Windows application manifest.
- `installer/`: NSIS installer files.

Windows packaging is present, but product-level Windows support should not be
claimed without platform validation evidence in `../docs/TEST_MATRIX.md`.
