# Format Probe

## Purpose

MusicUnlocker separates filename hints, binary content probing, future decoder validation, and decoding. A suffix is only a dispatch candidate; it is never proof of an encrypted format. The probe performs no decryption, key acquisition, output writing, metadata retrieval, or network activity.

## Detection pipeline

```text
filename -> extension hints --+
                              +-> bounded header/footer probe -> evidence -> Probe Result -> future Decoder Validate
file bytes -> structural probe+
```

## Currently confirmed structural probes

- KGM and VPR: their canonical 16-byte magic; KGM-compatible header fields at `0x10`, `0x14`, and `0x18` are reported when present.
- NCM: `CTENFDAM` at the beginning of the file.
- KWM: either canonical 16-byte `yeelion-kuwo-tme` or `yeelion-kuwo\0\0\0\0` header.
- QMC: final `STag` or `QTag`, and MusicEx version 1 with a plausible tag size (at least `0xC0`, no larger than the file, and no larger than the probe limit).
- Raw audio: `fLaC`, `OggS`, `RIFF....WAVE`, ID3 MP3, and structurally plausible ISO BMFF `ftyp` (including M4A brands).

## Limitations

Legacy QMC variants may lack a unique plaintext signature. Suffix-only QMC candidates remain unconfirmed until a future decoder-validation stage. Ximalaya, Xiami, TM, and other formats may initially rely on candidate dispatch until content evidence is formally implemented. Format Probe does not obtain DRM/media keys and does not decode audio.

## Security/resource limits

The probe reads at most 4096 bytes from the beginning and recognizes footer structures from small tail reads. MusicEx validation refuses sizes over 64 KiB before any tag allocation or read. Short and malformed files return an unknown or hint-only result without panicking. Reader position is restored on return.

## Developer diagnostic CLI

`musicprobe` is a development diagnostic tool, not the final end-user CLI. It accepts one or more file paths, does not recurse into directories, and never decodes or writes audio.

```powershell
go run ./cmd/musicprobe -- <file>
go run ./cmd/musicprobe --json -- <file>
```

Human output is intended to be copied into bug reports. JSON output contains a `files` array with path, size, probe result, and an operational `error` field when opening or probing fails. Unknown formats are successful probes and exit with code 0; a missing or unreadable file makes the command exit nonzero.

Filename hints conservatively normalize a terminal numeric copy counter such as `song.mgg0(1).flac` or `song.mgg0 (2).flac` only when removing it exposes a suffix registered by the decoder candidate system. This is hint extraction only and never changes the filesystem name; arbitrary title parentheses are not normalized.

In JSON, unavailable optional format fields are `null`, never numeric sentinel values. Evidence objects use the stable lowercase fields `kind` and `detail`.
