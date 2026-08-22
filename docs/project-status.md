# Project status

This is the fresh-session handoff for the current MusicUnlocker mainline.
Implementation claims below are based on source and commit history; real-file
claims are separately recorded in `docs/real-validation.md`.

## Current baseline

- Branch: `master`.
- Stage 7D implementation baseline: `bfdfe305dc1c5ecaa77d827d19004c109cc7d8e8`
  (`docs: reconcile project baseline and roadmap`). Stage 7D is completed by
  the commit containing this document.
- Product target: Windows 11 x64-first desktop utility.
- Stage 0-7 history: frozen; do not renumber completed work.
- Stage 5H: `IMPLEMENTED / AUTOMATED_TESTED / REAL_PASS`.
- Stage 6: `ENGINEERING_COMPLETE`; real QMC supplied-key and KGM V5/KGG
  database-backed validation remains `PENDING_EXTERNAL_MATERIAL`.
- Stage 7 Desktop Product Foundation: complete through Stage 7D.

The Stage 7D baseline reconciliation commit changes no production code, tests,
frontend code, dependencies, or build configuration.

## Frozen stage status

| Stage | Status | Historical meaning |
| --- | --- | --- |
| 0 | `COMPLETE` | Repository and toolchain bootstrap. |
| 1 | `COMPLETE` | Canonical upstream research and adapter boundary. |
| 2 | `COMPLETE` | Content probing and format identification. |
| 3 | `ENGINEERING_COMPLETE` | Decode pipeline and decoder integration; real-validation backlog remains. |
| 4 | `COMPLETE` | UI-independent batch engine, bounded concurrency, progress, deterministic ordering, and collision safety. |
| 5 | `COMPLETE` at engineering level | Desktop functional application: Wails shell, native file/folder workflow and drop, secure preview, Windows launch handoff, naming, localization, NCM container metadata, and metadata quality guard. |
| 6 | `ENGINEERING_COMPLETE` | Typed, backend-owned, offline external-material subsystem for QMC EKeys and KGG databases, including GUI per-item retry and closure hardening. |
| 7 | `COMPLETE` | Desktop Product Foundation through release identity, workflow layout, Fluent visual/accessibility foundation, Windows-native themes, and premium glass polish. |

Stage 7 sub-stages retain their original meanings:

- 7A — Release Identity Foundation: `COMPLETE`. This is not installer
  engineering.
- 7B-A — Workflow Layout: `COMPLETE`.
- 7B-B — Fluent Visual Foundation and Accessibility: `COMPLETE`.
- 7C — Windows Native Theme Experience: `COMPLETE`. This is not a
  library/history feature.
- 7D — Premium Glass & Interaction Polish: `COMPLETE`. Content glass is the
  supported visual system; native Mica remains disabled and optional.

## Toolchain

The known Windows validation environment is:

- Windows 11 x64
- Go 1.26.5
- Node.js 24.18.0
- npm 11.16.0
- Wails v2.15.0
- WebView2 151.0.4129.86
- canonical decoder module: `git.um-react.app/um/cli v0.2.19`

Go module versions are pinned in `go.mod`, and the frontend dependency graph is
captured by `frontend/package-lock.json`. No C compiler or race-detector
availability is claimed.

## Current architecture

The desktop call path is:

```text
React/TypeScript frontend
    -> Wails bindings
    -> internal/app
    -> internal/batch
    -> internal/decode
```

The frontend owns presentation and state mapping only. Go owns filesystem
dialogs and validation, discovery, probing, scheduling, decoding, output
planning, metadata handling, naming, atomic finalization, preview authorization,
and external-material selection.

Important subsystems are:

- `internal/probe` for bounded content evidence and filename candidate hints;
- `internal/decoder/upstream` as the only canonical upstream adapter boundary;
- `internal/audiofmt` and `internal/outputplan` for decoded-content sniffing and
  conservative output planning;
- `internal/audiometa`, `internal/containermeta`, and
  `internal/metadataquality` for offline decoded/container metadata and
  field-level quality fallback;
- `internal/naming` plus `internal/decode` finalization for validated naming,
  decoded-format extensions, and atomic no-replace commit;
- `internal/externalmaterial` for request-scoped QMC EKeys and read-only KGG
  database capabilities;
- `PreviewService` for opaque, in-memory, loopback-only preview authorization;
- `internal/app/launch.go` plus Wails single-instance handling for launch-path
  queue handoff.

## Current desktop capabilities

- Native multi-file, folder, and output-directory selection.
- Recursive deterministic batch discovery with 1-8 bounded workers.
- Native in-window file/folder drop.
- Startup-argument and drag-to-EXE path queueing plus later-instance handoff.
  Launch paths never auto-start decoding.
- Explicit output destination, no silent overwrite, immutable sources, and
  atomic output finalization.
- Keep-source, preset, and backend-validated custom naming templates.
- Offline decoded metadata, NCM container metadata, and the Stage 5H metadata
  quality guard.
- Bundled `SYSTEM`, `zh-CN`, and `en-US` localization.
- Secure local preview using opaque capabilities rather than frontend paths.
- Backend-only QMC EKey and KGG database selection with same-row retry.
- Persisted `SYSTEM`, `LIGHT`, and `DARK` themes with native title-bar
  coordination.
- Responsive workflow layout, Fluent visual tokens, shared SVG status
  presentation, keyboard focus, and reduced-motion handling.
- Theme-tuned semantic glass surfaces, atmospheric canvas depth, lightweight
  result rows, elevated preview/dialog presentation, and opaque/forced-colors
  fallbacks. Blur is limited to major surfaces rather than per-progress row.
- Localized About dialog backed by centralized application/version/build
  identity.

The GUI permits one active batch. Live events are transient; the final
`BatchComplete` snapshot is authoritative, and rows correlate by stable batch
index. Selected input rows remain distinct from recursively discovered files.

## Validation distinction

Implementation and automated tests do not imply real compatibility for every
format variant. Current key real-file evidence is:

- KGM V3 decode: `REAL_PASS`.
- Stage 5H real KGM metadata-quality fallback retest: `REAL_PASS`.
- NCM decode and metadata-aware naming: `REAL_PASS`.
- QMC STag probing: `PROBE_REAL_PASS`; supplied-key decode is
  `PENDING_EXTERNAL_MATERIAL`.
- KGM V5/KGG support: `ENGINEERING_COMPLETE`; real database-backed decode is
  `PENDING_EXTERNAL_MATERIAL`.
- KWM, TM, Xiami, Ximalaya, and self-contained QMC variants remain
  `REAL_PENDING`.

See `docs/real-validation.md` for exact filenames, byte lengths, hashes,
interpretation limits, and the remaining desktop manual-validation evidence.

## Product boundary

MusicUnlocker remains offline-first and source-preserving. It is not a streaming
client, downloader, cloud service, mandatory transcoder, or full music-library
manager. Network metadata, lyrics, automatic album art, telemetry, persistent
history/library storage, mandatory updating, and similar post-v1 ideas are not
current mainline stages.

## Next work

`docs/roadmap.md` is the sole current implementation roadmap. Stage 7D Premium
Glass & Interaction Polish is complete. Stage 8A Installer & Portable Packaging
is the first and next Windows release-engineering stage.
