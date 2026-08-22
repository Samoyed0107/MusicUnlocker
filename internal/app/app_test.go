package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"musicunlocker/internal/batch"
	"musicunlocker/internal/naming"
)

type fakeSink struct {
	mu       sync.Mutex
	events   []string
	payloads []any
}

func (s *fakeSink) Emit(name string, payload any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, name)
	s.payloads = append(s.payloads, payload)
}
func (s *fakeSink) count(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, e := range s.events {
		if e == name {
			n++
		}
	}
	return n
}

type fakeEngine struct {
	started  chan struct{}
	release  chan struct{}
	observer batch.Observer
	mu       sync.Mutex
	calls    int
	requests []batch.Request
}

func (e *fakeEngine) Run(ctx context.Context, req batch.Request) (batch.Summary, error) {
	e.mu.Lock()
	e.calls++
	e.requests = append(e.requests, req)
	e.mu.Unlock()
	if e.started != nil {
		close(e.started)
		e.started = nil
	}
	if e.observer != nil {
		_ = e.observer.Observe(batch.Event{Type: batch.EventDiscovered, Index: 0, Result: batch.Result{Index: 0, SourcePath: `C:\song.kgm`}})
	}
	if e.release != nil {
		select {
		case <-e.release:
		case <-ctx.Done():
			return batch.Summary{Cancelled: 1}, nil
		}
	}
	return batch.Summary{Results: []batch.Result{{Index: 0, SourcePath: `C:\song.kgm`, Status: batch.StatusSuccess}}}, nil
}
func (e *fakeEngine) RunWithObserver(ctx context.Context, req batch.Request, observer batch.Observer) (batch.Summary, error) {
	e.observer = observer
	return e.Run(ctx, req)
}
func validRequest() BatchStartRequest {
	return BatchStartRequest{Inputs: []string{`C:\song.kgm`}, OutputRoot: `C:\out`, MaxWorkers: 1}
}
func waitComplete(s *fakeSink, expected int) {
	deadline := time.After(time.Second)
	for {
		if s.count(EventComplete) == expected {
			return
		}
		select {
		case <-deadline:
			panic("timeout waiting for complete")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
func TestApplicationLifecycleAndSingleActive(t *testing.T) {
	e := &fakeEngine{started: make(chan struct{}), release: make(chan struct{})}
	sink := &fakeSink{}
	a := New(e, sink, nil)
	got, err := a.StartBatch(validRequest())
	if err != nil || got.BatchID != 1 {
		t.Fatalf("start: %#v %v", got, err)
	}
	<-e.started
	if _, err := a.StartBatch(validRequest()); !errors.Is(err, ErrBatchAlreadyRunning) {
		t.Fatalf("second start error=%v", err)
	}
	if err := a.CancelBatch(); err != nil {
		t.Fatal(err)
	}
	waitComplete(sink, 1)
	if _, err := a.StartBatch(validRequest()); err != nil {
		t.Fatal(err)
	}
	close(e.release)
	waitComplete(sink, 2)
	if sink.count(EventComplete) != 2 {
		t.Fatalf("complete count=%d", sink.count(EventComplete))
	}
}
func TestInvalidWorkers(t *testing.T) {
	a := New(&fakeEngine{}, &fakeSink{}, nil)
	for _, n := range []int{0, 9, -1} {
		if _, err := a.StartBatch(BatchStartRequest{Inputs: []string{"x"}, OutputRoot: "out", MaxWorkers: n}); err == nil {
			t.Fatalf("workers %d accepted", n)
		}
	}
}

func TestLaunchActivationQueuesPathsWithoutStartingBatch(t *testing.T) {
	engine := &fakeEngine{}
	sink := &fakeSink{}
	a := New(engine, sink, nil)

	a.ReceiveLaunchPaths(LaunchPathsEvent{Paths: []string{`C:\音乐\Artist - 曲名.ncm`}, Source: LaunchSourceInitial})
	pending := a.GetPendingLaunchPaths()
	if pending == nil || len(pending.Paths) != 1 || pending.Paths[0] != `C:\音乐\Artist - 曲名.ncm` {
		t.Fatalf("pending launch paths = %#v", pending)
	}
	a.ReceiveLaunchPaths(LaunchPathsEvent{Paths: []string{`C:\Music\Second Song.kgm`}, Source: LaunchSourceSecond})
	if sink.count(EventLaunchPaths) != 1 {
		t.Fatalf("launch event count = %d", sink.count(EventLaunchPaths))
	}
	engine.mu.Lock()
	calls := engine.calls
	engine.mu.Unlock()
	if calls != 0 {
		t.Fatalf("launch activation started %d batch calls", calls)
	}
}

func TestNamingValidationExamplesAndErrors(t *testing.T) {
	for _, tc := range []struct{ mode, template, example string }{
		{"ARTIST_TITLE", "", "Artist - Song Title.mp3"},
		{"TITLE_ARTIST", "", "Song Title - Artist.mp3"},
		{"TRACK_TITLE", "", "05 - Song Title.mp3"},
		{"CUSTOM", "{artist} - {title}", "Artist - Song Title.mp3"},
	} {
		got := ValidateNamingConfig(NamingConfigDTO{Mode: tc.mode, CustomTemplate: tc.template})
		if !got.Valid || got.Example != tc.example {
			t.Fatalf("%s: %+v", tc.mode, got)
		}
	}
	if got := ValidateNamingConfig(NamingConfigDTO{Mode: "CUSTOM", CustomTemplate: "{unknown}"}); got.Valid || got.Error == "" {
		t.Fatalf("invalid custom accepted: %+v", got)
	}
	if got := ValidateNamingConfig(NamingConfigDTO{}); !got.Valid || got.Example != "source-file.mp3" {
		t.Fatalf("omitted config: %+v", got)
	}
}

func TestBatchCapturesNamingConfig(t *testing.T) {
	e := &fakeEngine{}
	a := New(e, &fakeSink{}, nil)
	request := validRequest()
	request.Naming = &NamingConfigDTO{Mode: "CUSTOM", CustomTemplate: "{title}"}
	if _, err := a.StartBatch(request); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	for {
		e.mu.Lock()
		if len(e.requests) == 1 {
			got := e.requests[0].Naming
			e.mu.Unlock()
			if got.Mode != naming.Custom || got.CustomTemplate != "{title}" {
				t.Fatalf("naming=%+v", got)
			}
			return
		}
		e.mu.Unlock()
		select {
		case <-deadline:
			t.Fatal("timeout waiting for request")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

type emptyDialogs struct{}

func (emptyDialogs) SelectFiles() ([]string, error)      { return nil, nil }
func (emptyDialogs) SelectFolder() (string, error)       { return "", nil }
func (emptyDialogs) SelectOutputFolder() (string, error) { return "", nil }

type retryDialogs struct {
	emptyDialogs
	keyPath string
	err     error
}

func (d retryDialogs) SelectQMCExternalKeyFile() (string, error) { return d.keyPath, d.err }

type kggDialogs struct {
	emptyDialogs
	paths []string
	calls int
}

func (d *kggDialogs) SelectKGGDatabaseFile() (string, error) {
	if d.calls >= len(d.paths) {
		return "", nil
	}
	path := d.paths[d.calls]
	d.calls++
	return path, nil
}

func (d *kggDialogs) SelectQMCExternalKeyFile() (string, error) { return "", nil }

type kggRetryEngine struct {
	mu       sync.Mutex
	calls    int
	paths    []string
	database []string
	items    []batch.Result
}

func (e *kggRetryEngine) Run(ctx context.Context, req batch.Request) (batch.Summary, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls++
	if len(req.Inputs) == 1 && req.KGGDatabaseProvider != nil {
		db, ok, err := req.KGGDatabaseProvider.ResolveKGGDatabase(ctx, req.Inputs[0])
		if err != nil || !ok {
			return batch.Summary{}, errors.New("database provider unavailable")
		}
		e.paths = append(e.paths, req.Inputs[0])
		e.database = append(e.database, db.Path)
		return batch.Summary{Results: []batch.Result{{Index: 0, SourcePath: req.Inputs[0], Status: batch.StatusSuccess, OutputPath: filepath.Join(req.OutputRoot, filepath.Base(req.Inputs[0])+".mp3"), BytesWritten: 9}}}, nil
	}
	return batch.Summary{Results: append([]batch.Result(nil), e.items...)}, nil
}

func (e *kggRetryEngine) RunWithObserver(ctx context.Context, req batch.Request, _ batch.Observer) (batch.Summary, error) {
	return e.Run(ctx, req)
}

type retryEngine struct {
	mu       sync.Mutex
	calls    int
	paths    []string
	material []string
	initial  batch.Status
}

func (e *retryEngine) Run(_ context.Context, req batch.Request) (batch.Summary, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls++
	if len(req.Inputs) != 1 {
		return batch.Summary{}, errors.New("unexpected retry input count")
	}
	e.paths = append(e.paths, req.Inputs[0])
	if req.ExternalMaterialProvider != nil {
		key, ok, err := req.ExternalMaterialProvider.ResolveQMCExternalKey(context.Background(), req.Inputs[0])
		if err != nil || !ok {
			return batch.Summary{}, errors.New("retry provider unavailable")
		}
		e.material = append(e.material, string(key))
	}
	if e.calls == 1 {
		status := e.initial
		if status == "" {
			status = batch.StatusBlockedExternalKey
		}
		code := "QMC_EXTERNAL_KEY_REQUIRED"
		if status != batch.StatusBlockedExternalKey {
			code = "PROCESSING_ERROR"
		}
		return batch.Summary{Results: []batch.Result{{Index: 0, SourcePath: req.Inputs[0], Status: status, ErrorCode: code, ErrorDetail: "requires key"}}}, nil
	}
	return batch.Summary{Results: []batch.Result{{Index: 0, SourcePath: req.Inputs[0], Status: batch.StatusSuccess, OutputPath: `C:\out\song.mp3`, DecodedAudioFormat: "mp3", BytesWritten: 12}}}, nil
}
func (e *retryEngine) RunWithObserver(ctx context.Context, req batch.Request, _ batch.Observer) (batch.Summary, error) {
	return e.Run(ctx, req)
}

func latestComplete(s *fakeSink) BatchCompleteEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.events) - 1; i >= 0; i-- {
		if s.events[i] == EventComplete {
			return s.payloads[i].(BatchCompleteEvent)
		}
	}
	return BatchCompleteEvent{}
}

func TestRetryCancelAndMalformedFileDoNotDecode(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.key")
	if err := os.WriteFile(bad, []byte("not base64"), 0600); err != nil {
		t.Fatal(err)
	}
	for name, keyPath := range map[string]string{"cancel": "", "malformed": bad} {
		e := &retryEngine{}
		sink := &fakeSink{}
		a := New(e, sink, retryDialogs{keyPath: keyPath})
		id, err := a.StartBatch(BatchStartRequest{Inputs: []string{`C:\song.qmc0`}, OutputRoot: `C:\out`, MaxWorkers: 1})
		if err != nil {
			t.Fatal(err)
		}
		waitComplete(sink, 1)
		err = a.RetryQMCWithExternalKey(id.BatchID, 0)
		if name == "cancel" && err != nil {
			t.Fatalf("cancel err=%v", err)
		}
		if name == "malformed" && (err == nil || !strings.Contains(err.Error(), "QMC_EXTERNAL_KEY_FILE_INVALID")) {
			t.Fatalf("malformed err=%v", err)
		}
		e.mu.Lock()
		calls := e.calls
		e.mu.Unlock()
		if calls != 1 {
			t.Fatalf("%s decode calls=%d", name, calls)
		}
	}
}

func TestRetryReplacesStableRowAndScopesMaterial(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "valid.key")
	if err := os.WriteFile(keyPath, []byte("\xef\xbb\xbfQUJDREVGR0hJSktMTU5PUA==\n"), 0600); err != nil {
		t.Fatal(err)
	}
	e := &retryEngine{}
	sink := &fakeSink{}
	a := New(e, sink, retryDialogs{keyPath: keyPath})
	id, err := a.StartBatch(BatchStartRequest{Inputs: []string{`C:\song.qmc0`}, OutputRoot: `C:\out`, MaxWorkers: 1})
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(sink, 1)
	if err := a.RetryQMCWithExternalKey(id.BatchID, 0); err != nil {
		t.Fatal(err)
	}
	complete := latestComplete(sink)
	if complete.Snapshot.Successes != 1 || complete.Snapshot.BlockedExternalKey != 0 || len(complete.Snapshot.Results) != 1 {
		t.Fatalf("snapshot=%+v", complete.Snapshot)
	}
	item := complete.Snapshot.Results[0]
	if item.Index != 0 || item.Status != string(batch.StatusSuccess) || item.OutputPath != `C:\out\song.mp3` {
		t.Fatalf("item=%+v", item)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.material) != 1 || e.material[0] != "QUJDREVGR0hJSktMTU5PUA==" || e.paths[1] != `C:\song.qmc0` {
		t.Fatalf("retry calls paths=%v material=%v", e.paths, e.material)
	}
}

func TestRetryEligibilityAndStaleBatch(t *testing.T) {
	e := &retryEngine{}
	sink := &fakeSink{}
	a := New(e, sink, retryDialogs{})
	if err := a.RetryQMCWithExternalKey(99, 0); err == nil || !strings.Contains(err.Error(), "STALE_BATCH") {
		t.Fatalf("stale error=%v", err)
	}
	id, err := a.StartBatch(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(sink, 1)
	if err := a.RetryQMCWithExternalKey(id.BatchID, 8); err == nil || !strings.Contains(err.Error(), "RETRY_NOT_ELIGIBLE") {
		t.Fatalf("ineligible error=%v", err)
	}
}

func TestRetryRejectsSuccessAndUnrelatedFailure(t *testing.T) {
	for _, status := range []batch.Status{batch.StatusSuccess, batch.StatusFailed} {
		e := &retryEngine{initial: status}
		sink := &fakeSink{}
		a := New(e, sink, retryDialogs{})
		id, err := a.StartBatch(validRequest())
		if err != nil {
			t.Fatal(err)
		}
		waitComplete(sink, 1)
		if err := a.RetryQMCWithExternalKey(id.BatchID, 0); err == nil || !strings.Contains(err.Error(), "RETRY_NOT_ELIGIBLE") {
			t.Fatalf("status=%s error=%v", status, err)
		}
	}
}

func TestKGGRetryCancellationAndInvalidDatabaseDoNotDecode(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.db")
	if err := os.WriteFile(valid, []byte("SQLite format 3\x00synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	for name, selected := range map[string]string{"cancel": "", "invalid": filepath.Join(dir, "missing.db")} {
		e := &kggRetryEngine{items: []batch.Result{{Index: 0, SourcePath: `C:\song.kgm`, Status: batch.StatusBlockedExternalDatabase, ErrorCode: "KGG_DATABASE_REQUIRED"}}}
		d := &kggDialogs{paths: []string{selected}}
		sink := &fakeSink{}
		a := New(e, sink, d)
		id, err := a.StartBatch(validRequest())
		if err != nil {
			t.Fatal(err)
		}
		waitComplete(sink, 1)
		err = a.RetryKGGWithDatabase(id.BatchID, 0)
		if name == "cancel" && err != nil {
			t.Fatalf("cancel err=%v", err)
		}
		if name == "invalid" && (err == nil || !strings.Contains(err.Error(), "KGG_DATABASE_INVALID")) {
			t.Fatalf("invalid err=%v", err)
		}
		e.mu.Lock()
		calls := e.calls
		e.mu.Unlock()
		if calls != 1 {
			t.Fatalf("%s decode calls=%d", name, calls)
		}
	}
}

func TestKGGRetryUsesExactDatabaseAndReplacesSummary(t *testing.T) {
	dir := t.TempDir()
	dbA, dbB := filepath.Join(dir, "a.db"), filepath.Join(dir, "b.db")
	for _, path := range []string{dbA, dbB} {
		if err := os.WriteFile(path, []byte("SQLite format 3\x00synthetic"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	e := &kggRetryEngine{items: []batch.Result{
		{Index: 0, SourcePath: `C:\a.kgm`, Status: batch.StatusBlockedExternalDatabase, ErrorCode: "KGG_DATABASE_REQUIRED"},
		{Index: 1, SourcePath: `C:\b.kgm`, Status: batch.StatusBlockedExternalDatabase, ErrorCode: "KGG_DATABASE_REQUIRED"},
	}}
	d := &kggDialogs{paths: []string{dbA, dbB}}
	sink := &fakeSink{}
	a := New(e, sink, d)
	id, err := a.StartBatch(BatchStartRequest{Inputs: []string{`C:\a.kgm`, `C:\b.kgm`}, OutputRoot: `C:\out`, MaxWorkers: 1})
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(sink, 1)
	if err := a.RetryKGGWithDatabase(id.BatchID, 0); err != nil {
		t.Fatal(err)
	}
	if err := a.RetryKGGWithDatabase(id.BatchID, 1); err != nil {
		t.Fatal(err)
	}
	complete := latestComplete(sink)
	if complete.Snapshot.Successes != 2 || complete.Snapshot.BlockedExternalDatabase != 0 || len(complete.Snapshot.Results) != 2 {
		t.Fatalf("snapshot=%+v", complete.Snapshot)
	}
	if complete.Snapshot.Results[0].Index != 0 || complete.Snapshot.Results[1].Index != 1 {
		t.Fatalf("results=%+v", complete.Snapshot.Results)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.database) != 2 || e.database[0] != dbA || e.database[1] != dbB || e.paths[0] != `C:\a.kgm` || e.paths[1] != `C:\b.kgm` {
		t.Fatalf("paths=%v databases=%v", e.paths, e.database)
	}
}

func TestKGGRetryRejectsWrongMaterialRows(t *testing.T) {
	qmcEngine := &retryEngine{}
	qmcSink := &fakeSink{}
	qmcApp := New(qmcEngine, qmcSink, &kggDialogs{})
	qmcID, err := qmcApp.StartBatch(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(qmcSink, 1)
	if err := qmcApp.RetryKGGWithDatabase(qmcID.BatchID, 0); err == nil || !strings.Contains(err.Error(), "RETRY_NOT_ELIGIBLE") {
		t.Fatalf("QMC row accepted by KGG retry: %v", err)
	}
	kggEngine := &kggRetryEngine{items: []batch.Result{{Index: 0, SourcePath: `C:\song.kgm`, Status: batch.StatusBlockedExternalDatabase, ErrorCode: "KGG_DATABASE_REQUIRED"}}}
	kggSink := &fakeSink{}
	kggApp := New(kggEngine, kggSink, &kggDialogs{})
	kggID, err := kggApp.StartBatch(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(kggSink, 1)
	if err := kggApp.RetryQMCWithExternalKey(kggID.BatchID, 0); err == nil || !strings.Contains(err.Error(), "RETRY_NOT_ELIGIBLE") {
		t.Fatalf("KGG row accepted by QMC retry: %v", err)
	}
}

func TestDialogCancellationIsEmptyAndNotAnError(t *testing.T) {
	a := New(&fakeEngine{}, &fakeSink{}, emptyDialogs{})
	if files, err := a.SelectFiles(); err != nil || len(files) != 0 {
		t.Fatalf("files=%v err=%v", files, err)
	}
	if folder, err := a.SelectFolder(); err != nil || folder != "" {
		t.Fatalf("folder=%q err=%v", folder, err)
	}
	if folder, err := a.SelectOutputFolder(); err != nil || folder != "" {
		t.Fatalf("output=%q err=%v", folder, err)
	}
}

func TestEventsCarryMappedItemsAndIncreasingBatchIDs(t *testing.T) {
	e := &fakeEngine{}
	sink := &fakeSink{}
	a := New(e, sink, nil)
	first, err := a.StartBatch(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(sink, 1)
	second, err := a.StartBatch(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	waitComplete(sink, 2)
	if first.BatchID != 1 || second.BatchID != 2 {
		t.Fatalf("batch IDs=%d,%d", first.BatchID, second.BatchID)
	}
	sink.mu.Lock()
	defer sink.mu.Unlock()
	for i, name := range sink.events {
		if name != EventDiscovered {
			continue
		}
		payload, ok := sink.payloads[i].(struct {
			BatchID uint64    `json:"batchId"`
			Item    QueueItem `json:"item"`
		})
		if !ok || payload.Item.SourceDisplayName != "song.kgm" || payload.Item.Index != 0 {
			t.Fatalf("unexpected discovered payload %#v", sink.payloads[i])
		}
		return
	}
	t.Fatal("no discovered event")
}
