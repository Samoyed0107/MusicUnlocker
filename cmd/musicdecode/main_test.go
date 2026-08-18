package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"musicunlocker/internal/decode"
)

func TestTypedErrorCodeRenderedOnce(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "blocked.qmc0")
	if err := os.WriteFile(input, append([]byte("encrypted"), []byte("STag")...), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"--out-dir", filepath.Join(dir, "out"), input}, &out, &errOut); code == 0 {
		t.Fatal("blocked decode succeeded")
	}
	if got := errOut.String(); strings.Count(got, "QMC_EXTERNAL_KEY_REQUIRED") != 1 || !strings.HasPrefix(got, "QMC_EXTERNAL_KEY_REQUIRED: ") {
		t.Fatalf("unexpected diagnostic: %q", got)
	}
}

func TestOutputExistsErrorRenderedOnce(t *testing.T) {
	var out bytes.Buffer
	writeError(&out, &decode.Error{Code: decode.CodeOutputExists, Err: errors.New("output already exists: D:\\out\\foo.mp3")})
	if got := out.String(); got != "OUTPUT_EXISTS: output already exists: D:\\out\\foo.mp3\n" {
		t.Fatalf("unexpected diagnostic: %q", got)
	}
}
