package audiofmt

import "testing"

func TestDetectMPEGFrame(t *testing.T) {
	valid := []byte{0xff, 0xfb, 0x90, 0x64}
	if got, ok := Detect(valid); !ok || got != "mp3" {
		t.Fatalf("valid MPEG frame = %q, %v", got, ok)
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"reserved version", []byte{0xff, 0xe3, 0x90, 0x64}},
		{"invalid layer", []byte{0xff, 0xf9, 0x90, 0x64}},
		{"invalid bitrate", []byte{0xff, 0xfb, 0x00, 0x64}},
		{"reserved bitrate", []byte{0xff, 0xfb, 0xf0, 0x64}},
		{"invalid sample rate", []byte{0xff, 0xfb, 0x9c, 0x64}},
		{"random ff", []byte{0xff, 0x01, 0x02, 0x03}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, ok := Detect(tc.data); ok {
				t.Fatalf("detected %q from invalid MPEG data", got)
			}
		})
	}
}

func TestDetectID3MP3(t *testing.T) {
	if got, ok := Detect([]byte("ID3\x04\x00\x00")); !ok || got != "mp3" {
		t.Fatalf("ID3 = %q, %v", got, ok)
	}
}
