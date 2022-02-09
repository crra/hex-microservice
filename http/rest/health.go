package rest

import "net/http"

func (h *handler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := h.health.Health()
		writeResponse(w, contentTypeJson, []byte(health.Name), http.StatusOK)
	}
}
