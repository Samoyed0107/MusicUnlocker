# Batch engine

`internal/batch` is the UI-independent orchestration layer for the developer
CLI and future GUI. It depends on `internal/decode`, probing, and the output
planner; it does not import Wails, frontend code, or upstream decoder types.

## Execution model

`Engine.Run` accepts files/directories, an explicit `OutputRoot`, and
`MaxWorkers`:

- `0` means one worker for API compatibility;
- `1` preserves sequential Stage 4A execution;
- values greater than one use a fixed bounded worker pool;
- negative values are invalid.

The scheduler never creates one goroutine per file. It creates at most
`min(MaxWorkers, runnable target-key groups)` workers. Independent target keys
can overlap; items sharing a target key are processed by one ordered group.

Discovery remains recursive, deterministic, regular-file-only, canonical-path
deduplicated, and symlink-loop-safe. Normal raw audio is skipped and unknown
content is not decoded.

## Deterministic results and events

Every queue item has a stable zero-based `Index`. `Summary.Results` is assembled
from the indexed queue after all workers have joined, so it remains in
discovery order even when completion order differs. Runtime `FINISHED` event
order is intentionally nondeterministic under concurrency.

The optional observer receives structured events:

| Event | Contract |
| --- | --- |
| `DISCOVERED` | Emitted in deterministic discovery order. |
| `STARTED` | Identifies the item beginning work; events may overlap. |
| `PROGRESS` | Contains one item's stable index and bytes written so far. |
| `FINISHED` | Contains the final item result; completion order is not promised. |

Observer errors and panics are isolated. No event contains preformatted GUI
strings or percentages.

## Progress

`internal/decode.Options.Progress` accepts a UI-independent
`decode.Progress` callback. It reports:

- `BytesWritten`: bytes successfully written to the atomic temporary output;
- `SourceBytes` and `SourceBytesKnown`: optional source file size metadata.

Source size is not an exact decoded-output total. The decoder emits an initial
zero progress state and then bounded copy-chunk updates. Values are monotonic;
successful completion's final callback matches `decode.Result.BytesWritten`.
Callback panics are ignored and cannot corrupt output. Cancellation or stream
failure does not invent a successful final total.

Batch bridges these callbacks to `PROGRESS` events with queue index, source
path, bytes written, and optional source size. Progress events from different
items may interleave.

## Output paths and collisions

One directory root preserves relative directories. Multiple roots retain root
basenames; colliding basenames receive deterministic suffixes such as
`Music (1)` and `Music (2)`. Direct-file basename collisions use deterministic
parent namespaces.

Before decoding, the batch derives a target key from the destination directory
and the output planner's clean wrapper-stripped stem. The decoded extension is
deliberately omitted. Same-key items form an ordered group, so the earlier
queue item always gets the first opportunity. If it creates the same final
destination, a later item becomes `SKIPPED_OUTPUT_EXISTS`; if it uses another
extension, it may succeed after waiting.

Metadata naming obtains bounded local container metadata before decoding when a
provider is available, then prepares temporary files concurrently and waits on a
queue-index commit gate before resolving the final destination. Decoded tags are
merged with container Title/Artist/Album values only after the temporary decoded
file is closed; the decoder finalizes with an
atomic no-replace hard-link operation from its temporary file. A destination that appears during finalization maps to
`OUTPUT_EXISTS`; it is never silently replaced. The temporary file is cleaned
on failure. This uses the standard library and assumes the temporary and final
paths share a filesystem, as they do because both are inside the output
directory.

## Statuses and discovery errors

`DISCOVERY_ERROR` is not a separate status. It is the machine-readable
`ErrorCode` on a `FAILED` result for a top-level or walked filesystem item.
Such results are retained in `Summary.Results`, counted as failures, rendered
distinctly by the CLI, serialized distinctly in JSON, and produce a nonzero
exit code.

The statuses are:

| Status | Meaning |
| --- | --- |
| `SUCCESS` | Decoded and atomically written. |
| `SKIPPED_NOT_ENCRYPTED` | Confirmed normal raw audio. |
| `SKIPPED_OUTPUT_EXISTS` | Existing destination; equivalent to `OUTPUT_EXISTS`. |
| `BLOCKED_EXTERNAL_KEY` | `QMC_EXTERNAL_KEY_REQUIRED`. |
| `BLOCKED_EXTERNAL_DATABASE` | `KGG_DATABASE_REQUIRED`. |
| `UNSUPPORTED` | Unsupported encrypted decoder condition. |
| `UNKNOWN` | Content or decoded audio format is unknown. |
| `FAILED` | Discovery or other processing failure. |
| `CANCELLED` | Not started or interrupted by context cancellation. |

## Cancellation

Cancellation stops dispatching new groups, propagates the same context to all
active `DecodeFile` calls, waits for workers to join, and marks undispatched
items and same-key waiters `CANCELLED`. Successful results and outputs remain
untouched. `Run` does not return while workers can still mutate the filesystem.

Pause/resume is intentionally not implemented.

Stage 6A adds an application-owned provider to `batch.Request`. It resolves by
source path only when probing says external QMC material is needed and may be
called concurrently; the built-in in-memory provider is safe for concurrent
reads. Material is never copied into `Result` or `Summary`. The CLI deliberately
does not expose a single-key-for-all-files mapping; richer per-file CLI mapping
is deferred to Stage 6B.

The GUI's Stage 6B retry invokes this same engine for a single source path with
the captured output root, naming configuration, and a request-scoped provider.
It does not rerun the original batch or mutate other queue items.

## Developer CLI

```text
go run ./cmd/musicbatch --out-dir D:/Unlocked --jobs 1 input-dir
go run ./cmd/musicbatch --out-dir D:/Unlocked --jobs 2 input-dir another.kgm
go run ./cmd/musicbatch --out-dir D:/Unlocked --jobs 2 --json input-dir
```

The default is one worker. Human output is printed from final ordered results,
not worker completion logs. JSON uses the explicit `encoding/json` Summary
model; transient progress events are not serialized. Exit codes are `0` when
there are no `FAILED` results, `1` for failed results, `2` for invalid CLI
requests, and `3` for cancellation. Skips and blocked external material are
non-fatal.

Stage 4B does not change `docs/real-validation.md`, add decoder families,
acquire external keys/databases, use FFmpeg/network metadata, or require real
music samples. There is no GUI or live CLI progress display yet.
