# Real validation backlog

This document records validation status without storing real audio files.

| Family | Status | Evidence or note |
| --- | --- | --- |
| KGM V3 | `REAL_PASS` | Real decode, MP3 sniff despite `.flac` source suffix, 5,641,481-byte payload, playback/seeking, clean naming, and no-overwrite validated. |
| NCM | `REAL_PENDING` | Code path is wired and tested; 3,941,989-byte external regression sample is available for a later manual run. |
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
