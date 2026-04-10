package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"mini-kv-go/internal/api"
	"mini-kv-go/internal/memtable"
	"mini-kv-go/internal/service"
	"mini-kv-go/internal/wal"
)

func main() {
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data dir failed: %v", err)
	}

	walPath := filepath.Join(dataDir, "wal.log")

	store := memtable.New()

	walFile, err := wal.Open(walPath)
	if err != nil {
		log.Fatalf("open wal file failed: %v", err)
	}
	defer walFile.Close()

	svc := service.NewKVService(store, walFile)

	// 启动恢复
	if err := svc.Recover(); err != nil {
		log.Fatalf("recover from wal failed: %v", err)
	}

	handler := api.NewHandler(svc)
	router := api.NewRouter(handler)

	addr := ":8080"
	log.Printf("MiniKV server listening on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
