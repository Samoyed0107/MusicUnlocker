package decode

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"musicunlocker/internal/externalmaterial"
	"musicunlocker/internal/probe"
)

type fakeResolver struct {
	data []byte
	err  error
}

type recordingMaterialProvider struct {
	path     string
	material []byte
}

type recordingDBProvider struct{ calls int }

func (p *recordingDBProvider) ResolveKGGDatabase(context.Context, string) (externalmaterial.KGGDatabase, bool, error) {
	p.calls++
	return externalmaterial.KGGDatabase{}, false, nil
}

func (p *recordingMaterialProvider) ResolveQMCExternalKey(_ context.Context, path string) ([]byte, bool, error) {
	p.path = path
	return append([]byte(nil), p.material...), true, nil
}

func TestResolverControlledFallback(t *testing.T) {
	for _, tc := range []struct{ name, suffix, want string }{
		{"tm", ".tm2", "candidate decoder validation failed"},
		{"ximalaya", ".x2m", "candidate decoder validation failed"},
		{"ambiguous xm", ".xm", "candidate decoder validation failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (DefaultResolver{}).Resolve(&Input{Path: "sample" + tc.suffix, Probe: probe.Result{ExtensionHints: []string{tc.suffix}}, Reader: bytes.NewReader([]byte("invalid"))})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v", err)
			}
		})
	}
	_, err := (DefaultResolver{}).Resolve(&Input{Path: "sample.mp3", Probe: probe.Result{ExtensionHints: []string{".mp3"}}, Reader: bytes.NewReader([]byte("invalid"))})
	if CodeOf(err) != CodeUnknownFormat {
		t.Fatalf("generic raw candidate code = %q", CodeOf(err))
	}
}

func TestResolverTMRawPassthroughCandidate(t *testing.T) {
	stream, err := (DefaultResolver{}).Resolve(&Input{Path: "sample.tm0", Probe: probe.Result{Family: probe.FamilyRaw, ExtensionHints: []string{".tm0"}}, Reader: bytes.NewReader([]byte("ID3\x04\x00audio"))})
	if err != nil {
		t.Fatal(err)
	}
	prefix := make([]byte, 3)
	if _, err := stream.Read(prefix); err != nil {
		t.Fatal(err)
	}
	if string(prefix) != "ID3" {
		t.Fatalf("prefix = %q", prefix)
	}
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

func TestExternalMaterialProviderIsResolvedOnlyForRequiredQMC(t *testing.T) {
	in := writeInput(t, "needs.qmc0", append([]byte("encrypted"), []byte("STag")...))
	provider := &recordingMaterialProvider{material: []byte("synthetic-ekey")}
	_, err := DecodeFile(context.Background(), in, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: append([]byte("ID3"), []byte("audio")...)}, ExternalMaterialProvider: provider})
	if provider.path != in {
		t.Fatalf("provider path=%q want %q", provider.path, in)
	}
	if err != nil {
		t.Fatal(err)
	}
	plain := writeInput(t, "self.qmc0", []byte("encrypted"))
	provider.path = ""
	_, err = DecodeFile(context.Background(), plain, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: append([]byte("ID3"), []byte("audio")...)}, ExternalMaterialProvider: provider})
	if err != nil {
		t.Fatal(err)
	}
	if provider.path != "" {
		t.Fatalf("provider was consulted for self-contained input: %q", provider.path)
	}
}

func TestKGMV3DoesNotConsultDatabaseProvider(t *testing.T) {
	header := make([]byte, 64)
	copy(header, []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14})
	binary.LittleEndian.PutUint32(header[0x10:], 60)
	binary.LittleEndian.PutUint32(header[0x14:], 3)
	binary.LittleEndian.PutUint32(header[0x18:], 1)
	in := writeInput(t, "v3.kgm", header)
	provider := &recordingDBProvider{}
	_, err := DecodeFile(context.Background(), in, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: append([]byte("ID3"), []byte("audio")...)}, KGGDatabaseProvider: provider})
	if err != nil {
		t.Fatal(err)
	}
	if provider.calls != 0 {
		t.Fatalf("KGM V3 consulted database provider %d times", provider.calls)
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

func TestDecodeProgressIsMonotonicAndIncludesPrefix(t *testing.T) {
	in := writeInput(t, "progress.kgm", []byte("encrypted source"))
	var progress []Progress
	data := append([]byte("ID3"), bytes.Repeat([]byte{'x'}, 100)...)
	result, err := DecodeFile(context.Background(), in, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: data}, Progress: func(p Progress) { progress = append(progress, p) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(progress) < 2 || progress[len(progress)-1].BytesWritten != result.BytesWritten || result.BytesWritten != int64(len(data)) {
		t.Fatalf("progress=%#v result=%#v", progress, result)
	}
	last := int64(0)
	for _, p := range progress {
		if p.BytesWritten < last || p.BytesWritten > result.BytesWritten {
			t.Fatalf("non-monotonic progress=%#v", progress)
		}
		last = p.BytesWritten
	}
	if !progress[0].SourceBytesKnown || progress[0].SourceBytes == 0 {
		t.Fatalf("source size missing: %#v", progress[0])
	}
}

func TestDecodeProgressCancellationHasNoFakeFinalTotal(t *testing.T) {
	in := writeInput(t, "cancel-progress.kgm", []byte("encrypted source"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var progress []Progress
	_, err := DecodeFile(ctx, in, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: append([]byte("ID3"), bytes.Repeat([]byte{'x'}, 100)...)}, Progress: func(p Progress) { progress = append(progress, p) }})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	for _, p := range progress {
		if p.BytesWritten != 0 {
			t.Fatalf("fake final progress: %#v", progress)
		}
	}
}

func TestDecodeProgressCallbackPanicDoesNotCorruptOutput(t *testing.T) {
	in := writeInput(t, "panic-progress.kgm", []byte("encrypted source"))
	data := append([]byte("ID3"), []byte("payload")...)
	result, err := DecodeFile(context.Background(), in, Options{OutDir: t.TempDir(), Resolver: fakeResolver{data: data}, Progress: func(Progress) { panic("observer") }})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(result.OutputPath); err != nil || !bytes.Equal(got, data) {
		t.Fatalf("output=%q err=%v", got, err)
	}
}
