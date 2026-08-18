package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type recordingOpener struct{ path string }

func (o *recordingOpener) Open(path string) error { o.path = path; return nil }

func TestSuggestOutputFolder(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "song.kgm")
	if err := os.WriteFile(file, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := SuggestOutputFolder([]string{file})
	if err != nil || got != filepath.Join(root, "Unlocked") {
		t.Fatalf("file suggestion=%q err=%v", got, err)
	}
	dir := filepath.Join(root, "Music")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	got, err = SuggestOutputFolder([]string{dir})
	if err != nil || got != filepath.Join(root, "Music - Unlocked") {
		t.Fatalf("folder suggestion=%q err=%v", got, err)
	}
	got, err = SuggestOutputFolder([]string{file, dir})
	if err != nil || got != "" {
		t.Fatalf("ambiguous suggestion=%q err=%v", got, err)
	}
}

func TestOpenOutputFolderValidation(t *testing.T) {
	root := t.TempDir()
	opener := &recordingOpener{}
	if err := openOutputFolder(root, opener); err != nil {
		t.Fatal(err)
	}
	if opener.path != root {
		t.Fatalf("opened %q", opener.path)
	}
	file := filepath.Join(root, "decoded.flac")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := openOutputFolder(file, opener); err != nil || opener.path != root {
		t.Fatalf("file parent path=%q err=%v", opener.path, err)
	}
	if err := openOutputFolder(filepath.Join(root, "missing"), opener); !errors.Is(err, ErrInvalidOpenPath) {
		t.Fatalf("missing error=%v", err)
	}
	if err := openOutputFolder("relative", opener); !errors.Is(err, ErrInvalidOpenPath) {
		t.Fatalf("relative error=%v", err)
	}
}
