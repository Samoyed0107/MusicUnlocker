package app

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type LaunchSource string

const (
	LaunchSourceInitial LaunchSource = "INITIAL_PROCESS"
	LaunchSourceSecond  LaunchSource = "SECOND_INSTANCE"
)

type LaunchPathsEvent struct {
	RequestID string       `json:"requestId"`
	Paths     []string     `json:"paths"`
	Source    LaunchSource `json:"source"`
	Deferred  bool         `json:"deferred"`
}

type LaunchPathRejection struct {
	Argument string `json:"argument"`
	Reason   string `json:"reason"`
}

type LaunchParseResult struct {
	Paths    []string              `json:"paths"`
	Rejected []LaunchPathRejection `json:"rejected"`
}

func ParseLaunchPaths(args []string, workingDirectory string) LaunchParseResult {
	result := LaunchParseResult{}
	seen := map[string]struct{}{}
	for _, arg := range args {
		if arg == "" || strings.HasPrefix(arg, "-") {
			if arg != "" {
				result.Rejected = append(result.Rejected, LaunchPathRejection{Argument: arg, Reason: "OPTION_ARGUMENT"})
			}
			continue
		}
		candidate := arg
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(workingDirectory, candidate)
		}
		candidate = filepath.Clean(candidate)
		info, err := os.Stat(candidate)
		if err != nil {
			result.Rejected = append(result.Rejected, LaunchPathRejection{Argument: arg, Reason: "NOT_FOUND"})
			continue
		}
		if !info.Mode().IsRegular() && !info.IsDir() {
			result.Rejected = append(result.Rejected, LaunchPathRejection{Argument: arg, Reason: "UNSUPPORTED_ENTRY"})
			continue
		}
		key := normalizeLaunchPath(candidate)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result.Paths = append(result.Paths, candidate)
	}
	return result
}

func normalizeLaunchPath(path string) string {
	return strings.ToLower(strings.ReplaceAll(filepath.Clean(path), "/", "\\"))
}

type LaunchInbox struct {
	mu          sync.Mutex
	nextRequest uint64
	initial     *LaunchPathsEvent
	deferred    []LaunchPathsEvent
}

func (i *LaunchInbox) nextID() string {
	i.nextRequest++
	return "launch-" + formatUint(i.nextRequest)
}

func formatUint(value uint64) string {
	if value == 0 {
		return "0"
	}
	const digits = "0123456789"
	var buf [20]byte
	pos := len(buf)
	for value > 0 {
		pos--
		buf[pos] = digits[value%10]
		value /= 10
	}
	return string(buf[pos:])
}

func (i *LaunchInbox) NewEvent(paths []string, source LaunchSource) LaunchPathsEvent {
	i.mu.Lock()
	defer i.mu.Unlock()
	return LaunchPathsEvent{RequestID: i.nextID(), Paths: append([]string(nil), paths...), Source: source}
}

func (i *LaunchInbox) RetainInitial(event LaunchPathsEvent) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.initial == nil {
		event.Paths = append([]string(nil), event.Paths...)
		i.initial = &event
	}
}

func (i *LaunchInbox) ConsumeInitial() *LaunchPathsEvent {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.initial == nil {
		return nil
	}
	event := *i.initial
	event.Paths = append([]string(nil), event.Paths...)
	i.initial = nil
	return &event
}

func (i *LaunchInbox) Defer(event LaunchPathsEvent) {
	i.mu.Lock()
	defer i.mu.Unlock()
	event.Deferred = true
	for _, existing := range i.deferred {
		if sameLaunchPaths(existing.Paths, event.Paths) {
			return
		}
	}
	event.Paths = append([]string(nil), event.Paths...)
	i.deferred = append(i.deferred, event)
}

func (i *LaunchInbox) FlushDeferred() []LaunchPathsEvent {
	i.mu.Lock()
	defer i.mu.Unlock()
	flushed := append([]LaunchPathsEvent(nil), i.deferred...)
	i.deferred = nil
	for n := range flushed {
		flushed[n].Deferred = false
		flushed[n].Paths = append([]string(nil), flushed[n].Paths...)
	}
	return flushed
}

func sameLaunchPaths(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for n := range a {
		if normalizeLaunchPath(a[n]) != normalizeLaunchPath(b[n]) {
			return false
		}
	}
	return true
}
