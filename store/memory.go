package store

type Storage interface {
  Put(key string, value []byte)
}

type key_pairs map[string][]byte

type memory_storage struct {
  items key_pairs
}

func New() Storage {
  return & memory_storage {
    items: make(key_pairs),
  }
}

func (* memory_storage) Put(key string, value []byte) {
  // do nothing
}
