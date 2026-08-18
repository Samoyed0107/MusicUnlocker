package naming

import (
	"musicunlocker/internal/audiometa"
	"strings"
	"testing"
)

func TestPolicies(t *testing.T) {
	m := audiometa.Metadata{Title: "Salvation", Artist: "Gabrielle Aplin", Album: "English Rain", Track: 5}
	tests := []struct {
		mode Mode
		want string
	}{{Title, "Salvation"}, {ArtistTitle, "Gabrielle Aplin - Salvation"}, {TitleArtist, "Salvation - Gabrielle Aplin"}, {TrackTitle, "05 - Salvation"}, {ArtistAlbumTrackTitle, "Gabrielle Aplin - English Rain - 05 - Salvation"}}
	for _, tt := range tests {
		r, e := Resolve(Config{Mode: tt.mode}, "source", m, true)
		if e != nil || r.Stem != tt.want {
			t.Errorf("%s: %#v %v", tt.mode, r, e)
		}
	}
	r, e := Resolve(Config{Mode: Custom, CustomTemplate: "{artist} - {album} - {track} - {title}"}, "source", m, true)
	if e != nil || r.Stem != "Gabrielle Aplin - English Rain - 05 - Salvation" {
		t.Fatalf("custom: %#v %v", r, e)
	}
}
func TestFallbackAndTemplateErrors(t *testing.T) {
	r, e := Resolve(Config{Mode: Custom, CustomTemplate: "{source}-TEST"}, "Salvation", audiometa.Metadata{}, false)
	if e != nil || r.Fallback || r.Stem != "Salvation-TEST" {
		t.Fatalf("source-only custom template should not require metadata: %#v %v", r, e)
	}
	r, e = Resolve(Config{Mode: ArtistTitle}, "foo", audiometa.Metadata{Title: "Salvation"}, true)
	if e != nil || !r.Fallback || r.Stem != "foo" {
		t.Fatalf("fallback %#v %v", r, e)
	}
	for _, c := range []Config{{Mode: Custom}, {Mode: Custom, CustomTemplate: "{bad}"}, {Mode: Custom, CustomTemplate: "{title"}} {
		if e := Validate(c); e == nil {
			t.Fatalf("expected config error %#v", c)
		}
	}
}
func TestSanitizeWindows(t *testing.T) {
	for _, s := range []string{"a<b>c:d\"e/f\\g|h?i*j", "CON", "con", "PRN", "AUX", "NUL", "COM1", "LPT9"} {
		x := Sanitize(s)
		if strings.ContainsAny(x, "<>:\"/\\|?*") || x == s && s == "CON" {
			t.Fatalf("%q -> %q", s, x)
		}
	}
	if Sanitize("中文 日本") != "中文 日本" {
		t.Fatal("unicode changed")
	}
	x := Sanitize(strings.Repeat("界", 300))
	if len([]rune(x)) > 240 {
		t.Fatal("not truncated")
	}
}
