# Roadmap

## Completed

- Stage 5A: Wails GUI shell and application facade.
- Stage 5B: queue, native in-window file/folder drop, and batch lifecycle.
- Stage 5C: secure output preview, playback/seek, and real NCM GUI validation.
- Stage 5D: Windows launch arguments and Wails single-instance handoff are
  implemented; physical drag-to-EXE validation is still pending.

## Completed: Stage 5E1

Add offline metadata-based decoded-output naming. Planned presets include
cleaned source name, Title, Artist - Title, Title - Artist, and Track - Title,
plus a sanitized custom template using placeholders such as `{title}`, `{artist}`,
`{album}`, `{albumartist}`, `{track}`, `{disc}`, `{year}`, and `{source}`.

Stage 5E1 implements the offline core, sanitization, templates, fallback, and
queue-ordered post-decode finalization. GUI controls are not exposed yet.

## Planned: Stage 5E2

Expose naming modes and custom templates through GUI preferences and controls.
