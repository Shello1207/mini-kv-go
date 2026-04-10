package wal

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestWAL_AppendAndReplay(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.log")

	w, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if err1 := w.AppendPut("name", []byte("alice")); err1 != nil {
		t.Fatalf("AppendPut failed: %v", err1)
	}

	if err2 := w.AppendDelete("name"); err2 != nil {
		t.Fatalf("AppendDelete failed: %v", err2)
	}

	if err3 := w.Close(); err3 != nil {
		t.Fatalf("Close failed: %v", err3)
	}

	var records []Record
	w2, err := Open(path)
	if err != nil {
		t.Fatalf("Open second wal failed: %v", err)
	}
	defer w2.Close()

	err = w2.Replay(func(rec Record) error {
		records = append(records, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Replay failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("unexpected record count: got=%d want=%d", len(records), 2)
	}

	if records[0].Op != OpPut || records[0].Key != "name" {
		t.Fatalf("unexpected first record: %+v", records[0])
	}

	val, err := base64.StdEncoding.DecodeString(records[0].Value)
	if err != nil {
		t.Fatalf("decode base64 failed: %v", err)
	}
	if string(val) != "alice" {
		t.Fatalf("unexpected decoded value: got=%q want=%q", string(val), "alice")
	}

	if records[1].Op != OpDel || records[1].Key != "name" {
		t.Fatalf("unexpected second record: %+v", records[1])
	}
}

func TestWAL_ReplayEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.log")

	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	w, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer w.Close()

	count := 0
	err = w.Replay(func(rec Record) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Replay failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no records, got=%d", count)
	}
}
