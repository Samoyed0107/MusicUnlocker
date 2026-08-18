package decode

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"musicunlocker/internal/probe"
)

type fakeResolver struct {
	data []byte
	err  error
}

func (r fakeResolver) Resolve(*Input) (Stream, error) {
	if r.err != nil {
		return nil, r.err
	}
	return bytes.NewReader(r.data), nil
}
func writeInput(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDecodePreservesPrefixAndUsesDecodedFormat(t *testing.T) {
	in := writeInput(t, "secret.kgm", []byte("random encrypted input"))
	out := t.TempDir()
	data := append([]byte("fLaC"), []byte("payload")...)
	got, err := DecodeFile(context.Background(), in, Options{OutDir: out, Resolver: fakeResolver{data: data}})
	if err != nil {
		t.Fatal(err)
	}
	if got.DecodedFormat != "flac" || got.BytesWritten != int64(len(data)) {
		t.Fatalf("result %#v", got)
	}
	written, err := os.ReadFile(got.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, data) {
		t.Fatalf("output %q, want %q", written, data)
	}
}

func TestDecodeErrorsAndSafety(t *testing.T) {
	in := writeInput(t, "secret.kgm", []byte("unchanged"))
	original, _ := os.ReadFile(in)
	out := t.TempDir()
	_, err := DecodeFile(context.Background(), in, Options{OutDir: out, Resolver: fakeResolver{data: []byte("not audio")}})
	if CodeOf(err) != CodeDecodedAudioUnknown {
		t.Fatalf("code %q, err %v", CodeOf(err), err)
	}
	if got, _ := os.ReadFile(in); !bytes.Equal(got, original) {
		t.Fatal("source changed")
	}

	data := append([]byte("ID3"), bytes.Repeat([]byte{'x'}, 100)...)
	first, err := DecodeFile(context.Background(), in, Options{OutDir: out, Resolver: fakeResolver{data: data}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeFile(context.Background(), in, Options{OutDir: out, Resolver: fakeResolver{data: data}}); CodeOf(err) != CodeOutputExists {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(first.OutputPath); err != nil {
		t.Fatal(err)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = DecodeFile(cancelled, in, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: data}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestResolverTypedStates(t *testing.T) {
	cases := []struct {
		name   string
		result probe.Result
		code   ErrorCode
	}{
		{"unknown", probe.Result{}, CodeUnknownFormat},
		{"qmc", probe.Result{Family: probe.FamilyQMC, NeedsExternalKey: true}, CodeQMCExternalKeyRequired},
		{"kgm v5", probe.Result{Family: probe.FamilyKugou, CryptoVersion: 5}, CodeKGGDatabaseRequired},
		{"ncm", probe.Result{Family: probe.FamilyNetease}, CodeUnsupportedEncrypted},
		{"tm candidate", probe.Result{ExtensionHints: []string{".tm2"}}, CodeUnsupportedEncrypted},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (DefaultResolver{}).Resolve(&Input{Probe: tc.result})
			if CodeOf(err) != tc.code {
				t.Fatalf("code %q, err %v", CodeOf(err), err)
			}
		})
	}
}
