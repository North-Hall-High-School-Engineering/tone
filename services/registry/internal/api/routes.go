package api

import "net/http"

func Routes(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/models/{name}", h.GetManifest)

	return mux
}
