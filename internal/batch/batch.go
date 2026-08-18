// Package batch provides the UI-independent, sequential batch processing
// engine used by future command-line and GUI frontends.
package batch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"musicunlocker/internal/decode"
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
	Inputs     []string `json:"inputs"`
	OutputRoot string   `json:"outputRoot"`
}

type Result struct {
	SourcePath         string `json:"sourcePath"`
	SourceRoot         string `json:"sourceRoot,omitempty"`
	RelativePath       string `json:"relativePath,omitempty"`
	OutputPath         string `json:"outputPath,omitempty"`
	DetectedFamily     string `json:"detectedFamily,omitempty"`
	Variant            string `json:"variant,omitempty"`
	DecodedAudioFormat string `json:"decodedAudioFormat,omitempty"`
	BytesWritten       int64  `json:"bytesWritten"`
	Status             Status `json:"status"`
	ErrorCode          string `json:"errorCode,omitempty"`
	ErrorDetail        string `json:"errorDetail,omitempty"`
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
	EventFinished   EventType = "FINISHED"
)

type Event struct {
	Type   EventType `json:"type"`
	Result Result    `json:"result"`
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
	started := time.Now()
	var summary Summary
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(req.OutputRoot) == "" || len(req.Inputs) == 0 {
		return summary, errors.New("batch: at least one input and an output root are required")
	}
	outputRoot, err := filepath.Abs(filepath.Clean(req.OutputRoot))
	if err != nil {
		return summary, fmt.Errorf("batch: output root: %w", err)
	}
	files, discovery := discover(req.Inputs)
	summary.DiscoveredRegularFiles = len(files)
	for _, item := range files {
		e.emit(Event{Type: EventDiscovered, Result: item.result})
	}
	for _, item := range discovery {
		summary.Results = append(summary.Results, item)
		account(&summary, item)
	}
	if e.Decoder == nil {
		e.Decoder = decode.DecodeFile
	}
	for i, item := range files {
		if err := ctx.Err(); err != nil {
			for _, remaining := range files[i:] {
				r := cancelledResult(remaining.result)
				summary.Results = append(summary.Results, r)
				account(&summary, r)
				e.emit(Event{Type: EventFinished, Result: r})
			}
			break
		}
		r := item.result
		e.emit(Event{Type: EventStarted, Result: r})
		p, probeErr := probe.ProbeFile(item.path)
		if probeErr != nil {
			r.Status, r.ErrorCode, r.ErrorDetail = StatusFailed, CodeDiscoveryError, probeErr.Error()
			summary.Results = append(summary.Results, r)
			account(&summary, r)
			e.emit(Event{Type: EventFinished, Result: r})
			continue
		}
		r.DetectedFamily, r.Variant = string(p.Family), p.Variant
		if p.Family == probe.FamilyRaw {
			r.Status = StatusSkippedNotEncrypted
			summary.Results = append(summary.Results, r)
			account(&summary, r)
			e.emit(Event{Type: EventFinished, Result: r})
			continue
		}
		if p.Confidence != probe.ConfidenceConfirmed || p.Family == probe.FamilyUnknown {
			r.Status, r.ErrorCode, r.ErrorDetail = StatusUnknown, decodeCode(decode.CodeUnknownFormat), "content was not recognised as an encrypted format"
			summary.Results = append(summary.Results, r)
			account(&summary, r)
			e.emit(Event{Type: EventFinished, Result: r})
			continue
		}
		summary.EligibleEncrypted++
		outDir := filepath.Join(outputRoot, item.outputRelativeDir)
		decoded, decodeErr := e.Decoder(ctx, item.path, decode.Options{OutDir: outDir})
		r.OutputPath, r.DecodedAudioFormat, r.BytesWritten = decoded.OutputPath, decoded.DecodedFormat, decoded.BytesWritten
		if decodeErr != nil {
			r.Status, r.ErrorCode, r.ErrorDetail = mapError(decodeErr)
			if errors.Is(decodeErr, context.Canceled) || errors.Is(decodeErr, context.DeadlineExceeded) {
				r.Status = StatusCancelled
			}
			summary.Results = append(summary.Results, r)
			account(&summary, r)
			e.emit(Event{Type: EventFinished, Result: r})
			if r.Status == StatusCancelled {
				for _, remaining := range files[i+1:] {
					cr := cancelledResult(remaining.result)
					summary.Results = append(summary.Results, cr)
					account(&summary, cr)
					e.emit(Event{Type: EventFinished, Result: cr})
				}
				break
			}
			continue
		}
		r.Status = StatusSuccess
		summary.Results = append(summary.Results, r)
		account(&summary, r)
		e.emit(Event{Type: EventFinished, Result: r})
	}
	summary.Duration = time.Since(started)
	return summary, nil
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
