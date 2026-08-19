# Project status

| Stage | Status | Scope |
| --- | --- | --- |
| 5A | `COMPLETE` | Wails GUI shell and application facade. |
| 5B | `COMPLETE` | Input queue, batch lifecycle, and native in-window drop. |
| 5C | `COMPLETE` | Preview service and GUI playback; real NCM/MP3 validation recorded. |
| 5D | `IMPLEMENTATION_COMPLETE` / `REAL_MANUAL_VALIDATION_PENDING` | Windows argv launch, drag-to-EXE path delivery, and Wails single-instance handoff. |
| 5E1 | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Offline metadata-aware decoded-output naming core and deterministic finalization. |
| 5E2 | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_PASS` | GUI naming controls, backend validation, batch mapping, preference integration, and validated source-only CUSTOM template regression fix. |
| 5F | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_PASS` | Offline GUI localization with SYSTEM, zh-CN, and en-US modes, live switching, interpolation parity fix, runtime failure containment, and validated production startup. |
| 5F2 | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_RETEST_PENDING` | Playback continuity across language switching and batch status badge placement. |

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
now real-validated. Stage 5F2 playback continuity remains real-retest pending.
Supported GUI modes are `SYSTEM`, `zh-CN`, and `en-US`.

Stage 5G remains planned.
