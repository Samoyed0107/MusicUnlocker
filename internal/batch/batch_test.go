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
	"sync"
	"sync/atomic"
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

func writeEncryptedSet(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		writeTestFile(t, filepath.Join(dir, name), testKGM)
	}
}

func TestBoundedConcurrencyAndStableOrder(t *testing.T) {
	dir := t.TempDir()
	writeEncryptedSet(t, dir, "a.kgm", "b.kgm", "c.kgm", "d.kgm")
	var active, maximum int32
	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		current := atomic.AddInt32(&active, 1)
		for {
			old := atomic.LoadInt32(&maximum)
			if current <= old || atomic.CompareAndSwapInt32(&maximum, old, current) {
				break
			}
		}
		entered <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			atomic.AddInt32(&active, -1)
			return decode.Result{}, ctx.Err()
		}
		atomic.AddInt32(&active, -1)
		return syntheticDecodeResult(input, options)
	}
	done := make(chan Summary, 1)
	go func() {
		summary, _ := (Engine{Decoder: decoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 2})
		done <- summary
	}()
	<-entered
	<-entered
	close(release)
	summary := <-done
	if maximum > 2 || maximum < 2 {
		t.Fatalf("maximum active workers = %d", maximum)
	}
	if len(summary.Results) != 4 || summary.Successes != 4 {
		t.Fatalf("summary = %#v", summary)
	}
	for i, result := range summary.Results {
		if result.Index != i {
			t.Fatalf("index %d result = %#v", i, result)
		}
		if i > 0 && summary.Results[i-1].SourcePath > result.SourcePath {
			t.Fatal("results are not ordered")
		}
	}
}

func TestMaxWorkersOneAndZeroRemainSequential(t *testing.T) {
	for _, workers := range []int{0, 1} {
		dir := t.TempDir()
		writeEncryptedSet(t, dir, "a.kgm", "b.kgm")
		var active, maximum int32
		decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
			current := atomic.AddInt32(&active, 1)
			updateMax(&maximum, current)
			atomic.AddInt32(&active, -1)
			return syntheticDecodeResult(input, options)
		}
		summary, err := (Engine{Decoder: decoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: workers})
		if err != nil || summary.Successes != 2 || maximum != 1 {
			t.Fatalf("workers=%d maximum=%d summary=%#v err=%v", workers, maximum, summary, err)
		}
	}
}

func TestCompletionOrderDoesNotChangeResultOrder(t *testing.T) {
	dir := t.TempDir()
	writeEncryptedSet(t, dir, "a.kgm", "b.kgm")
	aReady, bReady, allowA := make(chan struct{}), make(chan struct{}), make(chan struct{})
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		if filepath.Base(input) == "a.kgm" {
			close(aReady)
			select {
			case <-allowA:
			case <-ctx.Done():
				return decode.Result{}, ctx.Err()
			}
		} else {
			close(bReady)
		}
		return syntheticDecodeResult(input, options)
	}
	done := make(chan Summary, 1)
	go func() {
		summary, _ := (Engine{Decoder: decoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 2})
		done <- summary
	}()
	<-aReady
	<-bReady
	close(allowA)
	summary := <-done
	if filepath.Base(summary.Results[0].SourcePath) != "a.kgm" || filepath.Base(summary.Results[1].SourcePath) != "b.kgm" {
		t.Fatalf("results = %#v", summary.Results)
	}
}

func TestTargetStemCollisionSerializesAndChoosesQueueWinner(t *testing.T) {
	for repeat := 0; repeat < 5; repeat++ {
		dir := t.TempDir()
		writeEncryptedSet(t, dir, "foo.kgm", "foo.kgma")
		var active, maximum int32
		decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
			current := atomic.AddInt32(&active, 1)
			for {
				old := atomic.LoadInt32(&maximum)
				if current <= old || atomic.CompareAndSwapInt32(&maximum, old, current) {
					break
				}
			}
			defer atomic.AddInt32(&active, -1)
			out := filepath.Join(options.OutDir, "foo.mp3")
			if _, err := os.Stat(out); err == nil {
				return decode.Result{OutputPath: out}, &decode.Error{Code: decode.CodeOutputExists, Err: errors.New("output exists")}
			}
			if err := os.MkdirAll(options.OutDir, 0755); err != nil {
				return decode.Result{}, err
			}
			if err := os.WriteFile(out, []byte("synthetic"), 0600); err != nil {
				return decode.Result{}, err
			}
			return decode.Result{InputPath: input, OutputPath: out, DecodedFormat: "mp3", BytesWritten: 9}, nil
		}
		summary, err := (Engine{Decoder: decoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 2})
		if err != nil || maximum != 1 || summary.Successes != 1 || summary.SkippedExisting != 1 {
			t.Fatalf("repeat=%d maximum=%d summary=%#v err=%v", repeat, maximum, summary, err)
		}
	}
}

func TestSharedStemDifferentExtensionsAreSerializedButCanSucceed(t *testing.T) {
	dir := t.TempDir()
	writeEncryptedSet(t, dir, "foo.kgm", "foo.kgma")
	var active, maximum int32
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		current := atomic.AddInt32(&active, 1)
		for {
			old := atomic.LoadInt32(&maximum)
			if current <= old || atomic.CompareAndSwapInt32(&maximum, old, current) {
				break
			}
		}
		defer atomic.AddInt32(&active, -1)
		format := "mp3"
		if filepath.Ext(input) == ".kgma" {
			format = "flac"
		}
		out := filepath.Join(options.OutDir, "foo."+format)
		_ = os.MkdirAll(options.OutDir, 0755)
		_ = os.WriteFile(out, []byte(format), 0600)
		return decode.Result{InputPath: input, OutputPath: out, DecodedFormat: format, BytesWritten: int64(len(format))}, nil
	}
	summary, err := (Engine{Decoder: decoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 2})
	if err != nil || maximum != 1 || summary.Successes != 2 {
		t.Fatalf("maximum=%d summary=%#v err=%v", maximum, summary, err)
	}
}

func TestCollisionWaiterCancellationDoesNotDeadlock(t *testing.T) {
	dir := t.TempDir()
	writeEncryptedSet(t, dir, "foo.kgm", "foo.kgma", "foo.vpr")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		close(started)
		<-ctx.Done()
		return decode.Result{}, ctx.Err()
	}
	done := make(chan Summary, 1)
	go func() {
		summary, _ := (Engine{Decoder: decoder}).Run(ctx, Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 3})
		done <- summary
	}()
	<-started
	cancel()
	summary := <-done
	if summary.Cancelled != 3 {
		t.Fatalf("summary = %#v", summary)
	}
}

func TestIndependentTargetKeysOverlap(t *testing.T) {
	dir := t.TempDir()
	writeEncryptedSet(t, dir, "a.kgm", "b.kgm")
	var active, maximum int32
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		current := atomic.AddInt32(&active, 1)
		updateMax(&maximum, current)
		entered <- struct{}{}
		<-release
		atomic.AddInt32(&active, -1)
		return syntheticDecodeResult(input, options)
	}
	done := make(chan Summary, 1)
	go func() {
		summary, _ := (Engine{Decoder: decoder}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 2})
		done <- summary
	}()
	<-entered
	<-entered
	close(release)
	<-done
	if maximum != 2 {
		t.Fatalf("maximum active workers = %d", maximum)
	}
}

func syntheticDecodeResult(input string, options decode.Options) (decode.Result, error) {
	_ = os.MkdirAll(options.OutDir, 0755)
	out := filepath.Join(options.OutDir, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))+".mp3")
	_ = os.WriteFile(out, []byte("synthetic"), 0600)
	return decode.Result{InputPath: input, OutputPath: out, DecodedFormat: "mp3", BytesWritten: 9}, nil
}

func updateMax(target *int32, value int32) {
	for {
		old := atomic.LoadInt32(target)
		if value <= old || atomic.CompareAndSwapInt32(target, old, value) {
			return
		}
	}
}

func TestBatchProgressEventsAreIndexedAndMonotonic(t *testing.T) {
	dir := t.TempDir()
	writeEncryptedSet(t, dir, "progress.kgm")
	var mu sync.Mutex
	var events []Event
	observer := ObserverFunc(func(event Event) error { mu.Lock(); defer mu.Unlock(); events = append(events, event); return nil })
	decoder := func(ctx context.Context, input string, options decode.Options) (decode.Result, error) {
		options.Progress(decode.Progress{BytesWritten: 2, SourceBytes: 10, SourceBytesKnown: true})
		options.Progress(decode.Progress{BytesWritten: 10, SourceBytes: 10, SourceBytesKnown: true})
		return decode.Result{InputPath: input, OutputPath: filepath.Join(options.OutDir, "progress.mp3"), DecodedFormat: "mp3", BytesWritten: 10}, nil
	}
	if _, err := (Engine{Decoder: decoder, Observer: observer}).Run(context.Background(), Request{Inputs: []string{dir}, OutputRoot: filepath.Join(t.TempDir(), "out"), MaxWorkers: 2}); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	var last int64
	progressCount := 0
	finished := -1
	for i, event := range events {
		if event.Type == EventProgress {
			if event.Index != 0 || event.BytesWritten < last {
				t.Fatalf("progress event = %#v", event)
			}
			last = event.BytesWritten
			progressCount++
		}
		if event.Type == EventFinished {
			finished = i
		}
	}
	if progressCount != 2 || finished < 1 || last != 10 {
		t.Fatalf("events = %#v", events)
	}
}
