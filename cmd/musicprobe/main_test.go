package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func syntheticKGM() []byte {
	b := make([]byte, 0x3c)
	copy(b, []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14})
	binary.LittleEndian.PutUint32(b[0x10:], 1024)
	binary.LittleEndian.PutUint32(b[0x14:], 3)
	binary.LittleEndian.PutUint32(b[0x18:], 1)
	return b
}

func syntheticSTag() []byte { return append([]byte("encrypted"), []byte("STag")...) }

func writeFixture(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHumanFormatting(t *testing.T) {
	for _, tt := range []struct {
		name string
		data []byte
		want []string
	}{
		{"kgm", syntheticKGM(), []string{"Family: kugou", "Variant: KGM", "AudioOffset: 1024", "CryptoVersion: 3", "CryptoSlot: 1", ".kgm"}},
		{"st tag", syntheticSTag(), []string{"Family: qmc", "Variant: STag", "FooterType: STag", "NeedsExternalKey: true"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixture(t, tt.name+".kgm.flac", tt.data)
			var out, errOut bytes.Buffer
			if code := run([]string{path}, &out, &errOut); code != 0 {
				t.Fatalf("exit %d: %s", code, errOut.String())
			}
			for _, want := range tt.want {
				if !strings.Contains(out.String(), want) {
					t.Fatalf("output missing %q:\n%s", want, out.String())
				}
			}
		})
	}
}

func TestJSONMultipleFilesAndUnknown(t *testing.T) {
	first := writeFixture(t, "one.kgm", syntheticKGM())
	second := writeFixture(t, "two.unknown", []byte("random"))
	var out, errOut bytes.Buffer
	if code := run([]string{"--json", second, first}, &out, &errOut); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	firstOutput := out.String()
	var secondOutput bytes.Buffer
	if code := run([]string{"--json", second, first}, &secondOutput, &errOut); code != 0 {
		t.Fatalf("second exit %d: %s", code, errOut.String())
	}
	if firstOutput != secondOutput.String() {
		t.Fatalf("JSON output is not deterministic:\n%s\n---\n%s", firstOutput, secondOutput.String())
	}
	var decoded struct {
		Files []struct {
			Path   string `json:"path"`
			Result *struct {
				Family string `json:"family"`
			} `json:"result"`
			Error string `json:"error"`
		} `json:"files"`
	}
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if len(decoded.Files) != 2 || decoded.Files[0].Result == nil || decoded.Files[0].Result.Family != "unknown" || decoded.Files[0].Error != "" {
		t.Fatalf("unexpected unknown entry: %#v", decoded.Files[0])
	}
	if decoded.Files[1].Result == nil || decoded.Files[1].Result.Family != "kugou" {
		t.Fatalf("unexpected KGM entry: %#v", decoded.Files[1])
	}
}

func TestMissingFileIsOperationalError(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"missing-file"}, &out, &errOut); code == 0 {
		t.Fatal("missing file returned success")
	}
	if !strings.Contains(out.String(), "Error:") {
		t.Fatalf("human output missing error: %s", out.String())
	}
	var jsonOut, jsonErr bytes.Buffer
	if code := run([]string{"--json", "missing-file"}, &jsonOut, &jsonErr); code == 0 {
		t.Fatal("missing JSON file returned success")
	}
	if !strings.Contains(jsonOut.String(), "\"error\"") {
		t.Fatalf("JSON output missing error: %s", jsonOut.String())
	}
}

func TestDiagnosticDoesNotCreateOutputBesideInput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input.kgm")
	if err := os.WriteFile(path, syntheticKGM(), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{path}, &out, &errOut); code != 0 {
		t.Fatal(errOut.String())
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "input.kgm" {
		t.Fatalf("unexpected files beside input: %#v", entries)
	}
}
