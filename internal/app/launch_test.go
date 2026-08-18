package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLaunchPaths(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "音樂 (live).ncm")
	if err := os.WriteFile(file, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(root, "日本語 文件")
	if err := os.Mkdir(folder, 0700); err != nil {
		t.Fatal(err)
	}
	result := ParseLaunchPaths([]string{file, folder, "missing", "--verbose", filepath.Join(".", filepath.Base(file))}, root)
	if len(result.Paths) != 2 {
		t.Fatalf("paths = %#v", result.Paths)
	}
	if len(result.Rejected) != 2 {
		t.Fatalf("rejected = %#v", result.Rejected)
	}
	if result.Paths[0] != filepath.Clean(file) {
		t.Fatalf("file = %q", result.Paths[0])
	}
}

func TestParseLaunchPathsDoesNotDiscoverDirectoryContents(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "inside.ncm"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	result := ParseLaunchPaths([]string{root}, root)
	if len(result.Paths) != 1 || result.Paths[0] != root {
		t.Fatalf("paths = %#v", result.Paths)
	}
}

func TestLaunchInboxRetainsAndFlushes(t *testing.T) {
	var inbox LaunchInbox
	event := inbox.NewEvent([]string{`C:\Music\Song.NCM`}, LaunchSourceSecond)
	inbox.Defer(event)
	inbox.Defer(event)
	flushed := inbox.FlushDeferred()
	if len(flushed) != 1 || flushed[0].Deferred {
		t.Fatalf("flushed = %#v", flushed)
	}
	inbox.RetainInitial(event)
	if inbox.ConsumeInitial() == nil || inbox.ConsumeInitial() != nil {
		t.Fatal("initial event was not consumed exactly once")
	}
}
