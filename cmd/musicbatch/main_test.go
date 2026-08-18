package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunJSONAndHumanModes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "normal.mp3"), []byte("ID3"), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--out-dir", out, "--json", dir}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var decoded struct {
		Statuses []struct{} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("json: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"--out-dir", out, dir}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[SKIP] normal.mp3 : not encrypted") || !strings.Contains(stdout.String(), "Summary:") {
		t.Fatalf("output=%q", stdout.String())
	}
}

func TestInvalidRequestExitCode(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run(nil, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "usage:") {
		t.Fatalf("code=%d err=%q", code, errOut.String())
	}
}
