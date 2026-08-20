# Roadmap

## Completed

- Stage 5A: Wails GUI shell and application facade.
- Stage 5B: queue, native in-window file/folder drop, and batch lifecycle.
- Stage 5C: secure output preview, playback/seek, and real NCM GUI validation.
- Stage 5D: Windows launch arguments and Wails single-instance handoff are
  implemented; physical drag-to-EXE validation is still pending.

## Stage 7A: Release identity and build foundation

Stage 7A provides one centralized application version and link-time build
identity for the Wails GUI and `musicprobe`, `musicdecode`, and `musicbatch`.
The GUI exposes localized About information, and Wails production builds carry
the application product/version metadata in the Windows executable resource.
The build identity fields are timestamp and commit/build identifier; local
development defaults are explicit and release builds can inject them with Go
`-ldflags -X`.

Stage 7A does not include an installer, file associations, updater,
history/library, network metadata, or decoder changes.

## Completed: Stage 7B-A

Reorganize the desktop UI around the existing batch workflow: choose music,
configure destination and naming, start unlocking, then inspect results and
use preview/retry actions. The React layout uses responsive input/destination
setup cards, a dedicated batch command surface, and one results workspace for
filters, rows, summary, output-folder action, and preview. The native title
bar, Wails behavior, DTOs, batch lifecycle, preview security, retry providers,
filesystem handling, and naming semantics remain unchanged. This phase does
not add the planned About dialog, explicit theme support, a full design system,
or motion polish.

## Completed: Stage 7B-B

Add a Fluent-inspired frontend visual system without changing desktop behavior.
Semantic CSS tokens cover canvas, surfaces, controls, text, borders, accents,
and success/warning/danger states with dark-mode counterparts prepared through
the system color-scheme media query. Shared status badges present a localized
label plus a non-color symbol and semantic tone. Controls use consistent
hierarchy and 32px targets; filter buttons expose their selected state, progress
bars expose byte values, and visually truncated paths retain accessible full
values. Explicit theme selection, native materials, animation, and backend
changes remain future work.

## Implemented: Stage 7C

Add Windows-native theme coordination and focused Fluent refinement without
changing the established desktop workflow. The persisted theme preference uses
stable `SYSTEM`, `LIGHT`, and `DARK` values, defaults to `SYSTEM`, updates the
React semantic CSS tokens, and synchronizes the Wails native title-bar theme.
Legacy lowercase theme values are migrated on load.

Status glyph text is replaced by a shared inline SVG icon abstraction. Controls
retain visible keyboard focus, the localized About information is available in
a modal dialog rather than a permanent card, and subtle transitions respect
the system reduced-motion preference. The dialog reads application name and
version through the existing centralized release identity facade and describes
supported formats and offline operation.

Wails v2.15 exposes native title-bar themes and an optional DWM backdrop without
requiring Windows App SDK. Native title-bar theme switching is enabled. Mica is
not enabled because Wails requires Windows 11 build 22621 or later plus a
translucent/transparent window composition path; the current opaque surfaces
provide a predictable fallback and preserve the portable executable. Installer,
updater, networking, cloud, Web, Android, and decoder scope remain excluded.

## Completed: Stage 5E1

Add offline metadata-based decoded-output naming. Planned presets include
cleaned source name, Title, Artist - Title, Title - Artist, and Track - Title,
plus a sanitized custom template using placeholders such as `{title}`, `{artist}`,
`{album}`, `{albumartist}`, `{track}`, `{disc}`, `{year}`, and `{source}`.

Stage 5E1 implements the offline core, sanitization, templates, fallback, and
queue-ordered post-decode finalization; its GUI exposure is delivered in 5E2.

## Completed: Stage 5E2

Expose naming modes and custom templates through GUI preferences and controls.
The GUI uses backend validation and examples, keeps KEEP_SOURCE as the default,
captures naming immutably per batch, and leaves real external-audio validation
pending. A source-only CUSTOM template regression was fixed; real retest is
still pending.

## Completed: Stage 5F

Add bundled GUI localization with live `SYSTEM`, `zh-CN`, and `en-US`
selection, persisted stable preference values, English fallback, and typed
resource/interpolation parity validation. The `batch.itemsAdded` parity defect
was fixed, runtime mismatches fall back safely to English, and the corrected
production GUI was real-validated. Stage 5F2 playback continuity is
real-validated. Machine statuses, error codes, naming modes, and template
tokens remain unchanged.

## Completed: Stage 5F2

Preserve active preview media across language changes and move the batch status
badge into the batch action area. Automated and manual playback-continuity
validation are complete.

## Completed: Stage 5G

Offline NCM container Title/Artist/Album metadata now fills missing decoded
metadata for the existing naming engine. Stage 5G is `COMPLETE / REAL_PASS`;
evidence is recorded in `docs/real-validation.md`.

## Completed: Stage 6A

Core accepts request-scoped, per-source external QMC EKeys through an
application-owned provider. `musicdecode --qmc-ekey-file` reads the pinned
upstream Base64 representation with a bounded, offline, in-memory lifetime.
Batch Core carries a per-file provider, but `musicbatch` deliberately has no
one-key-for-all-files interface. GUI key entry is implemented in Stage 6B;
KGG/KGM V5 database support is implemented in Stages 6C–6D.

## Completed: Stage 6B

Blocked QMC rows can open a native Windows EKey file chooser and retry only
that item through the normal batch/decode pipeline. The original output and
naming snapshot is retained in memory for the active completed batch. Valid
file parsing reuses the Stage 6A bounded Base64 parser; malformed files do not
start decoding. Results replace the original queue index and summary counts
are recalculated. Keys and key paths are never persisted or sent to React.
Real supplied-key QMC GUI validation remains pending; Stage 6C core is
complete and its real database-backed validation remains pending.

## Completed: Stage 6C

KGM V5/KGG now accepts an explicit, per-source application-owned database
provider. The adapter uses the pinned upstream path-based helper and retains
KGM V3's self-contained path. Database paths are validated as bounded regular
files without opening SQLite in the source directory; upstream imports data
into memory. `musicdecode` and `musicbatch` accept `--kgg-database` for offline
operation. Missing databases remain blocked, while a supplied
database with no matching `AudioHash` is reported separately. Real KGM V5/KGG
decode remains pending.

## Completed: Stage 6D

`RetryKGGWithDatabase(batchID, itemIndex)` is a backend-only native database
chooser and per-item retry. It reuses Stage 6C bounded read-only validation and
the Stage 6B completed-batch snapshot, while keeping KGG database capability
typed separately from QMC EKeys. Database paths are request-scoped and never
reach React or preferences. A selected row is replaced at the same queue
index, summaries are recomputed, normal naming/finalization and preview are
used, and cancellation or invalid databases leave the row retryable. Real
database-backed KGM V5/KGG GUI validation remains pending.

## Stage 6 closure

Stage 6A–6D engineering is complete and automated-tested. Remaining real
QMC supplied-key and KGM V5/KGG database-backed validation is tracked in
`docs/real-validation.md` as external-material backlog, not as active
development work. The next product stage may proceed without those materials.
