// Package containermeta reads small, offline metadata subsets from encrypted
// containers without exposing decoder-library types to application callers.
package containermeta

import (
	"context"
	"errors"
	"io"

	"musicunlocker/internal/audiometa"
	"musicunlocker/internal/decoder/upstream"
	"musicunlocker/internal/probe"
)

var ErrUnavailable = errors.New("container metadata unavailable")

func Read(ctx context.Context, family probe.Family, reader io.ReadSeeker, path string) (audiometa.Metadata, error) {
	if family != probe.FamilyNetease {
		return audiometa.Metadata{}, ErrUnavailable
	}
	meta, err := upstream.ReadNCMMetadata(ctx, reader, path)
	if err != nil {
		return audiometa.Metadata{}, ErrUnavailable
	}
	return audiometa.Metadata{
		Title:  meta.Title,
		Artist: meta.Artist,
		Album:  meta.Album,
	}, nil
}
