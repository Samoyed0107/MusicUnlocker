# GUI architecture

MusicUnlocker's desktop frontend is a presentation and state-mapping layer over
one shared Go core:

```text
React + TypeScript
    -> generated Wails bindings
    -> internal/app
    -> internal/batch
    -> internal/decode
```

`internal/app` owns the Wails-facing contract and maps core results to stable
JSON-shaped DTOs. It never exposes upstream decoder types or Go errors
directly. Filesystem access, discovery, probing, scheduling, decoding, output
planning, external-material selection, and finalization remain Go-owned.

## Core subsystem boundaries

- `internal/probe` performs bounded, content-first format probing. Extensions
  supply candidate hints only.
- `internal/decoder/upstream` is the sole adapter boundary to the pinned
  canonical decoder module.
- `internal/audiofmt` identifies the decrypted audio format; naming and
  encrypted filename suffixes cannot choose the final extension.
- `internal/outputplan` and `internal/naming` provide conservative source
  cleanup, validated naming presets/templates, and no-overwrite planning.
- `internal/audiometa` and `internal/containermeta` read offline metadata after
  decode or from the supported NCM container fields.
- `internal/metadataquality` validates Title, Artist, Album, and Album Artist
  independently before naming.
- `internal/externalmaterial` represents QMC EKeys and KGG databases as typed,
  request-scoped backend capabilities.
- `PreviewService` authorizes successful outputs with opaque, in-memory IDs.
- `internal/app/launch.go` and the Wails single-instance callback queue launch
  paths without starting work.

## Application facade and lifecycle

The facade exposes application identity, native file/folder/output dialogs,
input descriptions and output suggestions, preferences, naming validation,
batch start/cancel, launch-path retrieval, output-folder opening, secure preview,
and QMC/KGG retry operations.

The GUI permits one active batch. Each start receives a monotonically increasing
`batchId`. Transient events carry that ID so React can ignore older
generations. A second start is rejected with `BATCH_ALREADY_RUNNING`.
Cancellation propagates through the same context used by the batch engine and
does not return completion until worker and temporary-file cleanup finishes.

The event contract includes:

- `musicunlocker:batch:discovered`
- `musicunlocker:batch:started`
- `musicunlocker:batch:progress`
- `musicunlocker:batch:finished`
- `musicunlocker:batch:complete`
- `musicunlocker:launch:paths`

Live events support presentation only. The final `BatchComplete` result,
including its ordered rows and summary counters, is authoritative.

## Selected inputs, discovery, and launch paths

Selected inputs are user-selected files and directories. They remain distinct
from discovered files: recursive Go discovery begins only when the user presses
Start, and each discovered result receives a stable zero-based batch index.
Queue rows correlate by that index even when workers finish out of order.

Wails owns native multi-file, folder, and output-directory dialogs. Native
in-window file/folder drop supplies OS absolute paths; the frontend only adds
those paths to the selected-input queue. Drop never starts decoding.

Existing files or directories passed on process startup use the same queue
helper. Initial arguments are retained until the frontend asks for them; later
launches arrive through Wails `SingleInstanceLock`. Paths received while a
batch is running are deferred until authoritative completion. Startup
arguments, Explorer drag-to-EXE, and future file associations may queue inputs,
but they must always wait for explicit user Start.

## Batch, progress, naming, and output

The backend validates 1-8 workers. `internal/batch` uses bounded concurrency,
recursive deterministic discovery, stable result ordering, and serialized
queue-order ownership for items that share an output target namespace.
Per-file failure does not terminate the batch.

Progress reports bytes written and optional source size; it does not claim that
encrypted source size is an exact decoded-output percentage. Cancellation waits
for active workers and filesystem cleanup.

Naming configuration is validated and rendered by the Go `internal/naming`
backend. React does not implement a second template parser. Each batch captures
its validated naming configuration immutably. Decoded embedded metadata remains
primary, trusted NCM container metadata fills missing fields, and the metadata
quality guard rejects suspicious fields independently. Safe fallback remains a
successful decode.

The decrypted stream is content-sniffed before output planning. Workers write
and close temporary files before metadata naming, and final destinations are
committed in deterministic queue-index order with no replacement. Sources are
not deleted or rewritten.

## Secure preview

A successful completed row may request a preview URL using only
`batchId + stable index`. The frontend cannot provide an arbitrary local path
to `PreviewService`.

The application validates the authoritative result and registers its output
under an unpredictable, in-memory capability ID. A loopback HTTP service bound
to `127.0.0.1` serves only authorized `GET`/`HEAD` requests with byte-range
support, `nosniff`, and `no-store`. It has no directory browser or arbitrary
static-file route. Capability registrations are cleared when a new batch starts
and at shutdown.

React owns one native HTML audio element. Missing output, token expiry, or an
unsupported WebView codec makes preview unavailable but never changes a decoder
`SUCCESS` result into failure. Starting a new batch releases the prior preview;
language and theme changes must not restart preview or mutate batch semantics.

## External-material retries

Blocked QMC STag/MusicEx rows can invoke
`RetryQMCWithExternalKey(batchID, itemIndex)`. Go validates the completed
session and row before opening a native EKey file chooser. The selected path
and raw material never enter React, DTO results, preferences, or logs.

Blocked KGM V5/KGG rows can invoke
`RetryKGGWithDatabase(batchID, itemIndex)`. Go opens the database chooser,
performs bounded regular-file validation, and passes a read-only,
request-scoped database capability through the normal batch/decode path. The
database path never enters React or persisted preferences.

Both retries use the captured source, output root, and naming configuration.
They act on the same stable result row, recompute the authoritative summary,
and reuse normal atomic finalization and preview authorization. Cancellation or
pre-decode validation failure leaves the row unchanged. QMC EKey and KGG
database requirements remain separate typed states.

## Preferences, localization, theme, and About

Preferences are small per-user JSON containing worker count, theme,
output-suggestion mode, naming configuration, and language. They contain no
input history, raw filenames, keys, external-material paths, or database
content. Reads are bounded and invalid values recover to safe defaults; writes
use a same-directory temporary file and replacement.

Bundled localization supports `SYSTEM`, `zh-CN`, and `en-US`, with English as
the complete fallback. User-visible presentation strings are localized;
machine statuses, error codes, naming modes, and template tokens are stable and
untranslated.

Theme preferences use `SYSTEM`, `LIGHT`, and `DARK`. React applies semantic
tokens and asks Wails to coordinate the native title bar. Theme changes are
presentation-only: they do not recreate the batch, selected queue, result
snapshot, retry state, or preview element. Short transitions respect
`prefers-reduced-motion`.

The localized modal About dialog obtains application name, version, build
timestamp, and build identifier through `GetAppInfo`, backed by
`internal/version`. Stage 7A release identity does not imply installer,
association, or updater support.

## Current boundaries

There is no pause/resume, overwrite mode, source deletion, network key
acquisition, online metadata, FFmpeg transcoding, tray integration, installer,
file association, or automatic updater. QMC supplied-key and KGM V5/KGG
database-backed real decoding remain pending compatible external material.
Other real-file decoder coverage is tracked in `docs/real-validation.md`.

Native Mica is not enabled. Stage 7D requires frontend glass content surfaces
with an opaque fallback; Mica remains an optional Windows experiment rather
than an architectural dependency.
