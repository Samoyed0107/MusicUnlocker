package externalmaterial

import (
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
