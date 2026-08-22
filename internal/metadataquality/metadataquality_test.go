package metadataquality

import (
	"testing"

	"musicunlocker/internal/audiometa"
	"musicunlocker/internal/naming"
)

func TestValidMultilingualMetadata(t *testing.T) {
	values := map[string]string{
		"English":             "Midnight Drive",
		"Simplified Chinese":  "后来",
		"Traditional Chinese": "後來",
		"Japanese":            "君の知らない物語",
		"Korean":              "봄날",
		"Cyrillic":            "Группа крови",
		"accented Latin":      "Beyoncé & Sigur Rós",
		"music punctuation":   "Song (Live) feat. Aimer ♪",
	}
	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			if got, reason, ok := ValidateText(value); !ok || got != value || reason != "" {
				t.Errorf("ValidateText(%q) = %q, %q, %v", value, got, reason, ok)
			}
		})
	}
}

func TestSuspiciousMetadata(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  Reason
	}{
		{"empty", " \t ", ReasonEmpty},
		{"replacement", "secret � title", ReasonReplacement},
		{"invalid utf8", string([]byte{'b', 'a', 'd', 0xff}), ReasonInvalidUTF8},
		{"control", "bad\x00title", ReasonControlCharacter},
		{"non printable", "A\u200b\u200c\u200dB", ReasonNonPrintableRatio},
		{"garbage symbols", "@#$%`=<>\\|", ReasonSymbolRatio},
		{"low readable", "ab@#$\u00ad%", ReasonReadableRatio},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, reason, ok := ValidateText(tc.value); ok || got != "" || reason != tc.want {
				t.Fatalf("ValidateText() = %q, %q, %v; want reason %q", got, reason, ok, tc.want)
			}
		})
	}
}

func TestValidateRejectsOneFieldAndKeepsProvenance(t *testing.T) {
	source := Source{Origin: OriginEmbedded, Provider: ProviderKGM}
	result := Validate(audiometa.Metadata{Title: "bad � title", Artist: "Aimer", Album: "Anime Hits", AlbumArtist: "Various Artists", Track: 2}, source)
	if result.Metadata.Title != "" || result.Metadata.Artist != "Aimer" || result.Metadata.Track != 2 {
		t.Fatalf("metadata = %#v", result.Metadata)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Field != FieldTitle || result.Diagnostics[0].Reason != ReasonReplacement || result.Diagnostics[0].Source != source {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestResolveUsesFieldLevelTrustOrderAndFilenameTitle(t *testing.T) {
	result := Resolve(
		Candidate{Metadata: audiometa.Metadata{Title: "bad �", Artist: "ZONE", AlbumArtist: " \t"}, Source: Source{Origin: OriginEmbedded, Provider: ProviderKGM}},
		Candidate{Metadata: audiometa.Metadata{Title: "Trusted container title", Album: "Album", AlbumArtist: "Various Artists"}, Source: Source{Origin: OriginContainer, Provider: ProviderKGM}},
		Candidate{Metadata: audiometa.Metadata{Title: "Filename title"}, Source: Source{Origin: OriginFilename, Provider: ProviderKGM}},
	)
	if result.Metadata.Title != "Trusted container title" || result.Metadata.Artist != "ZONE" || result.Metadata.Album != "Album" || result.Metadata.AlbumArtist != "Various Artists" {
		t.Fatalf("metadata = %#v", result.Metadata)
	}

	result = Resolve(
		Candidate{Metadata: audiometa.Metadata{Title: "bad �", Artist: "ZONE"}, Source: Source{Origin: OriginEmbedded, Provider: ProviderKGM}},
		Candidate{Metadata: audiometa.Metadata{}, Source: Source{Origin: OriginContainer, Provider: ProviderKGM}},
		Candidate{Metadata: audiometa.Metadata{Title: "Filename title"}, Source: Source{Origin: OriginFilename, Provider: ProviderKGM}},
	)
	if result.Metadata.Title != "Filename title" || result.Metadata.Artist != "ZONE" {
		t.Fatalf("filename fallback metadata = %#v", result.Metadata)
	}
}

func TestResolvePreservesNCMNamingPrecedence(t *testing.T) {
	result := Resolve(
		Candidate{Metadata: audiometa.Metadata{Title: "Embedded title"}, Source: Source{Origin: OriginEmbedded, Provider: ProviderNCM}},
		Candidate{Metadata: audiometa.Metadata{Title: "Container title", Artist: "Container artist", Album: "Container album"}, Source: Source{Origin: OriginContainer, Provider: ProviderNCM}},
		Candidate{Metadata: audiometa.Metadata{Title: "Filename title"}, Source: Source{Origin: OriginFilename, Provider: ProviderNCM}},
	)
	if result.Metadata.Title != "Embedded title" || result.Metadata.Artist != "Container artist" || result.Metadata.Album != "Container album" {
		t.Fatalf("metadata = %#v", result.Metadata)
	}
	resolved, err := naming.Resolve(naming.Config{Mode: naming.ArtistTitle}, "source", result.Metadata, result.Metadata.Available())
	if err != nil || resolved.Stem != "Container artist - Embedded title" || resolved.Fallback {
		t.Fatalf("naming result = %#v, err = %v", resolved, err)
	}
}
