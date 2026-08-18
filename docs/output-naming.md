# Metadata-aware output naming

Stage 5E1 provides offline core naming. Stage 5E2 exposes the same naming
configuration through the desktop GUI.

Modes are `KEEP_SOURCE` (default), `TITLE`, `ARTIST_TITLE`, `TITLE_ARTIST`,
`TRACK_TITLE`, `ARTIST_ALBUM_TRACK_TITLE`, and `CUSTOM`. Custom templates accept
only `{title}`, `{artist}`, `{album}`, `{albumartist}`, `{track}`, `{disc}`,
`{year}`, and `{source}`. `{source}` is the existing output planner's cleaned
encrypted-source stem. Templates cannot select an extension or create directories.

Metadata is read after successful decryption, offline, from the temporary
decoded audio. The reader handles ID3v2 MP3, FLAC/Vorbis comments, Ogg/Vorbis
comments, and common MP4/M4A atoms. Unsupported, absent, or malformed tags use
the complete cleaned source stem; this does not fail decoding.

Generated stems replace Windows-invalid/control characters, trim trailing spaces
and periods, protect reserved device names, preserve Unicode, and are limited to
240 Unicode code points. This is a component bound, not a full-path guarantee.
The final extension always comes from decrypted-content sniffing.

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
