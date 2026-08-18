# MusicUnlocker

MusicUnlocker is a planned Windows desktop application for unlocking supported encrypted music formats while preserving source files and decoded audio formats where possible.

The Go core has been bootstrapped, with upstream decoding algorithms pinned to a fixed dependency. A minimal Wails v2 desktop shell is available for local development.

## GUI development

With the Wails CLI and Node.js installed:

```text
wails dev
wails build
```

The application uses the existing Go module and keeps the decoder core and
batch engine independent from the React presentation layer. See
`docs/gui-architecture.md` for the facade, event, and queue contract.

The desktop UI supports Add Files, Add Folder, native file/folder drag-and-drop,
output selection, Start, Cancel, result filters, and Open Output Folder. Drops
only add local filesystem paths to the queue; they never start decoding.
