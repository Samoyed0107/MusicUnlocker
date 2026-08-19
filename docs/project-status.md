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
| 6A | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` / `REAL_RETEST_PENDING` | Request-scoped external QMC EKey provider, offline STag/MusicEx adapter path, and single-file CLI file input. |
| 6B | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` / `REAL_QMC_EKEY_GUI_RETEST_PENDING` | Backend-only native EKey selection and per-item retry for blocked QMC rows. |
| 6C | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` / `REAL_RETEST_PENDING` | Application-owned KGG database provider, read-only bounded validation, KGM V5 adapter, and CLI database path. |

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

Stage 6A remains offline. Real QMC STag supplied-key decoding is
`REAL_RETEST_PENDING / EXTERNAL_KEY_MATERIAL_REQUIRED`. KGG/KGM V5 and the GUI
key-entry workflow remain deferred to later stages.

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
persisted, and the GUI database-selection workflow remains deferred.
