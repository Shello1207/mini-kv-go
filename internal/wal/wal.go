package wal

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type WAL struct {
	mu   sync.Mutex
	file *os.File
	path string
}

func Open(path string) (*WAL, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
		path: path,
	}, nil
}

func (w *WAL) AppendPut(key string, value []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	rec := Record{
		Op:    OpPut,
		Key:   key,
		Value: base64.StdEncoding.EncodeToString(value),
	}
	return w.writeRecord(rec)
}

func (w *WAL) AppendDelete(key string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	rec := Record{
		Op:  OpDel,
		Key: key,
	}
	return w.writeRecord(rec)
}

func (w *WAL) writeRecord(rec Record) error {
	if w.file == nil {
		return errors.New("wal file is closed")
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if _, err := w.file.Write(data); err != nil {
		return err
	}
	return w.file.Sync()
}

func (w *WAL) Replay(apply func(rec Record) error) error {
	f, err := os.Open(w.path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// 提高一点 scanner 的 buffer 上限，避免 value 稍大时出现问题
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var rec Record
		if err := json.Unmarshal(line, &rec); err != nil {
			return err
		}

		if err := apply(rec); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}
