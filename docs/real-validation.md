# Real validation backlog

This document records validation status without storing real audio files.

| Family | Status | Evidence or note |
| --- | --- | --- |
| KGM V3 | `REAL_PASS` | Real decode, MP3 sniff despite `.flac` source suffix, 5,641,481-byte payload, playback/seeking, clean naming, and no-overwrite validated. |
| NCM | `REAL_PASS` | Real NCM -> MP3 decode completed; output 3,977,343 bytes; SHA-256 `AAED5081FD7DD1D1C8CFE24E2E644A05A40419775A8A551837B2950A38201762`; ID3v2.4 and MPEG frame confirmed; GUI completion, playback, and seek confirmed. |
| KWM | `REAL_PENDING` | Code path is wired and tested; real sample still needed. |
| QMC STag | `PROBE_REAL_PASS_DECODE_BLOCKED` | Real STag sample was confirmed; external key is required. Stage 6A core is automated-tested, but no real key was available. |
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

## Stage 5E2 and 5F acceptance regression

- NCM decoding: `REAL_PASS` (existing evidence preserved).
- Audio integrity SHA-256: `REAL_PASS` (existing evidence preserved).
- Stage 5E2 metadata-dependent naming on the known NCM: metadata fallback was
  observed because decoded-audio metadata is unavailable under the current
  design; this is not by itself a Core failure.
- Stage 5E2 CUSTOM `{source}` GUI naming: `REAL_FAIL / FIX_IMPLEMENTED /
  REAL_RETEST_PENDING`. The source-only template was incorrectly forced through
  metadata-availability fallback; no real audio was used for the automated fix.
- Stage 5F production GUI: `REAL_FAIL / FIX_IMPLEMENTED /
  REAL_RETEST_PENDING`. Debug DevTools identified the first exception as
  `Localization interpolation mismatch: batch.itemsAdded`; the favicon 404 was
  non-causal. A corrected production EXE has not yet been physically confirmed.

## Stage 5F1 real validation

- Production GUI startup after the localization fix: `REAL_PASS`.
- English, Simplified Chinese, and follow-system localization: `REAL_PASS`.
- Stage 5E2 CUSTOM `{source}` naming: `REAL_PASS`; observed output was
  `Gabrielle Aplin - Salvation(1)-TEST.mp3`.
- Decoded audio SHA-256: `REAL_PASS`; `AAED5081FD7DD1D1C8CFE24E2E644A05A40419775A8A551837B2950A38201762`.

## Stage 5F2 playback continuity

- Language switching presentation: manually observed working.
- Playback continuity across language switching: `REAL_PASS`. Audio continued
  without restart, the paused position remained stable, and the revised batch
  status badge placement was manually accepted.

## Stage 5G NCM naming

- Implementation and automated synthetic validation: `IMPLEMENTATION_COMPLETE /
  AUTOMATED_TESTED`.
- Real NCM naming retest: `COMPLETE / REAL_PASS`.

Stage 5G real evidence: `KEEP_SOURCE` -> `Gabrielle Aplin - Salvation(1).mp3`;
`TITLE` -> `Salvation.mp3`; `ARTIST_TITLE` -> `Gabrielle Aplin - Salvation.mp3`;
`TITLE_ARTIST` -> `Salvation - Gabrielle Aplin.mp3`; `{title}-{artist}` ->
`Salvation-Gabrielle Aplin.mp3`; `{album} - {title}` ->
`English Rain - Salvation.mp3`; `{source}-TEST` ->
`Gabrielle Aplin - Salvation(1)-TEST.mp3`. All variants preserved identical
decoded bytes: 3,977,343 bytes, SHA-256
`AAED5081FD7DD1D1C8CFE24E2E644A05A40419775A8A551837B2950A38201762`.
Controlled isolated CUSTOM retests did not reproduce earlier ambiguous fallback
observations; no confirmed product bug remains.

## Stage 6B GUI retry validation status

- GUI external-key retry workflow: `IMPLEMENTATION_COMPLETE / AUTOMATED_TESTED`.
- Real QMC STag probe: `REAL_PASS`.
- Real supplied-key decode: `REAL_RETEST_PENDING / EXTERNAL_KEY_MATERIAL_REQUIRED`.
- No real QMC file or EKey was accessed during Stage 6B implementation.

## Stage 6C KGM V5/KGG validation status

- KGM V3: `REAL_PASS` (existing evidence unchanged).
- KGM V5/KGG database Core: `IMPLEMENTATION_COMPLETE / AUTOMATED_TESTED`.
- Synthetic plaintext SQLite lookup: passed; matching and missing `AudioHash`
  cases were exercised without copyrighted or user data.
- Real KGM V5/KGG supplied-database decode:
  `REAL_RETEST_PENDING / EXTERNAL_DATABASE_REQUIRED`.
- No real KGG database or KGM V5 file was accessed.

## Stage 6D KGG database GUI retry status

- KGG database GUI retry: `IMPLEMENTATION_COMPLETE / AUTOMATED_TESTED`.
- Native database selection, backend-only path lifetime, per-item result
  replacement, summary recomputation, localization, and preview continuity
  are covered by automated tests and smoke checks.
- Real KGM V5/KGG database-backed GUI decode:
  `REAL_RETEST_PENDING / EXTERNAL_DATABASE_REQUIRED`.
- No real KGG database, KGM V5 file, or installed Kugou application data was
  accessed during Stage 6D.

## Stage 6 external-material validation backlog

These items are validation backlog only and do not block Stage 6 engineering
closure:

- QMC STag supplied EKey: a compatible real file was probed, but a real EKey
  is unavailable; supplied-key decode remains `PENDING_EXTERNAL_MATERIAL`.
- KGM V5/KGG database-backed decode: Core and GUI are implemented and
  automated-tested, but a compatible real file/database is unavailable;
  database-backed decode remains `PENDING_EXTERNAL_MATERIAL`.

The debug production EXE was still running when `wails dev` was started during
the failed investigation. Because Stage 5D uses a fixed Wails
`SingleInstanceLock` ID, that run does not establish that `wails dev` alone
white-screened. A valid dev reproduction requires closing all MusicUnlocker
instances first, then running `wails dev` alone.

## Stage 7C Windows-native experience

- Theme preference migration/persistence and default: `AUTOMATED_TESTED`.
- Frontend build, localization, preview, retry, and theme/About smoke checks:
  `AUTOMATED_TESTED`.
- Browser-rendered dark theme and localized modal About interaction with
  a mocked Wails facade: `VISUAL_SMOKE_PASS`.
- Native Wails title-bar synchronization in `SYSTEM`, `LIGHT`, and `DARK`:
  `REAL_MANUAL_VALIDATION_PENDING`.
- Live Windows system-theme change while `SYSTEM` is selected:
  `REAL_MANUAL_VALIDATION_PENDING`.
- Windows 10/older-Windows opaque fallback: `REAL_MANUAL_VALIDATION_PENDING`.

No native Mica backdrop was enabled. The Wails v2.15 API supports it only on
Windows 11 build 22621 or later and requires changing transparent/translucent
window composition. The documented opaque fallback preserves portable EXE
behavior without Windows App SDK. No real audio was used and no existing real
decoder validation status changes in Stage 7C.
