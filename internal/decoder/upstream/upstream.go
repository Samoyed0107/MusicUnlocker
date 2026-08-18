// Package upstream provides the boundary between MusicUnlocker and the
// canonical upstream decoder registrations.
package upstream

import (
	"errors"
	"fmt"
	"git.um-react.app/um/cli/algo/common"
	"go.uber.org/zap"
	"io"

	_ "git.um-react.app/um/cli/algo/kgm"
	_ "git.um-react.app/um/cli/algo/kwm"
	_ "git.um-react.app/um/cli/algo/ncm"
	_ "git.um-react.app/um/cli/algo/qmc"
	_ "git.um-react.app/um/cli/algo/tm"
	_ "git.um-react.app/um/cli/algo/xiami"
	_ "git.um-react.app/um/cli/algo/ximalaya"
)

type Stream interface{ io.Reader }

// DecodeKGM constructs and validates only the canonical upstream KGM decoder.
// The application-facing decode package never depends on upstream decoder types.
func DecodeKGM(reader io.ReadSeeker, path string) (Stream, error) {
	if reader == nil {
		return nil, errors.New("upstream KGM: nil reader")
	}
	decoders := common.GetDecoder(path, true)
	var create common.NewDecoderFunc
	for _, decoder := range decoders {
		if decoder.Suffix == ".kgm" || decoder.Suffix == ".kgma" || decoder.Suffix == ".vpr" || decoder.Suffix == ".kgm.flac" || decoder.Suffix == ".vpr.flac" {
			create = decoder.Create
			break
		}
	}
	if create == nil {
		// Content probe has already established KGM/VPR; use the canonical
		// registration without making filename suffix the resolver decision.
		for _, decoder := range common.DecoderRegistry {
			if decoder.Suffix == ".kgm" {
				create = decoder.Create
				break
			}
		}
	}
	if create == nil {
		return nil, errors.New("upstream KGM decoder is not registered")
	}
	decoder := create(&common.DecoderParams{Reader: reader, FilePath: path, Logger: zap.NewNop()})
	if err := decoder.Validate(); err != nil {
		return nil, fmt.Errorf("upstream KGM validate: %w", err)
	}
	return decoder, nil
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
