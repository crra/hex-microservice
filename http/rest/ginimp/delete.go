package ginimp

import (
	"errors"
	"hex-microservice/invalidator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type redirectDeleteRequest struct {
	Code  string `uri:"code" binding:"required"`
	Token string `uri:"token" binding:"required"`
}

// RedirectGet implements the "delete" verb of the REST context that deletes an existing redirect.
func (h *handler) RedirectInvalidate(mappingUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r redirectDeleteRequest

		if err := c.BindUri(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "field validation failed"})
			return
		}

		err := h.invalidator.Invalidate(invalidator.RedirectQuery{
			Code:  r.Code,
			Token: r.Token,
		})
		if err != nil {
			status := http.StatusInternalServerError

			if errors.Is(err, invalidator.ErrNotFound) {
				status = http.StatusNotFound
			}

			c.JSON(status, gin.H{"error": http.StatusText(status)})
			return
		}

		c.Status(http.StatusNoContent)
		return
	}
}
