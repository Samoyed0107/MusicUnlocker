package fileassoc

import (
	"reflect"
	"testing"

	"musicunlocker/internal/decoder/upstream"
)

func TestWindowsExtensionsMatchEncryptedCandidateSurface(t *testing.T) {
	want := []string{
		".666c6163", ".6d3461", ".6d7033", ".6f6767", ".776176",
		".bkcape", ".bkcflac", ".bkcm4a", ".bkcmp3", ".bkcogg", ".bkcwav", ".bkcwma",
		".kgg", ".kgm", ".kgma", ".kwm",
		".mflac", ".mflac0", ".mflac1", ".mflaca", ".mflach", ".mflacl", ".mflacm",
		".mgg", ".mgg0", ".mgg1", ".mgga", ".mggh", ".mggl", ".mggm", ".mmp4",
		".ncm", ".qmc0", ".qmc2", ".qmc3", ".qmc4", ".qmc6", ".qmc8", ".qmcflac", ".qmcogg",
		".tkm", ".tm0", ".tm2", ".tm3", ".tm6", ".vpr", ".x2m", ".x3m", ".xm",
	}
	if got := WindowsExtensions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("WindowsExtensions() = %#v, want %#v", got, want)
	}

	registered := upstream.RegisteredCandidateSuffixes()
	for _, extension := range WindowsExtensions() {
		if !contains(registered, extension) {
			t.Errorf("%s is not an upstream filename candidate", extension)
		}
	}
}

func TestWindowsExtensionsExcludeDecodedAudioAndCompoundSuffixes(t *testing.T) {
	registered := WindowsExtensions()
	for _, extension := range []string{".aac", ".flac", ".m4a", ".mp3", ".ogg", ".wav", ".wma", ".kgm.flac", ".vpr.flac"} {
		if contains(registered, extension) {
			t.Errorf("decoded or compound extension %s must not be registered", extension)
		}
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
