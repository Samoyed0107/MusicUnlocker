# Metadata-aware output naming

Stage 5E1 provides offline core naming; GUI controls remain Stage 5E2.

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
