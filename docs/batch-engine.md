# Sequential batch engine

Stage 4A adds `internal/batch`, a UI-independent orchestration layer. Future
CLI and Wails callers depend on this package; it depends on `internal/decode`
and the existing probe/output-planning core. It does not import Wails,
frontend code, CLI formatting, or upstream decoder packages.

`Engine.Run` accepts one or more files or directories and an explicit
`OutputRoot`. Directory inputs are traversed recursively and processed in
deterministic path order. Stage 4A is deliberately sequential: there is no
worker pool or byte-level progress.

## Discovery and paths

Only regular files are selected. Directory symlinks/reparse-point-like links
are not followed, and top-level symlinks are reported as discovery failures.
Canonical input files are deduplicated. Missing top-level inputs, permission
errors, and other scan errors become structured `FAILED` results with
`DISCOVERY_ERROR`; they do not prevent other roots from being processed.

Content probing happens before decoding. Confirmed raw MP3/FLAC/etc. files are
`SKIPPED_NOT_ENCRYPTED`; unknown content is `UNKNOWN`. A misleading filename
does not override confirmed content evidence.

For one directory root, its root itself is omitted and relative subdirectories
are preserved:

```text
Encrypted/A/song.kgm -> Unlocked/A/song.mp3
```

For multiple directory roots, each root basename is retained. If root
basenames collide, canonical root paths are sorted and receive `Music (1)`,
`Music (2)`, and so on. Direct file inputs normally go directly under
`OutputRoot`; exact direct-file basename collisions use the source parent
basename as a namespace.

## Results and events

Each result includes source/root/relative paths, detected family and variant,
decoded format, bytes written, status, typed error code, and detail. The
summary retains ordered results and reconciled counts.

| Status | Meaning |
| --- | --- |
| `SUCCESS` | Decoded and atomically written. |
| `SKIPPED_NOT_ENCRYPTED` | Confirmed normal raw audio. |
| `SKIPPED_OUTPUT_EXISTS` | Existing output; equivalent to `OUTPUT_EXISTS`. |
| `BLOCKED_EXTERNAL_KEY` | `QMC_EXTERNAL_KEY_REQUIRED`. |
| `BLOCKED_EXTERNAL_DATABASE` | `KGG_DATABASE_REQUIRED`. |
| `UNSUPPORTED` | `UNSUPPORTED_ENCRYPTED_FORMAT`. |
| `UNKNOWN` | Content is not a recognised encrypted format. |
| `FAILED` | Discovery or other processing failure. |
| `CANCELLED` | Not started, or interrupted by context cancellation. |

An optional observer receives `DISCOVERED`, `STARTED`, and `FINISHED` events
with structured results. Observer errors and panics are isolated and cannot
affect processing or output files.

Cancellation stops starting new files, propagates to the current decoder, and
marks all remaining discovered files `CANCELLED`. Successful outputs already
written remain in place and sources are never deleted.

## Developer CLI

```text
go run ./cmd/musicbatch --out-dir D:/Unlocked input-dir another.kgm
go run ./cmd/musicbatch --out-dir D:/Unlocked --json input-dir
```

Human mode prints one concise line per result and a summary. JSON mode uses
the explicit `encoding/json` representation of `Summary`; statuses and typed
error codes are separate from human-readable details.

Exit code `0` means no `FAILED` results; skipped, blocked, unsupported, and
unknown results remain visible but are non-fatal. Exit code `1` means at least
one failed result, `3` means cancellation, and `2` means an invalid CLI
request.

Real-sample validation remains an independent backlog. Stage 4A does not
change any status in `docs/real-validation.md`, acquire external keys or
databases, add decoder families, or require copyrighted samples.
