package stdlib

import (
	"errors"
	"hex-microservice/lookup"
	"net/http"
)

// RedirectGet implements the "get" verb of the REST context that gets an existing redirect.
func (h *handler) RedirectGet(mappingUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := h.paramFn(r, UrlParameterCode)

		redirect, err := h.lookup.Lookup(
			lookup.RedirectQuery{Code: code},
		)
		if err != nil {
			status := http.StatusInternalServerError

			if errors.Is(err, lookup.ErrNotFound) {
				status = http.StatusNotFound
			}

			if status == http.StatusInternalServerError {
				h.log.Error(err, "Internal server error", "method", "RedirectGet", UrlParameterCode, code)
			}

			http.Error(w, http.StatusText(status), status)
			return
		}

		http.Redirect(w, r, redirect.URL, http.StatusTemporaryRedirect)
		return
	}
}
