// Package fileassoc defines the Windows Shell registration surface derived
// from the decoder adapter's authoritative filename-candidate registrations.
package fileassoc

import (
	"strings"

	"musicunlocker/internal/decoder/upstream"
)

const (
	ProgID                 = "MusicUnlocker.EncryptedMusic.1"
	FriendlyTypeName       = "MusicUnlocker Encrypted Music"
	ApplicationDescription = "Unlock supported encrypted music files offline"
)

var normalAudioExtensions = map[string]struct{}{
	".aac": {}, ".flac": {}, ".m4a": {}, ".mp3": {}, ".ogg": {}, ".wav": {}, ".wma": {},
}

// WindowsExtensions returns the simple encrypted-container candidate
// extensions that the installer may register. Compound decoder suffixes remain
// application-level filename hints because Windows associations match only the
// final extension. Normal decoded audio extensions are intentionally excluded.
func WindowsExtensions() []string {
	candidates := upstream.RegisteredCandidateSuffixes()
	result := make([]string, 0, len(candidates))
	for _, extension := range candidates {
		if strings.Count(extension, ".") != 1 {
			continue
		}
		if _, decoded := normalAudioExtensions[extension]; decoded {
			continue
		}
		result = append(result, extension)
	}
	return result
}
