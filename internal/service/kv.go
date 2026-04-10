package service

import (
	"encoding/base64"
	"errors"

	"mini-kv-go/internal/engine"
	"mini-kv-go/internal/wal"
)

type KVService struct {
	store engine.Store
	log   *wal.WAL
}

func NewKVService(store engine.Store, log *wal.WAL) *KVService {
	return &KVService{
		store: store,
		log:   log,
	}
}

func (s *KVService) Put(key string, value []byte) error {
	if key == "" {
		return errors.New("key is empty")
	}

	if err := s.log.AppendPut(key, value); err != nil {
		return err
	}

	return s.store.Put(key, value)
}

func (s *KVService) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("key is empty")
	}
	return s.store.Get(key)
}

func (s *KVService) Delete(key string) error {
	if key == "" {
		return errors.New("key is empty")
	}

	if err := s.log.AppendDelete(key); err != nil {
		return err
	}

	return s.store.Delete(key)
}

func (s *KVService) Exists(key string) bool {
	if key == "" {
		return false
	}
	return s.store.Exists(key)
}

func (s *KVService) Recover() error {
	return s.log.Replay(func(rec wal.Record) error {
		switch rec.Op {
		case wal.OpPut:
			val, err := base64.StdEncoding.DecodeString(rec.Value)
			if err != nil {
				return err
			}
			return s.store.Put(rec.Key, val)
		case wal.OpDel:
			return s.store.Delete(rec.Key)
		default:
			return nil
		}
	})
}
