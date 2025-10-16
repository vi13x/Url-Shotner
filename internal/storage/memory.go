package storage

import (
	"context"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (m *MemoryStorage) Save(ctx context.Context, id string, originalURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = originalURL
	return nil
}

func (m *MemoryStorage) Get(ctx context.Context, id string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[id]
	return val, ok, nil
}
