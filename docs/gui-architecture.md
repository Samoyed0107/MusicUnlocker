# GUI architecture

Stage 5A adds a Wails v2 desktop shell without moving business logic into the
frontend:

```text
React + TypeScript -> Wails bindings -> internal/app -> internal/batch -> internal/decode
```

`internal/app` owns the GUI contract and the single active batch lifecycle.
It maps core results into explicit DTOs and never exposes upstream decoder
types. The frontend only holds selected input paths, renders DTOs, and maps
machine statuses to labels. Discovery, output planning, collision handling,
and scheduling remain in Go.

## DTOs and lifecycle

The facade exposes `GetAppInfo`, `SelectFiles`, `SelectFolder`,
`SelectOutputFolder`, `DescribeInputs`, `SuggestOutputFolder`,
`GetPreferences`, `SavePreferences`, `OpenOutputFolder`, `StartBatch`,
`CancelBatch`, and `GetPendingLaunchPaths`. `BatchStartRequest`, `QueueItem`, `ProgressEvent`,
`BatchSnapshot`, and `BatchCompleteEvent` are the stable JSON-shaped contract.

Each start receives a monotonically increasing `batchId`. Every transient
event carries that ID, allowing the frontend to ignore events from an older
generation. The facade rejects a second active start with
`BATCH_ALREADY_RUNNING`. Cancellation calls the same context used by
`batch.Engine.Run`; completion waits for the engine to return and emits one
authoritative `musicunlocker:batch:complete` event.

Events are:

- `musicunlocker:batch:discovered`
- `musicunlocker:batch:started`
- `musicunlocker:batch:progress`
- `musicunlocker:batch:finished`
- `musicunlocker:batch:complete`
- `musicunlocker:launch:paths`

The first four carry a batch ID and item data (progress also carries byte
counters). The complete event carries the final ordered result list and
summary counters; it is authoritative rather than reconstructed from live
events.

## Queue, dialogs, and progress

The input queue contains only user-selected files and directories. It is kept
distinct from discovered files: recursive discovery begins only after Start,
and discovered rows use the stable batch index from `internal/batch`.

The worker selector currently accepts 1, 2, 4, or 8 jobs and the backend
validates the range 1..8. Progress displays bytes written. Source byte size is
metadata and is not presented as an exact decoded-output percentage.

Wails runtime dialogs support multi-file selection, folder selection, and
output-folder selection. Dialog cancellation returns an empty selection and
is not an application error. Stage 5B enables Wails v2's native
`options.DragAndDrop{EnableFileDrop:true, DisableWebViewDrop:true}` and the
frontend registers `runtime.OnFileDrop(callback, true)`. The drop zone uses
`--wails-drop-target: drop`; Wails supplies absolute OS paths and applies its
active drop-target class. The listener is removed with `OnFileDropOff` during
React cleanup. Drop adds queue entries only and never starts a batch.

## Queue and output policy

Files, folders, and native drops share one path-deduplicating queue helper.
Windows case and slash differences are treated conservatively as duplicates.
Folders remain one row until discovery begins, and the bounded queue displays
the type, basename, and shortened full path. Selection and drop controls are
locked while a batch is running or cancelling; drops during that time are
ignored with an inline message. Paths supplied to the executable use the same
queue helper and are never started automatically. Initial process arguments are
retained until `GetPendingLaunchPaths` is called; later launches use Wails
`SingleInstanceLock` and are deferred while a batch is running or cancelling,
then flushed after authoritative completion.

When auto-suggestion is enabled and output has not been explicitly selected,
one file suggests `<source directory>\Unlocked`; one directory suggests
`<parent>\<directory name> - Unlocked`. Multiple inputs are treated as
ambiguous and leave output unset. Browse marks the destination as an explicit
override; it is not changed by later queue edits. Reset suggestion clears the
override.

Preferences contain only `MaxWorkers`, `Theme`, `AutoSuggestOutput`,
`NamingMode`, `CustomNamingTemplate`, and stable `Language` (`SYSTEM`,
`zh-CN`, or `en-US`) and
are stored as JSON at the per-user `os.UserConfigDir()/MusicUnlocker/preferences.json`.
The file is bounded, unknown fields are ignored, invalid values use defaults,
malformed files do not crash the app, and writes use a same-directory
temporary file followed by rename (with a Windows replacement fallback).

## Presentation and controls

Machine statuses retain their backend values while the UI maps them to
Completed, Not encrypted, Already exists, Key required, Database required,
Unsupported, Unknown format, Failed, and Cancelled. Known error codes have
concise explanations in expandable row details. Results use stable Index
ordering and local All/Completed/Skipped/Blocked/Problems filters. Progress
uses actual bytes written; any source-size bar is explicitly secondary and
does not show a decoded percentage. Summary counters come from the final
complete DTO and include Completed, Skipped, Blocked, Failed/Cancelled, and
Written groups.

Open Output Folder is a backend operation. It accepts only an absolute path
that exists as a directory or file, resolves a file to its parent, and passes
the validated path as a separate argument to `explorer.exe`; the frontend
cannot run shell commands.

The UI follows the system `prefers-color-scheme`, with a light default and a
dark neutral surface treatment. It uses Segoe UI/system fonts, semantic
buttons, visible keyboard focus, text status labels, disabled states, and a
bounded processing-list scroll region. No explicit theme selector is exposed
in this stage. The frontend uses a small `LocalizationProvider` and typed
translation hook; locale changes are presentation-only and do not alter batch,
preview, queue, naming, or output state.

## Manual Stage 5B GUI checklist

1. Run `wails dev` and confirm the drop zone renders.
2. Drop one file, several files, and one folder; confirm each remains a queue
   row, duplicates are ignored, and no batch starts automatically.
3. Confirm a single-file and single-folder output suggestion, then Browse and
   verify the explicit destination is preserved.
4. Start a harmless synthetic batch, inspect filters, status details, output
   folder action, and cancellation.
5. Restart and confirm only worker count/output-suggestion preference returns;
   inputs and job history do not persist.

## Current limitations

There is no pause/resume, overwrite mode, source deletion, online external-key
acquisition, KGG database loading, metadata/network access, FFmpeg integration,
tray support, installer packaging, file association, or explicit theme
selector. Physical Explorer drag-to-EXE still requires manual verification in
a real Windows installation; automated tests do not perform WebView E2E or
decode real music.

## Stage 5E2 file naming

The facade exposes `NamingConfigDTO` and `ValidateNamingConfig`. The latter
returns validity, a concise backend error, and a sample filename rendered by
`internal/naming`; React does not parse templates. `BatchStartRequest` carries
the validated naming snapshot, and result DTOs preserve the authoritative
final output path plus structured metadata-fallback fields. The GUI disables
naming controls while a batch runs or cancels. Preferences default and recover
to `KEEP_SOURCE` when naming fields are missing or invalid.

## Stage 5C audio preview

Successful completed rows can request a media URL through the application
facade using `BatchID + queue Index`. The application retains only the latest
completed snapshot and validates that the authoritative result is `SUCCESS`,
has a non-empty output path, and still resolves to a regular file.

The application owns a short-lived HTTP service bound explicitly to
`127.0.0.1` on an OS-assigned port. URLs contain only unpredictable,
in-memory capability IDs; the frontend can never provide a filesystem path to
the endpoint. The registry is cleared when a new batch starts and on process
shutdown, so tokens do not persist. Every request reopens and revalidates the
authorized path, uses conservative `nosniff`/`no-store` headers, and serves
only `GET`/`HEAD` through `http.ServeContent`, including byte ranges for seek.
No directory browsing or arbitrary static-file route exists.

The UI has one native HTML audio element and one compact player. A new batch
stops and releases the old source; stale asynchronous URL responses are
discarded. Play is offered only for successful rows. A missing output or an
unsupported WebView codec makes preview unavailable without changing the
decoder result, which remains `SUCCESS`; Open Output Folder remains available.
Preview is native browser playback only and does not add metadata networking,
FFmpeg, transcoding, waveform, or cover-art support.

## Stage 6B external-key retry

`RetryQMCWithExternalKey(batchID, itemIndex)` is the only GUI retry operation.
It validates the retained completed-batch session and item eligibility before
opening a backend-owned native file dialog. The dialog returns a path only to
Go; Go reuses `externalmaterial.ReadEKeyFile`, builds a request-scoped provider,
and calls the normal `batch.Engine` with the exact source path and captured
output/naming configuration. React receives only retry state and the ordinary
non-secret result snapshot.

Cancellation or malformed input leaves the row unchanged and performs no
decode. A supplied-key decode failure becomes a cautious retryable failure,
not `QMC_EXTERNAL_KEY_REQUIRED`. On success, the result replaces the original
queue index, summary counts are recomputed, and the existing preview path is
usable through the normal `GetPreviewURL` authorization. The native dialog
title is currently a stable backend string rather than frontend-localized;
this avoids coupling native dialog code to React localization.
