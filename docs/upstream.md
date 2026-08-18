# Upstream Decoder Dependency

## Canonical dependency

MusicUnlocker uses the canonical Unlock Music CLI Go module:

- Project/module: Unlock Music CLI
- Module path: `git.um-react.app/um/cli`
- Pinned version: `v0.2.19`
- Upstream commit: `107aa2c3a843e6b30406d45d64789f3c155dd68f`

The fixed version makes the reviewed decoder surface reproducible. Upgrades must be explicit and reviewed.

## Integration model

MusicUnlocker currently consumes the upstream algorithms as a Go dependency through `internal/decoder/upstream`. Decoder implementation files are not copied into this repository, and future MusicUnlocker packages should not import the upstream module directly.

The adapter exposes filename candidate discovery and explicit family adapters.
Application packages never construct upstream `DecoderParams` directly. The
adapters call only `Validate` and `Read`; metadata, cover, and network APIs are
not part of the decode path.

## Decoder families

The v0.2.19 module contains these requested decoder families under `algo`:

- Kugou / KGM: `algo/kgm`
- Kuwo / KWM: `algo/kwm`
- NetEase Cloud Music / NCM: `algo/ncm`
- QQ Music / QMC: `algo/qmc`
- TM / QQ Music iOS: `algo/tm`
- Xiami: `algo/xiami`
- Ximalaya: `algo/ximalaya`

## Filename registrations

The upstream `algo/common` package stores registrations through `RegisterDecoder` and matches lower-cased base filenames with `strings.HasSuffix`. The adapter imports each decoder family for its registration side effects and reads the resulting non-noop registrations.

Verified registered suffixes are:

- KGM/Kugou: `.kgg`, `.kgm`, `.kgma`, `.vpr`, `.kgm.flac`, `.vpr.flac`
- KWM/Kuwo: `.kwm`
- NCM/NetEase: `.ncm`
- QMC/QQ Music: `.qmc0`, `.qmc3`, `.qmc2`, `.qmc4`, `.qmc6`, `.qmc8`, `.qmcflac`, `.qmcogg`, `.tkm`, `.bkcmp3`, `.bkcm4a`, `.bkcflac`, `.bkcwav`, `.bkcape`, `.bkcogg`, `.bkcwma`, `.666c6163`, `.6d7033`, `.6f6767`, `.6d3461`, `.776176`, `.mmp4`, `.mgg`, `.mgg0`, `.mgg1`, `.mgga`, `.mggh`, `.mggl`, `.mggm`, `.mflac`, `.mflac0`, `.mflac1`, `.mflaca`, `.mflach`, `.mflacl`, `.mflacm`
- TM: `.tm2`, `.tm6`, `.tm0`, `.tm3`
- Xiami: `.xm`, `.wav`, `.mp3`, `.flac`, `.m4a`
- Ximalaya: `.x2m`, `.x3m`, `.xm`

Compound suffixes confirmed by the source are `.kgm.flac` and `.vpr.flac`. QMC registers `.mgg0` and `.mflac0` as suffixes, but v0.2.19 does not register `.mgg0.flac` or `.mflac0.flac` as compound suffixes. Because matching is suffix-based, a name such as `track.kgm.flac` can also match the separate `.flac` registration.

These registrations are filename/suffix matching only. They are not binary format validation and do not establish that a file is encrypted, valid, or decodable. Binary validation belongs to a later content-probing design.

## Notable upstream mechanisms

- KGM includes v3 and v5 code. v5 uses an external KGG database path to obtain an audio key.
- QMC includes static, map, and RC4-related cipher paths; QMC parsing also contains QTag, STag, and MusicEx handling.
- QMC contains MMKV-related loaders for platform-specific legacy key sources;
  MusicUnlocker does not invoke them. Static, embedded-key, and QTag
  validation can proceed offline, while STag and keyless MusicEx remain
  externally blocked.
- TM validates the `QQMU` header and repairs it to an M4A-style `ftyp` header,
  or safely passes through ordinary audio for its distinctive TM suffixes.
- Xiami validates `ifmt`, a supported four-byte type, and `FEFEFEFE` at
  offsets `0x08`-`0x0B`.
- Ximalaya has no reliable simple plaintext magic; `.x2m`, `.x3m`, and
  ambiguous `.xm` candidates use controlled upstream validation, with Xiami
  attempted before Ximalaya for `.xm`.

## Upgrade policy

Future upstream upgrades must:

1. use an explicit version change;
2. review release and commit changes;
3. run MusicUnlocker regression tests; and
4. never silently track a moving branch such as `main`, `master`, or `latest`.
