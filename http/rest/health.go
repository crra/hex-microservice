package rest

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

func (h *handler) Health(now time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := h.health.Health(now)

		response, err := json.Marshal(healthResponse{
			Name:    health.Name,
			Version: health.Version,
			Uptime:  health.Uptime.String(),
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		writeResponse(w, contentTypeJson, response, http.StatusOK)
	}
}
