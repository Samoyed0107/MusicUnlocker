package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferencesDefaultsAndRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "MusicUnlocker", "preferences.json")
	store := &filePreferencesStore{path: path}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got != DefaultPreferences() {
		t.Fatalf("defaults=%+v", got)
	}
	if err := store.Save(Preferences{MaxWorkers: 4, Theme: "dark", AutoSuggestOutput: false}); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxWorkers != 4 || got.Theme != "dark" || got.AutoSuggestOutput {
		t.Fatalf("round trip=%+v", got)
	}
}

func TestPreferencesMalformedAndInvalidValuesFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte(`{"maxWorkers":99,"theme":"neon","future":true`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := (&filePreferencesStore{path: path}).Load()
	if err == nil || got != DefaultPreferences() {
		t.Fatalf("malformed=%+v err=%v", got, err)
	}
	if err := os.WriteFile(path, []byte(`{"maxWorkers":99,"theme":"neon","autoSuggestOutput":false}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err = (&filePreferencesStore{path: path}).Load()
	if err != nil || got.MaxWorkers != 1 || got.Theme != "system" || got.AutoSuggestOutput {
		t.Fatalf("sanitized=%+v err=%v", got, err)
	}
}
