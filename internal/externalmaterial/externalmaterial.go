// Package externalmaterial owns request-scoped decryption material.
package externalmaterial

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
)

var ErrUnavailable = errors.New("external material unavailable")

// Provider may be called concurrently. Implementations must not mutate or
// expose material based on a different source path.
type Provider interface {
	ResolveQMCExternalKey(context.Context, string) ([]byte, bool, error)
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

func canonical(path string) string { return filepath.Clean(path) }
