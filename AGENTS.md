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
