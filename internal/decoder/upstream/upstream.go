// Package upstream provides the boundary between MusicUnlocker and the
// canonical upstream decoder registrations.
package upstream

import (
	"git.um-react.app/um/cli/algo/common"

	_ "git.um-react.app/um/cli/algo/kgm"
	_ "git.um-react.app/um/cli/algo/kwm"
	_ "git.um-react.app/um/cli/algo/ncm"
	_ "git.um-react.app/um/cli/algo/qmc"
	_ "git.um-react.app/um/cli/algo/tm"
	_ "git.um-react.app/um/cli/algo/xiami"
	_ "git.um-react.app/um/cli/algo/ximalaya"
)

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
