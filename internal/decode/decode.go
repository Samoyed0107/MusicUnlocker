package decode

import (
	"bytes"
	"context"
	"errors"
	"io"
	"musicunlocker/internal/audiofmt"
	"musicunlocker/internal/audiometa"
	"musicunlocker/internal/containermeta"
	"musicunlocker/internal/externalmaterial"
	"musicunlocker/internal/naming"
	"musicunlocker/internal/outputplan"
	"musicunlocker/internal/probe"
	"os"
	"path/filepath"
	"strings"
)

type Stream interface{ io.Reader }
type Input struct {
	Path             string
	Probe            probe.Result
	Reader           io.ReadSeeker
	ExternalMaterial []byte
}
type Resolver interface{ Resolve(*Input) (Stream, error) }
type Progress struct {
	BytesWritten, SourceBytes int64
	SourceBytesKnown          bool
}
type Options struct {
	OutDir                   string
	Resolver                 Resolver
	Progress                 func(Progress)
	Naming                   naming.Config
	ExternalMaterialProvider externalmaterial.Provider
}
type Result struct {
	InputPath, OutputPath, EncryptedFamily, EncryptedVariant, DecodedFormat string
	CryptoVersion                                                           uint32
	BytesWritten                                                            int64
	Naming                                                                  naming.Result
}
type PreparedResult struct {
	Result
	TempPath, OutputDir, SourceStem string
	Probe                           probe.Result
	Metadata                        audiometa.Metadata
	MetadataAvailable               bool
	finalized                       bool
}

func DecodeFile(ctx context.Context, inputPath string, options Options) (Result, error) {
	p, e := PrepareFile(ctx, inputPath, options)
	if e != nil {
		return p.Result, e
	}
	return FinalizePrepared(&p)
}
func PrepareFile(ctx context.Context, inputPath string, options Options) (PreparedResult, error) {
	var r Result
	if options.OutDir == "" {
		return PreparedResult{Result: r}, coded(CodeInvalidInput, errors.New("output directory is required"))
	}
	f, e := os.Open(inputPath)
	if e != nil {
		return PreparedResult{Result: r}, e
	}
	defer f.Close()
	var sb int64
	known := false
	if i, se := f.Stat(); se == nil {
		sb, known = i.Size(), true
	}
	emitProgress(options.Progress, Progress{SourceBytes: sb, SourceBytesKnown: known})
	p, e := probe.Probe(f, filepath.Base(inputPath))
	if e != nil {
		return PreparedResult{Result: r}, e
	}
	r.InputPath = inputPath
	r.EncryptedFamily = string(p.Family)
	r.EncryptedVariant = p.Variant
	r.CryptoVersion = p.CryptoVersion
	containerMeta, _ := containermeta.Read(ctx, p.Family, f, inputPath)
	if options.Resolver == nil {
		options.Resolver = DefaultResolver{}
	}
	in := &Input{Path: inputPath, Probe: p, Reader: f}
	if p.NeedsExternalKey && options.ExternalMaterialProvider != nil {
		material, ok, providerErr := options.ExternalMaterialProvider.ResolveQMCExternalKey(ctx, inputPath)
		if providerErr != nil {
			return PreparedResult{Result: r}, coded(CodeQMCExternalKeyProvider, errors.New("external QMC key provider failed"))
		}
		if ok {
			in.ExternalMaterial = material
		}
	}
	stream, e := options.Resolver.Resolve(in)
	if e != nil {
		return PreparedResult{Result: r}, e
	}
	p = in.Probe
	r.EncryptedFamily = string(p.Family)
	r.EncryptedVariant = p.Variant
	r.CryptoVersion = p.CryptoVersion
	prefix := make([]byte, 64)
	n, re := io.ReadFull(stream, prefix)
	if re != nil && !errors.Is(re, io.ErrUnexpectedEOF) && !errors.Is(re, io.EOF) {
		return PreparedResult{Result: r}, re
	}
	prefix = prefix[:n]
	format, ok := audiofmt.Detect(prefix)
	if !ok {
		return PreparedResult{Result: r}, coded(CodeDecodedAudioUnknown, errors.New("decrypted stream has no recognised audio signature"))
	}
	r.DecodedFormat = format
	if e = os.MkdirAll(options.OutDir, 0755); e != nil {
		return PreparedResult{Result: r}, e
	}
	tmp, e := os.CreateTemp(options.OutDir, ".musicunlocker-*.tmp")
	if e != nil {
		return PreparedResult{Result: r}, e
	}
	name := tmp.Name()
	cleanup := func() { _ = tmp.Close(); _ = os.Remove(name) }
	r.BytesWritten, e = copyContext(ctx, tmp, io.MultiReader(bytes.NewReader(prefix), stream), func(n int64) {
		emitProgress(options.Progress, Progress{BytesWritten: n, SourceBytes: sb, SourceBytesKnown: known})
	})
	if e == nil {
		e = tmp.Sync()
	}
	if ce := tmp.Close(); e == nil {
		e = ce
	}
	if e != nil {
		cleanup()
		return PreparedResult{Result: r}, e
	}
	decodedMeta, _ := audiometa.ReadFile(name)
	meta := audiometa.Merge(decodedMeta, containerMeta)
	config := options.Naming
	if config.Mode == "" {
		config.Mode = naming.KeepSource
	}
	stem := outputplan.CleanStem(filepath.Base(inputPath), p)
	nr, ne := naming.Resolve(config, stem, meta, meta.Available())
	if ne != nil {
		cleanup()
		return PreparedResult{Result: r}, ne
	}
	r.Naming = nr
	return PreparedResult{Result: r, TempPath: name, OutputDir: options.OutDir, SourceStem: stem, Probe: p, Metadata: meta, MetadataAvailable: meta.Available()}, nil
}
func FinalizePrepared(p *PreparedResult) (Result, error) {
	if p == nil {
		return Result{}, errors.New("nil prepared result")
	}
	if p.finalized {
		return p.Result, errors.New("prepared result already finalized")
	}
	p.finalized = true
	defer func() { _ = os.Remove(p.TempPath) }()
	if p.TempPath == "" {
		return p.Result, errors.New("prepared result has no temporary file")
	}
	ext := strings.TrimPrefix(strings.ToLower(p.DecodedFormat), ".")
	out := filepath.Join(p.OutputDir, p.Naming.Stem+"."+ext)
	if samePath(out, p.InputPath) {
		return p.Result, coded(CodeOutputConflictsSource, outputplan.ErrSourceConflict)
	}
	p.Result.OutputPath = out
	if e := os.Link(p.TempPath, out); e != nil {
		if errors.Is(e, os.ErrExist) {
			return p.Result, coded(CodeOutputExists, errors.New("output already exists: "+out))
		}
		return p.Result, e
	}
	return p.Result, nil
}
func (p *PreparedResult) Abort() {
	if p == nil {
		return
	}
	p.finalized = true
	if p.TempPath != "" {
		_ = os.Remove(p.TempPath)
	}
}
func copyContext(ctx context.Context, dst io.Writer, src io.Reader, progress func(int64)) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
		n, e := src.Read(buf)
		if n > 0 {
			w, we := dst.Write(buf[:n])
			total += int64(w)
			if w > 0 {
				safeProgress(progress, total)
			}
			if we != nil {
				return total, we
			}
			if w != n {
				return total, io.ErrShortWrite
			}
		}
		if e != nil {
			if errors.Is(e, io.EOF) {
				return total, nil
			}
			return total, e
		}
	}
}
func emitProgress(cb func(Progress), p Progress) {
	if cb == nil {
		return
	}
	func() { defer func() { _ = recover() }(); cb(p) }()
}
func safeProgress(cb func(int64), n int64) {
	if cb == nil {
		return
	}
	func() { defer func() { _ = recover() }(); cb(n) }()
}
func samePath(a, b string) bool {
	aa, e1 := filepath.Abs(a)
	bb, e2 := filepath.Abs(b)
	return e1 == nil && e2 == nil && strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}
