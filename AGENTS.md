# Engineering Rules

- Make changes in small, reviewable stages.
- Preserve source audio files.
- Never silently overwrite outputs.
- Distinguish format detection from decoding.
- Do not infer the true encrypted format solely from the final filename extension.
- Keep decoder implementations modular.
- Prefer offline operation.
- Network metadata retrieval must remain optional.
- Avoid unnecessary audio transcoding.
- Every supported encrypted format should eventually have regression tests.
- Do not commit real copyrighted music samples.
- Access upstream decoder code through the adapter boundary; do not add direct upstream imports elsewhere without explicit architectural approval.
- Keep dependency versions pinned.
- Treat extension matching as filename candidate discovery only, never as the final format-detection mechanism.
- Format probing must use bounded I/O and preserve the reader position.
- Probe code must not perform network activity; content evidence outranks filename hints.
- Malformed files must never cause unbounded allocation or a panic.
