# Project status

This is the fresh-session handoff for the current MusicUnlocker release-candidate branch.
Implementation claims below are based on source and commit history; real-file
claims are separately recorded in `docs/real-validation.md`.

## Current baseline

- Branch: `codex/stage-8c-v2` during implementation; integrate by commit rather
  than treating the branch name as product identity.
- Stage 8A implementation predecessor: `ca740d377a14555b35bb9e766c7eff1bfd1edcf6`
  (`feat: add premium glass desktop polish`). Stage 8A is implemented by the
  commit containing this document.
- Product target: Windows 11 x64-first desktop utility.
- Stage 0-7 history: frozen; do not renumber completed work.
- Stage 5H: `IMPLEMENTED / AUTOMATED_TESTED / REAL_PASS`.
- Stage 6: `ENGINEERING_COMPLETE`; real QMC supplied-key and KGM V5/KGG
  database-backed validation remains `PENDING_EXTERNAL_MATERIAL`.
- Stage 7 Desktop Product Foundation: complete through Stage 7D.
- Stage 8A Installer & Portable Packaging:
  `IMPLEMENTED / PORTABLE_VALIDATED / LOCAL_INSTALLER_VALIDATED`.
  Clean-machine WebView2 and cross-version upgrade validation remain pending.
- Stage 8B Windows File Associations & Launch Lifecycle:
  `ENGINEERING_COMPLETE / LOCAL_ASSOCIATION_VALIDATED`.
  Clean-machine association validation remains Stage 8C work.
- Stage 8C Release Candidate, Standalone EXE & Product Readiness:
  `ENGINEERING_COMPLETE / RELEASE_CANDIDATE_LOCAL_VALIDATED / STANDALONE_EXE_VALIDATED`.
  No clean-machine, signing, SmartScreen-download, missing-WebView2, or
  cross-version-upgrade claim is made.
- Stage 8D CI and Windows Release Automation:
  `ENGINEERING_COMPLETE / CI_AUTOMATED / RELEASE_BUILD_AUTOMATED /
  REMOTE_CI_VALIDATED / GITHUB_RUNNER_RELEASE_BUILD_VALIDATED`.
  Feature-branch CI, master CI, and a manual artifact-only Windows release
  runner build have passed. Draft Release mutation and provenance attestation
  remain intentionally untested.

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

## Stage 8A release packaging

The Windows x64 release workflow now derives version and product identity from
`internal/version`, builds with Wails v2.15.0, embeds the official WebView2
Evergreen bootstrapper as the missing-runtime recovery path, and stages a
per-user NSIS installer plus a portable ZIP and SHA-256 checksums. The custom
Wails NSIS project creates a current-user Start Menu shortcut, offers an
optional desktop shortcut, and writes the uninstall entry under HKCU.

The portable ZIP was built, inspected, extracted outside the repository,
launched, and closed successfully on the development machine. Its executable
reported Windows product/file version `0.7.0`/`0.7.0.0` and carried the
`wv2runtime.embed` production build tag. No repository-relative runtime files
were required.

With NSIS 3.12 available, the production installer was built and lifecycle
validated locally. A custom-path installation and its registered uninstaller
removed only MusicUnlocker-owned files, shortcut, and HKCU registration while
retaining preferences and unrelated parent content. A fresh default install at
`%LOCALAPPDATA%\Programs\MusicUnlocker` launched directly and from its Start
Menu shortcut, reported version `0.7.0`, and wrote an HKCU `InstallLocation`
matching the actual directory. A same-version `0.7.0` reinstall retained one
working shortcut and correct registration; the final registered uninstall
removed the owned installation, shortcut, and HKCU entry while retaining
preferences and user music/output content. The installer requests no UAC
elevation. `CLEAN_MACHINE_WEBVIEW2_TEST_PENDING` and
`CROSS_VERSION_UPGRADE_TEST_PENDING` remain explicit; the installed WebView2
Runtime was not removed for testing.

## Stage 8B Windows file associations

The per-user installer now registers one shared versioned ProgID,
`MusicUnlocker.EncryptedMusic.1`, for the 49 simple encrypted-container
filename candidates derived from the upstream adapter registration surface.
Normal decoded audio extensions and compound application-level hints are
excluded. Registration uses HKCU `OpenWithProgids`, a MusicUnlocker-owned
Capabilities tree, and `HKCU\Software\RegisteredApplications`; it never writes
UserChoice or replaces another handler. The Shell open command passes one
quoted path into the existing Stage 5D startup/single-instance queue and never
starts a batch.

The production installer passed local default-path install, first-instance KGM
Explorer activation, existing-instance NCM handoff, duplicate activation,
Unicode/space-path, idle/no-output, registered uninstall, portable
registry-neutral, and same-version reinstall/activation checks. Both real-file
activations remained at `Ready` until explicit Start, the output directory was
never created, and source hashes remained unchanged. Uninstall removed all
MusicUnlocker-owned association data while retaining a controlled unrelated
`OpenWithProgids` sentinel. Reinstall restored 49 unique association values.
This Shell lifecycle check does not change decoder real-validation status.

## Stage 8C release candidate

The existing Windows release script now stages six public assets in one run:
the per-user installer, the actual production Wails executable under a
versioned standalone filename, the complete portable ZIP, top-level Scheme A
license and third-party notice files, and `SHA256SUMS`. The script preserves
the existing `windows/amd64`, `-trimpath`, `-webview2 embed`, user-scope NSIS,
and source-derived build identity path. It enforces an exact final artifact
allowlist before and after checksum generation.

Local RC validation used only the standalone EXE in an unrelated temporary
directory whose path contained spaces and Japanese characters. It launched,
rendered the frontend and icon, exposed About version `0.7.0`, and exited
cleanly without adjacent runtime files. A separate argument launch queued one
existing synthetic `.ncm` path through the normal Stage 5D inbox, remained
`Ready`, and created no output directory. After the installed build had been
uninstalled, standalone launch left the MusicUnlocker uninstall key, ProgID,
Capabilities, RegisteredApplications entry, sampled UserChoice keys, and all
installer-owned state absent.

The RC installer installed per-user without elevation, created its Start Menu
shortcut and `0.7.0` uninstall identity, registered 49 encrypted candidates,
queued a Unicode/space launch argument without auto-starting, and uninstalled
its owned files and registration. The portable ZIP retained its four-file
versioned layout and launched from a separate Unicode/space path with no
installer-owned registry state. Installer, standalone, portable, and installed
executables shared version `0.7.0`; the three application executables shared
identical production bytes. All inspected executables were `NotSigned`, so
`UNSIGNED_RELEASE_ARTIFACTS` remains explicit.

The development host had WebView2 Runtime `151.0.4129.101`, yielding
`DEVELOPMENT_HOST_WEBVIEW2_PRESENT_PASS`; it is not a clean-machine claim.
`CLEAN_MACHINE_CORE_TEST_PENDING`,
`CLEAN_MACHINE_WEBVIEW2_MISSING_TEST_PENDING`,
`CROSS_VERSION_UPGRADE_TEST_PENDING`, and
`SMARTSCREEN_PUBLIC_DOWNLOAD_TEST_PENDING` remain open. No authentic real KGM
or NCM sample was available on this host for a release-artifact decode/Preview
smoke, so that regression remains pending without changing
`docs/real-validation.md`.

The RC UI was observed at approximately 1100x760 and maximized, in explicit
Light and Dark themes, with `zh-CN`, a long Unicode/space input path, and
keyboard-accessible controls. The 900x600, snapped, live system-theme,
100/125/150 percent scaling, `en-US` RC-binary, exhaustive focus, long-name
stress, and real Preview-continuity checks remain pending rather than inferred
from automated frontend smoke tests or older evidence.

## Toolchain

The known Windows validation environment is:

- Windows 11 x64
- Go 1.26.5
- Node.js 24.18.0
- npm 11.16.0
- Wails v2.15.0
- WebView2 151.0.4129.101
- canonical decoder module: `git.um-react.app/um/cli v0.2.19`

Go module versions are pinned in `go.mod`, and the frontend dependency graph is
captured by `frontend/package-lock.json`. No C compiler or race-detector
availability is claimed.

## Stage 8D automation

CI and privileged release operations are structurally separate. `.github/workflows/ci.yml`
has read-only repository permission and runs on pull requests, manual dispatch,
and pushes to `master`, `main`, `develop`, and `codex/**`. It installs Go 1.26.5,
Node.js 24.18.0, npm 11.16.0, and Wails v2.15.0 exactly, uses `npm ci`, runs the
established Go/frontend checks, and smoke-builds the Windows production app.

`.github/workflows/windows-release.yml` is manual-only and defaults to an
artifact-only build from the default branch. An optional tag input checks out
an already existing tag and must equal `v` plus the authoritative
`internal/version` value. The build job uses `windows-2025`, installs exact
NSIS 3.12.0, calls `scripts/build-windows-release.ps1`, independently verifies
the exact six-file directory and every `SHA256SUMS` entry, then uploads only
that directory for 30 days. It has no repository-content write permission.

Draft Release creation is a separate conditional job with `contents: write`.
It re-downloads and re-verifies the workflow artifact, refuses non-draft
releases or drafts with unexpected assets, uses the checked-in RC note, and
only creates or updates a draft for an existing verified tag. Publication is
always human-controlled. Optional GitHub provenance attestation is wired for
the installer, standalone EXE, and portable ZIP, but is off by default because
repository visibility and account support could not be established locally:
`ARTIFACT_ATTESTATION_PENDING`. GitHub provenance attestation is not a Windows
publisher code signature.

Remote validation evidence is now recorded:

- feature branch `codex/stage-8c-v2` at `59b49f0` passed CI;
- primary branch `master` at `59b49f0` passed CI;
- the manual artifact-only `windows-2025` release workflow passed from a fresh
  GitHub runner and uploaded `MusicUnlocker-0.7.0-windows-release`;
- the downloaded artifact's five `SHA256SUMS` entries were independently
  recomputed and all matched after download.

This justifies `REMOTE_CI_VALIDATED`, `CLEAN_GITHUB_RUNNER_CI_PASS`,
`MASTER_CI_PASS`, `CLEAN_GITHUB_RUNNER_BUILD_PASS`, and
`GITHUB_RUNNER_RELEASE_BUILD_VALIDATED`. It does not justify
`CLEAN_MACHINE_CORE_VALIDATED`: the hosted runner is not interactive consumer
Windows validation. `DRAFT_RELEASE_REMOTE_TEST_PENDING` and
`ARTIFACT_ATTESTATION_PENDING` remain. The current boundary also remains
`CODE_SIGNING_NOT_CONFIGURED` and `UNSIGNED_RELEASE_ARTIFACTS`. Future signing
must occur after build but before final hashing: build, sign EXEs, verify
signatures, regenerate final `SHA256SUMS`, then upload/publish.

Downloaded artifact manifest evidence:

```text
ca0823e35b5de5df2dc468132a1273dfd07aceb1d7df86a341f57e895b577650  LICENSE.txt
2402b4306e15f2cabab781d0f43554beda7ead1628c247fef8b4e23972766f27  MusicUnlocker-0.7.0-win64-portable.zip
4b6e6382dee8b9b851c13e5db7c89bbd4b13ccae4b2d7ecb3fc5cc555470ee38  MusicUnlocker-0.7.0-win64.exe
31ee92ac5f8cb2039819f56a0519fa553abe8b8c8c25c1562671ca733879ca25  MusicUnlocker-Setup-0.7.0-win64.exe
3e450b04a1b0110fe3f204542ea97c20b6f3ce42762bcd09253062d2f325b71f  THIRD_PARTY_NOTICES.txt
```

These hashes belong to the remote workflow artifact at this commit and are
not expected to match the earlier Stage 8C local build.

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

`docs/roadmap.md` is the sole current implementation roadmap. Stage 8D
engineering is complete locally. Remote CI/release-workflow execution and the
clean-machine, missing-WebView2, SmartScreen, cross-version-upgrade, Stage 7D
environment, and real release-smoke evidence listed above remain pending.
