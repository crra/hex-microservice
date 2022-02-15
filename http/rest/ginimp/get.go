package ginimp

import (
	"errors"
	"hex-microservice/lookup"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) RedirectGet(mappingUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param(UrlParameterCode)

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

			c.JSON(status, gin.H{"error": http.StatusText(status)})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, redirect.URL)
		return
	}
}
