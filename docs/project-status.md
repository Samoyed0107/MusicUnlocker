# Project status

| Stage | Status | Scope |
| --- | --- | --- |
| 5A | `COMPLETE` | Wails GUI shell and application facade. |
| 5B | `COMPLETE` | Input queue, batch lifecycle, and native in-window drop. |
| 5C | `COMPLETE` | Preview service and GUI playback; real NCM/MP3 validation recorded. |
| 5D | `IMPLEMENTATION_COMPLETE` / `REAL_MANUAL_VALIDATION_PENDING` | Windows argv launch, drag-to-EXE path delivery, and Wails single-instance handoff. |
| 5E1 | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Offline metadata-aware decoded-output naming core and deterministic finalization. |
| 5E2 | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_RETEST_PENDING` | GUI naming controls, backend validation, batch mapping, preference integration, and source-only CUSTOM template regression fix. |
| 5F | `FIX_IMPLEMENTED` / `AUTOMATED_TESTED` / `REAL_RETEST_PENDING` | Offline GUI localization with SYSTEM, zh-CN, and en-US modes, live switching, interpolation parity fix, and runtime failure containment. |

Stage 5D accepts existing files and directories passed as process arguments,
retains initial arguments until the frontend retrieves them, and forwards later
launches through Wails `SingleInstanceLock`. Launches never start a batch.
Physical Explorer drag-to-EXE acceptance remains pending unless performed on a
real Windows installation.

Stage 5E2 automated validation is complete. The prior real CUSTOM naming
validation failed because source-only templates were incorrectly treated as
metadata-dependent; the fix is implemented and real retest remains pending.

Stage 5F had a real production white-screen failure caused by the
`batch.itemsAdded` interpolation mismatch. The fix is implemented and
automated-tested; a corrected production GUI launch remains required before
real acceptance. Supported GUI modes are `SYSTEM`, `zh-CN`, and `en-US`.
