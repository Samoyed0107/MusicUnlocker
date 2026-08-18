package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrPreviewNotAvailable      = &ApplicationError{Code: "PREVIEW_NOT_AVAILABLE", Detail: "preview is not available"}
	ErrPreviewResultNotFound    = &ApplicationError{Code: "PREVIEW_RESULT_NOT_FOUND", Detail: "the completed result was not found"}
	ErrPreviewFileMissing       = &ApplicationError{Code: "PREVIEW_FILE_MISSING", Detail: "the output file is no longer available"}
	ErrPreviewServerUnavailable = &ApplicationError{Code: "PREVIEW_SERVER_UNAVAILABLE", Detail: "the local preview server is unavailable"}
)

type PreviewResponse struct {
	Available bool   `json:"available"`
	URL       string `json:"url,omitempty"`
	Format    string `json:"format,omitempty"`
}

type previewEntry struct{ path, format string }

// PreviewService is application-layer infrastructure. It is deliberately not
// connected to discovery, decoding, or output planning.
type PreviewService struct {
	mu         sync.RWMutex
	entries    map[string]previewEntry
	baseURL    string
	server     *http.Server
	listener   net.Listener
	startError error
}

func NewPreviewService() *PreviewService {
	p := &PreviewService{entries: make(map[string]previewEntry)}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		p.startError = err
		return p
	}
	p.listener = listener
	p.baseURL = "http://" + listener.Addr().String()
	p.server = &http.Server{Handler: p, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		if err := p.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			p.mu.Lock()
			p.startError = err
			p.mu.Unlock()
		}
	}()
	return p
}

func (p *PreviewService) Available() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.server != nil && p.startError == nil
}

func (p *PreviewService) Invalidate() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = make(map[string]previewEntry)
}

func (p *PreviewService) Register(path string) (string, string, error) {
	if !p.Available() {
		return "", "", ErrPreviewServerUnavailable
	}
	canonical, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", "", ErrPreviewFileMissing
	}
	info, err := os.Lstat(canonical)
	if err != nil || !info.Mode().IsRegular() {
		return "", "", ErrPreviewFileMissing
	}
	tokenBytes := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, tokenBytes); err != nil {
		return "", "", fmt.Errorf("preview token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(canonical)), ".")
	p.mu.Lock()
	p.entries[token] = previewEntry{path: canonical, format: format}
	p.mu.Unlock()
	return token, format, nil
}

func (p *PreviewService) URL(token string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.entries[token]; !ok {
		return ""
	}
	return p.baseURL + "/preview/" + token
}

func (p *PreviewService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/preview/") || strings.Contains(strings.TrimPrefix(r.URL.Path, "/preview/"), "/") {
		http.NotFound(w, r)
		return
	}
	token := strings.TrimPrefix(r.URL.Path, "/preview/")
	p.mu.RLock()
	entry, ok := p.entries[token]
	p.mu.RUnlock()
	if !ok || token == "" {
		http.NotFound(w, r)
		return
	}
	info, err := os.Lstat(entry.path)
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(entry.path)
	if err != nil {
		http.Error(w, "preview unavailable", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	contentType := mime.TypeByExtension(filepath.Ext(entry.path))
	if contentType == "" {
		contentType = map[string]string{".mp3": "audio/mpeg", ".flac": "audio/flac", ".ogg": "audio/ogg", ".wav": "audio/wav", ".m4a": "audio/mp4", ".mp4": "audio/mp4"}[strings.ToLower(filepath.Ext(entry.path))]
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(w, r, filepath.Base(entry.path), info.ModTime(), f)
}

func (p *PreviewService) Shutdown(ctx context.Context) error {
	if p == nil || p.server == nil {
		return nil
	}
	return p.server.Shutdown(ctx)
}
