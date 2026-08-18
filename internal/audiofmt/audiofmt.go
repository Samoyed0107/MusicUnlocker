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
	case isMPEGFrame(b):
		return "mp3", true
	case len(b) >= 16 && bytes.Equal(b[4:8], []byte("ftyp")) && binary.BigEndian.Uint32(b[:4]) >= 16:
		if bytes.Equal(b[8:12], []byte("M4A ")) || bytes.Contains(b[16:], []byte("M4A ")) {
			return "m4a", true
		}
		return "mp4", true
	}
	return "", false
}

func isMPEGFrame(b []byte) bool {
	if len(b) < 4 || b[0] != 0xff || b[1]&0xe0 != 0xe0 {
		return false
	}
	version := (b[1] >> 3) & 0x03
	layer := (b[1] >> 1) & 0x03
	bitrate := (b[2] >> 4) & 0x0f
	sampleRate := (b[2] >> 2) & 0x03
	return version != 0 && version != 1 && layer != 0 && bitrate != 0 && bitrate != 15 && sampleRate != 3
}
