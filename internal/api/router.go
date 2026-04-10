package api

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	// 支持 /kv 和 /kv/* 路径
	mux.HandleFunc("/kv", h.HandleKV)
	mux.HandleFunc("/kv/", h.HandleKV)
	mux.HandleFunc("/health", h.HandleHealth)
	return mux
}
