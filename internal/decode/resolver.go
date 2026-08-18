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
			return nil, coded(CodeUnsupportedEncrypted, errors.New("Kugou crypto version is not supported in Stage 3C"))
		}
		return upstream.DecodeKGM(input.Reader, input.Path)
	case probe.FamilyQMC:
		if input.Probe.NeedsExternalKey {
			return nil, coded(CodeQMCExternalKeyRequired, errors.New("QMC format requires an external key"))
		}
		if stream, err := upstream.DecodeQMC(input.Reader, input.Path); err == nil {
			return stream, nil
		} else {
			return nil, err
		}
	case probe.FamilyNetease:
		return upstream.DecodeNCM(input.Reader, input.Path)
	case probe.FamilyKuwo:
		return upstream.DecodeKWM(input.Reader, input.Path)
	case probe.FamilyTM:
		return upstream.DecodeTM(input.Reader, input.Path)
	case probe.FamilyXiami:
		return upstream.DecodeXiami(input.Reader, input.Path)
	default:
		for _, candidate := range fallbackCandidates(input.Probe.ExtensionHints) {
			var stream Stream
			var err error
			switch candidate {
			case "qmc":
				stream, err = upstream.DecodeQMC(input.Reader, input.Path)
			case "tm":
				stream, err = upstream.DecodeTM(input.Reader, input.Path)
			case "xiami":
				stream, err = upstream.DecodeXiami(input.Reader, input.Path)
			case "ximalaya":
				stream, err = upstream.DecodeXimalaya(input.Reader, input.Path)
			case "ncm":
				stream, err = upstream.DecodeNCM(input.Reader, input.Path)
			case "kwm":
				stream, err = upstream.DecodeKWM(input.Reader, input.Path)
			}
			if err == nil {
				markCandidate(input, candidate)
				return stream, nil
			}
		}
		if len(fallbackCandidates(input.Probe.ExtensionHints)) > 0 {
			return nil, coded(CodeUnsupportedEncrypted, errors.New("candidate decoder validation failed"))
		}
		return nil, coded(CodeUnknownFormat, errors.New("input content is not a recognised encrypted format"))
	}
}

func markCandidate(input *Input, candidate string) {
	switch candidate {
	case "qmc":
		input.Probe.Family, input.Probe.Variant = probe.FamilyQMC, "QMC"
	case "tm":
		input.Probe.Family, input.Probe.Variant = probe.FamilyTM, "TM"
	case "xiami":
		input.Probe.Family, input.Probe.Variant = probe.FamilyXiami, "XM"
	case "ximalaya":
		input.Probe.Family, input.Probe.Variant = probe.FamilyXimalaya, "XIMALAYA"
	case "ncm":
		input.Probe.Family, input.Probe.Variant = probe.FamilyNetease, "NCM"
	case "kwm":
		input.Probe.Family, input.Probe.Variant = probe.FamilyKuwo, "KWM"
	}
}

func fallbackCandidates(hints []string) []string {
	seen := map[string]bool{}
	var result []string
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	for _, hint := range hints {
		switch hint {
		case ".qmc0", ".qmc2", ".qmc3", ".qmc4", ".qmc6", ".qmc8", ".qmcflac", ".qmcogg", ".tkm", ".bkcmp3", ".bkcm4a", ".bkcflac", ".bkcwav", ".bkcape", ".bkcogg", ".bkcwma", ".666c6163", ".6d7033", ".6f6767", ".6d3461", ".776176", ".mmp4", ".mgg", ".mgg0", ".mgg1", ".mgga", ".mggh", ".mggl", ".mggm", ".mflac", ".mflac0", ".mflac1", ".mflaca", ".mflach", ".mflacl", ".mflacm":
			add("qmc")
		case ".tm0", ".tm2", ".tm3", ".tm6":
			add("tm")
		case ".x2m", ".x3m":
			add("ximalaya")
		case ".xm":
			add("xiami")
			add("ximalaya")
		case ".ncm":
			add("ncm")
		case ".kwm":
			add("kwm")
		}
	}
	return result
}
