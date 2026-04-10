package engine

import (
	"errors"
)

var ErrKeyNotFound = errors.New("key not found")

type Store interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Exists(key string) bool
}
