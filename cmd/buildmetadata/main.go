// Command buildmetadata writes the Wails Windows version resource from the
// centralized application identity. Wails runs it from the frontend directory.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"musicunlocker/internal/fileassoc"
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

func newAssociationNSISInfo(identity version.Identity, extensions []string) string {
	var output strings.Builder
	fmt.Fprintf(&output, "!define FILE_ASSOC_PROGID %q\n", fileassoc.ProgID)
	fmt.Fprintf(&output, "!define FILE_ASSOC_CAPABILITIES_KEY \"Software\\%s\\Capabilities\"\n\n", identity.Name)
	output.WriteString("!macro RegisterFileAssociations\n")
	output.WriteString("    SetRegView 64\n")
	fmt.Fprintf(&output, "    WriteRegStr HKCU \"Software\\Classes\\${FILE_ASSOC_PROGID}\" \"\" %q\n", fileassoc.FriendlyTypeName)
	fmt.Fprintf(&output, "    WriteRegStr HKCU \"Software\\Classes\\${FILE_ASSOC_PROGID}\" \"FriendlyTypeName\" %q\n", fileassoc.FriendlyTypeName)
	output.WriteString("    WriteRegStr HKCU \"Software\\Classes\\${FILE_ASSOC_PROGID}\\DefaultIcon\" \"\" \"$\\\"$INSTDIR\\${PRODUCT_EXECUTABLE}$\\\",0\"\n")
	output.WriteString("    WriteRegStr HKCU \"Software\\Classes\\${FILE_ASSOC_PROGID}\\shell\\open\\command\" \"\" \"$\\\"$INSTDIR\\${PRODUCT_EXECUTABLE}$\\\" $\\\"%1$\\\"\"\n")
	fmt.Fprintf(&output, "    WriteRegStr HKCU \"${FILE_ASSOC_CAPABILITIES_KEY}\" \"ApplicationName\" %q\n", identity.Name)
	fmt.Fprintf(&output, "    WriteRegStr HKCU \"${FILE_ASSOC_CAPABILITIES_KEY}\" \"ApplicationDescription\" %q\n", fileassoc.ApplicationDescription)
	for _, extension := range extensions {
		fmt.Fprintf(&output, "    WriteRegStr HKCU \"Software\\Classes\\%s\\OpenWithProgids\" \"${FILE_ASSOC_PROGID}\" \"\"\n", extension)
		fmt.Fprintf(&output, "    WriteRegStr HKCU \"${FILE_ASSOC_CAPABILITIES_KEY}\\FileAssociations\" \"%s\" \"${FILE_ASSOC_PROGID}\"\n", extension)
	}
	fmt.Fprintf(&output, "    WriteRegStr HKCU \"Software\\RegisteredApplications\" %q \"${FILE_ASSOC_CAPABILITIES_KEY}\"\n", identity.Name)
	output.WriteString("    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0x0000, p 0, p 0)'\n")
	output.WriteString("!macroend\n\n")

	output.WriteString("!macro UnregisterFileAssociations\n")
	output.WriteString("    SetRegView 64\n")
	for _, extension := range extensions {
		fmt.Fprintf(&output, "    DeleteRegValue HKCU \"Software\\Classes\\%s\\OpenWithProgids\" \"${FILE_ASSOC_PROGID}\"\n", extension)
		fmt.Fprintf(&output, "    DeleteRegKey /ifempty HKCU \"Software\\Classes\\%s\\OpenWithProgids\"\n", extension)
		fmt.Fprintf(&output, "    DeleteRegKey /ifempty HKCU \"Software\\Classes\\%s\"\n", extension)
	}
	fmt.Fprintf(&output, "    DeleteRegValue HKCU \"Software\\RegisteredApplications\" %q\n", identity.Name)
	output.WriteString("    DeleteRegKey HKCU \"${FILE_ASSOC_CAPABILITIES_KEY}\"\n")
	output.WriteString("    DeleteRegKey HKCU \"Software\\Classes\\${FILE_ASSOC_PROGID}\"\n")
	output.WriteString("    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0x0000, p 0, p 0)'\n")
	output.WriteString("!macroend\n")
	return output.String()
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
	associations := newAssociationNSISInfo(identity, fileassoc.WindowsExtensions())
	if err := os.WriteFile(filepath.Join("..", "build", "windows", "associations.nsh"), []byte(associations), 0o644); err != nil {
		panic(err)
	}
}
