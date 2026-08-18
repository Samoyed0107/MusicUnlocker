# Product requirements

## Localized GUI language resources

Status: `IMPLEMENTED` (Stage 5F).

MusicUnlocker supports system-following GUI language selection plus Simplified
Chinese (`zh-CN`) and English (`en-US`). Resources are bundled and the
architecture is extensible to future mainstream languages, but only these two
production languages are implemented now. Localization is presentation-layer
behavior; machine status/error codes, naming modes, custom template tokens,
filenames, and metadata remain stable and untranslated.

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
