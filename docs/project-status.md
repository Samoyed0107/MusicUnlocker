# Project status

| Stage | Status | Scope |
| --- | --- | --- |
| 5A | `COMPLETE` | Wails GUI shell and application facade. |
| 5B | `COMPLETE` | Input queue, batch lifecycle, and native in-window drop. |
| 5C | `COMPLETE` | Preview service and GUI playback; real NCM/MP3 validation recorded. |
| 5D | `IMPLEMENTATION_COMPLETE` / `REAL_MANUAL_VALIDATION_PENDING` | Windows argv launch, drag-to-EXE path delivery, and Wails single-instance handoff. |
| 5E1 | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Offline metadata-aware decoded-output naming core and deterministic finalization. |
| 5E2 | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | GUI naming controls, backend validation, batch mapping, and preference integration. |
| 5F | `IMPLEMENTATION_COMPLETE` / `AUTOMATED_TESTED` | Offline GUI localization with SYSTEM, zh-CN, and en-US modes and live switching. |

Stage 5D accepts existing files and directories passed as process arguments,
retains initial arguments until the frontend retrieves them, and forwards later
launches through Wails `SingleInstanceLock`. Launches never start a batch.
Physical Explorer drag-to-EXE acceptance remains pending unless performed on a
real Windows installation.

Stage 5E2 automated validation is complete. Manual real naming validation with
the user's external NCM sample remains `PENDING`.

Stage 5F localization implementation and automated validation are complete.
Supported GUI modes are `SYSTEM`, `zh-CN`, and `en-US`.
