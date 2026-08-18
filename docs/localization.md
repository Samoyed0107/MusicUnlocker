# GUI localization

Stage 5F bundles GUI resources locally and supports `SYSTEM`, `zh-CN`, and
`en-US`. `SYSTEM` is the default and is persisted as `SYSTEM`, never as the
resolved locale. `zh-CN`, `zh-SG`, and `zh-Hans` resolve to Simplified Chinese;
other locales, including unsupported Traditional Chinese, resolve to English.
Resolution uses WebView browser language APIs at startup and never uses the
network.

Resources are `frontend/src/i18n/en-US.ts` and `zh-CN.ts`; the typed provider,
interpolation, resolution, English fallback, and key/interpolation-parity
validation are in `index.tsx` and `validation.ts`. Strict validation runs in
the frontend i18n smoke test; runtime initialization does not throw because of
one non-English resource mismatch. Such a key falls back to English and logs
diagnostics in Vite development mode. UI code uses semantic keys and
`t(key, vars)`. English is complete.

Switching language is live and presentation-only: it does not restart or
cancel a batch, clear the queue, change naming or output paths, or stop audio
preview. Missing/invalid preference values recover to `SYSTEM`.

Statuses, error codes, naming modes, paths, metadata, and custom naming tokens
remain machine/data values. Tokens remain `{title}`, `{artist}`, `{album}`,
`{albumartist}`, `{track}`, `{disc}`, `{year}`, and `{source}`.

Future languages should add a resource with the exact English key set, register
it, and add explicit preference/resolution policy. Stage 5F ships only Chinese
and English resources; all resources are bundled offline.
