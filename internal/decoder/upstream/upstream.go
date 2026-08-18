// Package upstream provides the boundary between MusicUnlocker and the
// canonical upstream decoder registrations.
package upstream

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"git.um-react.app/um/cli/algo/common"
	"go.uber.org/zap"

	_ "git.um-react.app/um/cli/algo/kgm"
	_ "git.um-react.app/um/cli/algo/kwm"
	_ "git.um-react.app/um/cli/algo/ncm"
	_ "git.um-react.app/um/cli/algo/qmc"
	_ "git.um-react.app/um/cli/algo/tm"
	_ "git.um-react.app/um/cli/algo/xiami"
	_ "git.um-react.app/um/cli/algo/ximalaya"
)

type Stream interface{ io.Reader }

func decoderFor(path, suffix string) (common.NewDecoderFunc, error) {
	for _, decoder := range common.GetDecoder(path, true) {
		if decoder.Suffix == suffix {
			return decoder.Create, nil
		}
	}
	for _, decoder := range common.DecoderRegistry {
		if decoder.Suffix == suffix {
			return decoder.Create, nil
		}
	}
	return nil, fmt.Errorf("upstream decoder %s is not registered", suffix)
}

func decodeWith(reader io.ReadSeeker, path, suffix string) (Stream, error) {
	if reader == nil {
		return nil, errors.New("upstream decoder: nil reader")
	}
	create, err := decoderFor(path, suffix)
	if err != nil {
		return nil, err
	}
	decoder := create(&common.DecoderParams{Reader: reader, FilePath: path, Logger: zap.NewNop()})
	if err := decoder.Validate(); err != nil {
		return nil, fmt.Errorf("upstream %s validate: %w", suffix, err)
	}
	return decoder, nil
}

// DecodeKGM constructs and validates only the canonical upstream KGM decoder.
// The application-facing decode package never depends on upstream decoder types.
func DecodeKGM(reader io.ReadSeeker, path string) (Stream, error) {
	return decodeWith(reader, path, ".kgm")
}

func DecodeNCM(reader io.ReadSeeker, path string) (Stream, error) {
	if reader == nil {
		return nil, errors.New("upstream NCM: nil reader")
	}
	if err := validateNCMBounds(reader); err != nil {
		return nil, err
	}
	return decodeWith(reader, path, ".ncm")
}

func DecodeKWM(reader io.ReadSeeker, path string) (Stream, error) {
	if reader == nil {
		return nil, errors.New("upstream KWM: nil reader")
	}
	if err := validateKWMBounds(reader); err != nil {
		return nil, err
	}
	return decodeWith(reader, path, ".kwm")
}

const maxNCMSection = 64 << 20

func validateNCMBounds(reader io.ReadSeeker) error {
	start, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	defer reader.Seek(start, io.SeekStart)
	size, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if size < 14 {
		return errors.New("upstream NCM: truncated header")
	}
	readU32 := func(offset int64) (uint32, error) {
		if _, err := reader.Seek(offset, io.SeekStart); err != nil {
			return 0, err
		}
		var b [4]byte
		if _, err := io.ReadFull(reader, b[:]); err != nil {
			return 0, err
		}
		return binary.LittleEndian.Uint32(b[:]), nil
	}
	keyLen, err := readU32(10)
	if err != nil || keyLen > maxNCMSection || int64(14)+int64(keyLen)+4 > size {
		return errors.New("upstream NCM: invalid key section length")
	}
	metaOffset := int64(14) + int64(keyLen)
	metaLen, err := readU32(metaOffset)
	if err != nil || metaLen > maxNCMSection || metaOffset+4+int64(metaLen)+5+8 > size {
		return errors.New("upstream NCM: invalid metadata section length")
	}
	coverFrameOffset := metaOffset + 4 + int64(metaLen) + 5
	coverFrameLen, err := readU32(coverFrameOffset)
	if err != nil || coverFrameLen > uint32(size) || coverFrameOffset+8 > size {
		return errors.New("upstream NCM: invalid cover section length")
	}
	coverLen, err := readU32(coverFrameOffset + 4)
	if err != nil || coverLen > maxNCMSection || coverFrameOffset+8+int64(coverLen) > size {
		return errors.New("upstream NCM: invalid cover data length")
	}
	return nil
}

func validateKWMBounds(reader io.ReadSeeker) error {
	start, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	defer reader.Seek(start, io.SeekStart)
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}
	header := make([]byte, 0x400)
	if _, err := io.ReadFull(reader, header); err != nil {
		return fmt.Errorf("upstream KWM: truncated header: %w", err)
	}
	if !(bytes.Equal(header[:16], []byte("yeelion-kuwo-tme")) || bytes.Equal(header[:16], []byte("yeelion-kuwo\x00\x00\x00\x00"))) {
		return errors.New("upstream KWM: magic header not matched")
	}
	field := strings.TrimRight(string(header[0x30:0x38]), "\x00")
	separator := -1
	for i, r := range field {
		if !unicode.IsDigit(r) {
			separator = i
			break
		}
	}
	if separator <= 0 {
		return errors.New("upstream KWM: invalid bitrate/format field")
	}
	return nil
}

// CandidateSuffixes returns upstream-registered filename suffixes that match
// filename. This is filename candidate discovery only; it does not validate
// file bytes and does not prove that a decoder can decode the file.
func CandidateSuffixes(filename string) []string {
	decoders := common.GetDecoder(filename, true)
	seen := make(map[string]struct{}, len(decoders))
	result := make([]string, 0, len(decoders))
	for _, decoder := range decoders {
		if _, ok := seen[decoder.Suffix]; ok {
			continue
		}
		seen[decoder.Suffix] = struct{}{}
		result = append(result, decoder.Suffix)
	}
	return result
}
