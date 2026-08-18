# Product requirements

## Metadata-based decoded-output naming

Status: `PLANNED / NOT IMPLEMENTED` (Stage 5E or equivalent).

Decoded outputs should eventually support selectable naming presets such as
cleaned source name, Title, Artist - Title, Title - Artist, and Track - Title,
as well as a custom template with `{title}`, `{artist}`, `{album}`,
`{albumartist}`, `{track}`, `{disc}`, `{year}`, and `{source}` placeholders.

The future implementation must keep the decoded extension under actual audio
format control, work offline by default, sanitize Windows filenames, use a
safe fallback for missing tags, preserve no-overwrite behavior, and retain
deterministic batch collision semantics. Metadata may only be available after
audio decryption, so this requirement needs coordinated output-planning work.

No naming templates, tag parsing, metadata networking, or output naming change
is implemented in Stage 5D.
