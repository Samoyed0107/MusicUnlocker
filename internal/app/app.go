// Package app contains the Wails-facing application facade. It owns GUI
// lifecycle state while leaving discovery, scheduling, and decoding in core.
package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"musicunlocker/internal/batch"
)

const (
	EventDiscovered  = "musicunlocker:batch:discovered"
	EventStarted     = "musicunlocker:batch:started"
	EventProgress    = "musicunlocker:batch:progress"
	EventFinished    = "musicunlocker:batch:finished"
	EventComplete    = "musicunlocker:batch:complete"
	EventLaunchPaths = "musicunlocker:launch:paths"
)

var (
	ErrBatchAlreadyRunning = &ApplicationError{Code: "BATCH_ALREADY_RUNNING", Detail: "another batch is already running"}
	ErrNoActiveBatch       = &ApplicationError{Code: "NO_ACTIVE_BATCH", Detail: "there is no active batch"}
)

type ApplicationError struct{ Code, Detail string }

func (e *ApplicationError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }

type BatchStartRequest struct {
	Inputs     []string `json:"inputs"`
	OutputRoot string   `json:"outputRoot"`
	MaxWorkers int      `json:"maxWorkers"`
}

type BatchStartResponse struct {
	BatchID uint64 `json:"batchId"`
}

type AppInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	MaxWorkers    int    `json:"maxWorkers"`
	SupportsDrop  bool   `json:"supportsDrop"`
	SupportsPause bool   `json:"supportsPause"`
}

type QueueItem struct {
	Index             int    `json:"index"`
	SourcePath        string `json:"sourcePath"`
	SourceDisplayName string `json:"sourceDisplayName"`
	RelativePath      string `json:"relativePath,omitempty"`
	OutputPath        string `json:"outputPath,omitempty"`
	Family            string `json:"family,omitempty"`
	Variant           string `json:"variant,omitempty"`
	DecodedFormat     string `json:"decodedFormat,omitempty"`
	SourceBytes       int64  `json:"sourceBytes"`
	SourceBytesKnown  bool   `json:"sourceBytesKnown"`
	BytesWritten      int64  `json:"bytesWritten"`
	Status            string `json:"status"`
	ErrorCode         string `json:"errorCode,omitempty"`
	ErrorDetail       string `json:"errorDetail,omitempty"`
}

type ProgressEvent struct {
	BatchID          uint64    `json:"batchId"`
	Item             QueueItem `json:"item"`
	BytesWritten     int64     `json:"bytesWritten"`
	SourceBytes      int64     `json:"sourceBytes"`
	SourceBytesKnown bool      `json:"sourceBytesKnown"`
}

type BatchSnapshot struct {
	DiscoveredRegularFiles  int         `json:"discoveredRegularFiles"`
	EligibleEncrypted       int         `json:"eligibleEncryptedCandidates"`
	Successes               int         `json:"successes"`
	SkippedNotEncrypted     int         `json:"skippedNotEncrypted"`
	SkippedExisting         int         `json:"skippedExisting"`
	BlockedExternalKey      int         `json:"blockedExternalKey"`
	BlockedExternalDatabase int         `json:"blockedExternalDatabase"`
	Unsupported             int         `json:"unsupported"`
	Unknown                 int         `json:"unknown"`
	Failed                  int         `json:"failed"`
	Cancelled               int         `json:"cancelled"`
	BytesWritten            int64       `json:"bytesWritten"`
	Duration                int64       `json:"durationNanoseconds"`
	Results                 []QueueItem `json:"results"`
}

type BatchCompleteEvent struct {
	BatchID     uint64        `json:"batchId"`
	Cancelled   bool          `json:"cancelled"`
	Snapshot    BatchSnapshot `json:"snapshot"`
	ErrorCode   string        `json:"errorCode,omitempty"`
	ErrorDetail string        `json:"errorDetail,omitempty"`
}

type EventSink interface {
	Emit(name string, payload any)
}
type Dialogs interface {
	SelectFiles() ([]string, error)
	SelectFolder() (string, error)
	SelectOutputFolder() (string, error)
}
type EngineRunner interface {
	Run(context.Context, batch.Request) (batch.Summary, error)
}
type ObserverEngine interface {
	RunWithObserver(context.Context, batch.Request, batch.Observer) (batch.Summary, error)
}

type Application struct {
	engine      EngineRunner
	sink        EventSink
	dialogs     Dialogs
	preferences PreferencesStore
	opener      FolderOpener
	mu          sync.Mutex
	running     bool
	activeID    uint64
	nextID      uint64
	cancel      context.CancelFunc
	preview     *PreviewService
	previewErr  error
	latestID    uint64
	latest      *BatchSnapshot
	launchInbox LaunchInbox
}

func (a *Application) ReceiveLaunchPaths(event LaunchPathsEvent) {
	if len(event.Paths) == 0 {
		return
	}
	if event.RequestID == "" {
		event = a.launchInbox.NewEvent(event.Paths, event.Source)
	}
	a.mu.Lock()
	running := a.running
	a.mu.Unlock()
	if event.Source == LaunchSourceInitial {
		a.launchInbox.RetainInitial(event)
		return
	}
	if running {
		a.launchInbox.Defer(event)
		event.Deferred = true
		a.sink.Emit(EventLaunchPaths, event)
		return
	}
	a.sink.Emit(EventLaunchPaths, event)
}

func (a *Application) GetPendingLaunchPaths() *LaunchPathsEvent {
	return a.launchInbox.ConsumeInitial()
}

func New(engine EngineRunner, sink EventSink, dialogs Dialogs) *Application {
	return NewWithDependencies(engine, sink, dialogs, newFilePreferencesStore(), nil)
}

func NewWithDependencies(engine EngineRunner, sink EventSink, dialogs Dialogs, preferences PreferencesStore, opener FolderOpener) *Application {
	preview := NewPreviewService()
	return &Application{engine: engine, sink: sink, dialogs: dialogs, preferences: preferences, opener: opener, preview: preview, previewErr: preview.startError}
}

// NewWithPreview is useful for application tests and controlled hosts.
func NewWithPreview(engine EngineRunner, sink EventSink, dialogs Dialogs, preferences PreferencesStore, opener FolderOpener, preview *PreviewService) *Application {
	return &Application{engine: engine, sink: sink, dialogs: dialogs, preferences: preferences, opener: opener, preview: preview}
}

func (a *Application) GetAppInfo() AppInfo {
	return AppInfo{Name: "MusicUnlocker", Version: "Stage 5B", MaxWorkers: 8, SupportsPause: false, SupportsDrop: true}
}

func (a *Application) SelectFiles() ([]string, error) {
	if a.dialogs == nil {
		return nil, errors.New("dialogs unavailable")
	}
	return a.dialogs.SelectFiles()
}
func (a *Application) SelectFolder() (string, error) {
	if a.dialogs == nil {
		return "", errors.New("dialogs unavailable")
	}
	return a.dialogs.SelectFolder()
}
func (a *Application) SelectOutputFolder() (string, error) {
	if a.dialogs == nil {
		return "", errors.New("dialogs unavailable")
	}
	return a.dialogs.SelectOutputFolder()
}

func (a *Application) SuggestOutputFolder(inputs []string) (string, error) {
	return SuggestOutputFolder(inputs)
}

func (a *Application) DescribeInputs(inputs []string) []SelectedInput { return DescribeInputs(inputs) }

func (a *Application) GetPreferences() (Preferences, error) {
	if a.preferences == nil {
		return DefaultPreferences(), nil
	}
	p, err := a.preferences.Load()
	return sanitizePreferences(p), err
}

func (a *Application) SavePreferences(p Preferences) error {
	if a.preferences == nil {
		return nil
	}
	return a.preferences.Save(sanitizePreferences(p))
}

func (a *Application) OpenOutputFolder(path string) error {
	return openOutputFolder(path, a.opener)
}

func (a *Application) StartBatch(req BatchStartRequest) (BatchStartResponse, error) {
	if len(req.Inputs) == 0 || filepath.Clean(req.OutputRoot) == "." || req.OutputRoot == "" {
		return BatchStartResponse{}, &ApplicationError{Code: "INVALID_REQUEST", Detail: "at least one input and an output folder are required"}
	}
	if req.MaxWorkers < 1 || req.MaxWorkers > 8 {
		return BatchStartResponse{}, &ApplicationError{Code: "INVALID_MAX_WORKERS", Detail: "maxWorkers must be between 1 and 8"}
	}
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return BatchStartResponse{}, ErrBatchAlreadyRunning
	}
	a.nextID++
	id := a.nextID
	if a.preview != nil {
		a.preview.Invalidate()
	}
	a.latestID, a.latest = 0, nil
	ctx, cancel := context.WithCancel(context.Background())
	a.running, a.activeID, a.cancel = true, id, cancel
	a.mu.Unlock()
	go a.run(ctx, id, req)
	return BatchStartResponse{BatchID: id}, nil
}

func (a *Application) CancelBatch() error {
	a.mu.Lock()
	cancel, running := a.cancel, a.running
	a.mu.Unlock()
	if !running || cancel == nil {
		return ErrNoActiveBatch
	}
	cancel()
	return nil
}

func (a *Application) run(ctx context.Context, id uint64, req BatchStartRequest) {
	observer := batch.ObserverFunc(func(event batch.Event) error { a.emitEvent(id, event); return nil })
	request := batch.Request{Inputs: append([]string(nil), req.Inputs...), OutputRoot: req.OutputRoot, MaxWorkers: req.MaxWorkers}
	engine := &observedEngine{runner: a.engine, observer: observer}
	summary, err := engine.Run(ctx, request)
	complete := BatchCompleteEvent{BatchID: id, Cancelled: errors.Is(ctx.Err(), context.Canceled), Snapshot: snapshotFrom(summary)}
	if err != nil {
		complete.ErrorCode, complete.ErrorDetail = "BATCH_ERROR", err.Error()
	}
	a.mu.Lock()
	if a.activeID == id {
		copySnapshot := complete.Snapshot
		copySnapshot.Results = append([]QueueItem(nil), complete.Snapshot.Results...)
		a.latestID, a.latest = id, &copySnapshot
	}
	a.mu.Unlock()
	a.sink.Emit(EventComplete, complete)
	a.mu.Lock()
	if a.activeID == id {
		a.running, a.cancel = false, nil
	}
	a.mu.Unlock()
	for _, event := range a.launchInbox.FlushDeferred() {
		a.sink.Emit(EventLaunchPaths, event)
	}
}

func (a *Application) GetPreviewURL(batchID uint64, index int) (PreviewResponse, error) {
	a.mu.Lock()
	if a.previewErr != nil || a.preview == nil {
		a.mu.Unlock()
		return PreviewResponse{}, ErrPreviewServerUnavailable
	}
	if a.latestID != batchID || a.latest == nil {
		a.mu.Unlock()
		return PreviewResponse{}, ErrPreviewResultNotFound
	}
	var item *QueueItem
	for i := range a.latest.Results {
		if a.latest.Results[i].Index == index {
			item = &a.latest.Results[i]
			break
		}
	}
	if item == nil {
		a.mu.Unlock()
		return PreviewResponse{}, ErrPreviewResultNotFound
	}
	if item.Status != string(batch.StatusSuccess) || item.OutputPath == "" {
		a.mu.Unlock()
		return PreviewResponse{}, ErrPreviewNotAvailable
	}
	path := item.OutputPath
	a.mu.Unlock()
	token, format, err := a.preview.Register(path)
	if err != nil {
		return PreviewResponse{}, err
	}
	url := a.preview.URL(token)
	if url == "" {
		return PreviewResponse{}, ErrPreviewResultNotFound
	}
	return PreviewResponse{Available: true, URL: url, Format: format}, nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	if a.preview == nil {
		return nil
	}
	return a.preview.Shutdown(ctx)
}

type observedEngine struct {
	runner   EngineRunner
	observer batch.Observer
}

func (e *observedEngine) Run(ctx context.Context, req batch.Request) (batch.Summary, error) {
	if r, ok := e.runner.(ObserverEngine); ok {
		return r.RunWithObserver(ctx, req, e.observer)
	}
	return e.runner.Run(ctx, req)
}

func (a *Application) emitEvent(id uint64, event batch.Event) {
	item := itemFrom(event.Result)
	switch event.Type {
	case batch.EventDiscovered:
		a.sink.Emit(EventDiscovered, struct {
			BatchID uint64    `json:"batchId"`
			Item    QueueItem `json:"item"`
		}{id, item})
	case batch.EventStarted:
		a.sink.Emit(EventStarted, struct {
			BatchID uint64    `json:"batchId"`
			Item    QueueItem `json:"item"`
		}{id, item})
	case batch.EventProgress:
		a.sink.Emit(EventProgress, ProgressEvent{BatchID: id, Item: item, BytesWritten: event.BytesWritten, SourceBytes: event.SourceBytes, SourceBytesKnown: event.SourceBytesKnown})
	case batch.EventFinished:
		a.sink.Emit(EventFinished, struct {
			BatchID uint64    `json:"batchId"`
			Item    QueueItem `json:"item"`
		}{id, item})
	}
}

func itemFrom(r batch.Result) QueueItem {
	return QueueItem{Index: r.Index, SourcePath: r.SourcePath, SourceDisplayName: filepath.Base(r.SourcePath), RelativePath: r.RelativePath, OutputPath: r.OutputPath, Family: r.DetectedFamily, Variant: r.Variant, DecodedFormat: r.DecodedAudioFormat, SourceBytes: r.SourceBytes, SourceBytesKnown: r.SourceBytesKnown, BytesWritten: r.BytesWritten, Status: string(r.Status), ErrorCode: r.ErrorCode, ErrorDetail: r.ErrorDetail}
}
func snapshotFrom(s batch.Summary) BatchSnapshot {
	out := BatchSnapshot{DiscoveredRegularFiles: s.DiscoveredRegularFiles, EligibleEncrypted: s.EligibleEncrypted, Successes: s.Successes, SkippedNotEncrypted: s.SkippedNotEncrypted, SkippedExisting: s.SkippedExisting, BlockedExternalKey: s.BlockedExternalKey, BlockedExternalDatabase: s.BlockedExternalDatabase, Unsupported: s.Unsupported, Unknown: s.Unknown, Failed: s.Failed, Cancelled: s.Cancelled, BytesWritten: s.BytesWritten, Duration: int64(s.Duration), Results: make([]QueueItem, len(s.Results))}
	for i, r := range s.Results {
		out.Results[i] = itemFrom(r)
	}
	return out
}
