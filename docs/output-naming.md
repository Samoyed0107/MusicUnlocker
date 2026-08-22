# Metadata-aware output naming

Stage 5E1 provides offline core naming. Stage 5E2 exposes the same naming
configuration through the desktop GUI. Stage 5H adds a metadata quality guard
between extraction and naming.

Modes are `KEEP_SOURCE` (default), `TITLE`, `ARTIST_TITLE`, `TITLE_ARTIST`,
`TRACK_TITLE`, `ARTIST_ALBUM_TRACK_TITLE`, and `CUSTOM`. Custom templates accept
only `{title}`, `{artist}`, `{album}`, `{albumartist}`, `{track}`, `{disc}`,
`{year}`, and `{source}`. `{source}` is the existing output planner's cleaned
encrypted-source stem. Templates cannot select an extension or create directories.

Metadata is read offline from two local sources when available: decoded-audio
tags and NCM container metadata. Decoded-audio metadata remains primary; NCM
fills only missing Title, Artist, and Album fields. The NCM provider is the only
container metadata provider currently implemented. It reads no network data,
does not read cover art, and does not write metadata into decoded audio.

Decoded-audio tags are read after successful decryption from the temporary
decoded audio. The reader handles ID3v2 MP3, FLAC/Vorbis comments, Ogg/Vorbis
comments, and common MP4/M4A atoms. Unsupported, absent, or malformed metadata
uses the complete cleaned source stem; this does not fail decoding.

Before merge, the quality guard evaluates Title, Artist, Album, and Album Artist
independently. It rejects invalid UTF-8, U+FFFD, control characters, excessive
non-printable content, abnormal symbol ratios, empty/whitespace-only values,
and suspiciously low human-readable ratios. The policy works on Unicode
character classes rather than an ASCII allowlist, preserving Chinese,
Japanese, Korean, accented Latin text, combining marks, and punctuation and
symbols commonly used in music titles.

Each candidate retains origin provenance (embedded decoded, container, or
filename) and provider provenance (generic, NCM, or KGM). Internal diagnostics
contain only field, rejection reason, and provenance; they never contain the
rejected value, encrypted material, keys, or database content. The naming
package receives only the already validated and merged metadata.

Fallback is per field. Valid decoded fields remain primary, valid container
fields fill rejected or missing decoded fields, and the cleaned source stem is
used as filename-derived Title when no trusted title remains. Thus a bad title
does not discard a valid artist. If the requested template still lacks another
required field, the existing naming fallback uses the complete cleaned source
stem. Metadata rejection never changes decode success or decoded bytes.

Generated stems replace Windows-invalid/control characters, trim trailing spaces
and periods, protect reserved device names, preserve Unicode, and are limited to
240 Unicode code points. This is a component bound, not a full-path guarantee.
The final extension always comes from decrypted-content sniffing; container
metadata and naming templates never control it.

Decode uses two phases: bounded-concurrency workers write and close unique temp
files in the destination directory; final naming and no-replace hard-link commit
are ordered by queue index. Thus identical metadata names are owned by the
earliest queue item, independent of worker completion timing. Cancellation and
failed finalization remove uncommitted temporary files.

## Desktop GUI

The File naming control offers Keep source name (the default), Title, Artist -
Title, Title - Artist, Track - Title, Artist - Album - Track - Title, and
Custom template. Custom templates use exactly the tokens listed above. The
decoded audio extension is added automatically; the GUI does not choose a
format or permit directory separators.

The custom field is validated by the Go naming backend. Unknown placeholders,
unclosed braces, empty templates, and overlong templates are shown inline and
cannot be used to start a batch. The GUI shows a backend-rendered illustrative
Example using neutral sample metadata; it is not a prediction of encrypted
input metadata. The finalized Batch Result output path remains authoritative.

Naming mode and the last valid custom template are persisted alongside worker,
theme, and output-suggestion preferences. Missing or invalid naming fields
recover to KEEP_SOURCE without preventing startup. Naming settings are
captured in each Batch Request, so later preference changes do not affect a
running batch. Metadata fallback remains a successful decode.
