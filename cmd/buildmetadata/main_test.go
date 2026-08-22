package main

import (
	"strings"
	"testing"

	"musicunlocker/internal/fileassoc"
	"musicunlocker/internal/version"
)

func TestGeneratedMetadataUsesCentralReleaseIdentity(t *testing.T) {
	identity := version.Current()
	metadata := newWindowsInfo(identity)
	if metadata.Fixed.FileVersion != identity.Version+".0" || metadata.Fixed.ProductVersion != identity.Version+".0" {
		t.Fatalf("fixed metadata=%+v", metadata.Fixed)
	}
	if len(metadata.Info) != 1 {
		t.Fatalf("language tables=%v", metadata.Info)
	}
	english := metadata.Info["0409"]
	for key, want := range map[string]string{
		"ProductVersion":   identity.Version,
		"FileVersion":      identity.Version,
		"ProductName":      identity.Name,
		"FileDescription":  identity.Name,
		"OriginalFilename": identity.Name + ".exe",
		"LegalCopyright":   version.ApplicationCopyright,
	} {
		if got := english[key]; got != want {
			t.Errorf("%s=%q, want %q", key, got, want)
		}
	}

	nsis := newNSISInfo(identity)
	for _, want := range []string{
		`!define INFO_PRODUCTNAME "` + identity.Name + `"`,
		`!define INFO_PRODUCTVERSION "` + identity.Version + `"`,
		`!define PRODUCT_EXECUTABLE "` + identity.Name + `.exe"`,
	} {
		if !strings.Contains(nsis, want) {
			t.Errorf("NSIS metadata missing %q", want)
		}
	}
}

func TestGeneratedAssociationRegistrationIsPerUserAndQuoted(t *testing.T) {
	identity := version.Current()
	nsis := newAssociationNSISInfo(identity, []string{".kgm", ".ncm"})
	for _, want := range []string{
		`!define FILE_ASSOC_PROGID "` + fileassoc.ProgID + `"`,
		`WriteRegStr HKCU "Software\Classes\.kgm\OpenWithProgids" "${FILE_ASSOC_PROGID}" ""`,
		`WriteRegStr HKCU "${FILE_ASSOC_CAPABILITIES_KEY}\FileAssociations" ".ncm" "${FILE_ASSOC_PROGID}"`,
		`WriteRegStr HKCU "Software\RegisteredApplications" "` + identity.Name + `" "${FILE_ASSOC_CAPABILITIES_KEY}"`,
		`WriteRegStr HKCU "Software\Classes\${FILE_ASSOC_PROGID}\shell\open\command" "" "$\"$INSTDIR\${PRODUCT_EXECUTABLE}$\" $\"%1$\""`,
		`DeleteRegValue HKCU "Software\Classes\.kgm\OpenWithProgids" "${FILE_ASSOC_PROGID}"`,
		`SHChangeNotify(i 0x08000000`,
	} {
		if !strings.Contains(nsis, want) {
			t.Errorf("association metadata missing %q", want)
		}
	}
	for _, forbidden := range []string{"HKLM", "UserChoice", `DeleteRegKey HKCU "Software\Classes\.kgm"`} {
		if strings.Contains(nsis, forbidden) {
			t.Errorf("association metadata contains unsafe registration %q", forbidden)
		}
	}
}
