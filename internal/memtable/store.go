package memtable

import (
	"sync"

	"mini-kv-go/internal/engine"
)

type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func New() *Store {
	return &Store{
		data: make(map[string][]byte),
	}
}

func (s *Store) Put(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := make([]byte, len(value))
	copy(buf, value)
	s.data[key] = buf
	return nil
}

func (s *Store) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return nil, engine.ErrKeyNotFound
	}

	buf := make([]byte, len(v))
	copy(buf, v)
	return buf, nil
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	return ok
}
