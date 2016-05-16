package store

import "sync"

type Storage interface {
	Put(key string, value []byte)
	Get(key string) ([]byte, bool)
	Delete(key string)
}

type keyPairs map[string][]byte

type memoryStorage struct {
	items keyPairs
	lock  sync.RWMutex
}

func New() Storage {
	return &memoryStorage{
		items: make(keyPairs),
	}
}

func (m *memoryStorage) Put(key string, value []byte) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.items[key] = value
}

func (m *memoryStorage) Get(key string) ([]byte, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	value, ok := m.items[key]
	return value, ok
}

func (m *memoryStorage) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.items, key)
}
