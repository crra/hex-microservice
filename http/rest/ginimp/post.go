package ginimp

import (
	"errors"
	"fmt"
	"hex-microservice/adder"
	"net/http"

	"github.com/gin-gonic/gin"
)

type redirectPostRequest struct {
	CustomCode string `json:"custom_code" binding:"omitempty,gte=5,lte=25"`
	URL        string `json:"url" binding:"required"`
}

type redirectResponse struct {
	Code string `json:"code" msgpack:"code"`
	URL  string `json:"url" msgpack:"url"`

	Links []link `json:"_links,omitempty"`
}

type link struct {
	Href string `json:"href,omitempty"`
	Rel  string `json:"rel,omitempty"`
	T    string `json:"type,omitempty"`
}

func (h *handler) RedirectPost(mappingUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := h.converters[c.ContentType()]
		if !ok {
			h.log.Error(nil, "unsupported content type", "contentType", c.ContentType())
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": http.StatusText(http.StatusUnsupportedMediaType)})
			return
		}

		var r redirectPostRequest

		if err := c.Bind(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "field validation failed"})
			return
		}

		results, err := h.adder.Add(
			adder.RedirectCommand{
				URL:        r.URL,
				CustomCode: r.CustomCode,
				ClientInfo: c.ClientIP(),
			})
		if err != nil {
			if r.CustomCode != "" && errors.Is(err, adder.ErrDuplicate) {
				c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf(customCodeAlreadyTaken, r.CustomCode)})
				return
			}

			h.log.Error(err, "error adding request", "request", r)
			c.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
			return
		}

		// response to client
		result := results[0]
		c.JSON(http.StatusCreated, redirectResponse{
			Code: result.Code,
			URL:  r.URL,
			Links: []link{
				{
					Href: urlForCode(mappingUrl, result.Code),
					Rel:  resourceName,
					T:    http.MethodGet,
				},
				{
					Href: urlForCodeAndToken(mappingUrl, result.Code, result.Token),
					Rel:  resourceName,
					T:    http.MethodDelete,
				},
			},
		})
		return
	}
}

/*
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Response setting gin.JSON
func (g *Gin) Response(httpCode, errCode int, data interface{}) {
	g.C.JSON(httpCode, Response{
		Code: errCode,
		Msg:  e.GetMsg(errCode),
		Data: data,
	})
	return
}
*/
