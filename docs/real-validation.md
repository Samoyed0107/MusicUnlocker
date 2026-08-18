# Real validation backlog

This document records validation status without storing real audio files.

| Family | Status | Evidence or note |
| --- | --- | --- |
| KGM V3 | `REAL_PASS` | Real decode, MP3 sniff despite `.flac` source suffix, 5,641,481-byte payload, playback/seeking, clean naming, and no-overwrite validated. |
| NCM | `REAL_PASS` | Real NCM -> MP3 decode completed; output 3,977,343 bytes; SHA-256 `AAED5081FD7DD1D1C8CFE24E2E644A05A40419775A8A551837B2950A38201762`; ID3v2.4 and MPEG frame confirmed; GUI completion, playback, and seek confirmed. |
| KWM | `REAL_PENDING` | Code path is wired and tested; real sample still needed. |
| QMC STag | `PROBE_REAL_PASS_DECODE_BLOCKED` | Real STag sample was confirmed; external key is required. |
| KGM v5/KGG | `REAL_PENDING_EXTERNAL_DATABASE` | Database support is intentionally deferred. |
| TM | `REAL_PENDING` | Synthetic content and adapter validation are tested. |
| Xiami | `REAL_PENDING` | Synthetic structural probe and malformed adapter tests are present. |
| Ximalaya | `REAL_PENDING` | Controlled `.x2m`, `.x3m`, and `.xm` candidate validation is tested synthetically. |
| QMC self-contained variants | `REAL_PENDING` | QTag/static/embedded-key paths remain offline-only and need real samples. |

The known external NCM regression sample is `normal_mp3-1773267.ncm`, with
SHA-256 `cde0d8bcbe4ebf08daba9f137c965248b0d2a4968415c14d5bac35a9b2f19dbf`.
It is not copied into this repository.

## Stage 5C real GUI validation

- NCM decoder: `REAL_PASS`.
- Wails GUI NCM pipeline: `REAL_PASS`.
- MP3 in-app Preview: `REAL_PASS`; playback was confirmed at approximately 4:08.
- HTTP Range/seek through PreviewService: `REAL_PASS`.
- Wails production Windows build: `REAL_PASS`.

No real audio or decoded output is stored in this repository.
