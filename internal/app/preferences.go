package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Preferences struct {
	MaxWorkers        int    `json:"maxWorkers"`
	Theme             string `json:"theme,omitempty"`
	AutoSuggestOutput bool   `json:"autoSuggestOutput"`
}

type PreferencesStore interface {
	Load() (Preferences, error)
	Save(Preferences) error
}

func DefaultPreferences() Preferences {
	return Preferences{MaxWorkers: 1, Theme: "system", AutoSuggestOutput: true}
}

func DefaultPreferencesStore() PreferencesStore { return newFilePreferencesStore() }

func sanitizePreferences(p Preferences) Preferences {
	d := DefaultPreferences()
	if p.MaxWorkers >= 1 && p.MaxWorkers <= 8 {
		d.MaxWorkers = p.MaxWorkers
	}
	if p.Theme == "system" || p.Theme == "light" || p.Theme == "dark" {
		d.Theme = p.Theme
	}
	d.AutoSuggestOutput = p.AutoSuggestOutput
	return d
}

type filePreferencesStore struct{ path string }

func newFilePreferencesStore() PreferencesStore {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return &filePreferencesStore{}
	}
	return &filePreferencesStore{path: filepath.Join(configDir, "MusicUnlocker", "preferences.json")}
}

func (s *filePreferencesStore) Load() (Preferences, error) {
	if s.path == "" {
		return DefaultPreferences(), errors.New("user config directory unavailable")
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultPreferences(), nil
		}
		return DefaultPreferences(), err
	}
	var p Preferences
	if err := json.Unmarshal(b, &p); err != nil {
		return DefaultPreferences(), err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return DefaultPreferences(), err
	}
	if _, ok := fields["autoSuggestOutput"]; !ok {
		p.AutoSuggestOutput = true
	}
	return sanitizePreferences(p), nil
}

func (s *filePreferencesStore) Save(p Preferences) error {
	if s.path == "" {
		return errors.New("user config directory unavailable")
	}
	p = sanitizePreferences(p)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".preferences-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.path); err == nil {
		return nil
	}
	// Windows cannot rename over an existing file. The fallback keeps the
	// temporary file in the same directory and avoids partial JSON writes.
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmpName, s.path)
}
