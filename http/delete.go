package service

import (
	"errors"
	"hex-microservice/deleter"
	"net/http"
)

// RedirectGet implements the "delete" verb of the REST context that deletes an existing redirect.
func (h *handler) RedirectDelete(mappingUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.deleter.Delete(deleter.RedirectQuery{
			Code:  h.paramFn(r, UrlParameterCode),
			Token: h.paramFn(r, UrlParameterToken),
		})
		if err != nil {
			status := http.StatusInternalServerError

			if errors.Is(err, deleter.ErrNotFound) {
				status = http.StatusNotFound
			}

			http.Error(w, http.StatusText(status), status)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
}
