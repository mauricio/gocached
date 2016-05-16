package store

type Storage interface {
	Put(key string, value []byte)
	Get(key string) ([]byte, bool)
	Delete(key string)
}

type keyPairs map[string][]byte

type memoryStorage struct {
	items keyPairs
}

func New() Storage {
	return &memoryStorage{
		items: make(keyPairs),
	}
}

func (m *memoryStorage) Put(key string, value []byte) {
	m.items[key] = value
}

func (m *memoryStorage) Get(key string) ([]byte, bool) {
	value, ok := m.items[key]
	return value, ok
}

func (m *memoryStorage) Delete(key string) {
	delete(m.items, key)
}
