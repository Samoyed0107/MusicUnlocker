package batch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"musicunlocker/internal/decode"
)

var testKGM = []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func fakeDecoder(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
	name := filepath.Base(input)
	if strings.Contains(name, "key") {
		return decode.Result{}, &decode.Error{Code: decode.CodeQMCExternalKeyRequired, Err: errors.New("key required")}
	}
	if strings.Contains(name, "db") {
		return decode.Result{}, &decode.Error{Code: decode.CodeKGGDatabaseRequired, Err: errors.New("database required")}
	}
	if strings.Contains(name, "unsupported") {
		return decode.Result{}, &decode.Error{Code: decode.CodeUnsupportedEncrypted, Err: errors.New("unsupported")}
	}
	if strings.Contains(name, "fail") {
		return decode.Result{}, errors.New("synthetic failure")
	}
	if strings.Contains(name, "cancel") {
		return decode.Result{}, context.Canceled
	}
	if err := os.MkdirAll(options.OutDir, 0755); err != nil {
		return decode.Result{}, err
	}
	out := filepath.Join(options.OutDir, strings.TrimSuffix(name, filepath.Ext(name))+".mp3")
	if _, err := os.Stat(out); err == nil {
		return decode.Result{OutputPath: out}, &decode.Error{Code: decode.CodeOutputExists, Err: errors.New("output exists")}
	}
	data := []byte("ID3 synthetic")
	if err := os.WriteFile(out, data, 0600); err != nil {
		return decode.Result{}, err
	}
	return decode.Result{InputPath: input, OutputPath: out, DecodedFormat: "mp3", BytesWritten: int64(len(data))}, nil
}

func TestRunDiscoveryMappingAndContentHandling(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Encrypted (测试)")
	kgm := filepath.Join(root, "A", "茅野愛衣 - test.kgm")
	raw := filepath.Join(root, "B", "already.mp3")
	random := filepath.Join(root, "notes.txt")
	writeTestFile(t, kgm, testKGM)
	writeTestFile(t, raw, []byte("ID3 normal"))
	writeTestFile(t, random, []byte("random"))
	original, _ := os.ReadFile(kgm)
	out := filepath.Join(t.TempDir(), "Unlocked")
	summary, err := (Engine{Decoder: fakeDecoder}).Run(context.Background(), Request{Inputs: []string{root}, OutputRoot: out})
	if err != nil {
		t.Fatal(err)
	}
	if summary.DiscoveredRegularFiles != 3 || summary.Successes != 1 || summary.SkippedNotEncrypted != 1 || summary.Unknown != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	if got, _ := os.ReadFile(kgm); !reflect.DeepEqual(got, original) {
		t.Fatal("source changed")
	}
	wantOut := filepath.Join(out, "A", "茅野愛衣 - test.mp3")
	if _, err := os.Stat(wantOut); err != nil {
		t.Fatalf("nested output missing: %v", err)
	}
	if summary.Results[0].SourcePath == "" {
		t.Fatal("missing source path")
	}
}

func TestRunDeterministicDeduplicatedAndMissingInput(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.kgm")
	b := filepath.Join(dir, "b.kgm")
	writeTestFile(t, a, testKGM)
	writeTestFile(t, b, testKGM)
	summary, err := (Engine{Decoder: fakeDecoder}).Run(context.Background(), Request{Inputs: []string{dir, a, filepath.Join(dir, "missing.kgm")}, OutputRoot: filepath.Join(t.TempDir(), "out")})
	if err != nil {
		t.Fatal(err)
	}
	if summary.DiscoveredRegularFiles != 2 || summary.Failed != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	paths := []string{summary.Results[len(summary.Results)-2].SourcePath, summary.Results[len(summary.Results)-1].SourcePath}
	if !sort.StringsAreSorted(paths) {
		t.Fatalf("results are not deterministic: %v", paths)
	}
	seen := map[string]bool{}
	for _, r := range summary.Results {
		if r.Status != StatusFailed {
			if seen[r.SourcePath] {
				t.Fatalf("duplicate result for %s", r.SourcePath)
			}
			seen[r.SourcePath] = true
		}
	}
}

func TestMultipleDirectoryRootsUseCollisionSafeNamespaces(t *testing.T) {
	parent := t.TempDir()
	first := filepath.Join(parent, "one", "Music")
	second := filepath.Join(parent, "two", "Music")
	a := filepath.Join(first, "a.kgm")
	b := filepath.Join(second, "b.kgm")
	writeTestFile(t, a, testKGM)
	writeTestFile(t, b, testKGM)
	out := filepath.Join(t.TempDir(), "out")
	if _, err := (Engine{Decoder: fakeDecoder}).Run(context.Background(), Request{Inputs: []string{second, first}, OutputRoot: out}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "Music (1)", "a.mp3")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "Music (2)", "b.mp3")); err != nil {
		t.Fatal(err)
	}
}

func TestStatusesContinueAndOutputExists(t *testing.T) {
	dir := t.TempDir()
	names := []string{"success.kgm", "key.kgm", "db.kgm", "unsupported.kgm", "fail.kgm", "later.kgm"}
	for _, name := range names {
		writeTestFile(t, filepath.Join(dir, name), testKGM)
	}
	out := filepath.Join(t.TempDir(), "out")
	writeTestFile(t, filepath.Join(out, "success.mp3"), []byte("existing"))
	summary, err := (Engine{Decoder: fakeDecoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: out})
	if err != nil {
		t.Fatal(err)
	}
	if summary.SkippedExisting != 1 || summary.BlockedExternalKey != 1 || summary.BlockedExternalDatabase != 1 || summary.Unsupported != 1 || summary.Failed != 1 || summary.Successes != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	if got, _ := os.ReadFile(filepath.Join(out, "success.mp3")); string(got) != "existing" {
		t.Fatal("existing output was replaced")
	}
	if _, err := os.Stat(filepath.Join(out, "later.mp3")); err != nil {
		t.Fatalf("later file did not run: %v", err)
	}
}

func TestCancellationMarksRemainingAndObserverPanicsAreIgnored(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.kgm", "b.kgm", "c.kgm"} {
		writeTestFile(t, filepath.Join(dir, name), testKGM)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		cancel()
		return decode.Result{}, context.Canceled
	}
	var events []EventType
	observer := ObserverFunc(func(e Event) error {
		events = append(events, e.Type)
		if e.Type == EventStarted {
			panic("observer must not affect processing")
		}
		return nil
	})
	summary, err := (Engine{Decoder: decoder, Observer: observer}).Run(ctx, Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out")})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Cancelled != 3 {
		t.Fatalf("summary = %#v", summary)
	}
	if len(events) != 3+4 {
		t.Fatalf("events = %v", events)
	}
}

func TestSourceAndOutputSafetyAfterFailure(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.kgm")
	writeTestFile(t, source, testKGM)
	out := filepath.Join(t.TempDir(), "out")
	_, err := (Engine{Decoder: fakeDecoder}).Run(context.Background(), Request{Inputs: []string{source}, OutputRoot: out})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(out)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".musicunlocker-") {
			t.Fatalf("temp leakage: %s", entry.Name())
		}
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatal(err)
	}
}

type invalidAudioResolver struct{}

func (invalidAudioResolver) Resolve(*decode.Input) (decode.Stream, error) {
	return bytes.NewReader([]byte("not an audio stream")), nil
}

func TestDecodeFailureLeavesNoTemporaryOutput(t *testing.T) {
	source := filepath.Join(t.TempDir(), "bad.kgm")
	writeTestFile(t, source, testKGM)
	out := filepath.Join(t.TempDir(), "out")
	engine := Engine{Decoder: func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		options.Resolver = invalidAudioResolver{}
		return decode.DecodeFile(ctx, input, options)
	}}
	summary, err := engine.Run(context.Background(), Request{Inputs: []string{source}, OutputRoot: out})
	if err != nil || summary.Unknown != 1 {
		t.Fatalf("summary=%#v err=%v", summary, err)
	}
	entries, err := os.ReadDir(out)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".musicunlocker-") {
			t.Fatalf("temp leakage: %s", entry.Name())
		}
	}
}

func TestSuccessfulOutputSurvivesLaterFailure(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a.kgm")
	second := filepath.Join(dir, "b-fail.kgm")
	writeTestFile(t, first, testKGM)
	writeTestFile(t, second, testKGM)
	out := filepath.Join(t.TempDir(), "out")
	summary, err := (Engine{Decoder: fakeDecoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: out})
	if err != nil || summary.Successes != 1 || summary.Failed != 1 {
		t.Fatalf("summary=%#v err=%v", summary, err)
	}
	if _, err := os.Stat(filepath.Join(out, "a.mp3")); err != nil {
		t.Fatalf("successful output disappeared: %v", err)
	}
}

func TestJSONFieldsAreExplicit(t *testing.T) {
	if got := fmt.Sprintf("%s", StatusSuccess); got != "SUCCESS" {
		t.Fatal(got)
	}
}
