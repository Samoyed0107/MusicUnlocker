# Stage 3C decode pipeline

## Pipeline

```text
Probe → Local Container Metadata → Resolver → Decoder → Decrypted Stream Sniff → Temporary Writer → Decoded Metadata → Merge → Naming → Ordered Atomic Commit
```

The core remains independent from the CLI and GUI. Content evidence has
priority over filename hints. When a family has no strong plaintext signature,
only a deterministic, explicitly allowed filename candidate can trigger
upstream `Validate`.

## Decoder coverage

| Family | Implementation | Automated tests | Real validation | External material |
| --- | --- | --- | --- | --- |
| KGM/VPR Crypto V3 | implemented | tested | validated | none |
| NCM | implemented offline | tested | pending | none |
| KWM | implemented offline | tested | pending | none |
| QMC static/embedded-key/QTag candidates | implemented offline | tested at adapter boundary | pending | none for self-contained files |
| QMC STag/MusicEx | supplied-key adapter; missing material blocked | provider/adapter tested | probe real-pass | real supplied-key validation pending |
| TM | implemented offline | tested | pending | none |
| Xiami | implemented offline | tested | pending | none |
| Ximalaya | controlled candidate validation | tested | pending | none |

KGM Crypto V5 remains database-dependent. Other encrypted families remain
unsupported for this stage.

## Detection and fallback policy

Strong TM and Xiami content signatures route directly to their adapters. QMC
footer evidence routes STag and MusicEx to the typed external-key state, while
QTag may use the canonical self-contained decoder.

For files without strong content evidence, only these distinctive candidates
are considered, in deterministic priority order:

1. QMC variants (`.qmc*`, `.mgg*`, `.mflac*`, `.tkm`)
2. TM (`.tm0`, `.tm2`, `.tm3`, `.tm6`)
3. Xiami then Ximalaya for ambiguous `.xm`
4. Ximalaya (`.x2m`, `.x3m`)
5. NCM (`.ncm`) and KWM (`.kwm`)

Generic raw extensions such as `.mp3`, `.flac`, `.wav`, `.m4a`, and `.mp4`
never trigger encrypted decoder brute force.

## Output Planner

`internal/outputplan` derives the extension from decrypted-content sniffing.
It removes only recognised encrypted wrappers and terminal duplicate counters:

- `foo.tm2` + M4A → `foo.m4a`
- `foo.x2m` + MP3 → `foo.mp3`
- `foo.xm` + FLAC → `foo.flac`
- `Artist - Song (Live).ncm` + MP3 → `Artist - Song (Live).mp3`

Ordinary title parentheses remain intact. Existing targets return
`OUTPUT_EXISTS`; a planned target equal to the source returns
`OUTPUT_CONFLICTS_SOURCE`. Both preserve the source file.

## Safety and offline guarantees

- source is read-only and never deleted or renamed;
- output is copied to a temporary file and atomically renamed;
- no transcoding is performed;
- NCM, KWM, QMC, TM, Xiami, and Ximalaya adapters call only local upstream validation and stream operations;
- metadata naming reads only local NCM container fields and bounded local tags from decoded temporary output; decoded tags are primary and NCM fills missing Title/Artist/Album only. Cover-art, lyrics, key acquisition, and network APIs are not used. Container metadata failure never invalidates a successful decode.

## Developer CLI

```powershell
go run ./cmd/musicdecode -- --out-dir "D:\decoded-test" "input.tm2"
```

The CLI accepts exactly one file and displays the clean planned output name.
