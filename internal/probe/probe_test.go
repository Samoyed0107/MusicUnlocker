package probe

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func kgm(magic []byte, version uint32) []byte {
	b := make([]byte, 0x3c)
	copy(b, magic)
	binary.LittleEndian.PutUint32(b[0x10:], 0x400)
	binary.LittleEndian.PutUint32(b[0x14:], version)
	binary.LittleEndian.PutUint32(b[0x18:], 7)
	return b
}

func musicEx(size, version uint32) []byte {
	if size < 16 {
		size = 16
	}
	b := make([]byte, size)
	binary.LittleEndian.PutUint32(b[len(b)-16:], size)
	binary.LittleEndian.PutUint32(b[len(b)-12:], version)
	copy(b[len(b)-8:], []byte("musicex\x00"))
	return b
}

func TestProbeStructuralFormats(t *testing.T) {
	kwm2 := []byte("yeelion-kuwo\x00\x00\x00\x00")
	tests := []struct {
		name    string
		data    []byte
		family  Family
		variant string
	}{
		{"kgm v3", kgm(kgmMagic, 3), FamilyKugou, "KGM"},
		{"kgm v5", kgm(kgmMagic, 5), FamilyKugou, "KGM"},
		{"vpr", kgm(vprMagic, 3), FamilyKugou, "VPR"},
		{"ncm", []byte("CTENFDAM"), FamilyNetease, "NCM"},
		{"kwm one", []byte("yeelion-kuwo-tme"), FamilyKuwo, "KWM"},
		{"kwm two", kwm2, FamilyKuwo, "KWM"},
		{"st tag", append([]byte("encrypted"), []byte("STag")...), FamilyQMC, "STag"},
		{"q tag", append([]byte("encrypted"), []byte("QTag")...), FamilyQMC, "QTag"},
		{"musicex", musicEx(0xc0, 1), FamilyQMC, "MusicEx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Probe(bytes.NewReader(tt.data), "misleading.jpg")
			if err != nil {
				t.Fatal(err)
			}
			if got.Family != tt.family || got.Variant != tt.variant || got.Confidence != ConfidenceConfirmed {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

func TestProbeRawFormats(t *testing.T) {
	tests := []struct {
		name, want string
		data       []byte
	}{
		{"flac", "flac", []byte("fLaC")}, {"ogg", "ogg", []byte("OggS")},
		{"wav", "wav", []byte("RIFFxxxxWAVE")}, {"mp3", "mp3", []byte("ID3")},
		{"m4a", "m4a", append(append([]byte{0, 0, 0, 20}, []byte("ftypM4A ")...), []byte{0, 0, 0, 0}...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Probe(bytes.NewReader(tt.data), "song.flac")
			if err != nil {
				t.Fatal(err)
			}
			if got.Family != FamilyRaw || got.RawAudioFormat != tt.want {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

func TestProbePreservesOffsetAndHints(t *testing.T) {
	r := bytes.NewReader(append(kgm(kgmMagic, 3), []byte("payload")...))
	_, _ = r.Seek(5, 0)
	got, err := Probe(r, "example.kgm.flac")
	if err != nil {
		t.Fatal(err)
	}
	pos, _ := r.Seek(0, 1)
	if pos != 5 {
		t.Fatalf("position = %d, want 5", pos)
	}
	if got.Confidence != ConfidenceConfirmed || got.Family != FamilyKugou {
		t.Fatalf("got %#v", got)
	}
	for _, want := range []string{".flac", ".kgm.flac", ".kgm"} {
		if !contains(got.ExtensionHints, want) {
			t.Fatalf("hints %v missing %s", got.ExtensionHints, want)
		}
	}
}

func TestExtensionChains(t *testing.T) {
	for _, tt := range []struct {
		name      string
		encrypted string
	}{
		{"example.kgm.flac", ".kgm"}, {"example.mgg0.flac", ".mgg0"}, {"example.mflac0.flac", ".mflac0"},
	} {
		got := extensionHints(tt.name)
		if len(got) < 2 || !contains(got, ".flac") || !contains(got, tt.encrypted) {
			t.Fatalf("%s: %v", tt.name, got)
		}
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func TestProbeMalformedAndSmallFiles(t *testing.T) {
	for _, data := range [][]byte{nil, {1, 2, 3}, musicEx(0xffffffff, 1), musicEx(0xc0, 2), []byte("random")} {
		got, err := Probe(bytes.NewReader(data), "file.kgm")
		if err != nil {
			t.Fatal(err)
		}
		if got.Confidence == ConfidenceConfirmed {
			t.Fatalf("unexpected confirmation for %x: %#v", data, got)
		}
	}
}
