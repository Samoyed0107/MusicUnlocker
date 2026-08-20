# Project status

| Stage | Status | Scope |
| --- | --- | --- |
| 5A | `COMPLETE` | Wails GUI shell and application facade. |
| 5B | `COMPLETE` | Input queue, batch lifecycle, and native in-window drop. |
| 5C | `COMPLETE` | Preview service and GUI playback; real NCM/MP3 validation recorded. |
| 5D | `IMPLEMENTATION_COMPLETE` / `REAL_MANUAL_VALIDATION_PENDING` | Windows argv launch, drag-to-EXE path delivery, and Wails single-instance handoff. |
| 5E1 | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Offline metadata-aware decoded-output naming core and deterministic finalization. |
| 5E2 | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_PASS` | GUI naming controls, backend validation, batch mapping, preference integration, and validated source-only CUSTOM template regression fix. |
| 5F1 | `COMPLETE` / `REAL_PASS` | Corrected source-only CUSTOM naming and validated production output/audio integrity. |
| 5F | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_PASS` | Offline GUI localization with SYSTEM, zh-CN, and en-US modes, live switching, interpolation parity fix, runtime failure containment, and validated production startup. |
| 5F2 | `COMPLETE` / `REAL_PASS` | Playback continuity across language switching and batch status badge placement. |
| 5G | `COMPLETE` / `REAL_PASS` | Offline NCM container Title/Artist/Album integration and validated metadata-aware naming. |
| 6A | `COMPLETE` / `AUTOMATED_TESTED` | Request-scoped external QMC EKey provider, offline STag/MusicEx adapter path, and single-file CLI file input. |
| 6B | `COMPLETE` / `AUTOMATED_TESTED` | Backend-only native EKey selection and per-item retry for blocked QMC rows. |
| 6C | `COMPLETE` / `AUTOMATED_TESTED` | Application-owned KGG database provider, read-only bounded validation, KGM V5 adapter, and CLI database path. |
| 6D | `COMPLETE` / `AUTOMATED_TESTED` | Backend-owned native KGG database chooser, typed per-item retry, result replacement, localization, and preview continuity. |
| 6E | `ENGINEERING_COMPLETE` | External-material architecture audit, hardening, security boundary review, and validation backlog closure. |
| 7A | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Release identity, CLI version reporting, localized GUI About information, and Windows executable metadata foundation. |
| 7B-A | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Desktop workflow layout: responsive setup workspace, batch command surface, and unified results/summary/preview area. |
| 7B-B | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Fluent visual tokens, shared status presentation, component normalization, and frontend accessibility improvements. |
| 7C | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` / `WINDOWS_MANUAL_VALIDATION_PENDING` | Explicit persisted theme modes, native title-bar theme coordination, SVG status icons, motion/accessibility refinement, localized About dialog, and documented Mica fallback. |

Stage 5D accepts existing files and directories passed as process arguments,
retains initial arguments until the frontend retrieves them, and forwards later
launches through Wails `SingleInstanceLock`. Launches never start a batch.
Physical Explorer drag-to-EXE acceptance remains pending unless performed on a
real Windows installation.

Stage 5E2 automated validation is complete. The prior real CUSTOM naming
validation failed because source-only templates were incorrectly treated as
metadata-dependent; the fix is implemented and the corrected real output and
audio hash are now validated.

Stage 5F had a real production white-screen failure caused by the
`batch.itemsAdded` interpolation mismatch. The fix and production startup are
now real-validated. Stage 5F2 playback continuity is also real-validated.
Supported GUI modes are `SYSTEM`, `zh-CN`, and `en-US`.

Stage 5G is `COMPLETE / REAL_PASS`; the validated naming outputs and audio hash
are recorded in `docs/real-validation.md`.

Stage 6 remains offline and engineering-complete. Real QMC STag supplied-key
decoding is `PENDING_EXTERNAL_MATERIAL`; KGG/KGM V5 database-backed decoding is
also `PENDING_EXTERNAL_MATERIAL`. These are explicit validation backlog items,
not active engineering stages.

Stage 6B adds a non-persistent native EKey file selection flow. Only the
selected blocked item is retried; its stable queue index is replaced in the
latest snapshot and summary counts are recomputed. Raw key material and the
selected key path remain backend-only. Real supplied-key QMC validation remains
pending.

Stage 6C supplies KGM V5 with an explicit per-source KGG database capability.
The pinned upstream helper reads the supplied database into memory, imports it
into an in-memory SQLite database, and queries `ShareFileItems` by the KGM
header's `AudioHash`. Database lookup is narrowly serialized around upstream
initialization because the upstream process cache is not concurrency-safe;
audio reads remain outside that lock. No database path or contents are
persisted. Stage 6D adds `RetryKGGWithDatabase(batchID, itemIndex)`: the
backend opens the native file dialog, reuses the same validation, and keeps the
selected path out of React, DTO results, preferences, and logs. Only the
eligible logical row is retried with its original output/naming snapshot;
result replacement recomputes the summary. Cancellation and validation failure
leave the original row unchanged. Real KGG/KGM V5 database-backed GUI decode
remains `REAL_RETEST_PENDING / EXTERNAL_DATABASE_REQUIRED`.

Stage 6 overall is `ENGINEERING_COMPLETE`. Missing real external material does
not block the next product stage.

Stage 7A is limited to release identity and build foundation. The centralized
Go identity is used by the Wails facade and all CLI tools; release builds may
inject timestamp and commit identifier values with `-ldflags -X`. Installer,
file-association, updater, history/library, network metadata, and decoder work
remain outside this stage.

Stage 7B-A reorganizes only the React presentation layer. The 900px minimum
window stacks the setup vertically; wider windows use input and destination
columns. App lifecycle, batch events, retry eligibility, preview capability
URLs, naming validation, DTOs, and backend behavior are unchanged. About build
identity is removed from the permanent workspace pending a later dialog; no
new theme system, animation, or native-window treatment is included.

Stage 7B-B adds a semantic CSS token layer with prepared light/dark values,
Windows-friendly typography, normalized controls and cards, and one shared
status presentation mapping. It adds visible focus, 32px minimum button
targets, selected filter semantics, byte-progress accessibility values, and
accessible full-path labels while preserving all batch, retry, preview, DTO,
and file-drop behavior. Explicit theme switching, Mica, and motion remain
outside this phase.

Stage 7C promotes the prepared palette into an explicit persisted theme system
with stable `SYSTEM`, `LIGHT`, and `DARK` values. React updates the root semantic
tokens and asks Wails to update the native title-bar theme; system mode continues
to follow Windows changes. Legacy lowercase preferences remain readable. Theme
and language changes are presentation-only and do not recreate or mutate batch,
preview, retry, queue, naming, or output state.

The permanent About card remains absent. A localized modal dialog obtains the
application name, version, build timestamp, and build identifier from
`GetAppInfo`, whose identity comes from `internal/version`, and adds supported
format and offline-processing information. Status presentation now uses a
shared inline SVG icon abstraction rather than font-dependent symbols. Motion
is limited to short presentation transitions and is disabled through
`prefers-reduced-motion`.

The pinned Wails version can switch native title-bar themes and exposes DWM
Mica as a Windows option without Windows App SDK. Mica is intentionally not
enabled: it is limited to Windows 11 build 22621 or later and requires changing
the window/WebView transparency composition. The opaque semantic surfaces are
the portable fallback. Native title-bar synchronization, real system-theme
change handling, and unsupported-Windows fallback remain manual validation
items; no installer or additional runtime dependency was added.
