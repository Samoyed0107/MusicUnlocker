# Stage 3A decode pipeline

## Pipeline

```text
Probe → Resolver → Decoder → Decrypted Stream Sniff → Atomic Writer
```

The decode core is independent from the CLI and GUI. It probes encrypted input
by content, resolves a supported decoder, sniffs the decrypted prefix, and
copies the original audio bytes without transcoding.

## Current Stage 3A support

- KGM/VPR Crypto V3: decode supported through the upstream adapter.
- KGM Crypto V5: recognised, but requires the Kugou database and is not implemented.
- QMC external-key formats: recognised, but the key is required and not implemented.
- NCM, KWM, and other encrypted families: recognised where probed, but not wired into the decode resolver.

Unknown decrypted streams fail with `DECODED_AUDIO_FORMAT_UNKNOWN`; they are
never assumed to be MP3.

## Safety guarantees

- source is opened read-only and is never deleted or renamed;
- existing outputs are rejected;
- output is written to a temporary file in the destination directory and renamed only after a successful copy;
- decode performs no network activity;
- decode does not transcode audio;
- the output format is determined from decrypted content, not the input filename.

## Developer CLI

```powershell
go run ./cmd/musicdecode -- --out-dir "D:\decoded-test" "input.kgm"
```

Stage 3A accepts exactly one input file and requires an explicit output
directory. Diagnostic output uses the temporary naming convention
`<original-base-name>.decoded.<actual-format>`. This is deliberately not the
final end-user filename planner.
