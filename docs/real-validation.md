# Real validation

This document separates implementation and automated evidence from tests using
real external media. No real audio, decoded output, EKey, or KGG database is
stored in the repository.

## Status vocabulary

- `IMPLEMENTED` — the code path exists; this alone claims neither automated nor
  real-file success.
- `AUTOMATED_TESTED` — synthetic/unit/integration or smoke coverage passed
  without treating copyrighted real media as a repository fixture.
- `ENGINEERING_COMPLETE` — the intended engineering scope and hardening are
  complete even if external-material validation remains.
- `PROBE_REAL_PASS` — a real file's family/variant evidence was identified; it
  does not prove decoding.
- `REAL_PASS` — the described behavior was observed with a real external file.
- `REAL_PENDING` — the implementation exists, but compatible real-file evidence
  has not yet been recorded.
- `PENDING_EXTERNAL_MATERIAL` — real validation cannot proceed without a
  compatible key, database, or paired external input.

Statuses apply only to the scope named in each row. A probe pass is not a decode
pass, and engineering completion is not silent promotion of real compatibility.

## Current real-validation matrix

| Scope | Engineering evidence | Real status | Evidence or remaining requirement |
| --- | --- | --- | --- |
| KGM V3 decode | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PASS` | Real MP3 output, content sniff despite a misleading suffix, playback/seek, source preservation, clean naming, and no-overwrite behavior validated. |
| Stage 5H KGM metadata quality/fallback | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PASS` | Suspicious KGM metadata no longer polluted the filename; safe fallback retained identical decoded bytes. |
| NCM decode | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PASS` | Real NCM-to-MP3 decode, GUI completion, playback/seek, and output hash validated. |
| NCM metadata naming | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PASS` | Real keep-source and metadata-aware names were produced with identical decoded bytes. |
| QMC STag probe | `IMPLEMENTED / AUTOMATED_TESTED` | `PROBE_REAL_PASS` | A real STag sample was identified. |
| QMC STag supplied-key decode | `ENGINEERING_COMPLETE / AUTOMATED_TESTED` | `PENDING_EXTERNAL_MATERIAL` | A compatible real EKey is unavailable. |
| KGM V5/KGG core and GUI retry | `ENGINEERING_COMPLETE / AUTOMATED_TESTED` | `PENDING_EXTERNAL_MATERIAL` | A compatible real KGM V5/KGG file and database are unavailable. |
| KWM | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PENDING` | Compatible real sample still needed. |
| TM | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PENDING` | Compatible real sample still needed. |
| Xiami | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PENDING` | Compatible real sample still needed. |
| Ximalaya | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PENDING` | Compatible real sample still needed. |
| QMC self-contained variants | `IMPLEMENTED / AUTOMATED_TESTED` | `REAL_PENDING` | QTag/static/embedded-key paths need compatible real samples. |

## Stage 5H real KGM metadata-quality retest

The real KGM V3 sample was retested after Stage 5H. Both `KEEP_SOURCE` and
metadata-aware naming produced the safe readable output name:

```text
茅野愛衣、戸松遥、早见沙织 - secret base ～君がくれたもの～ (secret base ~你给我的所有~)(10 years after Ver.)_MQ.mp3
```

Both outputs were:

- Length: 5,641,481 bytes
- SHA-256:
  `57BA83391720592FD2DED2380033162DE51C17FDFB80CA6461105766FC4273B6`

The previous corrupted metadata-derived filename containing replacement or
garbage Unicode was no longer produced. This is `REAL_PASS` evidence that the
Stage 5H guard prevents suspicious KGM metadata from polluting the final
filename for this sample and that safe fallback preserves decoded audio bytes.

This does **not** prove that the underlying malformed KGM metadata was
successfully re-decoded into corrected metadata. Stage 5H protects naming; no
separate KGM metadata-encoding repair is claimed.

## NCM real regression

The real NCM sample was retested with:

- `KEEP_SOURCE`:
  `Gabrielle Aplin - Salvation(1).mp3`
- metadata-aware `TITLE_ARTIST`:
  `Salvation - Gabrielle Aplin.mp3`

Both outputs were:

- Length: 3,977,343 bytes
- SHA-256:
  `AAED5081FD7DD1D1C8CFE24E2E644A05A40419775A8A551837B2950A38201762`

This reconfirms real NCM decode, NCM metadata naming, and byte identity across
naming modes. Earlier Stage 5G tests also observed `TITLE`,
`ARTIST_TITLE`, custom metadata templates, and source-only `{source}` naming
with the same decoded hash.

The known external NCM regression input is `normal_mp3-1773267.ncm` with
SHA-256
`cde0d8bcbe4ebf08daba9f137c965248b0d2a4968415c14d5bac35a9b2f19dbf`.
It is not copied into the repository.

## Desktop real evidence

Recorded real desktop evidence includes:

- Wails production Windows build: `REAL_PASS`.
- Real NCM GUI batch completion: `REAL_PASS`.
- MP3 PreviewService playback and HTTP Range/seek: `REAL_PASS`.
- Production GUI startup after the localization correction: `REAL_PASS`.
- English, Simplified Chinese, and system-language behavior:
  `REAL_PASS`.
- Source-only custom naming correction: `REAL_PASS`.
- Playback continuity across language switching: `REAL_PASS`.

Stage 7C theme persistence/migration and the frontend build, localization,
preview, retry, and theme/About smoke checks are `AUTOMATED_TESTED`. A
browser-rendered dark theme and localized About interaction with a mocked Wails
facade have visual-smoke evidence. Native title-bar synchronization across all
theme modes, live Windows system-theme changes, and older-Windows opaque
fallback remain manual validation items; these do not reopen Stage 7C
engineering status.

Physical Explorer launch/drag-to-EXE behavior remains part of the clean-machine
release-readiness checklist. Launch arguments and single-instance queue handoff
are implemented and automated-tested, and they never auto-start decoding.

## External-material backlog

QMC STag supplied-key support and its backend-owned GUI retry are
`ENGINEERING_COMPLETE / AUTOMATED_TESTED`. The real STag probe is
`PROBE_REAL_PASS`, but supplied-key decoding remains
`PENDING_EXTERNAL_MATERIAL` because no compatible EKey is available.

KGM V5/KGG database support and its backend-owned GUI retry are
`ENGINEERING_COMPLETE / AUTOMATED_TESTED`. Synthetic matching and missing
`AudioHash` database cases are covered. Real database-backed decoding remains
`PENDING_EXTERNAL_MATERIAL` because no compatible file/database pair is
available.

These are explicit validation backlog items and do not block the next product
stage. Material remains offline, request-scoped, non-persistent, and absent from
frontend state, results, preferences, and logs.
