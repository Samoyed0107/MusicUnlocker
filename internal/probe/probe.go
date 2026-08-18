// Package probe identifies container families from bounded structural evidence.
// It does not decrypt, decode, obtain keys, or perform network activity.
package probe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"musicunlocker/internal/decoder/upstream"
)

const (
	maxHeaderRead = 4096
	maxMusicExTag = 64 * 1024
)

type Family string

const (
	FamilyUnknown Family = "unknown"
	FamilyRaw     Family = "raw"
	FamilyKugou   Family = "kugou"
	FamilyNetease Family = "netease"
	FamilyKuwo    Family = "kuwo"
	FamilyQMC     Family = "qmc"
)

type Confidence string

const (
	ConfidenceUnknown   Confidence = "unknown"
	ConfidenceHint      Confidence = "hint"
	ConfidenceConfirmed Confidence = "confirmed"
)

type Evidence struct{ Kind, Detail string }

type Result struct {
	Family           Family
	Variant          string
	Confidence       Confidence
	ExtensionHints   []string
	Evidence         []Evidence
	AudioOffset      uint32
	CryptoVersion    uint32
	CryptoSlot       uint32
	FooterType       string
	NeedsExternalKey bool
	RawAudioFormat   string
	TagVersion       uint32
	TagSize          uint32
}

var (
	kgmMagic = []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14}
	vprMagic = []byte{0x05, 0x28, 0xbc, 0x96, 0xe9, 0xe4, 0x5a, 0x43, 0x91, 0xaa, 0xbd, 0xd0, 0x7a, 0xf5, 0x36, 0x31}
)

// Probe examines bounded regions and restores the reader's initial position.
func Probe(r io.ReadSeeker, filename string) (result Result, err error) {
	if r == nil {
		return result, errors.New("probe: nil reader")
	}
	result.Family = FamilyUnknown
	result.Confidence = ConfidenceUnknown
	start, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return result, fmt.Errorf("probe: record position: %w", err)
	}
	defer func() {
		if restoreErr := seekRestore(r, start); err == nil && restoreErr != nil {
			err = restoreErr
		}
	}()

	result.ExtensionHints = extensionHints(filename)
	if len(result.ExtensionHints) > 0 {
		result.Confidence = ConfidenceHint
		result.Evidence = append(result.Evidence, Evidence{Kind: "filename", Detail: strings.Join(result.ExtensionHints, ", ")})
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return result, err
	}
	header, readErr := readAtMost(r, maxHeaderRead)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return result, readErr
	}
	if detectContainer(&result, header) {
		return result, nil
	}
	if detectRaw(&result, header) {
		return result, nil
	}
	if err = detectFooter(&result, r); err != nil {
		return result, err
	}
	if result.Family == FamilyUnknown && len(result.ExtensionHints) == 0 {
		result.Confidence = ConfidenceUnknown
	}
	return result, nil
}

func ProbeFile(path string) (Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer f.Close()
	return Probe(f, filepath.Base(path))
}

func seekRestore(r io.Seeker, pos int64) error { _, err := r.Seek(pos, io.SeekStart); return err }

func readAtMost(r io.Reader, limit int) ([]byte, error) {
	b := make([]byte, limit)
	n, err := io.ReadFull(r, b)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return b[:n], nil
	}
	return b[:n], err
}

func confirm(result *Result, family Family, variant, detail string) {
	result.Family, result.Variant, result.Confidence = family, variant, ConfidenceConfirmed
	result.Evidence = append(result.Evidence, Evidence{Kind: "content", Detail: detail})
}

func detectContainer(result *Result, b []byte) bool {
	switch {
	case bytes.HasPrefix(b, kgmMagic):
		confirm(result, FamilyKugou, "KGM", "16-byte KGM magic")
	case bytes.HasPrefix(b, vprMagic):
		confirm(result, FamilyKugou, "VPR", "16-byte VPR magic")
	case bytes.HasPrefix(b, []byte("CTENFDAM")):
		confirm(result, FamilyNetease, "NCM", "CTENFDAM magic")
	case len(b) >= 16 && (bytes.Equal(b[:16], []byte("yeelion-kuwo-tme")) || bytes.Equal(b[:16], []byte("yeelion-kuwo\x00\x00\x00\x00"))):
		confirm(result, FamilyKuwo, "KWM", "16-byte KWM magic")
	default:
		return false
	}
	if len(b) >= 0x1c {
		result.AudioOffset = binary.LittleEndian.Uint32(b[0x10:0x14])
		result.CryptoVersion = binary.LittleEndian.Uint32(b[0x14:0x18])
		result.CryptoSlot = binary.LittleEndian.Uint32(b[0x18:0x1c])
	}
	return true
}

func detectRaw(result *Result, b []byte) bool {
	format := ""
	switch {
	case bytes.HasPrefix(b, []byte("fLaC")):
		format = "flac"
	case bytes.HasPrefix(b, []byte("OggS")):
		format = "ogg"
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WAVE")):
		format = "wav"
	case bytes.HasPrefix(b, []byte("ID3")):
		format = "mp3"
	case len(b) >= 16 && bytes.Equal(b[4:8], []byte("ftyp")) && binary.BigEndian.Uint32(b[:4]) >= 16:
		format = "mp4"
		brand := string(b[8:12])
		if brand == "M4A " || (len(b) >= 20 && bytes.Contains(b[16:min(len(b), 4096)], []byte("M4A "))) {
			format = "m4a"
		}
	}
	if format == "" {
		return false
	}
	result.Family, result.RawAudioFormat, result.Confidence = FamilyRaw, format, ConfidenceConfirmed
	result.Evidence = append(result.Evidence, Evidence{Kind: "content", Detail: format + " signature"})
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func detectFooter(result *Result, r io.ReadSeeker) error {
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if size >= 4 {
		if _, err = r.Seek(-4, io.SeekEnd); err != nil {
			return err
		}
		last, readErr := readAtMost(r, 4)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return readErr
		}
		if bytes.Equal(last, []byte("STag")) {
			confirm(result, FamilyQMC, "STag", "final STag footer")
			result.FooterType, result.NeedsExternalKey = "STag", true
			return nil
		}
		if bytes.Equal(last, []byte("QTag")) {
			confirm(result, FamilyQMC, "QTag", "final QTag footer")
			result.FooterType = "QTag"
			return nil
		}
	}
	if size < 16 {
		return nil
	}
	if _, err = r.Seek(-16, io.SeekEnd); err != nil {
		return err
	}
	tail, readErr := readAtMost(r, 16)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return readErr
	}
	if len(tail) < 16 || !bytes.Equal(tail[8:], []byte("musicex\x00")) {
		return nil
	}
	tagSize, version := binary.LittleEndian.Uint32(tail[:4]), binary.LittleEndian.Uint32(tail[4:8])
	if version != 1 || tagSize < 0xc0 || uint64(tagSize) > uint64(size) || tagSize > maxMusicExTag {
		return nil
	}
	confirm(result, FamilyQMC, "MusicEx", "valid MusicEx footer")
	result.FooterType, result.TagSize, result.TagVersion = "MusicEx", tagSize, version
	result.NeedsExternalKey = true
	return nil
}

func extensionHints(filename string) []string {
	seen := map[string]bool{}
	var out []string
	name := filepath.Base(filename)
	for name != "" {
		for _, suffix := range upstream.CandidateSuffixes(name) {
			if !seen[suffix] {
				seen[suffix] = true
				out = append(out, suffix)
			}
		}
		ext := filepath.Ext(name)
		if ext == "" {
			break
		}
		name = strings.TrimSuffix(name, ext)
	}
	return out
}
