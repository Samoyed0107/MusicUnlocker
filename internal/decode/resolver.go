package decode

import (
	"errors"

	"musicunlocker/internal/decoder/upstream"
	"musicunlocker/internal/probe"
)

type DefaultResolver struct{}

func (DefaultResolver) Resolve(input *Input) (Stream, error) {
	if input == nil {
		return nil, coded(CodeInvalidInput, errors.New("nil decode input"))
	}
	switch input.Probe.Family {
	case probe.FamilyKugou:
		if input.Probe.CryptoVersion == 5 {
			return nil, coded(CodeKGGDatabaseRequired, errors.New("KGM CryptoVersion 5 requires KGMusicV3.db"))
		}
		if input.Probe.CryptoVersion != 3 {
			return nil, coded(CodeUnsupportedEncrypted, errors.New("Kugou crypto version is not supported in Stage 3A"))
		}
		return upstream.DecodeKGM(input.Reader, input.Path)
	case probe.FamilyQMC:
		if input.Probe.NeedsExternalKey {
			return nil, coded(CodeQMCExternalKeyRequired, errors.New("QMC format requires an external key"))
		}
		return nil, coded(CodeUnsupportedEncrypted, errors.New("QMC variant is not wired into Stage 3A"))
	case probe.FamilyNetease:
		return upstream.DecodeNCM(input.Reader, input.Path)
	case probe.FamilyKuwo:
		return upstream.DecodeKWM(input.Reader, input.Path)
	default:
		for _, hint := range input.Probe.ExtensionHints {
			switch hint {
			case ".tm2", ".xm", ".x2m":
				return nil, coded(CodeUnsupportedEncrypted, errors.New("encrypted family is recognised by filename candidate but not wired into Stage 3A"))
			}
		}
		return nil, coded(CodeUnknownFormat, errors.New("input content is not a recognised encrypted format"))
	}
}
