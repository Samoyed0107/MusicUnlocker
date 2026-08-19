package externalmaterial

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestMapProviderIsPerFileAndConcurrent(t *testing.T) {
	p := NewMapProvider(map[string][]byte{"a": []byte("A"), "b": []byte("B")})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		for path, want := range map[string]string{"a": "A", "b": "B", "c": ""} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				got, ok, err := p.ResolveQMCExternalKey(context.Background(), path)
				if err != nil || string(got) != want || ok != (want != "") {
					t.Errorf("path=%s got=%q ok=%v err=%v", path, got, ok, err)
				}
			}()
		}
	}
	wg.Wait()
	got, _, _ := p.ResolveQMCExternalKey(context.Background(), "a")
	got[0] = 'X'
	got, _, _ = p.ResolveQMCExternalKey(context.Background(), "a")
	if string(got) != "A" {
		t.Fatalf("provider material was mutable")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := p.ResolveQMCExternalKey(ctx, "a"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}

func TestKGGDatabaseProviderIsPerFileAndConcurrent(t *testing.T) {
	p := NewMapKGGDatabaseProvider(map[string]KGGDatabase{"a": {Path: "db-a"}, "b": {Path: "db-b"}})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		for path, want := range map[string]string{"a": "db-a", "b": "db-b", "c": ""} {
			wg.Add(1)
			go func(path, want string) {
				defer wg.Done()
				got, ok, err := p.ResolveKGGDatabase(context.Background(), path)
				if err != nil || got.Path != want || ok != (want != "") {
					t.Errorf("path=%s got=%+v ok=%v err=%v", path, got, ok, err)
				}
			}(path, want)
		}
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := p.ResolveKGGDatabase(ctx, "a"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}

func TestValidateKGGDatabasePathIsBoundedAndReadOnly(t *testing.T) {
	dir := t.TempDir()
	db := filepath.Join(dir, "database.kgg")
	data := append([]byte("SQLite format 3\x00"), bytes.Repeat([]byte{0}, 32)...)
	if err := os.WriteFile(db, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateKGGDatabasePath(db); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "database.kgg" {
		t.Fatalf("database side effects: %+v", entries)
	}
	for name, path := range map[string]string{"missing": filepath.Join(dir, "missing"), "directory": dir} {
		if err := ValidateKGGDatabasePath(path); err == nil {
			t.Fatalf("%s path accepted", name)
		}
	}
}

func TestReadEKeyFileParsingAndBounds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ekey.txt")
	if err := os.WriteFile(path, []byte("\xef\xbb\xbf\nQUJDREVGR0hJSktMTU5PUA==\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadEKeyFile(path)
	if err != nil || string(got) != "QUJDREVGR0hJSktMTU5PUA==" {
		t.Fatalf("got=%q err=%v", got, err)
	}
	if err := os.WriteFile(path, []byte("not base64!"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadEKeyFile(path); err == nil {
		t.Fatal("malformed key accepted")
	}
	if err := os.WriteFile(path, make([]byte, MaxEKeyFileSize+1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadEKeyFile(path); err == nil {
		t.Fatal("oversized key accepted")
	}
}
