# Windows release packaging

Stages 8A and 8B provide one auditable Windows x64 workflow for the Wails v2
NSIS installer and portable ZIP, including per-user encrypted-container file
associations. They do not add an updater, signing, CI publishing, or a portable
preference mode.

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
MusicUnlocker-0.7.0-win64-portable.zip
SHA256SUMS
```

`SHA256SUMS` uses lowercase SHA-256 followed by two spaces and the artifact
filename. Checksums detect transfer corruption; they do not authenticate the
publisher and are not a substitute for code signing. Code signing remains
`NOT_DECIDED`.

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
`CLEAN_MACHINE_WEBVIEW2_TEST_PENDING`.

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

`CLEAN_MACHINE_WEBVIEW2_TEST_PENDING` and
`CLEAN_MACHINE_ASSOCIATION_TEST_PENDING` and
`CROSS_VERSION_UPGRADE_TEST_PENDING` remain open. Stage 8C owns the broader
clean-machine release-readiness matrix.
