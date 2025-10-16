package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"URL_shortener/internal/shortener"
)

type Handler struct {
	svc     *shortener.Service
	baseURL string
}

func NewHandler(svc *shortener.Service, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: strings.TrimRight(baseURL, "/")}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shorten", h.handleShorten)
	mux.HandleFunc("/s/", h.handleResolve)
	return mux
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"shortUrl"`
}

func (h *Handler) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req shortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	id, err := h.svc.Shorten(r.Context(), req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := shortenResponse{ShortURL: h.baseURL + "/s/" + id}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/s/")
	if id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	original, ok, err := h.svc.Resolve(context.Background(), id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(w, r, original, http.StatusFound)
}
