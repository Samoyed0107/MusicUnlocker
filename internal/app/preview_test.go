package app

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"musicunlocker/internal/batch"
)

func TestPreviewHTTPRangeAndAuthorization(t *testing.T) {
	p := NewPreviewService()
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.mp3")
	want := []byte("0123456789")
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}
	token, _, err := p.Register(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(token, path) || strings.Contains(p.URL(token), path) {
		t.Fatal("preview token leaked filesystem path")
	}
	resp, err := http.Get(p.URL(token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status=%d", resp.StatusCode)
	}
	got, _ := io.ReadAll(resp.Body)
	if string(got) != string(want) {
		t.Fatalf("GET bytes=%q", got)
	}
	head, err := http.Head(p.URL(token))
	if err != nil {
		t.Fatal(err)
	}
	if head.StatusCode != http.StatusOK {
		t.Fatalf("HEAD status=%d", head.StatusCode)
	}
	req, _ := http.NewRequest(http.MethodGet, p.URL(token), nil)
	req.Header.Set("Range", "bytes=2-5")
	rangeResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rangeResp.Body.Close()
	if rangeResp.StatusCode != http.StatusPartialContent {
		t.Fatalf("range status=%d", rangeResp.StatusCode)
	}
	ranged, _ := io.ReadAll(rangeResp.Body)
	if string(ranged) != "2345" {
		t.Fatalf("range bytes=%q", ranged)
	}
	badReq, _ := http.NewRequest(http.MethodGet, p.URL(token), nil)
	badReq.Header.Set("Range", "bytes=99-100")
	badResp, err := http.DefaultClient.Do(badReq)
	if err != nil {
		t.Fatal(err)
	}
	badResp.Body.Close()
	if badResp.StatusCode != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("bad range status=%d", badResp.StatusCode)
	}
	unknown, err := http.Get(p.baseURL + "/preview/not-a-token")
	if err != nil {
		t.Fatal(err)
	}
	unknown.Body.Close()
	if unknown.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown status=%d", unknown.StatusCode)
	}
	methodReq, _ := http.NewRequest(http.MethodPost, p.URL(token), nil)
	methodResp, err := http.DefaultClient.Do(methodReq)
	if err != nil {
		t.Fatal(err)
	}
	methodResp.Body.Close()
	if methodResp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("method status=%d", methodResp.StatusCode)
	}
}

func TestPreviewDoesNotServeDeletedFilesOrDirectories(t *testing.T) {
	p := NewPreviewService()
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })
	dir := t.TempDir()
	path := filepath.Join(dir, "out.wav")
	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	token, _, err := p.Register(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(p.URL(token))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("directory status=%d", resp.StatusCode)
	}
}

func TestPreviewFacadeValidatesLatestSuccessfulResult(t *testing.T) {
	p := NewPreviewService()
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })
	path := filepath.Join(t.TempDir(), "ok.flac")
	if err := os.WriteFile(path, []byte("audio"), 0600); err != nil {
		t.Fatal(err)
	}
	a := NewWithPreview(&fakeEngine{}, &fakeSink{}, nil, nil, nil, p)
	a.latestID = 7
	a.latest = &BatchSnapshot{Results: []QueueItem{{Index: 3, Status: string(batch.StatusSuccess), OutputPath: path}}}
	result, err := a.GetPreviewURL(7, 3)
	if err != nil || !result.Available || !strings.HasPrefix(result.URL, "http://127.0.0.1:") {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if _, err := a.GetPreviewURL(6, 3); err == nil || !strings.Contains(err.Error(), "PREVIEW_RESULT_NOT_FOUND") {
		t.Fatalf("stale batch err=%v", err)
	}
	for _, status := range []batch.Status{batch.StatusFailed, batch.StatusSkippedNotEncrypted, batch.StatusCancelled} {
		a.latest.Results[0].Status = string(status)
		if _, err := a.GetPreviewURL(7, 3); err == nil || !strings.Contains(err.Error(), "PREVIEW_NOT_AVAILABLE") {
			t.Fatalf("status %s err=%v", status, err)
		}
	}
	a.latest.Results[0].Status = string(batch.StatusSuccess)
	a.latest.Results[0].OutputPath = filepath.Join(t.TempDir(), "missing.mp3")
	if _, err := a.GetPreviewURL(7, 3); err == nil || !strings.Contains(err.Error(), "PREVIEW_FILE_MISSING") {
		t.Fatalf("missing output err=%v", err)
	}
	a.preview.Invalidate()
	a.latestID, a.latest = 0, nil
	if _, err := a.GetPreviewURL(7, 3); err == nil {
		t.Fatal("invalidated generation still authorized")
	}
	if err := a.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.Shutdown(shutdownCtx); err != nil {
		t.Fatal(err)
	}
}
