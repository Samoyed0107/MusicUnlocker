// Command releaseinfo exposes the centralized release identity to packaging
// scripts without duplicating version strings outside internal/version.
package main

import (
	"encoding/json"
	"os"

	"musicunlocker/internal/version"
)

func main() {
	identity := version.Current()
	output := struct {
		Name      string `json:"name"`
		Version   string `json:"version"`
		Copyright string `json:"copyright"`
	}{identity.Name, identity.Version, version.ApplicationCopyright}
	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
		panic(err)
	}
}
