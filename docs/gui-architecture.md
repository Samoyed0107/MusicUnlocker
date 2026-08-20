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

`GetAppInfo` also returns the centralized application version, build timestamp,
and build identifier. The localized About section renders these identity
fields only; it does not expose decoder internals.

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
`zh-CN`, or `en-US`). Theme uses stable `SYSTEM`, `LIGHT`, or `DARK` values.
Preferences are stored as JSON at the per-user
`os.UserConfigDir()/MusicUnlocker/preferences.json`.
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

The UI defaults to the persisted `SYSTEM` theme and follows
`prefers-color-scheme`, with explicit `LIGHT` and `DARK` overrides. It uses
Segoe UI/system fonts, semantic
buttons, visible keyboard focus, text status labels, disabled states, and a
bounded processing-list scroll region. The frontend uses a small
`LocalizationProvider` and typed
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
tray support, installer packaging, file association, or Mica backdrop.
Physical Explorer drag-to-EXE still requires manual verification in
a real Windows installation; automated tests do not perform WebView E2E or
decode real music.

## Stage 7B-A desktop workflow layout

Stage 7B-A changes presentation hierarchy only. `App.tsx` continues to own
batch ID filtering, running/cancelling state, retry state, preview selection,
preferences, and Wails event subscriptions. Presentation-only components split
the interface into a compact application header, responsive setup workspace,
batch command surface, and unified results workspace.

The setup workspace places file/folder selection and the selected-input queue
first, then output destination and backend-validated naming configuration;
the worker selector is in a native HTML `details` advanced section. At the
900px native minimum width the setup stacks vertically; at wider widths it uses
two columns. The command surface owns the existing ready/running/complete
label and Start action. Cancel is rendered only while a batch is running or
cancelling; this is presentation-only and does not change cancellation logic.

The results workspace keeps existing stable-index ordering, local filters,
row status calculation, retries, and single preview element. It now groups the
results title, filters, rows, final summary, one Open Output Folder action, and
preview within one card. The localized About identity fields are no longer a
permanent workspace card; a dialog is deliberately deferred to a later 7B
phase. No explicit theme support, native material, animation, or backend
contract change is included in this phase.

## Stage 7B-B visual system and accessibility

Stage 7B-B retains the Stage 7B-A component boundaries and changes only
frontend presentation. `style.css` defines semantic tokens for light-mode
surfaces, controls, text, borders, spacing, radius, typography, accent, and
status colors; the system dark-scheme media query supplies matching token
values for a later theme-selection phase. It does not add a preference control
or change the existing theme persistence behavior.

`StatusBadge` is the shared item-status presentation mapping used by result
rows. It leaves core status strings untouched and renders a localized label,
symbol, and semantic tone for completed, blocked, failed, skipped, neutral,
and waiting/running states. Result filters now expose `aria-pressed`, source
byte progress includes accessible values without claiming decoded percentage,
and standard button targets are at least 32px with visible focus styling.
Input and output paths remain visually truncated while their full values stay
available to assistive technology. Batch events, stable ordering, retry
eligibility, preview URL lifecycle, and Wails file-drop behavior are unchanged.

## Stage 7C Windows-native experience

`App.tsx` owns the theme preference alongside the existing language and naming
preferences. Loading normalizes legacy lowercase values to `SYSTEM`, `LIGHT`,
or `DARK`; applying a theme updates `document.documentElement.dataset.theme`,
the browser `color-scheme`, and the native Wails title-bar theme. Saving remains
part of the existing bounded preference object and is deferred while a batch is
locked, exactly like other presentation preferences. Theme changes do not alter
component identity or the batch, preview, retry, file-drop, or launch-path
lifecycle.

Light and dark semantic token sets remain in `style.css`. `SYSTEM` selects the
appropriate set through `prefers-color-scheme`; explicit modes override it.
Short color, border, control, and modal transitions provide presentation polish,
while `prefers-reduced-motion: reduce` collapses them. `Icon` supplies
maintainable, current-color inline SVG paths, and `StatusBadge` retains visible
localized text so the icon is supplemental rather than the sole status cue.

`AboutDialog` is a modal, localized surface opened from the compact header. It
uses the existing `GetAppInfo` DTO, backed by `internal/version`, for application
name and release identity. Supported-format and offline-operation copy stays in
the existing localization resources. Escape, the labeled close button, and
backdrop dismissal close the dialog; it is not persisted as workspace content.

The pinned Wails v2.15 Windows backend supports runtime title-bar theme changes
and offers `BackdropType=Mica` without Windows App SDK. Backdrop support requires
Windows 11 build 22621 or later and a translucent/transparent WebView/window
configuration. Stage 7C does not enable it because that composition change is
riskier than the presentation goal and would not improve older-Windows behavior.
Opaque semantic surfaces remain the fallback, the executable remains portable,
and no new native dependency is introduced.

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

## Stage 6D KGG database retry

`RetryKGGWithDatabase(batchID, itemIndex)` is the only GUI database retry
operation. The backend validates the retained completed-batch session, item
index, stored source path, and KGG-specific retry eligibility before opening a
native Windows file dialog. It reuses
`externalmaterial.ValidateKGGDatabasePath` and constructs a request-scoped
`KGGDatabaseProvider`; the database path is never a JavaScript argument and is
not included in DTOs, results, preferences, or logs.

The selected database is used only for the exact logical row. The captured
output root, naming configuration, source path, and stable queue index are
preserved. Cancellation or pre-decode validation failure does not mutate the
row. A compatible database with no matching entry remains a cautious,
independently retryable failure. QMC EKey and KGG database capabilities remain
separate typed backend paths.
