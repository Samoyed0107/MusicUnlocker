package naming

import (
	"errors"
	"fmt"
	"strings"

	"musicunlocker/internal/audiometa"
)

type Mode string

const (
	KeepSource            Mode = "KEEP_SOURCE"
	Title                 Mode = "TITLE"
	ArtistTitle           Mode = "ARTIST_TITLE"
	TitleArtist           Mode = "TITLE_ARTIST"
	TrackTitle            Mode = "TRACK_TITLE"
	ArtistAlbumTrackTitle Mode = "ARTIST_ALBUM_TRACK_TITLE"
	Custom                Mode = "CUSTOM"
)
const MaxTemplateLength = 4096

type Config struct {
	Mode           Mode   `json:"mode"`
	CustomTemplate string `json:"customTemplate,omitempty"`
}
type Result struct {
	Stem           string
	UsedMode       Mode
	RequestedMode  Mode
	Fallback       bool
	FallbackReason string
}

var ErrInvalidConfig = errors.New("invalid naming configuration")
var ErrInvalidName = errors.New("generated filename is empty")

var tokens = map[string]func(audiometa.Metadata) (string, bool){
	"title": func(m audiometa.Metadata) (string, bool) { return m.Title, m.Title != "" }, "artist": func(m audiometa.Metadata) (string, bool) { return m.Artist, m.Artist != "" }, "album": func(m audiometa.Metadata) (string, bool) { return m.Album, m.Album != "" }, "albumartist": func(m audiometa.Metadata) (string, bool) { return m.AlbumArtist, m.AlbumArtist != "" },
	"track": func(m audiometa.Metadata) (string, bool) {
		if m.Track <= 0 {
			return "", false
		}
		return fmt.Sprintf("%02d", m.Track), true
	}, "disc": func(m audiometa.Metadata) (string, bool) {
		if m.Disc <= 0 {
			return "", false
		}
		return fmt.Sprint(m.Disc), true
	}, "year": func(m audiometa.Metadata) (string, bool) {
		if m.Year <= 0 {
			return "", false
		}
		return fmt.Sprint(m.Year), true
	},
}

func Validate(c Config) error {
	if c.Mode == "" {
		c.Mode = KeepSource
	}
	switch c.Mode {
	case KeepSource, Title, ArtistTitle, TitleArtist, TrackTitle, ArtistAlbumTrackTitle:
		return nil
	case Custom:
		if strings.TrimSpace(c.CustomTemplate) == "" {
			return fmt.Errorf("%w: empty custom template", ErrInvalidConfig)
		}
		_, err := parse(c.CustomTemplate)
		return err
	default:
		return fmt.Errorf("%w: unknown mode %q", ErrInvalidConfig, c.Mode)
	}
}

func Resolve(c Config, source string, metadata audiometa.Metadata, available bool) (Result, error) {
	if c.Mode == "" {
		c.Mode = KeepSource
	}
	if err := Validate(c); err != nil {
		return Result{}, err
	}
	source = Sanitize(source)
	if source == "" {
		return Result{}, ErrInvalidName
	}
	r := Result{Stem: source, UsedMode: KeepSource, RequestedMode: c.Mode}
	if c.Mode == KeepSource {
		return r, nil
	}
	template := map[Mode]string{Title: "{title}", ArtistTitle: "{artist} - {title}", TitleArtist: "{title} - {artist}", TrackTitle: "{track} - {title}", ArtistAlbumTrackTitle: "{artist} - {album} - {track} - {title}"}[c.Mode]
	if c.Mode == Custom {
		template = c.CustomTemplate
	}
	stem, ok := render(template, source, metadata)
	if !ok || (requiresMetadata(template) && !available) {
		r.Fallback = true
		r.FallbackReason = "metadata incomplete; source filename used"
		return r, nil
	}
	stem = Sanitize(stem)
	if stem == "" {
		r.Fallback = true
		r.FallbackReason = "generated filename empty; source filename used"
		return r, nil
	}
	r.Stem = stem
	r.UsedMode = c.Mode
	return r, nil
}

func requiresMetadata(template string) bool {
	parts, err := parse(template)
	if err != nil {
		return true
	}
	for _, part := range parts {
		if part != "source" && tokens[part] != nil {
			return true
		}
	}
	return false
}

func parse(s string) ([]string, error) {
	if len(s) > MaxTemplateLength {
		return nil, fmt.Errorf("%w: template too long", ErrInvalidConfig)
	}
	var out []string
	for i := 0; i < len(s); {
		j := strings.IndexByte(s[i:], '{')
		if j < 0 {
			out = append(out, s[i:])
			break
		}
		j += i
		out = append(out, s[i:j])
		k := strings.IndexByte(s[j+1:], '}')
		if k < 0 {
			return nil, fmt.Errorf("%w: unclosed placeholder", ErrInvalidConfig)
		}
		k += j + 1
		name := s[j+1 : k]
		if name == "source" {
			out = append(out, "source")
		} else if _, ok := tokens[name]; !ok {
			return nil, fmt.Errorf("%w: unknown placeholder %q", ErrInvalidConfig, name)
		} else {
			out = append(out, name)
		}
		i = k + 1
	}
	return out, nil
}
func render(s, source string, m audiometa.Metadata) (string, bool) {
	parts, err := parse(s)
	if err != nil {
		return "", false
	}
	var b strings.Builder
	for _, p := range parts {
		if p == "source" {
			b.WriteString(source)
			continue
		}
		if f, ok := tokens[p]; ok {
			v, present := f(m)
			if !present {
				return "", false
			}
			b.WriteString(v)
		} else {
			b.WriteString(p)
		}
	}
	return b.String(), true
}

func Sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || strings.ContainsRune(`<>:"/\\|?*`, r) {
			b.WriteRune('_')
		} else {
			b.WriteRune(r)
		}
	}
	s = strings.Join(strings.Fields(b.String()), " ")
	s = strings.TrimRight(s, " .")
	if isReserved(s) {
		s = "_" + s
	}
	if len([]rune(s)) > 240 {
		s = string([]rune(s)[:240])
		s = strings.TrimRight(s, " .")
	}
	return s
}
func isReserved(s string) bool {
	u := strings.ToUpper(s)
	if i := strings.IndexByte(u, '.'); i >= 0 {
		u = u[:i]
	}
	switch u {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	if len(u) == 4 && (strings.HasPrefix(u, "COM") || strings.HasPrefix(u, "LPT")) {
		return u[3] >= '1' && u[3] <= '9'
	}
	return false
}
