package rest

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (h *handler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := h.health.Health()

		response, err := json.Marshal(healthResponse{
			Name:    health.Name,
			Version: health.Version,
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		writeResponse(w, contentTypeJson, response, http.StatusOK)
	}
}
