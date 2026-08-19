package audiometa

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestID3v23AndV24(t *testing.T) {
	for _, version := range []byte{3, 4} {
		b := []byte{'I', 'D', '3', version, 0, 0, 0, 0, 0, 0}
		frames := append(frame("TIT2", "Salvation", version), frame("TPE1", "Gabrielle Aplin", version)...)
		frames = append(frames, frame("TRCK", "5/12", version)...)
		size := len(frames)
		b[6] = byte(size >> 21)
		b[7] = byte(size>>14) & 127
		b[8] = byte(size>>7) & 127
		b[9] = byte(size) & 127
		b = append(b, frames...)
		p := filepath.Join(t.TempDir(), "x.mp3")
		if e := os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
		m, e := ReadFile(p)
		if e != nil || m.Title != "Salvation" || m.Artist != "Gabrielle Aplin" || m.Track != 5 || m.TrackTotal != 12 {
			t.Fatalf("v%d %#v %v", version, m, e)
		}
	}
}
func frame(id, value string, version byte) []byte {
	b := []byte{0}
	b = append(b, []byte(value)...)
	h := make([]byte, 10)
	copy(h, id)
	n := len(b)
	if version == 4 {
		h[4] = byte(n >> 21)
		h[5] = byte(n>>14) & 127
		h[6] = byte(n>>7) & 127
		h[7] = byte(n) & 127
	} else {
		binary.BigEndian.PutUint32(h[4:8], uint32(n))
	}
	return append(h, b...)
}
func TestVorbisComment(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.flac")
	comments := []string{"TITLE=Salvation", "ARTIST=Gabrielle Aplin", "TRACKNUMBER=5"}
	data := make([]byte, 4)
	copy(data, "fLaC")
	body := make([]byte, 0)
	binary.Write(sliceWriter{&body}, binary.LittleEndian, uint32(0))
	binary.Write(sliceWriter{&body}, binary.LittleEndian, uint32(len(comments)))
	for _, c := range comments {
		binary.Write(sliceWriter{&body}, binary.LittleEndian, uint32(len(c)))
		body = append(body, []byte(c)...)
	}
	head := []byte{0x84, byte(len(body) >> 16), byte(len(body) >> 8), byte(len(body))}
	data = append(data, head...)
	data = append(data, body...)
	if e := os.WriteFile(p, data, 0600); e != nil {
		t.Fatal(e)
	}
	m, e := ReadFile(p)
	if e != nil || m.Title != "Salvation" || m.Track != 5 {
		t.Fatalf("%#v %v", m, e)
	}
}

type sliceWriter struct{ p *[]byte }

func (w sliceWriter) Write(b []byte) (int, error) { *w.p = append(*w.p, b...); return len(b), nil }

func TestMergeDecodedMetadataWinsAndFillsMissingFields(t *testing.T) {
	container := Metadata{Title: "Container title", Artist: "Container artist", Album: "Container album"}
	for name, tc := range map[string]struct {
		decoded Metadata
		want    Metadata
	}{
		"all missing":           {decoded: Metadata{}, want: container},
		"decoded wins":          {decoded: Metadata{Title: "Embedded title", Artist: "Embedded artist", Album: "Embedded album", Track: 7, Year: 2020}, want: Metadata{Title: "Embedded title", Artist: "Embedded artist", Album: "Embedded album", Track: 7, Year: 2020}},
		"whitespace is missing": {decoded: Metadata{Title: "  ", Artist: "\t"}, want: Metadata{Title: "Container title", Artist: "Container artist", Album: "Container album"}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := Merge(tc.decoded, container); got != tc.want {
				t.Fatalf("Merge() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
