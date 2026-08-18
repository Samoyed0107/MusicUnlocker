package decode

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"musicunlocker/internal/audiofmt"
	"musicunlocker/internal/outputplan"
	"musicunlocker/internal/probe"
)

type Stream interface{ io.Reader }

type Input struct {
	Path   string
	Probe  probe.Result
	Reader io.ReadSeeker
}

type Resolver interface{ Resolve(*Input) (Stream, error) }

type Options struct {
	OutDir   string
	Resolver Resolver
}

type Result struct {
	InputPath        string
	OutputPath       string
	EncryptedFamily  string
	EncryptedVariant string
	CryptoVersion    uint32
	DecodedFormat    string
	BytesWritten     int64
}

func DecodeFile(ctx context.Context, inputPath string, options Options) (Result, error) {
	var result Result
	if options.OutDir == "" {
		return result, coded(CodeInvalidInput, errors.New("output directory is required"))
	}
	f, err := os.Open(inputPath)
	if err != nil {
		return result, err
	}
	defer f.Close()
	p, err := probe.Probe(f, filepath.Base(inputPath))
	if err != nil {
		return result, err
	}
	result.InputPath = inputPath
	result.EncryptedFamily, result.EncryptedVariant, result.CryptoVersion = string(p.Family), p.Variant, p.CryptoVersion
	if options.Resolver == nil {
		options.Resolver = DefaultResolver{}
	}
	input := &Input{Path: inputPath, Probe: p, Reader: f}
	stream, err := options.Resolver.Resolve(input)
	if err != nil {
		return result, err
	}
	result.EncryptedFamily, result.EncryptedVariant, result.CryptoVersion = string(input.Probe.Family), input.Probe.Variant, input.Probe.CryptoVersion

	prefix := make([]byte, 64)
	n, readErr := io.ReadFull(stream, prefix)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
		return result, readErr
	}
	prefix = prefix[:n]
	format, ok := audiofmt.Detect(prefix)
	if !ok {
		return result, coded(CodeDecodedAudioUnknown, errors.New("decrypted stream has no recognised audio signature"))
	}
	result.DecodedFormat = format
	if err := os.MkdirAll(options.OutDir, 0755); err != nil {
		return result, err
	}
	outputPath, planErr := outputplan.Plan(outputplan.Request{InputPath: inputPath, OutputDir: options.OutDir, Probe: input.Probe, DecodedFormat: format})
	if planErr != nil {
		if errors.Is(planErr, outputplan.ErrSourceConflict) {
			return result, coded(CodeOutputConflictsSource, planErr)
		}
		return result, planErr
	}
	result.OutputPath = outputPath
	if _, err := os.Stat(outputPath); err == nil {
		return result, coded(CodeOutputExists, errors.New("output already exists: "+outputPath))
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}

	tmp, err := os.CreateTemp(options.OutDir, ".musicunlocker-*.tmp")
	if err != nil {
		return result, err
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = tmp.Close(); _ = os.Remove(tmpName) }
	full := io.MultiReader(bytes.NewReader(prefix), stream)
	result.BytesWritten, err = copyContext(ctx, tmp, full)
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		cleanup()
		return result, err
	}
	if err = os.Rename(tmpName, outputPath); err != nil {
		cleanup()
		return result, err
	}
	return result, nil
}

func copyContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			written, err := dst.Write(buf[:n])
			total += int64(written)
			if err != nil {
				return total, err
			}
			if written != n {
				return total, io.ErrShortWrite
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return total, nil
			}
			return total, readErr
		}
	}
}
