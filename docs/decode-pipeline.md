# Stage 3A decode pipeline

## Pipeline

```text
Probe → Resolver → Decoder → Decrypted Stream Sniff → Atomic Writer
```

The decode core is independent from the CLI and GUI. It probes encrypted input
by content, resolves a supported decoder, sniffs the decrypted prefix, and
copies the original audio bytes without transcoding.

## Current Stage 3B support

- KGM/VPR Crypto V3: decode supported through the upstream adapter.
- NCM: wired through the offline upstream adapter.
- KWM: wired through the offline upstream adapter.
- KGM Crypto V5: recognised, but requires the Kugou database and is not implemented.
- QMC external-key formats: recognised, but the key is required and not implemented.
- Other encrypted families: recognised where probed, but not wired into the decode resolver.

Unknown decrypted streams fail with `DECODED_AUDIO_FORMAT_UNKNOWN`; they are
never assumed to be MP3.

## Safety guarantees

- source is opened read-only and is never deleted or renamed;
- existing outputs are rejected;
- output is written to a temporary file in the destination directory and renamed only after a successful copy;
- decode performs no network activity;
- decode does not transcode audio;
- the output format is determined from decrypted content, not the input filename.

## Output Planner

The reusable `internal/outputplan` package creates a clean destination name from
the source basename, encrypted-family probe result, and actual decrypted audio
format. It strips only recognised encrypted wrappers and terminal duplicate
counters such as `foo.kgm(1).flac`; arbitrary title parentheses remain intact.
Misleading outer audio extensions are removed when encrypted content is
confirmed. Thus `foo.kgm(1).flac` decoded as MP3 becomes `foo.mp3`, and an
extensionless `foo` becomes `foo.mp3` when the decrypted format is MP3.

The planner detects source-path conflicts, while the decode core rejects an
existing destination with `OUTPUT_EXISTS`. Neither case truncates the source.
NCM and KWM decoding only calls the local upstream `Validate` and `Read`
interfaces; metadata, cover-image, and network APIs are not invoked.

## Developer CLI

```powershell
go run ./cmd/musicdecode -- --out-dir "D:\decoded-test" "input.kgm"
```

Stage 3B accepts exactly one input file and requires an explicit output
directory. Output uses the clean planned name, for example `foo.mp3`, rather
than the old `.decoded` diagnostic suffix.
