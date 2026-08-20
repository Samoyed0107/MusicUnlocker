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
	if err := store.Save(Preferences{MaxWorkers: 4, Theme: ThemeDark, AutoSuggestOutput: false}); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxWorkers != 4 || got.Theme != ThemeDark || got.AutoSuggestOutput {
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
	if err != nil || got.MaxWorkers != 1 || got.Theme != ThemeSystem || got.AutoSuggestOutput {
		t.Fatalf("sanitized=%+v err=%v", got, err)
	}
}

func TestPreferencesNamingDefaultsAndInvalidRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	store := &filePreferencesStore{path: path}
	if err := os.WriteFile(path, []byte(`{"maxWorkers":4,"theme":"dark","autoSuggestOutput":false,"namingMode":"CUSTOM","customNamingTemplate":"{bad}"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got.NamingMode != "KEEP_SOURCE" || got.MaxWorkers != 4 || got.Theme != ThemeDark || got.AutoSuggestOutput {
		t.Fatalf("recovery=%+v err=%v", got, err)
	}
	if err := store.Save(Preferences{MaxWorkers: 2, Theme: ThemeLight, AutoSuggestOutput: true, NamingMode: "CUSTOM", CustomNamingTemplate: "{artist} - {title}"}); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil || got.NamingMode != "CUSTOM" || got.CustomNamingTemplate != "{artist} - {title}" {
		t.Fatalf("custom=%+v err=%v", got, err)
	}
}

func TestPreferencesLanguageCompatibilityAndRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	store := &filePreferencesStore{path: path}
	if err := os.WriteFile(path, []byte(`{"maxWorkers":4,"theme":"dark","autoSuggestOutput":false,"namingMode":"TITLE"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got.Language != "SYSTEM" || got.MaxWorkers != 4 || got.Theme != ThemeDark || got.AutoSuggestOutput {
		t.Fatalf("old config=%+v err=%v", got, err)
	}
	for _, language := range []string{"SYSTEM", "zh-CN", "en-US"} {
		if err := store.Save(Preferences{MaxWorkers: 2, Theme: ThemeLight, AutoSuggestOutput: true, NamingMode: "TITLE", Language: language}); err != nil {
			t.Fatal(err)
		}
		got, err = store.Load()
		if err != nil || got.Language != language || got.MaxWorkers != 2 || got.NamingMode != "TITLE" {
			t.Fatalf("language=%q got=%+v err=%v", language, got, err)
		}
	}
	if err := store.Save(Preferences{MaxWorkers: 3, Theme: ThemeDark, AutoSuggestOutput: false, NamingMode: "TITLE", Language: "xx"}); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil || got.Language != "SYSTEM" || got.MaxWorkers != 3 || got.Theme != ThemeDark || got.AutoSuggestOutput {
		t.Fatalf("invalid language=%+v err=%v", got, err)
	}
}

func TestPreferencesThemeCompatibilityAndRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	store := &filePreferencesStore{path: path}
	for legacy, want := range map[string]string{"system": ThemeSystem, "light": ThemeLight, "dark": ThemeDark} {
		if err := os.WriteFile(path, []byte(`{"maxWorkers":1,"theme":"`+legacy+`","autoSuggestOutput":true}`), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := store.Load()
		if err != nil || got.Theme != want {
			t.Fatalf("legacy theme=%q got=%+v err=%v", legacy, got, err)
		}
	}
	if err := store.Save(Preferences{MaxWorkers: 1, Theme: "NEON", AutoSuggestOutput: true}); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got.Theme != ThemeSystem {
		t.Fatalf("invalid theme=%+v err=%v", got, err)
	}
}
