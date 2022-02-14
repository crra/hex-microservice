package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ginRedirectPostRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *ginhandler) RedirectPost(mappingUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r ginRedirectPostRequest

		if err := c.ShouldBind(&r); err != nil {
			// fmt.Sprintf(titleProcessingFieldFormat, fieldName),
			c.JSON(http.StatusBadRequest, gin.H{"error": "field validation failed"})
			return
		}
	}
}
