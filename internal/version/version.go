// Package version contains the single application release identity.
package version

import "fmt"

const (
	ApplicationName    = "MusicUnlocker"
	ApplicationVersion = "0.7.0"
)

// These values are set by release builds with go -ldflags -X. Development
// builds retain safe, explicit defaults.
var (
	BuildTimestamp  = "unknown"
	BuildIdentifier = "dev"
)

type Identity struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	BuildTimestamp  string `json:"buildTimestamp"`
	BuildIdentifier string `json:"buildIdentifier"`
}

func Current() Identity {
	return Identity{ApplicationName, ApplicationVersion, BuildTimestamp, BuildIdentifier}
}

func CLI(tool string) string {
	i := Current()
	return fmt.Sprintf("%s %s %s (build %s, built %s)", i.Name, tool, i.Version, i.BuildIdentifier, i.BuildTimestamp)
}
