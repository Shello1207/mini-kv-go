package api

import (
	"errors"
	"io"
	"mini-kv-go/internal/engine"
	"mini-kv-go/internal/service"
	"net/http"
	"strings"
)

type Handler struct {
	svc *service.KVService
}

func NewHandler(svc *service.KVService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleKV(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv")
	if key == "" || key == r.URL.Path {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body failed", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := h.svc.Put(key, body); err != nil {
			http.Error(w, "put failed:"+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		val, err := h.svc.Get(key)
		if err != nil {
			if errors.Is(err, engine.ErrKeyNotFound) {
				http.Error(w, "key not found", http.StatusNotFound)
				return
			}
			http.Error(w, "get failed:"+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(val)

	case http.MethodDelete:
		if err := h.svc.Delete(key); err != nil {
			http.Error(w, "delete failed:"+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
