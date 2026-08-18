package audiofmt

import (
	"bytes"
	"encoding/binary"
)

// Detect identifies an audio container from a bounded prefix.
func Detect(b []byte) (string, bool) {
	switch {
	case bytes.HasPrefix(b, []byte("fLaC")):
		return "flac", true
	case bytes.HasPrefix(b, []byte("OggS")):
		return "ogg", true
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WAVE")):
		return "wav", true
	case bytes.HasPrefix(b, []byte("ID3")):
		return "mp3", true
	case len(b) >= 16 && bytes.Equal(b[4:8], []byte("ftyp")) && binary.BigEndian.Uint32(b[:4]) >= 16:
		if bytes.Equal(b[8:12], []byte("M4A ")) || bytes.Contains(b[16:], []byte("M4A ")) {
			return "m4a", true
		}
		return "mp4", true
	}
	return "", false
}
