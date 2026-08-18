package upstream

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func syntheticKGMV3() []byte {
	b := make([]byte, 60+4)
	copy(b, []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14})
	binary.LittleEndian.PutUint32(b[0x10:], 60)
	binary.LittleEndian.PutUint32(b[0x14:], 3)
	binary.LittleEndian.PutUint32(b[0x18:], 1)
	return b
}

func TestDecodeKGMRejectsMalformedInput(t *testing.T) {
	if _, err := DecodeKGM(bytes.NewReader([]byte("not kgm")), "input.bin"); err == nil {
		t.Fatal("malformed KGM was accepted")
	}
}

func TestDecodeKGMV3UsesCanonicalAdapter(t *testing.T) {
	stream, err := DecodeKGM(bytes.NewReader(syntheticKGMV3()), "input.with-unusual-name")
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	if _, err := stream.Read(buf); err != nil {
		t.Fatal(err)
	}
}

func TestCandidateSuffixesUsesUpstreamFilenameRegistrations(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     []string
	}{
		{name: "kgm", filename: "track.kgm", want: []string{".kgm"}},
		{name: "kgm flac compound", filename: "track.kgm.flac", want: []string{".flac", ".kgm.flac"}},
		{name: "vpr flac compound", filename: "track.vpr.flac", want: []string{".flac", ".vpr.flac"}},
		{name: "kwm", filename: "track.kwm", want: []string{".kwm"}},
		{name: "ncm", filename: "track.ncm", want: []string{".ncm"}},
		{name: "qmc", filename: "track.qmc0", want: []string{".qmc0"}},
		{name: "qmc mgg suffix", filename: "track.mgg0", want: []string{".mgg0"}},
		{name: "qmc mflac suffix", filename: "track.mflac0", want: []string{".mflac0"}},
		{name: "tm", filename: "track.tm2", want: []string{".tm2"}},
		{name: "xiami", filename: "track.xm", want: []string{".xm"}},
		{name: "ximalaya", filename: "track.x2m", want: []string{".x2m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CandidateSuffixes(tt.filename); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CandidateSuffixes(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCandidateSuffixesDoesNotInventUnsupportedCompoundSuffixes(t *testing.T) {
	for _, filename := range []string{"track.mgg0.flac", "track.mflac0.flac"} {
		got := CandidateSuffixes(filename)
		for _, suffix := range got {
			if suffix == ".mgg0.flac" || suffix == ".mflac0.flac" {
				t.Fatalf("CandidateSuffixes(%q) invented unsupported compound suffix %q", filename, suffix)
			}
		}
	}
}

func TestCandidateSuffixesDoesNotValidateContent(t *testing.T) {
	// A matching filename is only a candidate; this test intentionally supplies
	// no file and makes no claim about encrypted-format validity.
	if got := CandidateSuffixes("not-a-real-file.ncm"); !reflect.DeepEqual(got, []string{".ncm"}) {
		t.Fatalf("CandidateSuffixes returned %v, want filename-only candidate [.ncm]", got)
	}
}
