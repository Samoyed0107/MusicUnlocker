# MusicUnlocker

MusicUnlocker is a Windows 11 x64-first desktop utility for decoding supported
encrypted music files locally. It uses a shared Go core with a Wails v2 and
React/TypeScript desktop frontend; the same probe, decode, batch, naming, and
output rules are used by the GUI and developer CLIs.

Current format-family support includes KGM/VPR, NCM, KWM, QMC, TM, Xiami, and
Ximalaya paths. Support is not the same as real-file validation: some variants
remain validation-pending, and QMC STag or KGM V5/KGG inputs require compatible
external material supplied explicitly at runtime. See
[`docs/real-validation.md`](docs/real-validation.md) for the evidence matrix.

The application is offline-first, preserves source files, never silently
overwrites outputs, writes decoded output atomically, and avoids transcoding.
The final output extension comes from decoded-content sniffing rather than the
encrypted filename suffix. Metadata and naming failures fall back safely and
do not alter decoded audio bytes.

The desktop application provides native file/folder selection, recursive batch
processing, native drop, launch-argument and single-instance queue handoff,
metadata-aware naming, localization, secure local preview, external-material
retry flows, System/Light/Dark themes, and About/version identity.

## Developer build

Install the repository's Go and frontend dependencies plus the Wails v2 CLI,
then run:

```powershell
wails dev
wails build
```

Architecture and current handoff details are in
[`docs/gui-architecture.md`](docs/gui-architecture.md) and
[`docs/project-status.md`](docs/project-status.md). Planned work is tracked only
in [`docs/roadmap.md`](docs/roadmap.md).

MusicUnlocker is not a streaming client, downloader, cloud service, mandatory
transcoder, or full music-library manager.
