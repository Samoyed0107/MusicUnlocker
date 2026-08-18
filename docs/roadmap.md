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
queue-ordered post-decode finalization; its GUI exposure is delivered in 5E2.

## Completed: Stage 5E2

Expose naming modes and custom templates through GUI preferences and controls.
The GUI uses backend validation and examples, keeps KEEP_SOURCE as the default,
captures naming immutably per batch, and leaves real external-audio validation
pending.

## Completed: Stage 5F

Add bundled GUI localization with live `SYSTEM`, `zh-CN`, and `en-US`
selection, persisted stable preference values, English fallback, and typed
resource parity validation. Machine statuses, error codes, naming modes, and
template tokens remain unchanged.
