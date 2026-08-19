// Package batch provides the UI-independent batch processing engine used by
// command-line and future GUI frontends.
package batch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"musicunlocker/internal/decode"
	"musicunlocker/internal/externalmaterial"
	"musicunlocker/internal/naming"
	"musicunlocker/internal/outputplan"
	"musicunlocker/internal/probe"
)

type Status string

const (
	StatusSuccess                 Status = "SUCCESS"
	StatusSkippedNotEncrypted     Status = "SKIPPED_NOT_ENCRYPTED"
	StatusSkippedOutputExists     Status = "SKIPPED_OUTPUT_EXISTS"
	StatusBlockedExternalKey      Status = "BLOCKED_EXTERNAL_KEY"
	StatusBlockedExternalDatabase Status = "BLOCKED_EXTERNAL_DATABASE"
	StatusUnsupported             Status = "UNSUPPORTED"
	StatusUnknown                 Status = "UNKNOWN"
	StatusFailed                  Status = "FAILED"
	StatusCancelled               Status = "CANCELLED"
)

const (
	CodeDiscoveryError  = "DISCOVERY_ERROR"
	CodeProcessingError = "PROCESSING_ERROR"
)

type Request struct {
	Inputs                   []string                             `json:"inputs"`
	OutputRoot               string                               `json:"outputRoot"`
	MaxWorkers               int                                  `json:"maxWorkers"`
	Naming                   naming.Config                        `json:"naming,omitempty"`
	ExternalMaterialProvider externalmaterial.Provider            `json:"-"`
	KGGDatabaseProvider      externalmaterial.KGGDatabaseProvider `json:"-"`
}

type Result struct {
	Index                int    `json:"index"`
	SourcePath           string `json:"sourcePath"`
	SourceRoot           string `json:"sourceRoot,omitempty"`
	RelativePath         string `json:"relativePath,omitempty"`
	OutputPath           string `json:"outputPath,omitempty"`
	DetectedFamily       string `json:"detectedFamily,omitempty"`
	Variant              string `json:"variant,omitempty"`
	DecodedAudioFormat   string `json:"decodedAudioFormat,omitempty"`
	BytesWritten         int64  `json:"bytesWritten"`
	SourceBytes          int64  `json:"sourceBytes,omitempty"`
	SourceBytesKnown     bool   `json:"sourceBytesKnown,omitempty"`
	Status               Status `json:"status"`
	ErrorCode            string `json:"errorCode,omitempty"`
	ErrorDetail          string `json:"errorDetail,omitempty"`
	RequestedNamingMode  string `json:"requestedNamingMode,omitempty"`
	ActualNamingMode     string `json:"actualNamingMode,omitempty"`
	NamingFallback       bool   `json:"namingFallback,omitempty"`
	NamingFallbackReason string `json:"namingFallbackReason,omitempty"`
}

type Summary struct {
	DiscoveredRegularFiles  int           `json:"discoveredRegularFiles"`
	EligibleEncrypted       int           `json:"eligibleEncryptedCandidates"`
	Successes               int           `json:"successes"`
	SkippedNotEncrypted     int           `json:"skippedNotEncrypted"`
	SkippedExisting         int           `json:"skippedExisting"`
	BlockedExternalKey      int           `json:"blockedExternalKey"`
	BlockedExternalDatabase int           `json:"blockedExternalDatabase"`
	Unsupported             int           `json:"unsupported"`
	Unknown                 int           `json:"unknown"`
	Failed                  int           `json:"failed"`
	Cancelled               int           `json:"cancelled"`
	BytesWritten            int64         `json:"bytesWritten"`
	Duration                time.Duration `json:"duration"`
	Results                 []Result      `json:"results"`
}

type EventType string

const (
	EventDiscovered EventType = "DISCOVERED"
	EventStarted    EventType = "STARTED"
	EventProgress   EventType = "PROGRESS"
	EventFinished   EventType = "FINISHED"
)

type Event struct {
	Type             EventType `json:"type"`
	Index            int       `json:"index"`
	SourcePath       string    `json:"sourcePath"`
	BytesWritten     int64     `json:"bytesWritten"`
	SourceBytes      int64     `json:"sourceBytes,omitempty"`
	SourceBytesKnown bool      `json:"sourceBytesKnown,omitempty"`
	Result           Result    `json:"result"`
}

type Observer interface{ Observe(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Observe(e Event) error { return f(e) }

type DecodeFunc func(context.Context, string, decode.Options) (decode.Result, error)

type Engine struct {
	Decoder  DecodeFunc
	Observer Observer
}

func (e Engine) Run(ctx context.Context, req Request) (Summary, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if req.MaxWorkers < 0 {
		return Summary{}, errors.New("batch: max workers cannot be negative")
	}
	if strings.TrimSpace(req.OutputRoot) == "" || len(req.Inputs) == 0 {
		return Summary{}, errors.New("batch: at least one input and an output root are required")
	}
	started := time.Now()
	outputRoot, err := filepath.Abs(filepath.Clean(req.OutputRoot))
	if err != nil {
		return Summary{}, fmt.Errorf("batch: output root: %w", err)
	}
	files, discovery := discover(req.Inputs)
	items := make([]*queueItem, 0, len(discovery)+len(files))
	for _, result := range discovery {
		items = append(items, &queueItem{result: result})
	}
	for i := range files {
		items = append(items, &queueItem{file: &files[i], result: files[i].result})
	}
	for index, item := range items {
		item.result.Index = index
	}
	for _, item := range files {
		item.result.Index = indexOfFile(items, item.path)
		e.emit(eventFor(EventDiscovered, item.result))
	}
	if req.Naming.Mode == "" {
		req.Naming.Mode = naming.KeepSource
	}
	if err := naming.Validate(req.Naming); err != nil {
		return Summary{}, err
	}
	workers := req.MaxWorkers
	if workers == 0 {
		workers = 1
	}
	for i := range items {
		item := items[i]
		if item.file == nil {
			continue
		}
		if err := ctx.Err(); err != nil {
			for _, remaining := range items[i:] {
				if remaining.file != nil && remaining.result.Status == "" {
					finishCancelled(e, remaining)
				}
			}
			break
		}
		p, probeErr := probe.ProbeFile(item.file.path)
		if probeErr != nil {
			item.result.Status, item.result.ErrorCode, item.result.ErrorDetail = StatusFailed, CodeDiscoveryError, probeErr.Error()
			e.emit(eventFor(EventStarted, item.result))
			e.emit(eventFor(EventFinished, item.result))
			continue
		}
		item.probe = p
		item.result.DetectedFamily, item.result.Variant = string(p.Family), p.Variant
		if info, statErr := os.Stat(item.file.path); statErr == nil {
			item.result.SourceBytes, item.result.SourceBytesKnown = info.Size(), true
		}
		if p.Family == probe.FamilyRaw {
			item.result.Status = StatusSkippedNotEncrypted
			e.emit(eventFor(EventStarted, item.result))
			e.emit(eventFor(EventFinished, item.result))
			continue
		}
		if p.Confidence != probe.ConfidenceConfirmed || p.Family == probe.FamilyUnknown {
			item.result.Status, item.result.ErrorCode, item.result.ErrorDetail = StatusUnknown, decodeCode(decode.CodeUnknownFormat), "content was not recognised as an encrypted format"
			e.emit(eventFor(EventStarted, item.result))
			e.emit(eventFor(EventFinished, item.result))
			continue
		}
		item.runnable = true
		item.targetKey, _ = outputplan.TargetKey(outputplan.Request{InputPath: item.file.path, OutputDir: filepath.Join(outputRoot, item.file.outputRelativeDir), Probe: p})
	}
	if ctx.Err() == nil {
		runGroups(ctx, e, items, workers, outputRoot, req.Naming, req.ExternalMaterialProvider, req.KGGDatabaseProvider)
	} else {
		for _, item := range items {
			if item.file != nil && item.runnable && item.result.Status == "" {
				finishCancelled(e, item)
			}
		}
	}
	var summary Summary
	summary.DiscoveredRegularFiles = len(files)
	for _, item := range items {
		if item.result.Status == "" {
			item.result = cancelledResult(item.result)
		}
		summary.Results = append(summary.Results, item.result)
		if item.runnable {
			summary.EligibleEncrypted++
		}
		account(&summary, item.result)
	}
	summary.Duration = time.Since(started)
	return summary, nil
}

type queueItem struct {
	file      *discovered
	result    Result
	probe     probe.Result
	runnable  bool
	targetKey string
}

type workGroup struct {
	items []*queueItem
}

// commitGate makes metadata-derived final path ownership follow queue index,
// while decode and metadata work remain concurrent.
type commitGate struct {
	mu    sync.Mutex
	order []int
	pos   int
}

func newCommitGate(items []*queueItem) *commitGate {
	g := &commitGate{}
	for _, i := range items {
		if i.runnable {
			g.order = append(g.order, i.result.Index)
		}
	}
	sort.Ints(g.order)
	return g
}
func (g *commitGate) wait(ctx context.Context, index int) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		g.mu.Lock()
		ready := g.pos >= len(g.order) || g.order[g.pos] == index
		g.mu.Unlock()
		if ready {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
}
func (g *commitGate) complete(index int) {
	g.mu.Lock()
	if g.pos < len(g.order) && g.order[g.pos] == index {
		g.pos++
	}
	g.mu.Unlock()
}

func indexOfFile(items []*queueItem, path string) int {
	for _, item := range items {
		if item.file != nil && item.file.path == path {
			return item.result.Index
		}
	}
	return 0
}

func eventFor(kind EventType, result Result) Event {
	return Event{Type: kind, Index: result.Index, SourcePath: result.SourcePath, BytesWritten: result.BytesWritten, SourceBytes: result.SourceBytes, SourceBytesKnown: result.SourceBytesKnown, Result: result}
}

func finishCancelled(e Engine, item *queueItem) {
	if item.result.Status != "" {
		return
	}
	item.result = cancelledResult(item.result)
	e.emit(eventFor(EventFinished, item.result))
}

func runGroups(ctx context.Context, e Engine, items []*queueItem, maxWorkers int, outputRoot string, config naming.Config, provider externalmaterial.Provider, databaseProvider externalmaterial.KGGDatabaseProvider) {
	groupsByKey := map[string]*workGroup{}
	var groups []*workGroup
	for _, item := range items {
		if !item.runnable {
			continue
		}
		key := canonicalKey(item.targetKey)
		group := groupsByKey[key]
		if group == nil {
			group = &workGroup{}
			groupsByKey[key] = group
			groups = append(groups, group)
		}
		group.items = append(group.items, item)
	}
	if len(groups) == 0 {
		return
	}
	if maxWorkers > len(groups) {
		maxWorkers = len(groups)
	}
	gate := newCommitGate(items)
	jobs := make(chan *workGroup)
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for group := range jobs {
				processGroup(ctx, e, group, outputRoot, config, provider, databaseProvider, gate)
			}
		}()
	}
	dispatched := 0
dispatchLoop:
	for dispatched < len(groups) {
		if ctx.Err() != nil {
			break
		}
		select {
		case jobs <- groups[dispatched]:
			dispatched++
		case <-ctx.Done():
			break dispatchLoop
		}
	}
	close(jobs)
	wg.Wait()
	for _, group := range groups[dispatched:] {
		for _, item := range group.items {
			finishCancelled(e, item)
		}
	}
}

func processGroup(ctx context.Context, e Engine, group *workGroup, outputRoot string, config naming.Config, provider externalmaterial.Provider, databaseProvider externalmaterial.KGGDatabaseProvider, gate *commitGate) {
	for position, item := range group.items {
		if ctx.Err() != nil {
			for _, remaining := range group.items[position:] {
				finishCancelled(e, remaining)
				if config.Mode != naming.KeepSource {
					gate.complete(remaining.result.Index)
				}
			}
			return
		}
		e.emit(eventFor(EventStarted, item.result))
		outDir := filepath.Join(outputRoot, item.file.outputRelativeDir)
		progress := func(p decode.Progress) {
			item.result.SourceBytes, item.result.SourceBytesKnown = p.SourceBytes, p.SourceBytesKnown
			item.result.BytesWritten = p.BytesWritten
			e.emit(eventFor(EventProgress, item.result))
		}
		var decoded decode.Result
		var decodeErr error
		if e.Decoder == nil && config.Mode != naming.KeepSource {
			prepared, prepareErr := decode.PrepareFile(ctx, item.file.path, decode.Options{OutDir: outDir, Progress: progress, Naming: config, ExternalMaterialProvider: provider, KGGDatabaseProvider: databaseProvider})
			if prepareErr == nil {
				if waitErr := gate.wait(ctx, item.result.Index); waitErr != nil {
					prepared.Abort()
					decoded, decodeErr = prepared.Result, waitErr
					gate.complete(item.result.Index)
				} else {
					decoded, decodeErr = decode.FinalizePrepared(&prepared)
					gate.complete(item.result.Index)
				}
			} else {
				decoded, decodeErr = prepared.Result, prepareErr
				gate.complete(item.result.Index)
			}
		} else {
			decoder := e.Decoder
			if decoder == nil {
				decoder = decode.DecodeFile
			}
			decoded, decodeErr = decoder(ctx, item.file.path, decode.Options{OutDir: outDir, Progress: progress, Naming: config, ExternalMaterialProvider: provider, KGGDatabaseProvider: databaseProvider})
		}
		item.result.OutputPath, item.result.DecodedAudioFormat, item.result.BytesWritten = decoded.OutputPath, decoded.DecodedFormat, decoded.BytesWritten
		item.result.RequestedNamingMode, item.result.ActualNamingMode = string(decoded.Naming.RequestedMode), string(decoded.Naming.UsedMode)
		item.result.NamingFallback, item.result.NamingFallbackReason = decoded.Naming.Fallback, decoded.Naming.FallbackReason
		if decodeErr != nil {
			item.result.Status, item.result.ErrorCode, item.result.ErrorDetail = mapError(decodeErr)
			if errors.Is(decodeErr, context.Canceled) || errors.Is(decodeErr, context.DeadlineExceeded) {
				item.result.Status = StatusCancelled
			}
			e.emit(eventFor(EventFinished, item.result))
			if item.result.Status == StatusCancelled {
				for _, remaining := range group.items[position+1:] {
					finishCancelled(e, remaining)
				}
				return
			}
			continue
		}
		item.result.Status = StatusSuccess
		e.emit(eventFor(EventFinished, item.result))
	}
}

type discovered struct {
	path              string
	result            Result
	outputRelativeDir string
}

func discover(inputs []string) ([]discovered, []Result) {
	type root struct {
		path string
		dir  bool
	}
	var roots []root
	var errorsOut []Result
	seenRoots := map[string]bool{}
	for _, input := range inputs {
		abs, err := filepath.Abs(filepath.Clean(input))
		if err != nil {
			errorsOut = append(errorsOut, discoveryResult(input, err))
			continue
		}
		info, err := os.Lstat(abs)
		if err != nil {
			errorsOut = append(errorsOut, discoveryResult(abs, err))
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			errorsOut = append(errorsOut, discoveryResult(abs, errors.New("top-level symlink is not a regular file or directory")))
			continue
		}
		canonical, err := filepath.EvalSymlinks(abs)
		if err != nil {
			errorsOut = append(errorsOut, discoveryResult(abs, err))
			continue
		}
		key := canonicalKey(canonical)
		if seenRoots[key] {
			continue
		}
		seenRoots[key] = true
		roots = append(roots, root{path: canonical, dir: info.IsDir()})
	}
	dirCount := 0
	for _, r := range roots {
		if r.dir {
			dirCount++
		}
	}
	names := map[string]int{}
	dirNamespace := map[string]string{}
	for _, r := range roots {
		if !r.dir {
			continue
		}
		base := filepath.Base(filepath.Clean(r.path))
		names[canonicalKey(base)]++
		if dirCount > 1 {
			dirNamespace[canonicalKey(r.path)] = base
		}
	}
	if dirCount > 1 {
		byBase := map[string][]string{}
		for _, r := range roots {
			if r.dir {
				byBase[canonicalKey(filepath.Base(r.path))] = append(byBase[canonicalKey(filepath.Base(r.path))], r.path)
			}
		}
		for key := range byBase {
			sort.Slice(byBase[key], func(i, j int) bool { return canonicalKey(byBase[key][i]) < canonicalKey(byBase[key][j]) })
		}
		for _, r := range roots {
			if r.dir {
				key := canonicalKey(filepath.Base(r.path))
				if names[key] > 1 {
					idx := 1
					for _, candidate := range byBase[key] {
						if canonicalKey(candidate) < canonicalKey(r.path) {
							idx++
						}
					}
					dirNamespace[canonicalKey(r.path)] = fmt.Sprintf("%s (%d)", filepath.Base(r.path), idx)
				}
			}
		}
	}
	var files []discovered
	fileBases := map[string]int{}
	for _, r := range roots {
		if !r.dir {
			files = append(files, discovered{path: r.path, result: Result{SourcePath: r.path, SourceRoot: r.path}, outputRelativeDir: ""})
			fileBases[canonicalKey(strings.TrimSuffix(filepath.Base(r.path), filepath.Ext(r.path)))]++
			continue
		}
		err := filepath.WalkDir(r.path, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				errorsOut = append(errorsOut, discoveryResult(path, walkErr))
				if entry == nil {
					return nil
				}
				return nil
			}
			if path != r.path && entry.Type()&os.ModeSymlink != 0 {
				if target, statErr := os.Stat(path); statErr == nil && target.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				errorsOut = append(errorsOut, discoveryResult(path, err))
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(r.path, path)
			if err != nil {
				errorsOut = append(errorsOut, discoveryResult(path, err))
				return nil
			}
			dir := filepath.Dir(rel)
			if dir == "." {
				dir = ""
			}
			if dirNamespace[canonicalKey(r.path)] != "" {
				dir = filepath.Join(dirNamespace[canonicalKey(r.path)], dir)
			}
			files = append(files, discovered{path: path, result: Result{SourcePath: path, SourceRoot: r.path, RelativePath: rel}, outputRelativeDir: dir})
			return nil
		})
		if err != nil {
			errorsOut = append(errorsOut, discoveryResult(r.path, err))
		}
	}
	sort.Slice(files, func(i, j int) bool { return canonicalKey(files[i].path) < canonicalKey(files[j].path) })
	unique := files[:0]
	for _, file := range files {
		if len(unique) > 0 && canonicalKey(unique[len(unique)-1].path) == canonicalKey(file.path) {
			// Prefer the directory-root mapping when the same file was also
			// supplied directly.
			if unique[len(unique)-1].result.SourceRoot == unique[len(unique)-1].path && file.result.SourceRoot != file.path {
				unique[len(unique)-1] = file
			}
			continue
		}
		unique = append(unique, file)
	}
	files = unique
	// Exact direct-file basename collisions use deterministic parent namespaces.
	if dirCount == 0 {
		for i := range files {
			base := canonicalKey(strings.TrimSuffix(filepath.Base(files[i].path), filepath.Ext(files[i].path)))
			if fileBases[base] > 1 {
				files[i].outputRelativeDir = filepath.Base(filepath.Dir(files[i].path))
			}
		}
	}
	return files, errorsOut
}

func discoveryResult(path string, err error) Result {
	return Result{SourcePath: path, Status: StatusFailed, ErrorCode: CodeDiscoveryError, ErrorDetail: err.Error()}
}
func cancelledResult(r Result) Result {
	r.Status, r.ErrorCode, r.ErrorDetail = StatusCancelled, "CONTEXT_CANCELLED", "processing was cancelled before this file started"
	return r
}
func canonicalKey(path string) string {
	p := filepath.Clean(path)
	if filepath.Separator == '\\' {
		p = strings.ToLower(p)
	}
	return p
}
func decodeCode(c decode.ErrorCode) string { return string(c) }
func mapError(err error) (Status, string, string) {
	switch decode.CodeOf(err) {
	case decode.CodeOutputExists:
		return StatusSkippedOutputExists, string(decode.CodeOutputExists), err.Error()
	case decode.CodeQMCExternalKeyRequired:
		return StatusBlockedExternalKey, string(decode.CodeQMCExternalKeyRequired), err.Error()
	case decode.CodeKGGDatabaseRequired:
		return StatusBlockedExternalDatabase, string(decode.CodeKGGDatabaseRequired), err.Error()
	case decode.CodeKGGDatabaseEntryMissing:
		return StatusFailed, string(decode.CodeKGGDatabaseEntryMissing), err.Error()
	case decode.CodeUnsupportedEncrypted:
		return StatusUnsupported, string(decode.CodeUnsupportedEncrypted), err.Error()
	case decode.CodeUnknownFormat:
		return StatusUnknown, string(decode.CodeUnknownFormat), err.Error()
	case decode.CodeDecodedAudioUnknown:
		return StatusUnknown, string(decode.CodeDecodedAudioUnknown), err.Error()
	default:
		return StatusFailed, CodeProcessingError, err.Error()
	}
}
func (e Engine) emit(event Event) {
	if e.Observer == nil {
		return
	}
	func() { defer func() { _ = recover() }(); _ = e.Observer.Observe(event) }()
}
func account(s *Summary, r Result) {
	switch r.Status {
	case StatusSuccess:
		s.Successes++
		s.BytesWritten += r.BytesWritten
	case StatusSkippedNotEncrypted:
		s.SkippedNotEncrypted++
	case StatusSkippedOutputExists:
		s.SkippedExisting++
	case StatusBlockedExternalKey:
		s.BlockedExternalKey++
	case StatusBlockedExternalDatabase:
		s.BlockedExternalDatabase++
	case StatusUnsupported:
		s.Unsupported++
	case StatusUnknown:
		s.Unknown++
	case StatusFailed:
		s.Failed++
	case StatusCancelled:
		s.Cancelled++
	}
}
