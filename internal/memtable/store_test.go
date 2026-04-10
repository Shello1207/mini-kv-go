package memtable

import (
	"errors"
	"mini-kv-go/internal/engine"
	"sync"
	"testing"
)

func TestStore_PutGet(t *testing.T) {
	s := New()

	err := s.Put("name", []byte("alice"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	got, err := s.Get("name")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != "alice" {
		t.Fatalf("unexpected value: got=%q want=%q", string(got), "alice")
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := New()

	_, err := s.Get("missing")
	if !errors.Is(err, engine.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got=%v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	s := New()

	if err := s.Put("name", []byte("alice")); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if err := s.Delete("name"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if s.Exists("name") {
		t.Fatalf("expected key to be deleted")
	}
}

func TestStore_CopyOnWriteAndRead(t *testing.T) {
	s := New()

	src := []byte("alice")
	if err := s.Put("name", src); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	src[0] = 'b'

	got, err := s.Get("name")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != "alice" {
		t.Fatalf("store should not copy input slice, got=%q", string(got))
	}

	got[0] = 'c'
	got2, err := s.Get("name")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got2) != "alice" {
		t.Fatalf("store should copy on read, got=%q", string(got2))
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := New()

	const n = 100
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key"
			val := []byte("value")
			_ = s.Put(key, val)
			_, _ = s.Get(key)
		}(i)
	}

	wg.Wait()

	if !s.Exists("key") {
		t.Fatalf("expected key to exist after concurrent writes")
	}
}
