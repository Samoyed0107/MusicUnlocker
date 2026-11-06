# Windows release packaging

Stages 8A through 8D provide one auditable Windows x64 workflow for the Wails
v2 NSIS installer, actual single-file Wails executable, and portable ZIP,
including per-user encrypted-container file associations owned only by the
installer. They do not add an updater, public automatic publication, signing,
or a portable preference mode.

## Toolchain and command

Use Windows PowerShell from a clean Git worktree with the repository's pinned
Go/frontend dependencies, Wails v2.15.0, and NSIS 3.x `makensis` on PATH. The
script never installs system software and fails before installer construction
when `makensis` is unavailable.

```powershell
.\scripts\build-windows-release.ps1
```

For local engineering work on a machine without NSIS, the portable surface can
be built explicitly:

```powershell
.\scripts\build-windows-release.ps1 -PortableOnly
```

`-AllowDirty` exists only for pre-commit engineering validation. Normal release
builds reject a dirty worktree. The script recreates only its resolved
repository-local `artifacts` directory and Wails' normal `build/bin` output.

## GitHub Actions automation

`.github/workflows/ci.yml` runs on pull requests, manual dispatch, and pushes
to `master`, `main`, `develop`, and `codex/**`. It grants only `contents: read`, cancels
superseded runs for the same branch/ref, and has a 45-minute timeout. On the
explicit `windows-2025` x64 image it runs `git diff --check`, `go test ./...`,
`go vet ./...`, locked frontend install/build/smoke tests, and `wails build`.
Go module/build and npm package-download caches are permitted; correctness does
not depend on a hit, and neither `node_modules` nor product outputs are cached.

`.github/workflows/windows-release.yml` is `workflow_dispatch` only and does
not cancel a running build for the same ref/tag. Without a tag it checks out
the repository default branch and only uploads a workflow artifact. With a tag
it checks out the existing ref. The authoritative `internal/version`
remains `0.7.0` for RC binaries, Windows PE metadata, installer identity, About,
and filenames. The first RC tag is `v0.7.0-rc.1`; later RCs use
`v0.7.0-rc.N` (positive integers without leading zeros). The final tag can
be `v0.7.0`. Other tags fail strict case-sensitive validation. Untagged
artifact-only builds use `docs/releases/0.7.0-rc.md`; tagged RC builds require
`docs/releases/0.7.0-rc.N.md`; the eventual final requires
`docs/releases/0.7.0.md`. Missing notes are fatal. Source version never comes
from a tag. Tagged builds require the exact tag to identify checkout HEAD; the
Draft job rechecks matching tag/commit identity.

GitHub enables `workflow_dispatch` only after the workflow exists on the
repository default branch. Merge the Stage 8D workflows into that branch
before the first manual Windows release run; pushing only the feature branch
is not sufficient for final manual-workflow validation.

For the first RC, choose the **master** branch in the Actions "Run workflow"
branch selector before supplying the existing RC tag. The selected workflow
revision and its tag checkout are separate identities: do not assume that
the repository default branch is master or that a tag input selects the
workflow YAML revision.

The release build uses no build caches and bootstraps Go 1.26.5, Node.js
24.18.0, npm 11.16.0, Wails v2.15.0, and Chocolatey's exact NSIS 3.12.0 package
on `windows-2025`. All external official Actions are fixed to full release
commit SHAs. The job calls this document's authoritative release script rather
than reproducing packaging in YAML, then `scripts/verify-windows-release.ps1`
independently enforces the six-file allowlist, parses all five manifest lines,
recomputes every SHA-256, and rejects missing, duplicate, malformed, or
unexpected entries. Only that verified directory is uploaded as
`MusicUnlocker-<version>-windows-release` for 30 days.

The build job is read-only with respect to repository contents. Draft Release
creation is a second, opt-in job with only `contents: write`; it downloads and
re-verifies the workflow artifact, requires the existing matching tag, uses
the exact tag-derived release note, refuses to alter a published release or
delete unexpected assets, and creates/updates a draft with GitHub CLI. RC drafts
must have `isDraft=true` and `isPrerelease=true`. Any existing release with
mismatched identity/status is rejected. The workflow never promotes a draft to
a published/stable/latest release. It never
publishes, creates, or moves a tag. Fork pull requests can run only the
unprivileged CI workflow and cannot reach this manual release path.

Provenance attestation for the Setup EXE, standalone EXE, and portable ZIP is
available as a separate opt-in using the official `actions/attest` action. It
was intentionally disabled for the completed artifact-only validation, so the
current status remains `ARTIFACT_ATTESTATION_PENDING`.
GitHub provenance attestation is not Authenticode and is not a Windows
publisher code signature.

## Published RC1 and public distribution

The [public repository](https://github.com/Samoyed0107/MusicUnlocker)
uses default branch `master`. The existing
[MusicUnlocker 0.7.0 RC 1](https://github.com/Samoyed0107/MusicUnlocker/releases/tag/v0.7.0-rc.1)
(Release ID `431691311`) was published at `2026-11-06T09:36:33Z` as
`draft=false`, `prerelease=true`, not stable `v0.7.0`. Its annotated
`v0.7.0-rc.1` tag still points to
`bf6f8f73bd9feeff0adfe00ae5aa70b37018a635`. Docs-only commits on
`master` do not alter this frozen RC1 build.

The published Release provides exactly six assets, including:

- [Windows installer (Setup EXE)](https://github.com/Samoyed0107/MusicUnlocker/releases/download/v0.7.0-rc.1/MusicUnlocker-Setup-0.7.0-win64.exe)
- [Standalone executable](https://github.com/Samoyed0107/MusicUnlocker/releases/download/v0.7.0-rc.1/MusicUnlocker-0.7.0-win64.exe)
- [Portable ZIP](https://github.com/Samoyed0107/MusicUnlocker/releases/download/v0.7.0-rc.1/MusicUnlocker-0.7.0-win64-portable.zip)
- `LICENSE.txt`, `THIRD_PARTY_NOTICES.txt`, and
  [SHA256SUMS](https://github.com/Samoyed0107/MusicUnlocker/releases/download/v0.7.0-rc.1/SHA256SUMS).

Expand **Assets** on the tagged Release detail page. The Actions
workflow artifact is a separate temporary build distribution surface.

Tagged workflow run `41480338162` passed the release build and Draft
Release jobs before human-authorized publication. All six public files
downloaded anonymously and matched frozen sizes and SHA-256 digests.
The five non-manifest files matched published `SHA256SUMS`. Native
Defender scanning recorded no matching detection on these exact files
in that test environment; this is not a universal malware guarantee.
Installer and standalone remain `NotSigned`, consistent with
`UNSIGNED_RELEASE_ARTIFACTS` and `CODE_SIGNING_NOT_CONFIGURED`.
CLI downloads do not validate normal browser Mark-of-the-Web/Explorer
SmartScreen UX: `SMARTSCREEN_PUBLIC_BROWSER_TEST_PENDING`. Clean-machine
installation and association, missing WebView2, upgrade, and attestation
tests remain pending.

## Release identity and artifacts

`internal/version` remains authoritative for the application name and version.
`cmd/buildmetadata` generates the Windows executable resource, release-identity
NSIS include, and association NSIS include. Association extensions are derived
from the decoder adapter registration surface and filtered by
`internal/fileassoc`. `cmd/releaseinfo` exposes the same identity to the
PowerShell workflow. The Git commit timestamp and 12-character commit ID become
the embedded build timestamp and build identifier, so GUI About and release
build provenance use stable source-derived values.

For version `0.7.0`, the default output directory contains:

```text
MusicUnlocker-Setup-0.7.0-win64.exe
MusicUnlocker-0.7.0-win64.exe
MusicUnlocker-0.7.0-win64-portable.zip
LICENSE.txt
THIRD_PARTY_NOTICES.txt
SHA256SUMS
```

`SHA256SUMS` uses lowercase SHA-256 followed by two spaces and the artifact
filename. The Stage 8C workflow hashes all five non-manifest release assets
after their final form is staged. Checksums detect transfer corruption; they
do not authenticate the publisher and are not a substitute for code signing.
Current RC binaries report `NotSigned`, recorded as
`UNSIGNED_RELEASE_ARTIFACTS`.

The signing boundary is `CODE_SIGNING_NOT_CONFIGURED`. No PFX, password,
certificate, Azure credential, or SignTool secret is stored or consumed. A
future signing integration must change ordering to: build, sign EXEs, verify
signatures, regenerate final `SHA256SUMS`, then upload and publish. Hashing the
current unsigned final files before upload remains correct for Stage 8D.

The ZIP writer sorts paths and applies the source commit timestamp to every
entry. Dependency versions are pinned, and Wails builds with `-trimpath`.
Byte-for-byte installer reproducibility still depends on using the same Go,
Node/npm, Wails, and NSIS toolchain versions; Stage 8A does not claim a
hermetic build environment.

## Installer behavior

The canonical command used by the script is equivalent to:

```powershell
wails build -clean -trimpath -platform windows/amd64 -webview2 embed -installscope user -ldflags <source-derived-values> -nsis
```

The installer is current-user scoped:

- install directory: `%LOCALAPPDATA%\Programs\MusicUnlocker`;
- UAC elevation: not requested (`RequestExecutionLevel user`);
- Start Menu: `%APPDATA%\Microsoft\Windows\Start Menu\Programs\MusicUnlocker.lnk`;
- desktop shortcut: optional installer component, off by default;
- uninstall registration: current-user HKCU uninstall key, including
  `InstallLocation` set to the actual install directory;
- file associations: current-user encrypted-container handler registration;
- URL protocols: none.

Installing the same or a newer version targets the same executable, shortcuts,
directory, and uninstall key, so files are replaced predictably and duplicate
Start Menu entries are not designed into the workflow. Local same-version
reinstall validation passed with one working Start Menu shortcut.

Uninstall deletes only the installed executable, packaged notices, installer
shortcuts, uninstaller, uninstall key, and then the install directory if empty.
It also removes only MusicUnlocker-owned ProgID, Capabilities,
RegisteredApplications, and per-extension `OpenWithProgids` values. It never
deletes an entire extension tree, another handler, or UserChoice.
It does not delete source music, decoded outputs, output directories, external
key/database material, or `%APPDATA%\MusicUnlocker\preferences.json`.
Preferences and WebView2 user data are retained conservatively.

## Standalone executable

`MusicUnlocker-0.7.0-win64.exe` is a renamed copy of the same normal production
Wails binary used by the installer and portable ZIP. It is not an NSIS wrapper,
self-extractor, launcher, UPX-compressed file, or second product. A user can
download and double-click this EXE directly without installing or extracting
MusicUnlocker. Normal per-user preferences and WebView2 user data remain in
their existing locations; nothing is relocated beside the executable.

The standalone executable never invokes the NSIS registration macros. It does
not create the uninstall key, ProgID, Capabilities, RegisteredApplications,
OpenWithProgids, or UserChoice state. Windows may pass manually selected
"Open with" paths to it, and explicit path arguments use the existing Stage 5D
queue/single-instance flow without starting a batch.

## File associations

The shared versioned ProgID is `MusicUnlocker.EncryptedMusic.1`, with the
installed executable as its friendly icon and this quoted open command:

```text
"%LOCALAPPDATA%\Programs\MusicUnlocker\MusicUnlocker.exe" "%1"
```

The exact 49 registered extensions are:

```text
.666c6163 .6d3461 .6d7033 .6f6767 .776176
.bkcape .bkcflac .bkcm4a .bkcmp3 .bkcogg .bkcwav .bkcwma
.kgg .kgm .kgma .kwm
.mflac .mflac0 .mflac1 .mflaca .mflach .mflacl .mflacm
.mgg .mgg0 .mgg1 .mgga .mggh .mggl .mggm .mmp4
.ncm .qmc0 .qmc2 .qmc3 .qmc4 .qmc6 .qmc8 .qmcflac .qmcogg
.tkm .tm0 .tm2 .tm3 .tm6 .vpr .x2m .x3m .xm
```

Normal decoded formats (`.aac`, `.flac`, `.m4a`, `.mp3`, `.ogg`, `.wav`, and
`.wma`) are excluded. Compound `.kgm.flac` and `.vpr.flac` hints remain
application logic because Windows associates only the final extension.

The installer writes under HKCU only:

- `Software\Classes\MusicUnlocker.EncryptedMusic.1`;
- each `Software\Classes\.<ext>\OpenWithProgids` value owned by MusicUnlocker;
- `Software\MusicUnlocker\Capabilities` and its `FileAssociations` values;
- the `MusicUnlocker` value under `Software\RegisteredApplications`.

Both install and uninstall send supported `SHCNE_ASSOCCHANGED` Shell
notifications. No UserChoice hash, competing ProgID, or default application is
written. Explorer activation therefore passes a path into the existing Stage
5D launch inbox and waits for explicit Start. Portable extraction and launch
never run these installer-owned macros.

## WebView2 strategy

Release builds use Wails v2's `embed` strategy with the official Microsoft
Edge WebView2 Evergreen bootstrapper. If a suitable runtime exists, the app
launches normally. The NSIS path checks for the runtime and invokes the bundled
bootstrapper when missing; the portable executable also prompts before invoking
its embedded bootstrapper when launch-time preflight finds no suitable runtime.
Runtime acquisition is therefore an explicit Microsoft-supported recovery
step and may require network access, while MusicUnlocker's decoding and
metadata paths remain offline.

No Fixed Version runtime, custom downloader, unofficial mirror, or TLS bypass
is included. The current machine already has WebView2, and it was not removed.
Missing-runtime behavior on a clean machine is
`CLEAN_MACHINE_WEBVIEW2_MISSING_TEST_PENDING`.

## Portable package

The portable ZIP has one top-level versioned directory containing only:

```text
MusicUnlocker.exe
LICENSE.txt
THIRD_PARTY_NOTICES.txt
README.txt
```

The executable embeds the frontend and application assets. No source tree,
frontend sources, Node modules, development scripts, samples, external keys or
databases, preferences, WebView2 user data, or build caches are packaged.
Preferences keep their existing per-user Windows configuration location; the
portable executable does not write them beside itself.
It does not register file associations on extraction or launch.

## Release directory allowlist and legal assets

After transient ZIP staging is removed, the release script accepts exactly the
six public files listed above and rejects any directory or unexpected entry.
The standalone executable is accompanied by separate `LICENSE.txt` and
`THIRD_PARTY_NOTICES.txt` assets (legal Scheme A). The same legal files remain
inside the portable ZIP and are installed by NSIS. The notice covers the
runtime-linked Go modules and bundled React production dependencies under
MIT, Apache-2.0, BSD-2-Clause, and BSD-3-Clause terms. The official Evergreen
bootstrapper remains governed by Microsoft's distribution/software terms; no
Fixed Version runtime or upstream source tree is redistributed.

## Current validation boundary

On the Stage 8A development machine, the production portable build succeeded,
its contents and executable metadata were inspected, and it launched and
closed after extraction into a clean temporary directory outside the
repository. With NSIS 3.12, the production installer was rebuilt and validated
through custom-path uninstall, default-path install, installed and Start Menu
launch, About version `0.7.0`, HKCU `InstallLocation`, same-version reinstall,
and final registered uninstall. The installer requested no UAC elevation.
Both Stage 8A uninstalls removed only MusicUnlocker-owned files, shortcuts, and
HKCU registration; preferences and user music/output content were retained.

Stage 8B additionally passed local first-instance KGM activation,
existing-instance NCM handoff, duplicate activation, Unicode/space paths,
idle/no-output confirmation, conservative association uninstall, portable
registry-neutral launch, and same-version reinstall/reactivation. Uninstall
retained a controlled unrelated OpenWith value, and reinstall produced 49
unique association values. No UserChoice key was created or removed, and no
decoder real-validation status changed.

Stage 8C locally rebuilt and allowlist-audited the installer, standalone EXE,
portable ZIP, top-level legal files, and checksum manifest. After uninstalling
the installed build, the standalone executable alone launched and closed from
an unrelated path containing spaces and Japanese characters, rendered its
frontend/icon, and reported About version `0.7.0`. It generated no adjacent
files. A separate Unicode/space `.ncm` argument entered the queue, remained
`Ready`, and created no output directory. Installer-owned registry keys and
sampled UserChoice keys were absent before and after both launches.

The RC installer passed per-user install, Start Menu, version, 49-association,
argument queue/idle, and uninstall regression on the development host. The
portable ZIP retained exactly its documented internal files, launched from a
separate Unicode/space path, and remained registry-neutral. WebView2
`151.0.4129.101` was present. This machine is not clean, so
`CLEAN_MACHINE_CORE_TEST_PENDING`,
`CLEAN_MACHINE_WEBVIEW2_MISSING_TEST_PENDING`, and
`CLEAN_MACHINE_ASSOCIATION_TEST_PENDING` remain. No representative
Mark-of-the-Web download was used, so `SMARTSCREEN_PUBLIC_DOWNLOAD_TEST_PENDING`
remains. No authentic older installer was available, so
`CROSS_VERSION_UPGRADE_TEST_PENDING` remains. No real KGM/NCM samples were
available for a release-artifact decode/Preview smoke; decoder evidence in
`docs/real-validation.md` is unchanged.

Historical Stage 8D artifact-only remote validation is recorded:

- feature-branch CI passed for `codex/stage-8c-v2` at `59b49f0`;
- master CI passed for `master` at `59b49f0`;
- manual artifact-only release validation passed on a fresh `windows-2025`
  runner and uploaded `MusicUnlocker-0.7.0-windows-release`;
- the downloaded artifact's five `SHA256SUMS` entries were independently
  recomputed and matched exactly.

This establishes `REMOTE_CI_VALIDATED`, `CLEAN_GITHUB_RUNNER_CI_PASS`,
`MASTER_CI_PASS`, `CLEAN_GITHUB_RUNNER_BUILD_PASS`, and
`GITHUB_RUNNER_RELEASE_BUILD_VALIDATED`. The downloaded hashes are
historical commit-specific evidence from `59b49f0`, not the frozen public
RC1 manifest linked above.

The hosted runner proves a fresh checkout, workflow-bootstrapped toolchain, and
production artifact construction. It is not an interactive Windows 11
consumer environment and does not validate Explorer integration, DPI/snap,
WebView2 GUI lifecycle, consumer UAC behavior, or SmartScreen reputation.
The later RC1 Draft job, manual public publication, and anonymous asset
integrity checks passed; `DRAFT_RELEASE_REMOTE_TEST_PENDING` is superseded.
Remaining open: `CLEAN_MACHINE_CORE_TEST_PENDING`,
`CLEAN_MACHINE_ASSOCIATION_TEST_PENDING`,
`CLEAN_MACHINE_WEBVIEW2_MISSING_TEST_PENDING`,
`SMARTSCREEN_PUBLIC_BROWSER_TEST_PENDING`,
`CROSS_VERSION_UPGRADE_TEST_PENDING`, and
`ARTIFACT_ATTESTATION_PENDING`.
