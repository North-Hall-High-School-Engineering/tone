package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/North-Hall-High-School-Engineering/tone/services/registry/internal/store"
)

type Handler struct {
	Store *store.FS
}

// GET /v1/models/<name>?version=<version>
func (h *Handler) GetManifest(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	version := r.URL.Query().Get("version")
	if len(strings.TrimSpace(version)) == 0 {
		http.Error(w, "version required", http.StatusBadRequest)
		return
	}

	m, err := h.Store.Load(name, version)
	if err != nil {
		http.Error(w, "model not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}
