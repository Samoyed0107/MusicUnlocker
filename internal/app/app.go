// Package app contains the Wails-facing application facade. It owns GUI
// lifecycle state while leaving discovery, scheduling, and decoding in core.
package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"musicunlocker/internal/audiometa"
	"musicunlocker/internal/batch"
	"musicunlocker/internal/externalmaterial"
	"musicunlocker/internal/naming"
	"musicunlocker/internal/version"
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
	Inputs     []string         `json:"inputs"`
	OutputRoot string           `json:"outputRoot"`
	MaxWorkers int              `json:"maxWorkers"`
	Naming     *NamingConfigDTO `json:"naming,omitempty"`
}

type NamingConfigDTO struct {
	Mode           string `json:"mode"`
	CustomTemplate string `json:"customTemplate,omitempty"`
}

type NamingValidation struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Example string `json:"example,omitempty"`
}

type BatchStartResponse struct {
	BatchID uint64 `json:"batchId"`
}

type AppInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	BuildTimestamp  string `json:"buildTimestamp"`
	BuildIdentifier string `json:"buildIdentifier"`
	MaxWorkers      int    `json:"maxWorkers"`
	SupportsDrop    bool   `json:"supportsDrop"`
	SupportsPause   bool   `json:"supportsPause"`
}

type QueueItem struct {
	Index                     int    `json:"index"`
	SourcePath                string `json:"sourcePath"`
	SourceDisplayName         string `json:"sourceDisplayName"`
	RelativePath              string `json:"relativePath,omitempty"`
	OutputPath                string `json:"outputPath,omitempty"`
	Family                    string `json:"family,omitempty"`
	Variant                   string `json:"variant,omitempty"`
	DecodedFormat             string `json:"decodedFormat,omitempty"`
	SourceBytes               int64  `json:"sourceBytes"`
	SourceBytesKnown          bool   `json:"sourceBytesKnown"`
	BytesWritten              int64  `json:"bytesWritten"`
	Status                    string `json:"status"`
	ErrorCode                 string `json:"errorCode,omitempty"`
	ErrorDetail               string `json:"errorDetail,omitempty"`
	RetryableExternalKey      bool   `json:"retryableExternalKey,omitempty"`
	RetryableExternalDatabase bool   `json:"retryableExternalDatabase,omitempty"`
	NamingFallback            bool   `json:"namingFallback,omitempty"`
	NamingFallbackReason      string `json:"namingFallbackReason,omitempty"`
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
type EKeyDialog interface{ SelectQMCExternalKeyFile() (string, error) }
type KGGDatabaseDialog interface{ SelectKGGDatabaseFile() (string, error) }
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
	retry       *retrySession
	retrying    bool
	retryCancel context.CancelFunc
	launchInbox LaunchInbox
}

type retrySession struct {
	batchID    uint64
	outputRoot string
	naming     naming.Config
	qmcItems   map[int]bool
	kggItems   map[int]bool
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
	identity := version.Current()
	return AppInfo{Name: identity.Name, Version: identity.Version, BuildTimestamp: identity.BuildTimestamp, BuildIdentifier: identity.BuildIdentifier, MaxWorkers: 8, SupportsPause: false, SupportsDrop: true}
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

func namingConfigFromDTO(dto *NamingConfigDTO) naming.Config {
	if dto == nil || dto.Mode == "" {
		return naming.Config{Mode: naming.KeepSource}
	}
	return naming.Config{Mode: naming.Mode(dto.Mode), CustomTemplate: dto.CustomTemplate}
}

func ValidateNamingConfig(dto NamingConfigDTO) NamingValidation {
	config := namingConfigFromDTO(&dto)
	if err := naming.Validate(config); err != nil {
		return NamingValidation{Error: err.Error()}
	}
	metadata := audiometa.Metadata{Title: "Song Title", Artist: "Artist", Album: "Album", AlbumArtist: "Album Artist", Track: 5, Disc: 1, Year: 2026}
	result, err := naming.Resolve(config, "source-file", metadata, true)
	if err != nil {
		return NamingValidation{Error: err.Error()}
	}
	return NamingValidation{Valid: true, Example: result.Stem + ".mp3"}
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
	namingConfig := namingConfigFromDTO(req.Naming)
	if err := naming.Validate(namingConfig); err != nil {
		return BatchStartResponse{}, &ApplicationError{Code: "INVALID_NAMING_CONFIG", Detail: err.Error()}
	}
	req.Naming = &NamingConfigDTO{Mode: string(namingConfig.Mode), CustomTemplate: namingConfig.CustomTemplate}
	a.mu.Lock()
	if a.running || a.retrying {
		a.mu.Unlock()
		return BatchStartResponse{}, ErrBatchAlreadyRunning
	}
	a.nextID++
	id := a.nextID
	if a.preview != nil {
		a.preview.Invalidate()
	}
	a.latestID, a.latest = 0, nil
	a.retry = nil
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
	request := batch.Request{Inputs: append([]string(nil), req.Inputs...), OutputRoot: req.OutputRoot, MaxWorkers: req.MaxWorkers, Naming: namingConfigFromDTO(req.Naming)}
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
		retry := &retrySession{batchID: id, outputRoot: req.OutputRoot, naming: namingConfigFromDTO(req.Naming), qmcItems: map[int]bool{}, kggItems: map[int]bool{}}
		for _, item := range copySnapshot.Results {
			if item.Status == string(batch.StatusBlockedExternalKey) && item.ErrorCode == string("QMC_EXTERNAL_KEY_REQUIRED") {
				retry.qmcItems[item.Index] = true
				for i := range copySnapshot.Results {
					if copySnapshot.Results[i].Index == item.Index {
						copySnapshot.Results[i].RetryableExternalKey = true
					}
				}
			}
			if item.Status == string(batch.StatusBlockedExternalDatabase) && item.ErrorCode == string("KGG_DATABASE_REQUIRED") {
				retry.kggItems[item.Index] = true
				for i := range copySnapshot.Results {
					if copySnapshot.Results[i].Index == item.Index {
						copySnapshot.Results[i].RetryableExternalDatabase = true
					}
				}
			}
		}
		a.retry = retry
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

// RetryKGGWithDatabase opens the native picker and retries exactly one KGM V5
// or KGG source. The selected path remains backend-owned for this request.
func (a *Application) RetryKGGWithDatabase(batchID uint64, index int) error {
	dialog, ok := a.dialogs.(KGGDatabaseDialog)
	if !ok {
		return &ApplicationError{Code: "DIALOG_UNAVAILABLE", Detail: "KGG database file dialog is unavailable"}
	}
	a.mu.Lock()
	if a.running || a.retrying {
		a.mu.Unlock()
		return ErrBatchAlreadyRunning
	}
	if a.retry == nil || a.retry.batchID != batchID {
		a.mu.Unlock()
		return &ApplicationError{Code: "STALE_BATCH", Detail: "the batch is no longer available for retry"}
	}
	if !a.retry.kggItems[index] {
		a.mu.Unlock()
		return &ApplicationError{Code: "RETRY_NOT_ELIGIBLE", Detail: "this item is not eligible for KGG database retry"}
	}
	var item QueueItem
	for _, candidate := range a.latest.Results {
		if candidate.Index == index {
			item = candidate
			break
		}
	}
	if item.SourcePath == "" {
		a.mu.Unlock()
		return &ApplicationError{Code: "INVALID_ITEM_INDEX", Detail: "the requested batch item does not exist"}
	}
	session := *a.retry
	a.retrying = true
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.retrying = false; a.retryCancel = nil; a.mu.Unlock() }()
	retryCtx, retryCancel := context.WithCancel(context.Background())
	defer retryCancel()
	a.mu.Lock()
	a.retryCancel = retryCancel
	a.mu.Unlock()

	dbPath, err := dialog.SelectKGGDatabaseFile()
	if err != nil {
		return &ApplicationError{Code: "KGG_DATABASE_SELECTION_FAILED", Detail: "the KGG database could not be selected"}
	}
	if dbPath == "" {
		return nil
	}
	if err := externalmaterial.ValidateKGGDatabasePath(dbPath); err != nil {
		return &ApplicationError{Code: "KGG_DATABASE_INVALID", Detail: "the selected KGG database is invalid"}
	}
	provider := externalmaterial.StaticKGGDatabaseProvider{Database: externalmaterial.KGGDatabase{Path: dbPath}}
	summary, runErr := a.engine.Run(retryCtx, batch.Request{Inputs: []string{item.SourcePath}, OutputRoot: session.outputRoot, MaxWorkers: 1, Naming: session.naming, KGGDatabaseProvider: provider})
	retryCancel()
	if runErr != nil {
		return &ApplicationError{Code: "RETRY_FAILED", Detail: "the KGG database retry could not be completed"}
	}
	if len(summary.Results) == 0 {
		return &ApplicationError{Code: "RETRY_FAILED", Detail: "the KGG database retry returned no result"}
	}
	replacement := itemFrom(summary.Results[0])
	replacement.Index = index
	replacement.SourcePath = item.SourcePath
	replacement.RetryableExternalDatabase = replacement.Status != string(batch.StatusSuccess)
	if replacement.Status == string(batch.StatusSuccess) {
		delete(session.kggItems, index)
	} else {
		session.kggItems[index] = true
	}
	a.mu.Lock()
	if a.latestID != batchID || a.latest == nil {
		a.mu.Unlock()
		return &ApplicationError{Code: "STALE_BATCH", Detail: "the batch is no longer available for retry"}
	}
	a.latest = replaceSnapshot(a.latest, replacement)
	a.retry = &session
	snapshot := *a.latest
	snapshot.Results = append([]QueueItem(nil), a.latest.Results...)
	a.mu.Unlock()
	a.sink.Emit(EventFinished, struct {
		BatchID uint64    `json:"batchId"`
		Item    QueueItem `json:"item"`
	}{batchID, replacement})
	a.sink.Emit(EventComplete, BatchCompleteEvent{BatchID: batchID, Snapshot: snapshot})
	return nil
}

// RetryQMCWithExternalKey opens the native picker and retries exactly one
// blocked source using the normal batch/decode pipeline. The key path and
// material never leave the backend.
func (a *Application) RetryQMCWithExternalKey(batchID uint64, index int) error {
	dialog, ok := a.dialogs.(EKeyDialog)
	if !ok {
		return &ApplicationError{Code: "DIALOG_UNAVAILABLE", Detail: "external-key file dialog is unavailable"}
	}
	a.mu.Lock()
	if a.running || a.retrying {
		a.mu.Unlock()
		return ErrBatchAlreadyRunning
	}
	if a.retry == nil || a.retry.batchID != batchID {
		a.mu.Unlock()
		return &ApplicationError{Code: "STALE_BATCH", Detail: "the batch is no longer available for retry"}
	}
	if !a.retry.qmcItems[index] {
		a.mu.Unlock()
		return &ApplicationError{Code: "RETRY_NOT_ELIGIBLE", Detail: "this item is not eligible for external-key retry"}
	}
	var item QueueItem
	for _, candidate := range a.latest.Results {
		if candidate.Index == index {
			item = candidate
			break
		}
	}
	if item.SourcePath == "" {
		a.mu.Unlock()
		return &ApplicationError{Code: "INVALID_ITEM_INDEX", Detail: "the requested batch item does not exist"}
	}
	session := *a.retry
	a.retrying = true
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.retrying = false; a.retryCancel = nil; a.mu.Unlock() }()
	retryCtx, retryCancel := context.WithCancel(context.Background())
	defer retryCancel()
	a.mu.Lock()
	a.retryCancel = retryCancel
	a.mu.Unlock()

	keyPath, err := dialog.SelectQMCExternalKeyFile()
	if err != nil {
		return &ApplicationError{Code: "EKEY_FILE_INVALID", Detail: "the external-key file could not be selected"}
	}
	if keyPath == "" {
		return nil
	}
	key, err := externalmaterial.ReadEKeyFile(keyPath)
	if err != nil {
		return &ApplicationError{Code: "QMC_EXTERNAL_KEY_FILE_INVALID", Detail: err.Error()}
	}
	provider := externalmaterial.NewMapProvider(map[string][]byte{item.SourcePath: key})
	defer func() {
		provider.Clear()
		for i := range key {
			key[i] = 0
		}
	}()

	summary, runErr := a.engine.Run(retryCtx, batch.Request{Inputs: []string{item.SourcePath}, OutputRoot: session.outputRoot, MaxWorkers: 1, Naming: session.naming, ExternalMaterialProvider: provider})
	retryCancel()
	if runErr != nil {
		return &ApplicationError{Code: "RETRY_FAILED", Detail: "the external-key retry could not be completed"}
	}
	if len(summary.Results) == 0 {
		return &ApplicationError{Code: "RETRY_FAILED", Detail: "the external-key retry returned no result"}
	}
	replacement := itemFrom(summary.Results[0])
	replacement.Index = index
	replacement.SourcePath = item.SourcePath
	replacement.RetryableExternalKey = true
	if replacement.Status == string(batch.StatusSuccess) {
		replacement.RetryableExternalKey = false
		delete(session.qmcItems, index)
	} else {
		replacement.ErrorCode = "QMC_EXTERNAL_KEY_ATTEMPT_FAILED"
		replacement.ErrorDetail = "the supplied external QMC key could not unlock this file"
		session.qmcItems[index] = true
	}
	a.mu.Lock()
	if a.latestID != batchID || a.latest == nil {
		a.mu.Unlock()
		return &ApplicationError{Code: "STALE_BATCH", Detail: "the batch is no longer available for retry"}
	}
	a.latest = replaceSnapshot(a.latest, replacement)
	a.retry = &session
	snapshot := *a.latest
	snapshot.Results = append([]QueueItem(nil), a.latest.Results...)
	a.mu.Unlock()
	a.sink.Emit(EventFinished, struct {
		BatchID uint64    `json:"batchId"`
		Item    QueueItem `json:"item"`
	}{batchID, replacement})
	a.sink.Emit(EventComplete, BatchCompleteEvent{BatchID: batchID, Snapshot: snapshot})
	return nil
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
	a.mu.Lock()
	retryCancel := a.retryCancel
	a.mu.Unlock()
	if retryCancel != nil {
		retryCancel()
	}
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
	return QueueItem{Index: r.Index, SourcePath: r.SourcePath, SourceDisplayName: filepath.Base(r.SourcePath), RelativePath: r.RelativePath, OutputPath: r.OutputPath, Family: r.DetectedFamily, Variant: r.Variant, DecodedFormat: r.DecodedAudioFormat, SourceBytes: r.SourceBytes, SourceBytesKnown: r.SourceBytesKnown, BytesWritten: r.BytesWritten, Status: string(r.Status), ErrorCode: r.ErrorCode, ErrorDetail: r.ErrorDetail, NamingFallback: r.NamingFallback, NamingFallbackReason: r.NamingFallbackReason}
}

func replaceSnapshot(snapshot *BatchSnapshot, replacement QueueItem) *BatchSnapshot {
	out := *snapshot
	out.Results = append([]QueueItem(nil), snapshot.Results...)
	for i := range out.Results {
		if out.Results[i].Index == replacement.Index {
			out.Results[i] = replacement
		}
	}
	out.Successes, out.SkippedNotEncrypted, out.SkippedExisting, out.BlockedExternalKey, out.BlockedExternalDatabase, out.Unsupported, out.Unknown, out.Failed, out.Cancelled, out.BytesWritten = 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
	for _, item := range out.Results {
		switch item.Status {
		case string(batch.StatusSuccess):
			out.Successes++
			out.BytesWritten += item.BytesWritten
		case string(batch.StatusSkippedNotEncrypted):
			out.SkippedNotEncrypted++
		case string(batch.StatusSkippedOutputExists):
			out.SkippedExisting++
		case string(batch.StatusBlockedExternalKey):
			out.BlockedExternalKey++
		case string(batch.StatusBlockedExternalDatabase):
			out.BlockedExternalDatabase++
		case string(batch.StatusUnsupported):
			out.Unsupported++
		case string(batch.StatusUnknown):
			out.Unknown++
		case string(batch.StatusFailed):
			out.Failed++
		case string(batch.StatusCancelled):
			out.Cancelled++
		}
	}
	return &out
}
func snapshotFrom(s batch.Summary) BatchSnapshot {
	out := BatchSnapshot{DiscoveredRegularFiles: s.DiscoveredRegularFiles, EligibleEncrypted: s.EligibleEncrypted, Successes: s.Successes, SkippedNotEncrypted: s.SkippedNotEncrypted, SkippedExisting: s.SkippedExisting, BlockedExternalKey: s.BlockedExternalKey, BlockedExternalDatabase: s.BlockedExternalDatabase, Unsupported: s.Unsupported, Unknown: s.Unknown, Failed: s.Failed, Cancelled: s.Cancelled, BytesWritten: s.BytesWritten, Duration: int64(s.Duration), Results: make([]QueueItem, len(s.Results))}
	for i, r := range s.Results {
		out.Results[i] = itemFrom(r)
	}
	return out
}
