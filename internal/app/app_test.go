package app

import (
	"context"
	"errors"
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
