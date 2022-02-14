package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *ginhandler) Health(now time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := h.health.Health(now)
		c.JSON(http.StatusOK, gin.H{
			"name":    h.Name,
			"version": h.Version,
			"uptime":  h.Uptime.String(),
		})
	}
}
