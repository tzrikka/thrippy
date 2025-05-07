package secrets

import (
	"context"
	"sync"
)

const (
	inMemoryOption = "in-memory"
)

type inMemoryProvider struct {
	store map[string]string
	mu    sync.RWMutex
}

func newInMemoryProvider() (Manager, error) {
	return &inMemoryProvider{store: make(map[string]string)}, nil
}

func (p *inMemoryProvider) Set(_ context.Context, key, value string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.store[key] = value
	return nil
}

func (p *inMemoryProvider) Get(_ context.Context, key string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	v, ok := p.store[key]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (p *inMemoryProvider) Delete(_ context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.store, key)
	return nil
}
