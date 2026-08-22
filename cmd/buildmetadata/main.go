// Command buildmetadata writes the Wails Windows version resource from the
// centralized application identity. Wails runs it from the frontend directory.
package main

import (
	"encoding/json"
	"fmt"
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

func newWindowsInfo(identity version.Identity) windowsInfo {
	metadata := windowsInfo{Info: map[string]map[string]string{
		"0409": {
			"ProductVersion":   identity.Version,
			"FileVersion":      identity.Version,
			"CompanyName":      identity.Name,
			"FileDescription":  identity.Name,
			"ProductName":      identity.Name,
			"OriginalFilename": identity.Name + ".exe",
			"LegalCopyright":   version.ApplicationCopyright,
			"Comments":         "Local desktop utility",
		},
	}}
	metadata.Fixed.FileVersion = identity.Version + ".0"
	metadata.Fixed.ProductVersion = identity.Version + ".0"
	return metadata
}

func newNSISInfo(identity version.Identity) string {
	return fmt.Sprintf("!define INFO_PROJECTNAME %q\n!define INFO_COMPANYNAME %q\n!define INFO_PRODUCTNAME %q\n!define INFO_PRODUCTVERSION %q\n!define INFO_COPYRIGHT %q\n!define PRODUCT_EXECUTABLE %q\n", identity.Name, identity.Name, identity.Name, identity.Version, version.ApplicationCopyright, identity.Name+".exe")
}

func main() {
	identity := version.Current()
	metadata := newWindowsInfo(identity)
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
	nsis := newNSISInfo(identity)
	if err := os.WriteFile(filepath.Join("..", "build", "windows", "version.nsh"), []byte(nsis), 0o644); err != nil {
		panic(err)
	}
}
