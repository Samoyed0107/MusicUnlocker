# MusicUnlocker Roadmap v2

This is the sole current implementation roadmap. Stages 0-7 are frozen
historical meanings and must not be renumbered. Implementation status and
real-file validation are separate; the evidence matrix lives in
`docs/real-validation.md`.

## Product direction

MusicUnlocker remains a Windows 11 x64-first, offline-first desktop utility
built from a shared Go core, Wails v2, and a React/TypeScript frontend. Sources
remain immutable, outputs are committed atomically without silent overwrite,
decoded format is content-sniff driven, and transcoding is avoided.

The product is not a streaming client, downloader, cloud service, mandatory
transcoder, or full music-library manager.

## Frozen Stage 0-7 history

| Stage | Status | Frozen meaning |
| --- | --- | --- |
| 0 — Repository/toolchain bootstrap | `COMPLETE` | Repository layout, pinned toolchain, and initial build foundation. |
| 1 — Upstream research/adapter boundary | `COMPLETE` | Canonical decoder dependency research and the enforced upstream adapter boundary. |
| 2 — Content probing/format identification | `COMPLETE` | Bounded structural probing, content-first evidence, and filename hints as candidate discovery only. |
| 3 — Decode pipeline/decoder integration | `ENGINEERING_COMPLETE` | Shared offline decode pipeline, adapters, decoded-content format sniffing, and safe output behavior. Real-validation backlog remains. |
| 4 — Batch engine | `COMPLETE` | Recursive discovery, bounded concurrency, progress, deterministic results, target collision serialization, and cancellation cleanup. |
| 5 — Desktop functional application | `COMPLETE` at engineering level | Wails shell, native workflow/drop, secure preview, launch handoff, naming, localization, NCM metadata, and metadata quality. |
| 6 — External-material subsystem | `ENGINEERING_COMPLETE` | Typed backend-only QMC EKey and KGG database capabilities, per-item GUI retries, offline/non-persistent lifetimes, and closure hardening. |
| 7 — Desktop Product Foundation | `COMPLETE` | Release identity, workflow layout, Fluent visual/accessibility foundation, native themes, and premium glass polish. |

### Stage 3 validation boundary

KGM V3 and NCM have real-pass evidence. KWM, TM, Xiami, Ximalaya, and
self-contained QMC variants retain real-validation backlog. QMC STag
supplied-key decode and KGM V5/KGG database-backed decode remain
`PENDING_EXTERNAL_MATERIAL`. These facts do not reopen or renumber Stage 3.

### Stage 5 historical scope

Stage 5 includes:

- 5A — Wails application shell and facade;
- 5B — native file/folder workflow, recursive queue, and in-window drop;
- 5C — secure preview;
- 5D — Windows startup arguments and single-instance handoff;
- 5E — offline metadata-aware naming core and GUI controls;
- 5F — bundled localization and playback-continuity corrections;
- 5G — offline NCM container metadata for naming;
- 5H — Metadata Quality Guard.

Stage 5H is `IMPLEMENTED / AUTOMATED_TESTED / REAL_PASS`. It rejects
suspicious fields independently, preserves valid multilingual siblings, falls
through trusted metadata sources deterministically, and protects naming without
changing decoded bytes or repairing/re-encoding the underlying tags.

### Stage 6 historical scope

Stage 6 includes request-scoped QMC supplied-EKey core support and GUI retry,
read-only KGG database core support and GUI retry, typed blocked states,
backend-only material and path handling, offline/non-persistent lifetimes, and
the external-material closure audit.

Real supplied-key QMC decoding and real KGG database-backed decoding remain
`PENDING_EXTERNAL_MATERIAL` because compatible material is unavailable. This
is validation backlog, not unfinished Stage 6 engineering.

### Stage 7 historical scope

- 7A — Release Identity Foundation: `COMPLETE`. It centralized GUI, CLI, and
  executable build identity; it was not installer engineering.
- 7B-A — Workflow Layout: `COMPLETE`.
- 7B-B — Fluent Visual Foundation and Accessibility: `COMPLETE`.
- 7C — Windows Native Theme Experience: `COMPLETE`. It added persisted
  System/Light/Dark themes, native title-bar coordination, SVG status icons,
  reduced-motion behavior, and the localized About dialog. It was not
  library/history work.

Native Windows Mica was investigated but not enabled. The existing opaque
surfaces remain the supported fallback.

## Stage 7D — Premium Glass & Interaction Polish

Status: `COMPLETE`.

Stage 7D completed the explicit pre-v1 visual requirement without changing core
decode, batch, preview, retry, naming, or output semantics.

Required direction:

- premium translucent/glass content surfaces;
- layered cards and clear depth hierarchy;
- soft background and glow treatment;
- coordinated light and dark glass tokens;
- refined drop-zone interaction;
- restrained interaction motion;
- results and preview visual polish;
- reduced-motion accessibility;
- opaque/no-backdrop-filter and forced-colors fallbacks;
- blur limited to major surfaces, with lightweight non-blurred result rows.

Glass content surfaces are implemented independently of the native window
backdrop. Native Windows Mica was not attempted and is not enabled; the opaque
Wails window remains the supported configuration.

## Stage 8 — Windows Release Engineering

Stage 8 follows the completed Stage 7D visual pass. Stage 8A is the first
release-engineering implementation stage.

### Stage 8A — Installer & Portable Packaging

Status: `IMPLEMENTED / PORTABLE_VALIDATED / LOCAL_INSTALLER_VALIDATED`.

Target artifacts:

- `MusicUnlocker-Setup-<version>-win64.exe`
- `MusicUnlocker-<version>-win64-portable.zip`

Implemented scope:

- Wails v2 NSIS project and canonical `wails build -nsis` release path;
- per-user installation under LocalAppData without UAC elevation;
- current-user Start Menu entry and optional desktop shortcut;
- conservative uninstall of explicitly installed files and shortcuts;
- centralized application/version/icon identity across executable, installer,
  and artifact names;
- minimal portable package with license and third-party notices;
- embedded official WebView2 Evergreen bootstrapper recovery strategy;
- deterministic ZIP entry ordering/timestamps and `SHA256SUMS` generation.

The portable artifact has passed local clean-directory extraction and launch.
With NSIS 3.12, the installer passed local custom-path and default-path
install, installed launch, Start Menu launch, same-version reinstall, HKCU
registration, and registered-uninstall validation. The HKCU entry now writes
`InstallLocation` explicitly and it matched the default installation path.
No UAC elevation was requested. Clean-machine missing-WebView2 validation and
cross-version upgrade validation remain pending as
`CLEAN_MACHINE_WEBVIEW2_TEST_PENDING` and
`CROSS_VERSION_UPGRADE_TEST_PENDING`, respectively.

### Stage 8B — File Associations & Windows Lifecycle

Status: `ENGINEERING_COMPLETE / LOCAL_ASSOCIATION_VALIDATED`.

File associations may include supported encrypted-container extensions. Launch
safety is mandatory:

```text
Explorer double-click
    -> launch or single-instance handoff
    -> queue input
    -> WAIT FOR USER
    -> Start
```

An association launch must never auto-decode by itself.

The Stage 8B installer registers the 49 simple encrypted-container filename
candidates derived from the upstream adapter surface under one versioned
`MusicUnlocker.EncryptedMusic.1` ProgID. Normal decoded-audio extensions and
compound filename hints are excluded. Registration is per-user and uses
`OpenWithProgids`, RegisteredApplications, and Capabilities without writing
UserChoice or forcing a default.

Local lifecycle validation passed for real KGM first-instance activation, real
NCM existing-instance handoff, duplicate queue deduplication, Unicode and
spaces, explicit idle/no-output safety, MusicUnlocker-owned association
cleanup, unrelated-value retention, portable registry neutrality, and
same-version reinstall/reactivation. This does not reclassify decoder real-file
coverage. `CLEAN_MACHINE_ASSOCIATION_TEST_PENDING` and
`CROSS_VERSION_UPGRADE_TEST_PENDING` remain open for Stage 8C.

### Stage 8C — Release Artifacts & Product Readiness

Status: `ENGINEERING_COMPLETE / RELEASE_CANDIDATE_LOCAL_VALIDATED / STANDALONE_EXE_VALIDATED`.

Completed engineering scope:

- one production Wails build now yields the per-user installer, the actual
  single-file Wails executable, and the documented portable ZIP;
- the final release directory uses an exact six-file allowlist and normalized,
  deterministic `SHA256SUMS` entries;
- top-level Scheme A license and third-party notice assets accompany the
  standalone executable, and the portable/installer notice layouts remain;
- local installer, association activation, standalone Unicode/space-path
  direct launch, standalone argument queueing/idle behavior, registry
  neutrality, portable extraction/launch, version resources, and unsigned
  Authenticode state were validated.

This host is a development machine, not a clean environment. Actual
clean-machine core testing, missing-WebView2 recovery, public-download
SmartScreen behavior, cross-version upgrade, remaining Stage 7D environment
matrix items, and real KGM/NCM release smoke remain explicit pending evidence.

### Stage 8D — CI / Release Automation

Status: `ENGINEERING_COMPLETE / CI_AUTOMATED / RELEASE_BUILD_AUTOMATED`.

Completed engineering scope:

- read-only PR/push CI with pinned Go, Node/npm, Wails, and immutable official
  Action revisions;
- manual clean `windows-2025` release-candidate builds through the existing
  authoritative PowerShell packaging script and exact NSIS 3.12.0;
- independent six-file allowlist and checksum-manifest verification followed
  by a deterministic, 30-day workflow artifact upload;
- optional existing-tag validation, opt-in provenance attestation, and a
  separately permissioned Draft Release job that never publishes or creates a
  tag.

Remote execution is not yet evidence-backed:
`DRAFT_RELEASE_REMOTE_TEST_PENDING`, `ARTIFACT_ATTESTATION_PENDING`, and no
`GITHUB_RUNNER_RELEASE_BUILD_VALIDATED` claim. A successful GitHub-hosted build
will prove fresh checkout/toolchain bootstrap and release construction on a
fresh Windows Server VM, not interactive Windows 11 installer, Explorer,
WebView2 lifecycle, DPI/snap, UAC, or SmartScreen behavior.

Code signing is `CODE_SIGNING_NOT_CONFIGURED`; current outputs remain
`UNSIGNED_RELEASE_ARTIFACTS`. An automatic updater is `DEFERRED`.

## Deferred or post-v1

Do not create near-term mainline stages for:

- music-library management;
- persistent listening/history databases;
- online metadata;
- lyrics retrieval;
- automatic album-art fetching;
- telemetry;
- a mandatory updater.

Optional future research must preserve offline-first behavior, privacy, source
immutability, deterministic output, and the shared-core architecture.
