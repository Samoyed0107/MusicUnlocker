// Package metadataquality validates metadata before it is used for output naming.
// It is deliberately independent of decoder and presentation concerns.
package metadataquality

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"musicunlocker/internal/audiometa"
)

type Origin string

const (
	OriginEmbedded  Origin = "EMBEDDED_DECODED"
	OriginContainer Origin = "CONTAINER"
	OriginFilename  Origin = "FILENAME"
)

type Provider string

const (
	ProviderGeneric Provider = "GENERIC"
	ProviderNCM     Provider = "NCM"
	ProviderKGM     Provider = "KGM"
)

type Source struct {
	Origin   Origin
	Provider Provider
}

type Field string

const (
	FieldTitle       Field = "title"
	FieldArtist      Field = "artist"
	FieldAlbum       Field = "album"
	FieldAlbumArtist Field = "albumArtist"
)

type Reason string

const (
	ReasonEmpty             Reason = "EMPTY"
	ReasonInvalidUTF8       Reason = "INVALID_UTF8"
	ReasonReplacement       Reason = "UNICODE_REPLACEMENT"
	ReasonControlCharacter  Reason = "CONTROL_CHARACTER"
	ReasonNonPrintableRatio Reason = "EXCESSIVE_NON_PRINTABLE"
	ReasonSymbolRatio       Reason = "ABNORMAL_SYMBOL_RATIO"
	ReasonReadableRatio     Reason = "LOW_HUMAN_READABLE_RATIO"
)

// Diagnostic intentionally excludes the rejected metadata value.
type Diagnostic struct {
	Field  Field
	Reason Reason
	Source Source
}

type Result struct {
	Metadata    audiometa.Metadata
	Diagnostics []Diagnostic
}

type Candidate struct {
	Metadata audiometa.Metadata
	Source   Source
}

// Resolve validates candidates independently and merges valid text fields in
// caller-supplied trust order. Numeric tags remain sourced from the first
// candidate, which is the decoded-audio metadata in the decode pipeline.
func Resolve(candidates ...Candidate) Result {
	var result Result
	for i, candidate := range candidates {
		validated := Validate(candidate.Metadata, candidate.Source)
		result.Diagnostics = append(result.Diagnostics, validated.Diagnostics...)
		if i == 0 {
			result.Metadata = validated.Metadata
			continue
		}
		fillMissingText(&result.Metadata, validated.Metadata)
	}
	return result
}

// Validate removes only suspicious text fields. Numeric tag fields pass
// through unchanged because they do not influence the text quality ratios.
func Validate(metadata audiometa.Metadata, source Source) Result {
	result := Result{Metadata: metadata}
	fields := []struct {
		field Field
		value *string
	}{
		{FieldTitle, &result.Metadata.Title},
		{FieldArtist, &result.Metadata.Artist},
		{FieldAlbum, &result.Metadata.Album},
		{FieldAlbumArtist, &result.Metadata.AlbumArtist},
	}
	for _, item := range fields {
		cleaned, reason, ok := ValidateText(*item.value)
		if ok {
			*item.value = cleaned
			continue
		}
		*item.value = ""
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Field: item.field, Reason: reason, Source: source})
	}
	return result
}

// ValidateText validates one metadata field without normalizing its language or
// transliterating it. Valid Unicode letters, marks, numbers, and common title
// punctuation are preserved.
func ValidateText(value string) (string, Reason, bool) {
	if !utf8.ValidString(value) {
		return "", ReasonInvalidUTF8, false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ReasonEmpty, false
	}

	var total, printable, readable, abnormalSymbols int
	for _, r := range value {
		total++
		if r == unicode.ReplacementChar {
			return "", ReasonReplacement, false
		}
		if unicode.IsControl(r) {
			return "", ReasonControlCharacter, false
		}
		if unicode.IsPrint(r) {
			printable++
		}
		if isHumanReadable(r) {
			readable++
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			abnormalSymbols++
		}
	}

	nonPrintable := total - printable
	if nonPrintable >= 3 || nonPrintable*5 > total {
		return "", ReasonNonPrintableRatio, false
	}
	if total >= 6 && abnormalSymbols*100 > total*70 {
		return "", ReasonSymbolRatio, false
	}
	if total >= 6 && readable*100 < total*35 {
		return "", ReasonReadableRatio, false
	}
	return value, "", true
}

func isHumanReadable(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
		return true
	}
	return strings.ContainsRune("&'+-.,!！?？()（）[]【】{}「」『』·・~～_/:：♪♫♥♡★☆×^", r)
}

func fillMissingText(target *audiometa.Metadata, fallback audiometa.Metadata) {
	if target.Title == "" {
		target.Title = fallback.Title
	}
	if target.Artist == "" {
		target.Artist = fallback.Artist
	}
	if target.Album == "" {
		target.Album = fallback.Album
	}
	if target.AlbumArtist == "" {
		target.AlbumArtist = fallback.AlbumArtist
	}
}
