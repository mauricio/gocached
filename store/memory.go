package store

type Storage interface {
	Put(key string, value []byte)
	Get(key string) ([]byte, bool)
	Delete(key string)
}

type key_pairs map[string][]byte

type memory_storage struct {
	items key_pairs
}

func New() Storage {
	return &memory_storage{
		items: make(key_pairs),
	}
}

func (m *memory_storage) Put(key string, value []byte) {
	m.items[key] = value
}

func (m *memory_storage) Get(key string) ([]byte, bool) {
	value, ok := m.items[key]
	return value, ok
}

func (m *memory_storage) Delete(key string) {
	delete(m.items, key)
}
