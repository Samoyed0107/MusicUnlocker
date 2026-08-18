# Roadmap

## Completed

- Stage 5A: Wails GUI shell and application facade.
- Stage 5B: queue, native in-window file/folder drop, and batch lifecycle.
- Stage 5C: secure output preview, playback/seek, and real NCM GUI validation.
- Stage 5D: Windows launch arguments and Wails single-instance handoff are
  implemented; physical drag-to-EXE validation is still pending.

## Planned: Stage 5E

Add offline metadata-based decoded-output naming. Planned presets include
cleaned source name, Title, Artist - Title, Title - Artist, and Track - Title,
plus a sanitized custom template using placeholders such as `{title}`, `{artist}`,
`{album}`, `{albumartist}`, `{track}`, `{disc}`, `{year}`, and `{source}`.

This is `PLANNED / NOT IMPLEMENTED`. The decoded extension must continue to be
determined by the actual decoded audio format; no metadata extraction,
template engine, or naming behavior is part of Stage 5D.
