package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{Status: "ok"}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
}

func (h *Handler) favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.Redirect(w, r, "/static/icons/favicon.svg", http.StatusPermanentRedirect)
}
