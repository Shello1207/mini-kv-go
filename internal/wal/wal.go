package wal

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type WAL struct {
	mu        sync.Mutex
	file      *os.File
	path      string
	maxSize   int64
	currentSz int64
	seq       int
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

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &WAL{
		file:      file,
		path:      path,
		maxSize:   1 << 20, // 1MB
		currentSz: info.Size(),
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

	// 检查写入新数据后是否会超过大小限制
	newSize := w.currentSz + int64(len(data))
	if newSize > w.maxSize {
		if err := w.rotate(); err != nil {
			return err
		}
		// 旋转后重新计算数据大小（因为currentSz被重置为0）
		newSize = int64(len(data))
	}

	if _, err := w.file.Write(data); err != nil {
		return err
	}

	w.currentSz = newSize

	return w.file.Sync()
}

func (w *WAL) Replay(apply func(rec Record) error) error {
	// 获取所有WAL文件，包括当前活跃的文件和分段文件
	pattern := filepath.Join(filepath.Dir(w.path), "*.log")
	files, _ := filepath.Glob(pattern)

	// 按文件名排序，确保按正确顺序处理
	sort.Strings(files)

	// 首先处理当前活跃的WAL文件
	if _, err := os.Stat(w.path); err == nil {
		file, err := os.Open(w.path)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var rec Record
			if err := json.Unmarshal(line, &rec); err != nil {
				file.Close()
				return err
			}

			if err := apply(rec); err != nil {
				file.Close()
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}

	// 然后处理所有分段文件
	for _, filename := range files {
		// 跳过当前活跃的文件，因为已经处理过了
		if filename == w.path {
			continue
		}

		file, err := os.Open(filename)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var rec Record
			if err := json.Unmarshal(line, &rec); err != nil {
				file.Close()
				return err
			}

			if err := apply(rec); err != nil {
				file.Close()
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}
	return nil
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

func (w *WAL) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}

	w.seq++

	newPath := filepath.Join(filepath.Dir(w.path), fmt.Sprintf("wal-%d.log", w.seq))

	file, err := os.OpenFile(newPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.currentSz = 0
	w.path = newPath

	return nil
}
