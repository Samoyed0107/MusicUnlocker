// Package externalmaterial owns request-scoped decryption material.
package externalmaterial

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var ErrUnavailable = errors.New("external material unavailable")

// Provider may be called concurrently. Implementations must not mutate or
// expose material based on a different source path.
type Provider interface {
	ResolveQMCExternalKey(context.Context, string) ([]byte, bool, error)
}

// KGGDatabase is an application-owned reference to a user-supplied runtime
// database. The path is never serialized into batch results or frontend DTOs.
type KGGDatabase struct{ Path string }

type KGGDatabaseProvider interface {
	ResolveKGGDatabase(context.Context, string) (KGGDatabase, bool, error)
}

const MaxKGGDatabaseSize int64 = 256 << 20

// ValidateKGGDatabasePath performs bounded, read-only validation. The file is
// intentionally not opened through SQLite here because the upstream format may
// be encrypted and is decoded by the upstream adapter.
func ValidateKGGDatabasePath(path string) error {
	if path == "" {
		return errors.New("KGG database path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return errors.New("KGG database is unavailable")
	}
	if !info.Mode().IsRegular() {
		return errors.New("KGG database is not a regular file")
	}
	if info.Size() <= 0 || info.Size() > MaxKGGDatabaseSize {
		return fmt.Errorf("KGG database size is outside the supported bound")
	}
	f, err := os.Open(path)
	if err != nil {
		return errors.New("KGG database could not be opened read-only")
	}
	defer f.Close()
	var header [16]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return errors.New("KGG database is truncated")
	}
	if string(header[:]) != "SQLite format 3\x00" && info.Size()%0x400 != 0 {
		return errors.New("KGG database format is not recognized")
	}
	return nil
}

// MapProvider is an immutable, in-memory provider suitable for a single
// request or batch. It never persists material and returns a defensive copy.
type MapProvider struct {
	mu   sync.RWMutex
	keys map[string][]byte
}

func NewMapProvider(keys map[string][]byte) *MapProvider {
	p := &MapProvider{keys: make(map[string][]byte, len(keys))}
	for path, key := range keys {
		p.keys[canonical(path)] = append([]byte(nil), key...)
	}
	return p
}

func (p *MapProvider) ResolveQMCExternalKey(ctx context.Context, sourcePath string) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if p == nil {
		return nil, false, nil
	}
	p.mu.RLock()
	key, ok := p.keys[canonical(sourcePath)]
	p.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	return append([]byte(nil), key...), true, nil
}

// Clear releases the provider's mutable copies after a request completes.
// This is lifetime minimization, not a guarantee of memory zeroization.
func (p *MapProvider) Clear() {
	if p == nil {
		return
	}
	p.mu.Lock()
	for path, key := range p.keys {
		for i := range key {
			key[i] = 0
		}
		delete(p.keys, path)
	}
	p.mu.Unlock()
}

type MapKGGDatabaseProvider struct {
	mu  sync.RWMutex
	dbs map[string]KGGDatabase
}

type StaticKGGDatabaseProvider struct{ Database KGGDatabase }

func (p StaticKGGDatabaseProvider) ResolveKGGDatabase(ctx context.Context, _ string) (KGGDatabase, bool, error) {
	if err := ctx.Err(); err != nil {
		return KGGDatabase{}, false, err
	}
	if p.Database.Path == "" {
		return KGGDatabase{}, false, nil
	}
	return p.Database, true, nil
}

func NewMapKGGDatabaseProvider(databases map[string]KGGDatabase) *MapKGGDatabaseProvider {
	p := &MapKGGDatabaseProvider{dbs: make(map[string]KGGDatabase, len(databases))}
	for source, database := range databases {
		p.dbs[canonical(source)] = database
	}
	return p
}

func (p *MapKGGDatabaseProvider) ResolveKGGDatabase(ctx context.Context, sourcePath string) (KGGDatabase, bool, error) {
	if err := ctx.Err(); err != nil {
		return KGGDatabase{}, false, err
	}
	if p == nil {
		return KGGDatabase{}, false, nil
	}
	p.mu.RLock()
	database, ok := p.dbs[canonical(sourcePath)]
	p.mu.RUnlock()
	return database, ok, nil
}

func canonical(path string) string { return filepath.Clean(path) }
