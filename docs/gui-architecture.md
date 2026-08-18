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
`SelectOutputFolder`, `StartBatch`, and `CancelBatch`. `BatchStartRequest`,
`QueueItem`, `ProgressEvent`, `BatchSnapshot`, and `BatchCompleteEvent` are
the stable JSON-shaped contract.

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
is not an application error. Native drag/drop is deferred to Stage 5B because
the installed Wails path does not provide a small, already-integrated drop
surface in this shell; the frontend does not scan folders itself.

## Current limitations

There is no pause/resume, persistence, overwrite mode, source deletion,
external-key acquisition, KGG database loading, metadata/network access,
FFmpeg integration, tray support, or installer packaging. The GUI has no
real-file automated smoke test; manual testing remains intentionally separate
from automated tests.
