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
- Never guess a decoded audio format or default an unknown decrypted stream to MP3.
- Decode outputs must be written atomically; source audio is never deleted by default.
- The decoding core remains independent from any GUI.
- External-key and database requirements must be represented as typed states or errors.
- The Batch Engine must remain UI-independent and sequential until a later stage.
- Per-file failures must not terminate a batch; output-exists is a skip, not a fatal batch error.
- External-material requirements are batch statuses, not generic failures.
- Batch directory traversal must not follow symlink loops and result ordering must be deterministic.
- Batch processing preserves immutable sources and atomically written outputs.
- Parallel batch execution must preserve deterministic result ordering.
- Items sharing an output target namespace must be serialized in queue order.
- Batch observer and progress APIs remain UI-independent.
- Cancellation must wait for worker and filesystem cleanup before returning.
- Frontend code contains presentation and state-mapping logic only; core filesystem, discovery, output, and decode semantics stay in Go.
- Wails-facing DTOs must not expose upstream decoder types or Go errors directly.
- The GUI permits only one active batch at a time and uses the final batch-complete result as authoritative.
- GUI queue rows correlate by the stable batch index; selected input paths remain distinct from discovered files.
- Native Wails file-drop paths remain OS-supplied absolute paths; frontend drops only mutate the selected-input queue and never start work.
- GUI preferences are small, non-sensitive, per-user JSON and must not contain input history, keys, or raw filenames.
- Preview URLs use authorized opaque in-memory IDs; frontend calls cannot provide arbitrary local paths.
- Preview failure never changes a decoder result status, and preview must not send complete audio through Wails/base64 IPC.
- Major development stages end with project-status/roadmap synchronization; product-requirements changes are made only when requirements change, and real-validation changes only when real validation occurs.
- Decoded audio metadata is read offline only after the decoded temporary file is closed; metadata naming failure falls back to the cleaned source stem without invalidating a successful decode.
- Naming templates never control the decoded extension or create directories; final extensions remain authoritative from decoded-content sniffing.
- Metadata-derived destinations are committed in deterministic queue-index order after concurrent preparation, with uncommitted temporary files cleaned on cancellation or conflict.
- GUI naming validation and rendering examples must come from the Go `internal/naming` backend; the frontend must not implement a second template parser.
- Each active batch captures its validated naming configuration immutably; later preference edits do not affect that batch.
- User-visible frontend strings use bundled localization keys; English is the complete fallback.
- Core status/error enum values and naming template tokens are never translated.
- Decoded embedded metadata is primary; trusted offline container metadata may
  fill only missing naming fields.
- Container metadata never controls the decoded extension and is not written
  into output audio.
- Container metadata failure must not invalidate an otherwise successful decode.
- External decryption material must be explicitly request-scoped, per-file,
  offline, and non-persistent; it must never appear in logs, results, or user
  errors. Missing material is a blocked state, not a generic decode failure.
- Frontend code must never receive raw external decryption material or its
  selected file path; native selection and consumption remain backend-owned.
- External decoder databases are explicit runtime inputs only. Database access
  must be read-only with no automatic user-data scanning or sidecar mutation;
  KGM V3 must not depend on KGG database support.
