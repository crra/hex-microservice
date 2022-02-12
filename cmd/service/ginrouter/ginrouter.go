package ginrouter

import (
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	router := gin.Default()
	router.HandleMethodNotAllowed = false

	s := rest.NewGin(log, h, a, l, d, func(r *http.Request, key string) string {
		panic("use 'gin.ShouldBind' instead")
	})

	router.GET("/health", s.Health())

	return router
}
