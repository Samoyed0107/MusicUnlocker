// Command buildmetadata writes the Wails Windows version resource from the
// centralized application identity. Wails runs it from the frontend directory.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"musicunlocker/internal/version"
)

type windowsInfo struct {
	Fixed struct {
		FileVersion    string `json:"file_version"`
		ProductVersion string `json:"product_version"`
	} `json:"fixed"`
	Info map[string]map[string]string `json:"info"`
}

func main() {
	identity := version.Current()
	metadata := windowsInfo{Info: map[string]map[string]string{
		"0000": {
			"ProductVersion":  identity.Version,
			"CompanyName":     identity.Name,
			"FileDescription": identity.Name,
			"ProductName":     identity.Name,
			"Comments":        "Local desktop utility",
		},
		"0409": {
			"ProductVersion":  identity.Version,
			"CompanyName":     identity.Name,
			"FileDescription": identity.Name,
			"ProductName":     identity.Name,
			"Comments":        "Local desktop utility",
		},
	}}
	metadata.Fixed.FileVersion = identity.Version + ".0"
	metadata.Fixed.ProductVersion = identity.Version + ".0"
	data, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		panic(err)
	}
	path := filepath.Join("..", "build", "windows", "info.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		panic(err)
	}
}
