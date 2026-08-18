package outputplan

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"musicunlocker/internal/probe"
)

var ErrSourceConflict = errors.New("planned output conflicts with source")

type Request struct {
	InputPath     string
	OutputDir     string
	Probe         probe.Result
	DecodedFormat string
}

// Plan returns a clean, conservative destination name. It removes only
// recognised encrypted wrappers and their terminal copy counters.
func Plan(request Request) (string, error) {
	if request.InputPath == "" || request.OutputDir == "" || request.DecodedFormat == "" {
		return "", errors.New("input path, output directory, and decoded format are required")
	}
	format := strings.TrimPrefix(strings.ToLower(request.DecodedFormat), ".")
	if strings.ContainsAny(format, `/\\`) || format == "" {
		return "", fmt.Errorf("invalid decoded format %q", request.DecodedFormat)
	}
	base := cleanBase(filepath.Base(request.InputPath), request.Probe)
	output := filepath.Join(request.OutputDir, base+"."+format)
	if samePath(output, request.InputPath) {
		return "", ErrSourceConflict
	}
	return output, nil
}

var copyCounter = regexp.MustCompile(`^(.*?)(?:\s*\(\d+\))$`)

func cleanBase(name string, result probe.Result) string {
	removedWrapper := false
	if wrapper := knownWrapper(filepath.Ext(name), result); wrapper != "" {
		name = strings.TrimSuffix(name, wrapper)
		removedWrapper = true
	}
	if !removedWrapper {
		if ext := filepath.Ext(name); ext != "" && knownWrapper(ext, result) == "" {
			name = strings.TrimSuffix(name, ext)
		}
	}
	if match := copyCounter.FindStringSubmatch(name); len(match) == 2 && knownWrapper(filepath.Ext(match[1]), result) != "" {
		name = match[1]
	}
	if ext := filepath.Ext(name); ext != "" && knownWrapper(ext, result) != "" {
		name = strings.TrimSuffix(name, ext)
	}
	if name == "" {
		return "decoded"
	}
	return name
}

func knownWrapper(ext string, result probe.Result) string {
	ext = strings.ToLower(ext)
	switch result.Family {
	case probe.FamilyKugou:
		if ext == ".kgm" || ext == ".kgma" || ext == ".vpr" || ext == ".kgg" {
			return ext
		}
	case probe.FamilyNetease:
		if ext == ".ncm" {
			return ext
		}
	case probe.FamilyKuwo:
		if ext == ".kwm" {
			return ext
		}
	case probe.FamilyQMC:
		if isQMCWrapper(ext) {
			return ext
		}
	case probe.FamilyTM:
		if ext == ".tm0" || ext == ".tm2" || ext == ".tm3" || ext == ".tm6" {
			return ext
		}
	case probe.FamilyXiami:
		if ext == ".xm" {
			return ext
		}
	case probe.FamilyXimalaya:
		if ext == ".x2m" || ext == ".x3m" || ext == ".xm" {
			return ext
		}
	}
	for _, hint := range result.ExtensionHints {
		if strings.EqualFold(hint, ext) && ext != ".flac" && ext != ".mp3" && ext != ".ogg" && ext != ".wav" && ext != ".m4a" && ext != ".mp4" {
			return ext
		}
	}
	return ""
}

func isQMCWrapper(ext string) bool {
	switch ext {
	case ".qmc0", ".qmc2", ".qmc3", ".qmc4", ".qmc6", ".qmc8", ".qmcflac", ".qmcogg", ".tkm", ".bkcmp3", ".bkcm4a", ".bkcflac", ".bkcwav", ".bkcape", ".bkcogg", ".bkcwma", ".666c6163", ".6d7033", ".6f6767", ".6d3461", ".776176", ".mmp4", ".mgg", ".mgg0", ".mgg1", ".mgga", ".mggh", ".mggl", ".mggm", ".mflac", ".mflac0", ".mflac1", ".mflaca", ".mflach", ".mflacl", ".mflacm":
		return true
	default:
		return false
	}
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	return errA == nil && errB == nil && strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}
